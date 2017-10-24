package main

import (
	"sync"
	"time"
)

type GameSession struct {
	Mutex   sync.Mutex
	Game    Game
	UserMap map[rune]*UserSession
	Winner  rune
}

func (gs *GameSession) Broadcast() {
	gs.Mutex.Lock()
	defer gs.Mutex.Unlock()
	gs.broadcast()
}

// broadcast assumes the mutex is already locked
func (gs *GameSession) broadcast() {
	players := make([]rune, 0, len(gs.UserMap))
	for pid := range gs.UserMap {
		players = append(players, pid)
	}

	var winner string
	if gs.Game.Winner != 0 {
		winner = string(gs.Game.Winner)
	}
	solvedTimes := make(map[string]time.Duration, len(gs.Game.SolvedTimes))
	for pid, duration := range gs.Game.SolvedTimes {
		solvedTimes[string(pid)] = duration
	}

	for pid, session := range gs.UserMap {
		session.NotifyUserState(UserState{
			Mode: ModeGame,
			GameState: &GameState{
				Token:       pid,
				Window:      gs.Game.PlayerWindow(pid),
				Players:     players,
				Winner:      winner,
				SolvedTimes: solvedTimes,
				GameStart:   gs.Game.Start,
			},
		})
	}
}

func (gs *GameSession) PlayerMove(pid rune, dir Dir) {
	gs.Mutex.Lock()
	defer gs.Mutex.Unlock()
	gs.Game = gs.Game.PlayerMove(pid, dir)
	// TODO: Move these into the user session loop?
	gs.broadcast()
}

func (gs *GameSession) DropPlayer(user *UserSession) {
	gs.Mutex.Lock()
	defer gs.Mutex.Unlock()
	for pid, u := range gs.UserMap {
		if u == user {
			gs.Game = gs.Game.DropPlayer(pid)
			delete(gs.UserMap, pid)
			return
		}
	}
	// TODO: Move these into the user session loop?
	gs.broadcast()
}

func (gs *GameSession) AddPlayer(pid rune, user *UserSession) {
	gs.Mutex.Lock()
	defer gs.Mutex.Unlock()
	gs.Game = gs.Game.AddPlayer(pid)
	gs.UserMap[pid] = user
	user.GameStart(PlayerSession{Token: pid, GameSession: gs})
}
