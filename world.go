package main

import (
	"math/rand"
	"time"
)

/**
 * @file world.go
 * @brief Defines the core data structures and initialization logic for the Wa-Tor world.
 *
 * This file contains the definitions for the World grid, Cells, and Agent states (Fish/Shark),
 * as well as helper functions for coordinate wrapping (toroidal geometry) and neighbor calculation.
 */

// Types and Constants

/**
 * @brief Enumeration for the type of object occupying a cell.
 */
type CellType int

const (
	Empty CellType = iota ///< Represents an empty ocean cell.
	Fish                  ///< Represents a cell occupied by a Fish.
	Shark                 ///< Represents a cell occupied by a Shark.
)

// Global configuration variables (set via command line flags in main.go)
var (
	FishBreedSteps  = 3 ///< Chronons a fish must survive to reproduce.
	SharkBreedSteps = 8 ///< Chronons a shark must survive to reproduce.
	SharkEnergyInit = 5 ///< Initial energy of a shark (and energy gained from eating).
	SharkEnergyGain = 2 ///< Energy gained by a shark when eating a fish.
	SharkEnergyLoss = 1 ///< Energy lost by a shark each chronon.
)

/**
 * @brief State information specific to a Fish.
 */
type FishState struct {
	BreedTimer int ///< Chronons survived since last reproduction.
}

/**
 * @brief State information specific to a Shark.
 */
type SharkState struct {
	BreedTimer int ///< Chronons survived since last reproduction.
	Energy     int ///< Current energy level. Shark dies if this reaches 0.
}

/**
 * @brief Represents a single grid unit in the simulation.
 */
type Cell struct {
	Type  CellType    ///< The type of the cell (Empty, Fish, or Shark).
	Fish  *FishState  ///< Pointer to fish state (nil if Type != Fish).
	Shark *SharkState ///< Pointer to shark state (nil if Type != Shark).
}

/**
 * @brief Represents the Wa-Tor planet.
 *
 * Contains the grid of cells and the dimensions of the world.
 * The grid is a 2D slice where coordinates are accessed as Grid[y][x].
 */
type World struct {
	Size int      ///< The width and height of the square grid.
	Grid [][]Cell ///< 2D matrix representing the world state.
}

/**
 * @brief Represents a 2D coordinate point (X, Y).
 */
type P struct {
	X int
	Y int
}

// Initialization Functions

/**
 * @brief Creates a new Wa-Tor world grid.
 *
 * Initializes a 2D grid of Cells with the specified size. Memory is allocated
 * for the rows and columns.
 *
 * @param size The dimension of the grid (N x N).
 * @return A pointer to the initialized World struct.
 */
func NewWorld(size int) *World {
	g := make([][]Cell, size)
	for i := range g {
		g[i] = make([]Cell, size)
	}
	return &World{Size: size, Grid: g}
}

/**
 * @brief Initializes the global random number generator.
 *
 * Runs automatically when the package is loaded. Sets the seed to current time
 * to ensure different results on each run (unless overridden in main).
 */
func init() { rand.Seed(time.Now().UnixNano()) }

/**
 * @brief Randomly populates the world with Fish and Sharks.
 *
 * Uses a Fisher-Yates shuffle on a list of all possible coordinates to ensure
 * agents are placed randomly without collision (two agents on the same spot).
 *
 * @param w Pointer to the World to populate.
 * @param numFish Number of fish to place.
 * @param numShark Number of sharks to place.
 */
func SeedRandom(w *World, numFish, numShark int) {
	total := w.Size * w.Size
	idx := make([]int, total)
	// Create a list of all possible indices 0 to total-1
	for i := 0; i < total; i++ {
		idx[i] = i
	}
	// Shuffle the indices
	rand.Shuffle(total, func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })

	pos := 0
	// Helper closure to place a specific type of agent
	place := func(t CellType) {
		if pos >= total {
			return
		}
		i := idx[pos]
		pos++
		x, y := i%w.Size, i/w.Size
		c := &w.Grid[y][x]
		c.Type = t
		if t == Fish {
			c.Fish = &FishState{}
		}
		if t == Shark {
			c.Shark = &SharkState{Energy: SharkEnergyInit}
		}
	}
	// Place agents
	for i := 0; i < numFish; i++ {
		place(Fish)
	}
	for i := 0; i < numShark; i++ {
		place(Shark)
	}
}

// Helper Functions (Geometry & Stats)

/**
 * @brief Wraps a coordinate around the grid boundaries (Toroidal geometry).
 *
 * Implements the "donut" shape logic. If an index goes off the left edge (-1),
 * it wraps to the right edge (n-1). If it goes off the right edge (n),
 * it wraps to 0.
 *
 * @param i The coordinate to wrap.
 * @param n The size of the dimension.
 * @return The wrapped coordinate index.
 */
func wrap(i, n int) int {
	if i < 0 {
		return n - 1
	}
	if i >= n {
		return 0
	}
	return i
}

/**
 * @brief Calculates the 4 von Neumann neighbors (North, East, South, West).
 *
 * Handles boundary wrapping automatically using the wrap() function.
 *
 * @param x The X coordinate of the center cell.
 * @param y The Y coordinate of the center cell.
 * @param n The size of the grid.
 * @return An array of 4 Point structs representing the neighbors.
 */
func neigh4(x, y, n int) [4]P {
	return [4]P{
		{x, wrap(y-1, n)}, // North
		{wrap(x+1, n), y}, // East
		{x, wrap(y+1, n)}, // South
		{wrap(x-1, n), y}, // West
	}
}

/**
 * @brief Counts the total number of Fish and Sharks in the world.
 *
 * Iterates through the entire grid to perform a census.
 *
 * @param w Pointer to the World to count.
 * @return fish The total number of fish.
 * @return sharks The total number of sharks.
 */
func Count(w *World) (fish, sharks int) {
	for y := 0; y < w.Size; y++ {
		for x := 0; x < w.Size; x++ {
			switch w.Grid[y][x].Type {
			case Fish:
				fish++
			case Shark:
				sharks++
			}
		}
	}
	return
}

/**
 * @brief Prints a text representation of the world to the console.
 *
 * Useful for debugging small grids.
 *
 * @param w Pointer to the World to print.
 * @param max The maximum grid size to print (to avoid flooding the console).
 */
func PrintWorld(w *World, max int) {
	n := w.Size
	if n > max {
		n = max
	}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			switch w.Grid[y][x].Type {
			case Fish:
				print("F")
			case Shark:
				print("S")
			default:
				print(".")
			}
		}
		println()
	}
}
