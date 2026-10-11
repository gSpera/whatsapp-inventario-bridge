package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

	// Update token if necessary
	for _, inv := range s.Inventario {
		// If not AuthToken is found the zero value will be used,
		// for time.Time this is timestamp 0, which is in the past.
		s.updateTokenIfNecessary(inv)
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

func (s *Server) updateTokenIfNecessary(inv Inventario) error {
	authToken, _ := s.AuthTokens[inv.Name]
	maxAllowedExpire := time.Now().AddDate(0, 0, 1)

	if authToken.Expiration.Before(maxAllowedExpire) {
		s.Log.Info("Updating AuthToken", "name", inv.Name)
		authToken, err := inv.updateToken()

		if err != nil {
			s.Log.Error("While updating", "name", inv.Name, "err", err)
		} else {
			s.AuthTokens[inv.Name] = authToken
		}
	}

	return nil
}

func (s *Server) inventarioFromPrefixSuffix(prefix, suffix string) ServerInterface {

	for i, inv := range s.Inventario {
		if inv.Prefix == prefix && inv.Suffix == suffix {
			return serverInventario{
				s:   s,
				inv: &s.Inventario[i],
			}
		}
	}

	s.Log.Warn("Requested inventario not found", "prefix", prefix, "suffix", suffix)
	return nil
}

type serverInventario struct {
	s   *Server
	inv *Inventario
}

var _ ServerInterface = serverInventario{}

func (s serverInventario) OrderInfo(orderID int) (order, companyName, city string, err error) {
	url := fmt.Sprintf("%s/api/info-order?id=%d", s.inv.Addr, orderID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("cannot create request to inventario: %w", err)
		return
	}

	req.AddCookie(&http.Cookie{
		Name:  "user",
		Value: s.s.AuthTokens[s.inv.Name].Token,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("cannot get order info: %w", err)
		return
	}
	var orderInfo struct {
		Ordine             string `json:"Ordine"`
		CompanyName        string `json:"CompanyName"`
		ComuneDestinatione string `json:"ComuneDestinazione"`
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&orderInfo)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal response: %w", err)
		return
	}

	return orderInfo.Ordine, orderInfo.CompanyName, orderInfo.ComuneDestinatione, nil
}

func (s serverInventario) DeliverOrder(orderID int, note string) error {
	url := fmt.Sprintf("%s/api/update-order", s.inv.Addr)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("cannot create request to inventario: %w", err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "user",
		Value: s.s.AuthTokens[s.inv.Name].Token,
	})

	// Body
	requestBody := struct {
		OrderID int
		State   int `json:",string"`
		When    time.Time
		Note    string
	}{
		OrderID: orderID,
		State:   7, // TODO: Remove hardcoded values
		When:    time.Now(),
		Note:    note,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("cannot marshal request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot get order info: %w", err)

	}
	_ = resp
	return nil
}
