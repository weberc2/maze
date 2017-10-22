package main

import (
	"fmt"
)

type Game struct {
	Board      Board
	Players    []Player
	WindowSize Point
	Winner     rune // non-zero value indicates the game has been won
}

func (g Game) InitPlayer(pid rune) Player {
	return Player{ID: pid, Pos: g.Board.Start}
}

func (g Game) SetPlayers(players []Player) Game {
	return Game{
		Board:      g.Board,
		WindowSize: g.WindowSize,
		Players:    players,
		Winner:     g.Winner,
	}
}

func (g Game) PlayerWindow(pid rune) string {
	for _, p := range g.Players {
		if p.ID == pid {
			rect := RectFromCenterAndSize(
				p.Pos,
				g.WindowSize,
			)
			windowRect := g.Board.WindowRect(rect)

			// get copy of window for the player
			window := windowCopy(g.Board.Slice(windowRect))

			// add in all players
			for _, player := range g.Players {
				if windowRect.Contains(player.Pos) {
					relPos := player.Pos.Rel(windowRect.TopLeft)
					window[relPos.Y][relPos.X] = player.ID
				}
			}

			// make sure the requested player is on top
			relPos := p.Pos.Rel(windowRect.TopLeft)
			window[relPos.Y][relPos.X] = p.ID
			return windowToString(window)
		}
	}
	panic(fmt.Sprintf("Player not found: %s", string(pid)))
}

func (g Game) MapPlayer(pid rune, f func(p Player) Player) Game {
	players := make([]Player, len(g.Players))
	var found bool
	for i, p := range g.Players {
		if p.ID == pid {
			found = true
			players[i] = f(p)
			continue
		}
		players[i] = p
	}
	if !found {
		panic(fmt.Sprintf("Player not found: %#v", pid))
	}
	return g.SetPlayers(players)
}

func (g Game) SetWinner(pid rune) Game {
	return Game{
		Players:    g.Players,
		Board:      g.Board,
		WindowSize: g.WindowSize,
		Winner:     pid,
	}
}

func (g Game) PlayerMove(pid rune, dir Dir) Game {
	players := make([]Player, len(g.Players))
	copy(players, g.Players)
	for i, p := range players {
		if p.ID == pid {
			if proposed := p.Pos.Translate(dir); g.Board.IsPath(proposed) {
				players[i].Pos = proposed
				if g.Winner == 0 && proposed == g.Board.End {
					game := g.SetPlayers(players)
					game.Winner = pid
					return game
				}
				return g.SetPlayers(players)
			}
			return g
		}
	}
	panic(fmt.Sprintf("Player not found: %s", string(pid)))
}

func (g Game) PlayerMoveLeft(pid rune) Game {
	return g.PlayerMove(pid, Left)
}

func (g Game) PlayerMoveRight(pid rune) Game {
	return g.PlayerMove(pid, Right)
}

func (g Game) PlayerMoveUp(pid rune) Game {
	return g.PlayerMove(pid, Up)
}

func (g Game) PlayerMoveDown(pid rune) Game {
	return g.PlayerMove(pid, Down)
}

func (g Game) AddPlayer(pid rune) Game {
	players := make([]Player, len(g.Players))
	copy(players, g.Players)
	return g.SetPlayers(append(players, Player{ID: pid, Pos: g.Board.Start}))
}

func (g Game) DropPlayer(pid rune) Game {
	players := make([]Player, 0, len(g.Players)-1)
	found := false
	for _, p := range g.Players {
		if p.ID != pid {
			found = true
			players = append(players, p)
		}
	}
	if !found {
		panic(fmt.Sprintf("Player not found: %#v", pid))
	}
	return g.SetPlayers(players)
}
