package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	u := c.baseURL + path
	resp, err := c.http.Get(u)
	if err != nil {
		return fmt.Errorf("API injoignable (%s) : %w", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("l'API a répondu %d pour %s", resp.StatusCode, u)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// Games interroge /games avec les filtres optionnels q et genre.
func (c *Client) Games(q, genre string) ([]catalog.Game, error) {
	params := url.Values{}
	if q != "" {
		params.Set("q", q)
	}
	if genre != "" {
		params.Set("genre", genre)
	}
	path := "/games"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	var games []catalog.Game
	err := c.get(path, &games)
	return games, err
}

// Tournaments récupère tous les tournois en un seul appel.
func (c *Client) Tournaments() ([]catalog.Tournament, error) {
	var ts []catalog.Tournament
	err := c.get("/tournaments", &ts)
	return ts, err
}
