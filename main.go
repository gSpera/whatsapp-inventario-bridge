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
	cfg.AuthTokenStorePath = "authtokens.json"
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
	s.AuthTokenFilePath = cfg.AuthTokenStorePath
	log.Info("Loading auth tokens")
	authTokensBytes, err := os.ReadFile(s.AuthTokenFilePath)
	if err != nil {
		log.Warn("Reading auth tokens", "file-name", s.AuthTokenFilePath, "err", err)
	}
	err = json.Unmarshal(authTokensBytes, &s.AuthTokens)
	if err != nil {
		log.Error("Decoding auth tokens", "file-name", s.AuthTokenFilePath, "err", err)
	}

	log.Info("Logging in")
	err = s.updateTokens()
	if err != nil {
		panic(err)
	}

	s.Whatsapp = NewWhatsappAPI(log.With("what", "whatsapp"), cfg.WhatsAppWebhookVerifyCode, cfg.WhatsAppToken, cfg.WhatsAppPhoneNumberID)
	http.Handle("/", s.Whatsapp)

	log.Info("Starting", "addr", cfg.ListenAddr, "authtoken-store", cfg.AuthTokenStorePath)
	err = http.ListenAndServe(cfg.ListenAddr, nil)
	if err != nil {
		log.Error("Error while serving", "err", err)
	} else {
		log.Info("Exiting")
	}
}
