package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

// Server mantains the state of the service
type Server struct {
	Log        *slog.Logger
	Whatsapp   *WhatsappAPI
	Inventario []Inventario

	AuthTokens map[string]AuthToken

	AuthTokenFilePath string
}

type AuthToken struct {
	Token      string    `json:"token"`
	Expiration time.Time `json:"expireAt"`
}

func (s *Server) updateTokens() error {
	log := s.Log.With("what", "authtoken-updating")
	log.Debug("Updating AuthTokens")
	errs := make([]error, len(s.Inventario))
	maxAllowedExpire := time.Now().AddDate(0, 0, 1)

	// Update token if necessary
	for i, inv := range s.Inventario {
		// If not AuthToken is found the zero value will be used,
		// for time.Time this is timestamp 0, which is in the past.
		authToken, _ := s.AuthTokens[inv.Name]

		if authToken.Expiration.Before(maxAllowedExpire) {
			log.Info("Updating AuthToken", "name", inv.Name)
			errs[i] = s.updateToken(inv)

			if errs[i] != nil {
				log.Error("While updating", "name", inv.Name, "err", errs[i])
			}
		}
	}

	updateTokensErr := errors.Join(errs...)

	// Store the updated tokens
	fl, err := os.Create(s.AuthTokenFilePath)
	if err != nil {
		return errors.Join(updateTokensErr, fmt.Errorf("cannot open AuthToken store: %w", err))
	}

	enc := json.NewEncoder(fl)
	enc.SetIndent("", "\t")
	err = enc.Encode(s.AuthTokens)
	if err != nil {
		return errors.Join(updateTokensErr, fmt.Errorf("cannot encode AuthTokens: %w", err))
	}
	return fl.Close()
}

func (s *Server) updateToken(inv Inventario) error {
	s.Log.Debug("Updating token", "inventario-name", inv.Name)

	cj, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("cannot create cookie jar: %w", err)
	}

	client := http.Client{Jar: cj}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.PostForm(inv.Addr+"/login", url.Values{
		"username": []string{inv.Username},
		"password": []string{inv.Password},
	})
	if err != nil {
		return fmt.Errorf("cannot POST login, inventario: %s: %w", inv.Name, err)
	}
	cookies := resp.Cookies()

	var cookieValue string
	var cookieExpiration time.Time
	for _, c := range cookies {
		if c.Name == "user" {
			cookieValue = c.Value
			cookieExpiration = c.Expires
		}
	}

	if cookieValue == "" {
		return fmt.Errorf("cookie not present, are credentials correct??")
	}

	// Update Server AuthToken
	s.AuthTokens[inv.Name] = AuthToken{
		Token:      cookieValue,
		Expiration: cookieExpiration,
	}

	return nil
}
