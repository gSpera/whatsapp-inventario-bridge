package main

import (
	"encoding/json"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"

	"github.com/gSpera/whatsapp-business-api"
)

var qrCodeRegex *regexp.Regexp = regexp.MustCompile(`(?P<prefix>\D*)(?P<orderID>\d+)(?P<suffix>\D*)`)

type ServerInterface interface {
	OrderInfo(orderID int) (order, companyName, city string, err error)
	DeliverOrder(orderID int) error
}

type WhatsappAPI struct {
	log                        *slog.Logger
	webhookToken               string
	phoneNumberID              string
	client                     whatsapp.Client
	imageFolderPath            string
	inventarioFromPrefixSuffix func(prefix, suffix string) ServerInterface
	// allowedPhones   map[string]struct{}
}

func NewWhatsappAPI(log *slog.Logger, webhookToken string, accessToken string, phoneNumberID string, imageFolderPath string, inventarioFromPrefixSuffix func(prefix, suffix string) ServerInterface) (*WhatsappAPI, error) {
	err := os.MkdirAll(imageFolderPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot create image folder: %w", err)
	}

	return &WhatsappAPI{
		log:                        log,
		webhookToken:               webhookToken,
		phoneNumberID:              phoneNumberID,
		client:                     *whatsapp.NewClient(accessToken),
		imageFolderPath:            imageFolderPath,
		inventarioFromPrefixSuffix: inventarioFromPrefixSuffix,
	}, nil
}

func (wa *WhatsappAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wa.log.Debug("Whatsapp request")

	// Check for webhook setup
	r.ParseForm()
	if r.FormValue("hub.mode") != "" {
		wa.setupWebhook(w, r)
		return
	}

	var content whatsapp.WebhookEntryObject
	body, err := io.ReadAll(r.Body)
	if err != nil {
		wa.log.Error("Cannot read body", "err", err)
	}

	err = json.Unmarshal(body, &content)
	if err != nil {
		wa.log.Error("Cannot decode body", "err", err, "body", string(body))
	}
	// wa.log.Debug("Received body", "body", string(body))

	// fmt.Println()
	// fmt.Println("Body:")
	// fmt.Println(string(body))
	// fmt.Println("Decoded:")
	// reveal.Dump(content)
	// fmt.Println()

	for _, entry := range content.Entry {
		wa.log.Debug("Received entry", "entry-id", entry.ID)
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				wa.handleMessage(msg)
			}
		}
	}
}

func (wa *WhatsappAPI) setupWebhook(w http.ResponseWriter, r *http.Request) {
	wa.log.Info("Setup webhook")
	if r.FormValue("hub.mode") != "subscribe" {
		wa.log.Warn("Unkown hub.mode")
		return
	}
	verifyToken := r.FormValue("hub.verify_token")
	if verifyToken != wa.webhookToken {
		wa.log.Error("Wrong verify token", "received-token", verifyToken, "own-token", wa.webhookToken)
	}

	wa.log.Info("Setup webhook challenge", "challenge", r.FormValue("hub.challenge"))
	fmt.Fprintln(w, r.FormValue("hub.challenge"))
}

func (wa *WhatsappAPI) handleMessage(msg whatsapp.WebhookMessage) {
	log := wa.log.With("msg-id", msg.ID)
	log.Info("Received message", "from", msg.From, "type", msg.Type, "timestamp", msg.Timestamp)

	err := wa.client.SetMessageAsRead(wa.phoneNumberID, msg.ID)
	if err != nil {
		log.Error("Cannot mark message message as read", "err", err)
	}

	switch msg.Type {
	case "image":
		err := wa.onReceivedImage(msg)
		if err != nil {
			log.Error("Cannot handle image", "err", err)
			wa.client.SendSimpleMessage(wa.phoneNumberID, msg.From, "Sfortunatamente c'è stato qualche errore, potresti mandare un altra foto??")
		}
	default:
		wa.client.SendSimpleMessage(wa.phoneNumberID, msg.From, "Non ho capito")
	}
}

func (wa *WhatsappAPI) onReceivedImage(msg whatsapp.WebhookMessage) error {
	log := wa.log.With("msg-id", msg.ID, "image-id", msg.Image.ID)

	// Download image
	imgReader, err := wa.client.SendGETRequestWithoutBase(msg.Image.URL)
	if err != nil {
		return fmt.Errorf("cannot retrieve image: %w", err)
	}
	defer imgReader.Close()

	// Read image
	imgPath := path.Join(wa.imageFolderPath, msg.ID+".jpg")
	fl, err := os.Create(imgPath)
	if err != nil {
		return fmt.Errorf("cannot create image file: %w", err)
	}
	_, err = io.Copy(fl, imgReader)
	if err != nil {
		return fmt.Errorf("cannot write image: %w", err)
	}
	fl.Close()

	cmd := exec.Command("ZXingReader", "-bytes", "-1", imgPath)
	res, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot scan image: %w", err)
	}

	if len(res) == 0 { // No QR Codes found
		log.Warn("Image with no qr code")
		return fmt.Errorf("no qr code found")
	}

	text := string(res)
	log.Info("Found QR Code", "text", text)

	var orderID int
	var prefix, suffix string
	matches := qrCodeRegex.FindStringSubmatch(text)
	if len(matches) != 4 {
		log.Warn("Invalid QR code, invalid submatch", "len", len(matches), "matches", matches, "text", text)
		return fmt.Errorf("invalid QR code, content %s: %w", text, err)
	}
	prefix = matches[1]
	orderID, err = strconv.Atoi(matches[2])
	suffix = matches[3]

	if err != nil {
		log.Warn("Invalid qr code, cannot parse order ID", "text", text, "err", err, "prefix", prefix, "orderID", matches[1], "suffix", suffix)
		return fmt.Errorf("invalid qr code, content: %s: %w", text, err)
	}
	log = log.With("order-id", orderID)

	log.Info("Found order", "prefix", prefix, "suffix", suffix)
	si := wa.inventarioFromPrefixSuffix(prefix, suffix)
	order, companyName, city, err := si.OrderInfo(orderID)
	if err != nil {
		log.Error("Cannot get order info", "err", err)
		return fmt.Errorf("cannot get order info: %w", err)
	}

	err = wa.client.SendSimpleMessage(wa.phoneNumberID, msg.From, fmt.Sprintf("Consegnato ordine %s con successo\n*%s* %s\nPrefisso: %s\nOrdine: %d\nSuffisso: %s", order, companyName, city, prefix, orderID, suffix))
	if err != nil {
		log.Error("Cannot send success message", "err", err)
	}

	return nil
}
