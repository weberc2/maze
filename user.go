package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type PlayerSession struct {
	Token       rune
	GameSession *GameSession
}

func (ps *PlayerSession) Move(dir Dir) {
	ps.GameSession.PlayerMove(ps.Token, dir)
}

type Mode int

const (
	ModeMatchMaking Mode = iota
	ModeLobby
	ModeGame
)

func (m Mode) MarshalJSON() ([]byte, error) {
	return json.Marshal(func() string {
		switch m {
		case ModeMatchMaking:
			return "MODE_MATCHMAKING"
		case ModeLobby:
			return "MODE_LOBBY"
		case ModeGame:
			return "MODE_GAME"
		default:
			panic(fmt.Sprint("Invalid mode:", m))
		}
	}())
}

type LobbyState struct {
	Players    int  `json:"players"`
	Total      int  `json:"total"`
	InProgress bool `json:"in_progress"`
}

type GameState struct {
	Token   rune   `json:"token"`
	Window  string `json:"window"`
	Players []rune `json:"players"`
	Winner  string `json:"winner,omitempty"`
}

type UserState struct {
	Mode       Mode        `json:"mode"`
	LobbyState *LobbyState `json:"lobby_state,omitempty"`
	GameState  *GameState  `json:"game_state,omitempty"`
}

type UserSession struct {
	writeLock     sync.Mutex
	conn          *websocket.Conn
	lock          sync.Mutex
	playerSession *PlayerSession
	gameManager   *GameManager
	logger        *Logger
}

func NewUserSession(
	gm *GameManager,
	conn *websocket.Conn,
	logger *Logger,
) *UserSession {
	return &UserSession{conn: conn, gameManager: gm, logger: logger}
}

func (user *UserSession) GameStart(playerSession PlayerSession) {
	user.lock.Lock()
	user.playerSession = &playerSession
	user.lock.Unlock()
}

func (user *UserSession) NotifyUserState(userState UserState) {
	user.writeLock.Lock()
	defer user.writeLock.Unlock()
	if err := user.conn.WriteJSON(userState); err != nil {
		user.logger.Logf(
			"Error writing UserState to websocket (terminating): %v",
			err,
		)
		// There shouldn't be any errors marshaling json, so any errors must be
		// I/O errors; if there is an I/O error, we should probably just quit
		user.quit()
	}
}

func (user *UserSession) quit() {
	user.gameManager.Drop(user)
	user.writeLock.Lock()
	defer user.writeLock.Unlock()
	user.conn.Close()
}

func (user *UserSession) ClearGame() {
	user.lock.Lock()
	user.playerSession = nil
	user.lock.Unlock()
}

func (user *UserSession) returnToMatchMaking(lobby *Lobby) error {
	user.gameManager.Drop(user)
	return user.lobbyMode(user.gameManager.Join(user))
}

func (user *UserSession) isGameMode() bool {
	user.lock.Lock()
	defer user.lock.Unlock()
	return user.playerSession != nil
}

func (user *UserSession) gameMode(lobby *Lobby) error {
	for {
		_, data, err := user.conn.ReadMessage()
		if err != nil {
			user.logger.Logf("Error reading message: %v", err)
			user.quit()
			return err
		}

		switch ev := string(data); ev {
		case "rtmm":
			return user.returnToMatchMaking(lobby)
		case "left":
			user.playerSession.Move(Left)
		case "right":
			user.playerSession.Move(Right)
		case "up":
			user.playerSession.Move(Up)
		case "down":
			user.playerSession.Move(Down)
		}
	}
}

func (user *UserSession) lobbyMode(lobby *Lobby) error {
	lobby.Broadcast()
	for {
		if user.isGameMode() {
			return user.gameMode(lobby)
		}

		_, data, err := user.conn.ReadMessage()
		if err != nil {
			user.quit()
			return err
		}

		if string(data) == "rtmm" {
			return user.returnToMatchMaking(lobby)
		}
	}
}

func (user *UserSession) Run() error {
	return user.lobbyMode(user.gameManager.Join(user))
}
