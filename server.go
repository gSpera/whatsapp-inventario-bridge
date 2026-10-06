package main

import (
	"log/slog"
)

// Server mantains the state of the service
type Server struct {
	Log        *slog.Logger
	Whatsapp   *WhatsappAPI
	Inventario []Inventario
}
