package main

import (
	"sync"
	"time"
)

var seed = time.Now().UnixNano()

const boardWidth = 10
const boardHeight = 5

var playerTokens = []rune("@$")

type GameManager struct {
	Mutex   sync.RWMutex
	Lobbies []*Lobby
}

func (gm *GameManager) State() []LobbyState {
	gm.Mutex.RLock()
	defer gm.Mutex.RUnlock()
	lobbies := make([]LobbyState, len(gm.Lobbies))
	for i, lobby := range gm.Lobbies {
		lobbies[i] = *lobby.lobbyState()
	}
	return lobbies
}

func (gm *GameManager) Join(user *UserSession) *Lobby {
	gm.Mutex.Lock()
	defer gm.Mutex.Unlock()
	for _, lobby := range gm.Lobbies {
		if lobby.Add(user) {
			return lobby
		}
	}
	lobby := &Lobby{MaxSize: len(playerTokens)}
	gm.Lobbies = append(gm.Lobbies, lobby)
	if !lobby.Add(user) {
		// shouldn't get here unless playerTokens is empty (and thus
		// lobby.MaxSize is zero), which shouldn't happen.
		panic("Couldn't add user to lobby")
	}
	return lobby
}

func (gm *GameManager) Drop(user *UserSession) {
	gm.Mutex.Lock()
	defer gm.Mutex.Unlock()

	for i, lobby := range gm.Lobbies {
		if success, count := lobby.Drop(user); success {
			// If there are no more players in the lobby, remove it
			if count < 1 {
				gm.Lobbies = append(gm.Lobbies[:i], gm.Lobbies[i+1:]...)
				return
			}
			// Since there is at least one player left in the lobby, broadcast
			// the player drop to remaining players
			lobby.Broadcast()
			return
		}
	}

	// We should never get here
	panic("User not found")
}
