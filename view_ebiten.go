package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const pixelScale = 5 // Pixels per cell, increase for better visibility

var (
	colBg    = color.RGBA{20, 40, 90, 255}    // Ocean color (Background)
	colFish  = color.RGBA{255, 230, 120, 255} // Fish color: Yellow
	colShark = color.RGBA{220, 60, 60, 255}   // Shark color: Red
)


type game struct {
	w       *World // Pointer to the simulation world
	mode    string // Execution mode: "seq" or "par"
	workers int    // Number of threads for parallel mode
	tick    int    // Frame counter to control simulation speed
}


func (g *game) Update() error {
	// Advance one simulation step every two frames
	if g.tick%2 != 0 {
		g.tick++
		return nil
	}

	if g.mode == "par" {
		// Use current time + tick as seed for variety in visualization
		seed := time.Now().UnixNano() + int64(g.tick)
		g.w = StepPar(g.w, g.workers, seed)
	} else {
		g.w = StepSeq(g.w)
	}

	g.tick++
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(colBg)
	n := g.w.Size
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			var c color.Color
			switch g.w.Grid[y][x].Type {
			case Fish:
				c = colFish
			case Shark:
				c = colShark
			default:
				continue
			}
			// Draw a block of pixels for each cell based on pixelScale
			for dy := 0; dy < pixelScale; dy++ {
				for dx := 0; dx < pixelScale; dx++ {
					screen.Set(x*pixelScale+dx, y*pixelScale+dy, c)
				}
			}
		}
	}
}


func (g *game) Layout(outW, outH int) (int, int) {
	return g.w.Size * pixelScale, g.w.Size * pixelScale
}


func runGUI(w *World, mode string, workers int) error {
	g := &game{w: w, mode: mode, workers: workers}
	f0, s0 := Count(w)
	ebiten.SetWindowSize(w.Size*pixelScale, w.Size*pixelScale)
	ebiten.SetWindowTitle(fmt.Sprintf(
		"Wa-Tor | size=%d fish=%d sharks=%d | FB=%d SB=%d Starve=%d | mode=%s workers=%d",
		w.Size, f0, s0, FishBreedSteps, SharkBreedSteps, SharkEnergyInit, mode, workers,
	))
	return ebiten.RunGame(g)
}
