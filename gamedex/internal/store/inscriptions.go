package store

import "sync"

var (
	muInscr      sync.Mutex
	inscriptions = map[int][]string{} // numéro du tournoi -> pseudos
)

// AddInscription inscrit un pseudo à un tournoi et renvoie la liste à jour.
func AddInscription(tournamentID int, pseudo string) []string {
	muInscr.Lock()
	defer muInscr.Unlock()
	inscriptions[tournamentID] = append(inscriptions[tournamentID], pseudo)
	return append([]string(nil), inscriptions[tournamentID]...)
}

// Inscrits renvoie une copie des pseudos inscrits à un tournoi.
func Inscrits(tournamentID int) []string {
	muInscr.Lock()
	defer muInscr.Unlock()
	return append([]string(nil), inscriptions[tournamentID]...)
}
