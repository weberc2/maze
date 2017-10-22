package main

import "sync"

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
