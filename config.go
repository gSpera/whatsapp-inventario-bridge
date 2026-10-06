package main

// Config contains the configuration of the bridge.
type Config struct {
	// Listening address for the WhatsApp Webhook
	ListenAddr string `json:"listenAddr"`
	// File path for the authentication tokens
	JwtStore string `json:"jwtStore"`
	// Enable deeper logging info
	Debug bool `json:"debug"`

	WhatsAppToken             string `json:"whatsapp_token"`
	WhatsAppWebhookVerifyCode string `json:"whatsapp_webhook_verifycode"`
	WhatsAppPhoneNumberID     string `json:"whatsapp_phonenumber_id"`

	// List of Inventario to bridge
	Inventario []Inventario `json:"inventario"`
}

// Inventario describes a service to bridge
type Inventario struct {
	// Identification name
	Name string `json:"name"`

	Addr   string `json:"addr"`
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}
