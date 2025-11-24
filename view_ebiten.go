package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

/**
 * @file view_ebiten.go
 * @brief Graphical User Interface (GUI) implementation using Ebiten.
 *
 * This file handles the visualization of the Wa-Tor simulation. It defines
 * the game loop (Update/Draw) and renders the grid state to the window.
 */

const pixelScale = 5 // Pixels per cell, increase for better visibility

var (
	colBg    = color.RGBA{20, 40, 90, 255}    // Ocean color (Background)
	colFish  = color.RGBA{255, 230, 120, 255} // Fish color: Yellow
	colShark = color.RGBA{220, 60, 60, 255}   // Shark color: Red
)

/**
 * @brief Represents the game state for the Ebiten engine.
 *
 * Implements the ebiten.Game interface required to run the simulation loop.
 */
type game struct {
	w       *World // Pointer to the simulation world
	mode    string // Execution mode: "seq" or "par"
	workers int    // Number of threads for parallel mode
	tick    int    // Frame counter to control simulation speed
}

/**
 * @brief Updates the game state. Called every tick (default 60Hz).
 *
 * Advances the Wa-Tor simulation. To control the simulation speed and allow
 * visualization, the world state is only updated every few ticks (currently every 2 frames).
 *
 * @return error Returns nil to continue the game.
 */
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

/**
 * @brief Renders the game screen. Called every frame.
 *
 * Iterates through the world grid and draws pixels corresponding to
 * Empty (Ocean), Fish, or Shark cells based on the defined color palette.
 *
 * @param screen The Ebiten image target to draw onto.
 */
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

/**
 * @brief Defines the logical screen size.
 *
 * @param outW Outer window width (ignored).
 * @param outH Outer window height (ignored).
 * @return The logical width and height of the screen.
 */
func (g *game) Layout(outW, outH int) (int, int) {
	return g.w.Size * pixelScale, g.w.Size * pixelScale
}

/**
 * @brief Initializes and runs the GUI simulation.
 *
 * Sets up the window properties (size, title) and starts the Ebiten game loop.
 *
 * @param w Pointer to the initial World state.
 * @param mode Simulation mode ("seq" or "par").
 * @param workers Number of threads for parallel mode.
 * @return error If the game fails to run.
 */
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
