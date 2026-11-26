package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"gamedex/internal/api"
	"gamedex/internal/catalog"
	"gamedex/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "port d'écoute du serveur")
	apiURL := flag.String("api", "http://localhost:8081/api", "adresse de la mini API")
	webDir := flag.String("web", "web", "dossier contenant templates/ et static/")
	flag.Parse()

	// Chargement unique au démarrage
	client := api.New(*apiURL)
	games, err := client.Games()
	if err != nil {
		log.Fatalf("chargement des jeux impossible : %v", err)
	}
	tournaments, err := client.Tournaments()
	if err != nil {
		log.Fatalf("chargement des tournois impossible : %v", err)
	}
	log.Printf("%d jeux et %d tournois chargés", len(games), len(tournaments))

	cat := catalog.New(games, tournaments)
	srv, err := server.New(cat, *webDir)
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serveur démarré sur http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, srv.Handler()))
}
