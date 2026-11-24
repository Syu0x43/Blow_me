package main

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"os"

	"golang.org/x/term"
)

const (
	W = 16
	H = 12
)

type cell struct {
	mine, open, flaged bool
	neighboors         int
	face               string
}

type game struct {
	minesQuant int
	g          [H][W]cell
	placed     bool
	flags      int
	status     string
}

type cursor struct {
	x int
	y int
}

func (g *game) reset() {
	mines := 12 + rand.IntN(15)
	g.minesQuant = mines
	g.placed = true
	g.flags = mines

	for i := range g.g {
		for j := range g.g[i] {
			g.g[i][j] = cell{open: false, neighboors: 0, face: "[]"}
		}
	}

	g.generateMines()
	g.computeNeighbors()
}

func (g *game) checkWin() bool {
	flaggedMines := 0
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			c := g.g[y][x]
			if c.flaged && !c.mine {
				return false
			}
			if c.mine && c.flaged {
				flaggedMines++
			}
		}
	}
	return flaggedMines == g.minesQuant
}

func (g *game) generateMines() {
	for placed := 0; placed < g.minesQuant; {
		x, y := rand.IntN(W), rand.IntN(H)
		if g.g[y][x].mine {
			continue
		}
		g.g[y][x].mine = true
		placed++
	}
}

func (g *game) computeNeighbors() {
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if g.g[y][x].mine {
				continue
			}

			count := 0
			// Check all 8 surrounding cells
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < W && ny >= 0 && ny < H && g.g[ny][nx].mine {
						count++
					}
				}
			}
			g.g[y][x].neighboors = count
		}
	}
}

func (g *game) reveal(x, y int) bool {
	if x < 0 || x >= W || y < 0 || y >= H {
		return false
	}

	c := &g.g[y][x]

	if c.open || c.flaged {
		return false
	}

	c.open = true

	if c.mine {
		c.face = "* "
		return true
	}

	if c.neighboors > 0 {
		c.face = fmt.Sprintf("%d ", c.neighboors)
		return false
	}

	c.face = "  "
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			g.reveal(x+dx, y+dy)
		}
	}
	return false
}

func (g *game) swap(cur cursor, flagIt bool) {
	// Map 1-based double-width terminal cursor to 0-based grid coordinates
	gridX := (cur.x - 1) / 2
	gridY := cur.y - 1

	if gridX < 0 || gridX >= W || gridY < 0 || gridY >= H {
		return
	}

	c := &g.g[gridY][gridX]

	g.status = ""

	if flagIt {
		if c.open {
			return
		}

		if c.flaged {
			c.flaged = false
			c.face = "[]"
			g.flags++
		} else {
			if g.flags <= 0 {
				return
			}
			c.flaged = true
			c.face = "x "
			g.flags--
		}

		if g.flags == 0 {
			won := true
			for y := 0; y < H && won; y++ {
				for x := 0; x < W; x++ {
					if g.g[y][x].mine != g.g[y][x].flaged {
						won = false
						break
					}
				}
			}
			if won {
				g.status = "U WON"
				g.reset()
			}
		}
	} else {
		hitMine := g.reveal(gridX, gridY)
		if hitMine {
			g.status = "K-boom"
			g.reset()
		}
	}
}

func (g *game) draw() {
	cur := cursor{1, 10}

	buf := make([]byte, 1)
	var outBuffer bytes.Buffer
	outBuffer.WriteString("\033[2J")
	os.Stdout.Write(outBuffer.Bytes())

	for {
		outBuffer.Reset()

		outBuffer.WriteString("\x1b[1;1H")

		for i, row := range g.g {
			for column := range row {
				outBuffer.WriteString(g.g[i][column].face)
			}
			outBuffer.WriteString("\r\n")

		}

		fmt.Fprintf(&outBuffer, "Flags: %d  %s\x1b[K\r\n", g.flags, g.status)
		fmt.Fprintf(&outBuffer, "\x1b[%d;%dH", cur.y, cur.x)
		os.Stdout.Write(outBuffer.Bytes())
		// TODO debug fmt.Fprintf(&outBuffer, "\r\n CurCount: %d x , %d y", cur.x, cur.y)

		fmt.Fprintf(&outBuffer, "\x1b[%d;%dH", cur.y, cur.x)

		os.Stdout.Write(outBuffer.Bytes())

		if _, err := os.Stdin.Read(buf); err != nil {
			break
		}
		if buf[0] == 'q' || buf[0] == 3 {
			break
		}
		switch buf[0] {
		case 'w':
			if cur.y-1 > 0 {
				cur.y -= 1
			}
		case 's':
			if cur.y+1 <= H {
				cur.y += 1
			}

		case 'a':
			if cur.x-1 > 0 {
				cur.x -= 2
			}
		case 'd':
			if cur.x+2 <= (W * 2) { // Width times two cause each [] owns 2 spaces
				cur.x += 2
			}
		case 'f':
			g.swap(cur, true)
		case ' ':
			g.swap(cur, false)
		}
	}

	outBuffer.WriteString("\033[2J")
	os.Stdout.Write(outBuffer.Bytes())
}

func main() {
	fd := int(os.Stdin.Fd())

	// Ensure Stdin is actually a terminal before switching modes
	if !term.IsTerminal(fd) {
		fmt.Println("Not running in a terminal")
		return
	}

	// Switch to raw mode and save the previous state
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	// Restore the terminal state before exiting
	defer term.Restore(fd, oldState)

	var g game
	g.reset()
	g.draw()
}
