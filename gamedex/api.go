package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const apiURL = "http://localhost:8081/api"

type Game struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Studio string   `json:"studio"`
	Genres []string `json:"genres"`
}

type Tournament struct {
	ID     int     `json:"id"`
	GameID int     `json:"gameId"`
	Name   string  `json:"name"`
	City   string  `json:"city"`
	Date   string  `json:"date"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
}

// APIError : l'API a répondu, mais avec un code autre que 200.
type APIError struct {
	Status int
	URL    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("l'API a répondu %d pour %s", e.Status, e.URL)
}

var apiClient = &http.Client{Timeout: 3 * time.Second}

func fetchJSON(u string, dst any) error {
	resp, err := apiClient.Get(u)
	if err != nil {
		return fmt.Errorf("API injoignable : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Status: resp.StatusCode, URL: u}
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// FetchGames : q est optionnel, l'API sait filtrer avec ?q=
func FetchGames(q string) ([]Game, error) {
	u := apiURL + "/games"
	if q != "" {
		u += "?" + url.Values{"q": {q}}.Encode()
	}
	var games []Game
	err := fetchJSON(u, &games)
	return games, err
}

func FetchGame(id int) (Game, error) {
	var g Game
	err := fetchJSON(fmt.Sprintf("%s/games/%d", apiURL, id), &g)
	return g, err
}

func FetchTournaments(gameID int) ([]Tournament, error) {
	var ts []Tournament
	err := fetchJSON(fmt.Sprintf("%s/games/%d/tournaments", apiURL, gameID), &ts)
	return ts, err
}
