package main

type Player struct {
	ID  rune
	Pos Point
}

// Replaced by Game.PlayerWindow()
// func (p *Player) Window(g *Game) [][]rune {
// 	// get copy of window for the player
// 	windowRect := g.Board.WindowRect(RectFromCenterAndSize(
// 		p.Pos,
// 		g.WindowSize,
// 	))
//
// 	window := windowCopy(g.Board.Slice(windowRect))
//
// 	// add in all players
// 	for _, player := range g.Players {
// 		if windowRect.Contains(player.Pos) {
// 			relPos := player.Pos.Rel(windowRect.TopLeft)
// 			window[relPos.Y][relPos.X] = player.ID
// 		}
// 	}
//
// 	// make sure the requested player is on top
// 	relPos := p.Pos.Rel(windowRect.TopLeft)
// 	window[relPos.Y][relPos.X] = p.ID
// 	return window
// }

// Replaced by Game.PlayerMoveLeft()
// func (p *Player) MoveLeft(board Board) bool {
// 	if p.Pos.X != board.Left() {
// 		if board.IsPath(Point{p.Pos.X - 1, p.Pos.Y}) {
// 			p.Pos.X -= 1
// 			return true
// 		}
// 	}
// 	return false
// }

// Replaced by Game.PlayerMoveRight()
// func (p *Player) MoveRight(board Board) bool {
// 	if p.Pos.X != board.Right() {
// 		log_(p.Pos.X + 1)
// 		if board.IsPath(Point{p.Pos.X + 1, p.Pos.Y}) {
// 			p.Pos.X += 1
// 			return true
// 		}
// 	}
// 	return false
// }

// Replaced by Game.PlayerMoveUp()
// func (p *Player) MoveUp(board Board) bool {
// 	if p.Pos.Y != board.Top() {
// 		if board.IsPath(Point{p.Pos.X, p.Pos.Y - 1}) {
// 			p.Pos.Y -= 1
// 			return true
// 		}
// 	}
// 	return false
// }

// Replaced by Game.PlayerMoveDown()
// func (p *Player) MoveDown(board Board) bool {
// 	if p.Pos.Y != board.Bottom() {
// 		if board.IsPath(Point{p.Pos.X, p.Pos.Y + 1}) {
// 			p.Pos.Y += 1
// 			return true
// 		}
// 	}
// 	return false
// }
