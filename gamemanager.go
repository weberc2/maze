package main

import (
	"sync"
	"time"
)

var seed = time.Now().UnixNano()

const boardWidth = 30
const boardHeight = 15

type Lobby struct {
	Mutex   sync.RWMutex
	Users   []*UserSession
	Game    *GameSession
	MaxSize int
}

func (l *Lobby) Broadcast() {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	var userState UserState
	if l.Game != nil {
		l.Game.Broadcast()
	} else {
		userState = UserState{
			Mode:       ModeLobby,
			LobbyState: l.lobbyState(),
		}
		for _, user := range l.Users {
			user.NotifyUserState(userState)
		}
	}
}

func (l *Lobby) lobbyState() *LobbyState {
	return &LobbyState{
		Players:    len(l.Users),
		Total:      l.MaxSize,
		InProgress: l.Game != nil,
	}
}

// Add adds a user to the lobby and starts the game if the lobby is full.
// Starting the game implies notifying the players that the game has started.
// The return value indicates whether or not the player was successfully added.
// In partiuclar, `false` means the lobby is full.
func (l *Lobby) Add(user *UserSession) bool {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	if l.Game == nil {
		l.Users = append(l.Users, user)
		if len(l.Users) >= l.MaxSize {
			l.startGame()
		}
		return true
	}
	return false
}

// Drops the user if found in the lobby. Returns a `bool` indicating whether or
// not the user was found (and consequently whether or not the drop succeeded)
// and an `int` representing the count of remaining users.
func (l *Lobby) Drop(user *UserSession) (bool, int) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	for i, u := range l.Users {
		if u == user {
			if l.Game != nil {
				l.Game.DropPlayer(user)
				user.ClearGame()
			}
			l.Users = append(l.Users[:i], l.Users[i+1:]...)
			return true, len(l.Users)
		}
	}
	return false, len(l.Users)
}

// This method assumes the mutex is already locked; it has a side effect of
// creating the game session and notifying all players that the game has
// started
func (l *Lobby) startGame() {
	l.Game = &GameSession{
		Game: Game{
			Board:      GenerateBoard(seed, boardWidth, boardHeight),
			Players:    make([]Player, len(l.Users)),
			WindowSize: Point{41, 21},
		},
		UserMap: make(map[rune]*UserSession, len(l.Users)),
	}
	for i, user := range l.Users {
		l.Game.AddPlayer(playerTokens[i], user)
	}
}

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
