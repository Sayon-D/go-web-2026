package server

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"gamedex/internal/catalog"
)

type Server struct {
	catalog *catalog.Catalog
	tmpl    *template.Template
	webDir  string
}

// PageData : toutes les données dont les templates ont besoin.
type PageData struct {
	Title       string
	Q           string
	Games       []catalog.Game
	Game        catalog.Game
	Tournaments []catalog.Tournament
	Status      int
	Message     string
}

func New(c *catalog.Catalog, webDir string) (*Server, error) {
	tmpl, err := template.ParseGlob(filepath.Join(webDir, "templates", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("chargement des templates : %w", err)
	}
	return &Server{catalog: c, tmpl: tmpl, webDir: webDir}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.accueil)
	mux.HandleFunc("GET /jeux", s.listeJeux)
	mux.HandleFunc("GET /jeux/{id}", s.ficheJeu)
	static := http.FileServer(http.Dir(filepath.Join(s.webDir, "static")))
	mux.Handle("GET /static/", http.StripPrefix("/static/", static))
	mux.HandleFunc("/", s.notFound)
	return mux
}

func (s *Server) render(w http.ResponseWriter, status int, name string, data PageData) {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Println("template :", err)
		http.Error(w, "erreur interne", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) renderError(w http.ResponseWriter, status int, msg string) {
	s.render(w, status, "erreur.html", PageData{Title: "Erreur", Status: status, Message: msg})
}

func (s *Server) accueil(w http.ResponseWriter, r *http.Request) {
	s.render(w, http.StatusOK, "accueil.html", PageData{Title: "Accueil"})
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.renderError(w, http.StatusNotFound, "Cette page n'existe pas.")
}

func (s *Server) listeJeux(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	s.render(w, http.StatusOK, "jeux.html", PageData{
		Title: "Les jeux",
		Q:     q,
		Games: s.catalog.Search(q),
	})
}

func (s *Server) ficheJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.renderError(w, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre.")
		return
	}
	game, ok := s.catalog.Game(id)
	if !ok {
		s.renderError(w, http.StatusNotFound, "Ce jeu n'existe pas.")
		return
	}
	s.render(w, http.StatusOK, "jeu.html", PageData{
		Title:       game.Name,
		Game:        game,
		Tournaments: s.catalog.TournamentsOf(id),
	})
}
