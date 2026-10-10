package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// Inventario describes a service to bridge
type Inventario struct {
	// Identification name
	Name string `json:"name"`

	Username string `json:"username"`
	Password string `json:"password"`

	Addr   string `json:"addr"`
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}

func (inv *Inventario) updateToken() (AuthToken, error) {
	cj, err := cookiejar.New(nil)
	if err != nil {
		return AuthToken{}, fmt.Errorf("cannot create cookie jar: %w", err)
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
		return AuthToken{}, fmt.Errorf("cannot POST login, inventario: %s: %w", inv.Name, err)
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
		return AuthToken{}, fmt.Errorf("cookie not present, are credentials correct??")
	}

	return AuthToken{
		Token:      cookieValue,
		Expiration: cookieExpiration,
	}, nil
}
