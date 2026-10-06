package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Authorization Token
type AccessToken = string

const BASE_URL = "https://graph.facebook.com/v25.0"

// A Client is used to make requests
type Client struct {
	http http.Client

	accessToken AccessToken
}

func NewClient(accessToken AccessToken) *Client {
	return &Client{
		http:        *http.DefaultClient,
		accessToken: accessToken,
	}
}

// SendRequest sends arbitrary requests
func (c *Client) SendPOSTRequest(endpoint string, body any) error {
	content, err := json.MarshalIndent(body, "", "\t")
	if err != nil {
		return fmt.Errorf("cannot marshal body: %w", err)
	}
	reader := bytes.NewBuffer(content)

	req, err := http.NewRequest("POST", BASE_URL+endpoint, reader)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("error in request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		decoder := json.NewDecoder(res.Body)
		defer res.Body.Close()
		var err ErrorWrapper
		decodeErr := decoder.Decode(&err)
		if decodeErr != nil {
			return fmt.Errorf("cannot decode error: %w", decodeErr)
		}

		return err.Error
	}
	return nil
}

func (c Client) SendGETRequest(endpoint string) (io.ReadCloser, error) {
	return c.SendGETRequestWithoutBase(BASE_URL + endpoint)
}

func (c Client) SendGETRequestWithoutBase(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error in request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		decoder := json.NewDecoder(res.Body)
		defer res.Body.Close()
		var err ErrorWrapper
		decodeErr := decoder.Decode(&err)
		if decodeErr != nil {
			return nil, fmt.Errorf("cannot decode error: %w", decodeErr)
		}

		return nil, err.Error
	}

	return res.Body, nil
}
