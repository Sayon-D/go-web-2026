package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

// PageData : toutes les données dont les templates ont besoin.
type PageData struct {
	Title       string
	Q           string
	Games       []Game
	Game        Game
	Tournaments []Tournament
	Status      int
	Message     string
}

// render exécute le template dans un tampon : si le template échoue,
// on n'a rien envoyé à moitié et on peut répondre proprement.
func render(w http.ResponseWriter, status int, name string, data PageData) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Println("template :", err)
		http.Error(w, "erreur interne", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func renderError(w http.ResponseWriter, status int, msg string) {
	render(w, status, "erreur.html", PageData{Title: "Erreur", Status: status, Message: msg})
}

func accueil(w http.ResponseWriter, r *http.Request) {
	render(w, http.StatusOK, "accueil.html", PageData{Title: "Accueil"})
}

func notFound(w http.ResponseWriter, r *http.Request) {
	renderError(w, http.StatusNotFound, "Cette page n'existe pas.")
}

// Étapes 2 et 5
func listeJeux(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	games, err := FetchGames(q)
	if err != nil {
		log.Println(err)
		renderError(w, http.StatusServiceUnavailable, "L'API des jeux est indisponible, réessayez plus tard.")
		return
	}
	render(w, http.StatusOK, "jeux.html", PageData{Title: "Les jeux", Q: q, Games: games})
}

// Étapes 3 et 4
func ficheJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		renderError(w, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre.")
		return
	}

	game, err := FetchGame(id)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
			renderError(w, http.StatusNotFound, "Ce jeu n'existe pas.")
			return
		}
		log.Println(err)
		renderError(w, http.StatusServiceUnavailable, "L'API des jeux est indisponible, réessayez plus tard.")
		return
	}

	tournaments, err := FetchTournaments(id)
	if err != nil {
		log.Println(err)
		renderError(w, http.StatusServiceUnavailable, "L'API des jeux est indisponible, réessayez plus tard.")
		return
	}

	render(w, http.StatusOK, "jeu.html", PageData{Title: game.Name, Game: game, Tournaments: tournaments})
}

func main() {
	port := flag.Int("port", 8080, "port d'écoute du serveur")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", accueil)
	mux.HandleFunc("GET /jeux", listeJeux)
	mux.HandleFunc("GET /jeux/{id}", ficheJeu)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", notFound) // toute autre URL : page d'erreur 404

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serveur démarré sur http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
