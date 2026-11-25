package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const baseURL = "http://localhost:8081/api"

// Étape 1 : structs
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

// Étape 2 : client avec timeout de 3 secondes
var client = &http.Client{Timeout: 3 * time.Second}

func fetchJSON(url string, dst any) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("appel impossible vers %s : %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("l'API a répondu %d (%s) pour %s",
			resp.StatusCode, http.StatusText(resp.StatusCode), url)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("JSON illisible reçu de %s : %w", url, err)
	}
	return nil
}

func FetchGames() ([]Game, error) {
	var games []Game
	err := fetchJSON(baseURL+"/games", &games)
	return games, err
}

func FetchTournaments(gameID int) ([]Tournament, error) {
	var ts []Tournament
	err := fetchJSON(fmt.Sprintf("%s/games/%d/tournaments", baseURL, gameID), &ts)
	return ts, err
}

func main() {
	games, err := FetchGames()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}

	// Étape 3 : chaque jeu avec son année et son nombre de tournois
	for _, g := range games {
		ts, err := FetchTournaments(g.ID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erreur :", err)
			os.Exit(1)
		}
		fmt.Printf("%s (%d) : %d tournoi(s)\n", g.Name, g.Year, len(ts))
	}
}
