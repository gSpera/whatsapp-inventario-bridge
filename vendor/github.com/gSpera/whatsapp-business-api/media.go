package whatsapp

import (
	"encoding/json"
	"fmt"
	"io"
)

// Documentation at: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/media

// MediaID uniquely references a media
type MediaID = string

// RetrieveMedia extracts the media url and return a io.ReadCloser to the media body
func (c *Client) RetrieveMedia(mediaID MediaID) (io.ReadCloser, error) {
	response, err := c.SendGETRequest("/" + mediaID)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve media url: %w", err)
	}
	var mediaURL struct {
		// ID  string `json:"id"`
		URL string `json:"url"`
		// MimeType string `json:"mime_type"`
		// FileSize string `json:"file_size"`
	}
	err = json.NewDecoder(response).Decode(&mediaURL)
	if err != nil {
		return nil, fmt.Errorf("cannot decode media url: %w", err)
	}

	body, err := c.SendGETRequestWithoutBase(mediaURL.URL)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve media: %w", err)
	}
	return body, nil
}
