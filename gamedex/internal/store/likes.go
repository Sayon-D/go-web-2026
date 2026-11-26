// snippets/likes.go
package store

import "sync"

// Le serveur traite plusieurs requêtes en même temps.
// Le mutex garantit qu'une seule à la fois touche la map.
var (
	mu    sync.Mutex
	likes = map[int]int{}
)

// AddLike ajoute un like au jeu et renvoie le nouveau total.
func AddLike(gameID int) int {
	mu.Lock()
	defer mu.Unlock()
	likes[gameID]++
	return likes[gameID]
}

// Likes renvoie le nombre de likes d'un jeu.
func Likes(gameID int) int {
	mu.Lock()
	defer mu.Unlock()
	return likes[gameID]
}

// LikesMap renvoie une copie des likes de tous les jeux (pour trier).
func LikesMap() map[int]int {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[int]int, len(likes))
	for id, n := range likes {
		out[id] = n
	}
	return out
}
