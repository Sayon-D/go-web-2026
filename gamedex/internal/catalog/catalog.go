package catalog

import (
	"slices"
	"strings"
)

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

// Catalog garde les données en mémoire, chargées une seule fois.
type Catalog struct {
	games       []Game
	tournaments []Tournament
}

func New(games []Game, tournaments []Tournament) *Catalog {
	return &Catalog{games: games, tournaments: tournaments}
}

// Search renvoie les jeux dont le nom contient q (sans tenir compte des majuscules).
// Un q vide renvoie tous les jeux.
func (c *Catalog) Search(q string) []Game {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return slices.Clone(c.games)
	}
	result := []Game{}
	for _, g := range c.games {
		if strings.Contains(strings.ToLower(g.Name), q) {
			result = append(result, g)
		}
	}
	return result
}

func (c *Catalog) Game(id int) (Game, bool) {
	for _, g := range c.games {
		if g.ID == id {
			return g, true
		}
	}
	return Game{}, false
}

func (c *Catalog) TournamentsOf(gameID int) []Tournament {
	result := []Tournament{}
	for _, t := range c.tournaments {
		if t.GameID == gameID {
			result = append(result, t)
		}
	}
	return result
}

// Genres renvoie la liste triée des genres, sans doublon.
func (c *Catalog) Genres() []string {
	var out []string
	for _, g := range c.games {
		out = append(out, g.Genres...)
	}
	slices.Sort(out)
	return slices.Compact(out)
}
