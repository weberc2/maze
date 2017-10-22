package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kr/pretty"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	GameManager GameManager
}

func (s *Server) Stats(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket connection:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		<-t.C
		pretty.Println(s.GameManager.State())
		if err := conn.WriteJSON(s.GameManager.State()); err != nil {
			log.Println("Error writing to websocket:", err)
			return
		}
	}
}

func (s *Server) User(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket connection:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	userSession := NewUserSession(&s.GameManager, conn)
	userSession.Run()
}
