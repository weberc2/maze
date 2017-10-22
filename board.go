package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/itchyny/maze"
)

const (
	tileWall  = '#'
	tileSpace = ' '
)

type Dir int

const (
	Left Dir = iota
	Right
	Up
	Down
)

type Point struct {
	X, Y int
}

func (p Point) Translate(d Dir) Point {
	switch d {
	case Left:
		return p.Left()
	case Right:
		return p.Right()
	case Up:
		return p.Up()
	case Down:
		return p.Down()
	default:
		panic(fmt.Sprint("Not a dir:", d))
	}
}

func (p Point) Left() Point {
	return Point{p.X - 1, p.Y}
}

func (p Point) Right() Point {
	return Point{p.X + 1, p.Y}
}

func (p Point) Up() Point {
	return Point{p.X, p.Y - 1}
}

func (p Point) Down() Point {
	return Point{p.X, p.Y + 1}
}

func (p Point) Offset(offset Point) Point {
	return Point{p.X + offset.X, p.Y + offset.Y}
}

func (p Point) Rel(other Point) Point {
	return Point{p.X - other.X, p.Y - other.Y}
}

type Rect struct {
	TopLeft     Point
	BottomRight Point
}

func (r Rect) Contains(p Point) bool {
	return p.X >= r.TopLeft.X &&
		p.X <= r.BottomRight.X &&
		p.Y >= r.TopLeft.Y &&
		p.Y <= r.BottomRight.Y
}

func RectFromCenterAndSize(center Point, size Point) Rect {
	return Rect{
		TopLeft:     Point{center.X - (size.X / 2), center.Y - (size.Y / 2)},
		BottomRight: Point{center.X + (size.X / 2), center.Y + (size.Y / 2)},
	}
}

type Board struct {
	Rows  [][]rune
	Start Point
	End   Point
}

func (b *Board) Width() int {
	if b.Height() > 0 {
		return len(b.Rows[0])
	}
	return 0
}
func (b *Board) Height() int { return len(b.Rows) }
func (b *Board) Left() int   { return 0 }
func (b *Board) Right() int  { return b.Width() - 1 }
func (b *Board) Top() int    { return 0 }
func (b *Board) Bottom() int { return b.Height() - 1 }
func (b *Board) IsPath(p Point) bool {
	return p.X >= b.Left() &&
		p.X <= b.Right() &&
		p.Y >= b.Top() &&
		p.Y <= b.Bottom() &&
		b.Rows[p.Y][p.X] != tileWall
}

func (b *Board) Slice(r Rect) [][]rune {
	rows := b.Rows[r.TopLeft.Y : r.BottomRight.Y+1]
	outrows := make([][]rune, len(rows))
	for i, row := range rows {
		outrows[i] = row[r.TopLeft.X : r.BottomRight.X+1]
	}
	return outrows
}

func (b *Board) WindowRect(r Rect) Rect {
	if r.TopLeft.X < b.Left() {
		r.TopLeft.X = b.Left()
	}
	if r.BottomRight.X > b.Right() {
		r.BottomRight.X = b.Right()
	}
	if r.TopLeft.Y < b.Top() {
		r.TopLeft.Y = b.Top()
	}
	if r.BottomRight.Y > b.Bottom() {
		r.BottomRight.Y = b.Bottom()
	}
	return r
}

func windowCopy(window [][]rune) [][]rune {
	copy := make([][]rune, len(window))
	for y, row := range window {
		copy[y] = make([]rune, len(row))
		for x, r := range row {
			copy[y][x] = r
		}
	}
	return copy
}

func windowToString(window [][]rune) string {
	out := ""
	for _, row := range window {
		out += string(row) + "\n"
	}
	return out
}

func ParseBoard(r io.Reader) (Board, error) {
	var start *Point
	var end *Point

	scanner := bufio.NewScanner(r)
	var rows [][]rune
	for scanner.Scan() {
		rows = append(rows, []rune(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return Board{}, nil
	}

	w := len(rows[0])
	for y, row := range rows {
		if len(row) != w {
			return Board{}, fmt.Errorf(
				"Wanted row size %d; got %d",
				w,
				len(row),
			)
		}
		for x, r := range row {
			if r == 'E' {
				if end != nil {
					return Board{}, fmt.Errorf(
						"Start position already exists at (%d, %d), but "+
							"encountered a second end token at (%d, %d)",
						end.X,
						end.Y,
						x,
						y,
					)
				}
				end = &Point{X: x, Y: y}
				continue
			}
			if r == 'S' {
				if start != nil {
					return Board{}, fmt.Errorf(
						"Start position already exists at (%d, %d), but "+
							"encountered a second start token at (%d, %d)",
						start.X,
						start.Y,
						x,
						y,
					)
				}
				start = &Point{X: x, Y: y}
				continue
			}
			if r != '#' && r != ' ' {
				return Board{}, fmt.Errorf(
					"Illegal character at (%d, %d): %#v",
					x,
					y,
					r,
				)
			}
		}
	}

	if start == nil {
		return Board{}, fmt.Errorf("Start position not found")
	}
	if end == nil {
		return Board{}, fmt.Errorf("End position not found")
	}

	return Board{Rows: rows, Start: *start, End: *end}, nil
}

func GenerateBoard(seed int64, w, h int) Board {
	rand.Seed(seed)
	buf := bytes.NewBuffer([]byte{})
	m := maze.NewMaze(h, w)
	m.Generate()
	m.Print(buf, &maze.Format{
		Path:      " ",
		Wall:      "#",
		StartLeft: "S",
		GoalRight: "E",
	})
	lines := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	for i, line := range lines {
		lines[i] = strings.Trim(line, " ")
	}
	board, err := ParseBoard(strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		panic("Error parsing generated board: " + err.Error())
	}
	return board
}
