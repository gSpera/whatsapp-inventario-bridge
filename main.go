package main

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"os"

	"github.com/agentine/reveal"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	log.Info("Whatsapp Inventario Bridge")

	cfgBytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Error("Cannot open config.json file", "err", err)
		os.Exit(1)
	}

	var cfg Config
	// Default configuration
	cfg.ListenAddr = ":8081"
	cfg.JwtStore = "jwt-store.json"
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		log.Error("Cannot parse configuration", "err", err)
		os.Exit(1)
	}

	loggerLevel := slog.LevelInfo
	if cfg.Debug {
		loggerLevel = slog.LevelDebug
	}

	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		// AddSource: cfg.Debug,
		Level: loggerLevel,
	}))
	reveal.Config.MaxDepth = 0
	reveal.Config.Indent = "    "

	var s Server
	s.Log = log
	s.Inventario = cfg.Inventario
	s.AuthTokens = make(map[string]AuthToken)
	log.Info("Logging in")
	i := Inventario{
		Name:     "TEST",
		Username: "gs",
		Password: "gs",
		Addr:     "http://localhost:8080",
	}
	reveal.Dump(s)
	_ = i
	err = s.updateTokens()
	if err != nil {
		panic(err)
	}
	reveal.Dump(s)

	return
	s.Whatsapp = NewWhatsappAPI(log.With("what", "whatsapp"), cfg.WhatsAppWebhookVerifyCode, cfg.WhatsAppToken, cfg.WhatsAppPhoneNumberID)
	http.Handle("/", s.Whatsapp)

	log.Info("Starting", "addr", cfg.ListenAddr, "jwt-store", cfg.JwtStore)
	err = http.ListenAndServe(cfg.ListenAddr, nil)
	if err != nil {
		log.Error("Error while serving", "err", err)
	} else {
		log.Info("Exiting")
	}
}
