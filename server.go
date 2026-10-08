package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// Server mantains the state of the service
type Server struct {
	Log        *slog.Logger
	Whatsapp   *WhatsappAPI
	Inventario []Inventario

	AuthTokens map[string]AuthToken
}

type AuthToken struct {
	Token      string    `json:"token"`
	Expiration time.Time `json:"expireAt"`
}

func (s *Server) updateTokens() error {
	s.Log.Debug("Updating AuthTokens")
	errs := make([]error, len(s.Inventario))
	maxAllowedExpire := time.Now().AddDate(0, 0, 7)

	for i, inv := range s.Inventario {
		// If not AuthToken is found the zero value will be used,
		// for time.Time this is timestamp 0, which is in the past.
		authToken, _ := s.AuthTokens[inv.Name]

		if authToken.Expiration.Before(maxAllowedExpire) {
			s.Log.Info("Updating AuthToken", "name", inv.Name)
			errs[i] = s.updateToken(inv)
		}
	}

	return errors.Join(errs...)
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
