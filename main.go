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
}

type game struct {
	minesQuant int
	g          [H][W]cell
	placed     bool
	flags      int
}

type cursor struct {
	x int
	y int
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

func (g *game) reveal(x, y int) {
}

func (g *game) draw() {
	buf := make([]byte, 1)
	var outBuffer bytes.Buffer

	for {

		outBuffer.Reset()

		for _, row := range g.g {
			for range row {
				outBuffer.WriteString("[]")
			}
			outBuffer.WriteString("\r\n")
		}

		os.Stdout.Write(outBuffer.Bytes())

		if _, err := os.Stdin.Read(buf); err != nil {
			break
		}

		if buf[0] == 'q' || buf[0] == 3 {
			break
		}
	}
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

	g := game{minesQuant: rand.IntN(12), placed: true, flags: 1}

	// Row 0: Open cells with neighbor counts
	g.g[0][0] = cell{open: true, neighboors: 0}
	g.g[0][1] = cell{open: true, neighboors: 1}
	g.g[0][2] = cell{flaged: true, mine: true} // Flagged mine

	// Row 1: Revealed mine (for debugging) & hidden cells
	g.g[1][1] = cell{open: true, mine: true}
	g.g[1][2] = cell{open: true, neighboors: 2}

	g.draw()

	// Exit on 'q' or Ctrl+C (ASCII 3)
}
