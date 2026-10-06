package whatsapp

import "fmt"

// Endpoint /PHONE_NUMBER_ID/messages
// Documentation at: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages

// SendMessage sends a message to a phone number id
func (c *Client) SendMessage(phoneNumberID string, msg SendMessage) error {
	return c.SendPOSTRequest(fmt.Sprintf("/%s/messages", phoneNumberID), msg)
}

// SendSimpleMessage is a wrapper around SendMessage, this way sending a simple text message is clearer
func (c *Client) SendSimpleMessage(phoneNumberID string, to string, text string) error {
	return c.SendMessage(phoneNumberID, SendMessage{
		MessagingProduct: "whatsapp",
		Type:             "text",
		To:               to,

		Text: &SendMessageText{
			Body: text,
		},
	})
}

// SetMessageAsRead marks the message with ID msgID as read by the client
func (c *Client) SetMessageAsRead(phoneNumberID string, msgID string) error {
	return c.SendPOSTRequest(fmt.Sprintf("/%s/messages", phoneNumberID), struct {
		MessagingProduct string `json:"messaging_product"`
		Status           string `json:"status"`
		MessageID        string `json:"message_id"`
	}{
		"whatsapp",
		"read",
		msgID,
	})
}

type SendMessage struct {
	MessagingProduct string                  `json:"messaging_product"`
	Type             string                  `json:"type"`
	To               string                  `json:"to"`
	Template         *SendMessageTemplate    `json:"template,omitempty"`
	Text             *SendMessageText        `json:"text,omitempty"`
	Media            *SendMessageMedia       `json:"media,omitempty"`
	Interactive      *SendMessageInteractive `json:"interactive,omitempty"`
	Reaction         *SendMessageReaction    `json:"reaction,omitempty"`
	Location         *SendMessageLocation    `json:"location,omitempty"`

	// TTL deprecated
}

type SendMessageTemplate struct {
	// Namespace  string   `json:"namespace"`
	Name       string   `json:"name"`
	Language   string   `json:"language"`
	Components struct{} `json:"components"`
}

type SendMessageText struct {
	Body       string `json:"body"`
	PreviewURL string `json:"preview_url"`
}

type SendMessageMedia struct {
	ID string `json:"id"`
	// Required when type is audio, document, image, sticker or video,
	// do not use when type is text
	Link string `json:"link"`
	// Optional. Do not use with audio or sticker media
	Caption string `json:"caption"`
	// Optional.
	Filename string `json:"filename"`
	// Optional. Used only in On-Premises API
	Provider string `json:"provider"`
}

type SendMessageInteractive struct {
	// Required.
	Action struct {
		// Required for List Messages
		Button            string     `json:"button"`
		Buttons           []struct{} `json:"buttons"`
		CatalogID         string     `json:"catalog_id"`
		ProductRetailerID string     `json:"product_retailer_id"`
		Sections          []struct{} `json:"sections"`
	} `json:"action"`

	Type struct {
	} `json:"type"`

	// Optional for product
	Body struct {
		Text string `json:"text"`
	} `json:"body"`

	// Optional
	Footer struct {
		// Max 60 characthers
		Text string `json:"text"`
	} `json:"footer"`

	// Required for product_list
	Header struct{} `json:"header"`
}

type SendMessageReaction struct {
	MessageID string `json:"message_id"`
	Emoji     string `json:"emoji"`
}

type SendMessageLocation struct {
	Longitute float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
}
