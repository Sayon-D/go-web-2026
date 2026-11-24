package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	port := flag.Int("port", 8080, "port d'écoute du serveur")
	flag.Parse()

	mux := http.NewServeMux()

	// Étape 2 : trois routes qui renvoient du texte
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bienvenue sur GameDex")
	})

	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Liste des jeux")
	})

	// Étape 3 : 400 si l'id n'est pas un nombre, 404 s'il est inconnu
	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "id invalide", http.StatusBadRequest)
			return
		}
		if id < 1 || id > 3 { // exemple : seuls les jeux 1 à 3 existent
			http.Error(w, "jeu introuvable", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "Détail du jeu %d\n", id)
	})

	// Étape 4 : fichiers statiques
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Étape 5 : port configurable
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serveur démarré sur http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
