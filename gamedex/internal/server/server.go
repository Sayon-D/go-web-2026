package server

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gamedex/internal/catalog"
	"gamedex/internal/store"
)

// GameSource : ce dont le serveur a besoin pour chercher des jeux.
// Le serveur ne sait pas que c'est une API derrière.
type GameSource interface {
	Games(q, genre string) ([]catalog.Game, error)
}

type Server struct {
	catalog *catalog.Catalog
	source  GameSource
	tmpl    *template.Template
	webDir  string
}

// PageData : toutes les données dont les templates ont besoin.
type PageData struct {
	Title       string
	Q           string
	Genre       string
	Genres      []string
	Games       []catalog.Game
	Game        catalog.Game
	Likes       int
	Tournaments []catalog.Tournament
	Inscrits    map[int][]string
	Status      int
	Message     string
}

func New(c *catalog.Catalog, src GameSource, webDir string) (*Server, error) {
	tmpl, err := template.ParseGlob(filepath.Join(webDir, "templates", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("chargement des templates : %w", err)
	}
	return &Server{catalog: c, source: src, tmpl: tmpl, webDir: webDir}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.accueil)
	mux.HandleFunc("GET /jeux", s.listeJeux)
	mux.HandleFunc("GET /jeux/{id}", s.ficheJeu)
	mux.HandleFunc("POST /jeux/{id}/like", s.likeJeu)
	mux.HandleFunc("POST /jeux/{id}/tournois/{tid}/inscription", s.inscription)
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
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	genre := r.URL.Query().Get("genre")

	games, err := s.source.Games(q, genre)
	if err != nil {
		log.Println(err)
		s.renderError(w, http.StatusServiceUnavailable, "L'API des jeux est indisponible, réessayez plus tard.")
		return
	}

	// Bonus A : les jeux les plus likés en premier (tri stable)
	counts := store.LikesMap()
	slices.SortStableFunc(games, func(a, b catalog.Game) int {
		return counts[b.ID] - counts[a.ID]
	})

	s.render(w, http.StatusOK, "jeux.html", PageData{
		Title:  "Les jeux",
		Q:      q,
		Genre:  genre,
		Genres: s.catalog.Genres(),
		Games:  games,
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

	tournaments := s.catalog.TournamentsOf(id)
	inscrits := map[int][]string{}
	for _, t := range tournaments {
		inscrits[t.ID] = store.Inscrits(t.ID)
	}

	s.render(w, http.StatusOK, "jeu.html", PageData{
		Title:       game.Name,
		Game:        game,
		Likes:       store.Likes(id),
		Tournaments: tournaments,
		Inscrits:    inscrits,
	})
}

func (s *Server) likeJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.renderError(w, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre.")
		return
	}
	if _, ok := s.catalog.Game(id); !ok {
		s.renderError(w, http.StatusNotFound, "Ce jeu n'existe pas.")
		return
	}
	store.AddLike(id)
	// 303 : le navigateur refait un GET sur la fiche, donc F5 ne reposte pas le like.
	http.Redirect(w, r, fmt.Sprintf("/jeux/%d", id), http.StatusSeeOther)
}

// Bonus B : inscription à un tournoi avec un pseudo
func (s *Server) inscription(w http.ResponseWriter, r *http.Request) {
	id, err1 := strconv.Atoi(r.PathValue("id"))
	tid, err2 := strconv.Atoi(r.PathValue("tid"))
	if err1 != nil || err2 != nil {
		s.renderError(w, http.StatusBadRequest, "Identifiant invalide.")
		return
	}
	if _, ok := s.catalog.Game(id); !ok {
		s.renderError(w, http.StatusNotFound, "Ce jeu n'existe pas.")
		return
	}
	found := false
	for _, t := range s.catalog.TournamentsOf(id) {
		if t.ID == tid {
			found = true
		}
	}
	if !found {
		s.renderError(w, http.StatusNotFound, "Ce tournoi n'existe pas.")
		return
	}
	pseudo := strings.TrimSpace(r.FormValue("pseudo"))
	if pseudo == "" || len(pseudo) > 30 {
		s.renderError(w, http.StatusBadRequest, "Le pseudo doit faire entre 1 et 30 caractères.")
		return
	}
	store.AddInscription(tid, pseudo)
	http.Redirect(w, r, fmt.Sprintf("/jeux/%d", id), http.StatusSeeOther)
}
