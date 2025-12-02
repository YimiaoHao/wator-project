package main

import (
	"math/rand"
	"time"
)

// Types and Constants

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

type FishState struct {
	BreedTimer int ///< Chronons survived since last reproduction.
}

type SharkState struct {
	BreedTimer int ///< Chronons survived since last reproduction.
	Energy     int ///< Current energy level. Shark dies if this reaches 0.
}

type Cell struct {
	Type  CellType    ///< The type of the cell (Empty, Fish, or Shark).
	Fish  *FishState  ///< Pointer to fish state (nil if Type != Fish).
	Shark *SharkState ///< Pointer to shark state (nil if Type != Shark).
}


type World struct {
	Size int      ///< The width and height of the square grid.
	Grid [][]Cell ///< 2D matrix representing the world state.
}

type P struct {
	X int
	Y int
}

// Initialization Functions

func NewWorld(size int) *World {
	g := make([][]Cell, size)
	for i := range g {
		g[i] = make([]Cell, size)
	}
	return &World{Size: size, Grid: g}
}


func init() { rand.Seed(time.Now().UnixNano()) }

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

func wrap(i, n int) int {
	if i < 0 {
		return n - 1
	}
	if i >= n {
		return 0
	}
	return i
}

func neigh4(x, y, n int) [4]P {
	return [4]P{
		{x, wrap(y-1, n)}, // North
		{wrap(x+1, n), y}, // East
		{x, wrap(y+1, n)}, // South
		{wrap(x-1, n), y}, // West
	}
}


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
