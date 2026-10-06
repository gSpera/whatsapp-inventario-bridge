package whatsapp

import "fmt"

// Documentation at: https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codeshttps://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes

type ErrorCode struct {
	Code             int    `json:"code"`
	Details          string `json:"details"`
	FBTraceID        string `json:"fbtrace_id"`
	Message          string `json:"message"`
	MessagingProduct string `json:"messaging_product"`
	Type             string `json:"type"`
}

// ErrorWrapper wraps the error code, this is used by the api,
// it should not be used  outside of decoding responses, see ErrorCode
type ErrorWrapper struct {
	Error ErrorCode `json:"error"`
}

// ErrorCode implements the error interface, it can be passed as an error
func (e ErrorCode) Error() string {
	return fmt.Sprintf("(%d) %s %s", e.Code, e.Type, e.Message)
}
