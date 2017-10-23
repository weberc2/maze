package main

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	GameManager GameManager
}

func (s *Server) Stats(
	w http.ResponseWriter,
	r *http.Request,
	logger *Logger,
) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logf("Error upgrading to websocket connection: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		<-t.C
		if err := conn.WriteJSON(s.GameManager.State()); err != nil {
			logger.Logf("Error writing to websocket: %v", err)
			return
		}
	}
}

func (s *Server) User(w http.ResponseWriter, r *http.Request, logger *Logger) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logf("Error upgrading to websocket connection: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	userSession := NewUserSession(&s.GameManager, conn, logger)
	userSession.Run()
}
