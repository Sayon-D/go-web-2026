package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gamedex/internal/catalog"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *Client) get(path string, dst any) error {
	url := c.baseURL + path
	resp, err := c.http.Get(url)
	if err != nil {
		return fmt.Errorf("API injoignable (%s) : %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("l'API a répondu %d pour %s", resp.StatusCode, url)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func (c *Client) Games() ([]catalog.Game, error) {
	var games []catalog.Game
	err := c.get("/games", &games)
	return games, err
}

// Tournaments récupère tous les tournois en un seul appel.
func (c *Client) Tournaments() ([]catalog.Tournament, error) {
	var ts []catalog.Tournament
	err := c.get("/tournaments", &ts)
	return ts, err
}
