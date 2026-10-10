package main

// Config contains the configuration of the bridge.
type Config struct {
	// Listening address for the WhatsApp Webhook
	ListenAddr string `json:"listenAddr"`
	// File path for the authentication tokens
	AuthTokenStorePath string `json:"authTokenFile"`
	// Enable deeper logging info
	Debug bool `json:"debug"`

	// Path where to store received images
	ImageFolderPath string `json:"imageFolder"`

	// Whatsapp secret Tokens
	WhatsAppToken             string `json:"whatsapp_token"`
	WhatsAppWebhookVerifyCode string `json:"whatsapp_webhook_verifycode"`
	WhatsAppPhoneNumberID     string `json:"whatsapp_phonenumber_id"`

	// List of Inventario to bridge
	Inventario []Inventario `json:"inventario"`
}
