package main

import (
	"math/rand"
	"sync"
)

/**
 * @file step_par.go
 * @brief Implementation of the parallel Wa-Tor simulation step.
 *
 * This file contains the logic for updating the simulation world using multiple
 * threads (goroutines). It uses spatial decomposition to divide the grid into
 * horizontal strips, assigning each strip to a worker.
 */

/**
 * @brief Splits the grid rows into segments for parallel processing.
 *
 * Divides the total number of rows (nRows) as evenly as possible among
 * the specified number of workers.
 *
 * @param nRows Total number of rows in the grid.
 * @param workers Number of worker threads available.
 * @return A slice of [start, end) row indices for each worker.
 */
func splitRows(nRows, workers int) [][2]int {
	if workers < 1 {
		workers = 1
	}
	if workers > nRows {
		workers = nRows
	}
	base := nRows / workers
	rem := nRows % workers

	segs := make([][2]int, 0, workers)
	y := 0
	for i := 0; i < workers; i++ {
		h := base
		if rem > 0 {
			h++
			rem--
		}
		y0 := y
		y += h
		segs = append(segs, [2]int{y0, y}) // [y0, y1)
	}
	return segs
}

/**
 * @brief Updates the world state in parallel using spatial decomposition.
 *
 * Divides the grid into horizontal strips and assigns each to a worker thread.
 * It uses a double-buffering approach (curr -> next) and row-level locks
 * to prevent race conditions when agents move between segments.
 *
 * @param curr Pointer to the current World state.
 * @param workers Number of goroutines to use.
 * @param stepSeed Random seed for this step to ensure determinism across runs.
 * @return Pointer to the new World state.
 */
func StepPar(curr *World, workers int, stepSeed int64) *World {
	n := curr.Size
	next := NewWorld(n)

	var wg sync.WaitGroup
	// Row-level locks to manage concurrent writes to the 'next' grid
	rowLocks := make([]sync.Mutex, n)

	segs := splitRows(n, workers)
	for i, seg := range segs {
		y0, y1 := seg[0], seg[1]
		// Create an independent random number generator for each segment to ensure determinism
		rng := rand.New(rand.NewSource(stepSeed + int64(i)))

		wg.Add(1)
		go func(y0, y1 int, rr *rand.Rand) {
			defer wg.Done()

			// Shuffle the processing order of cells in this segment to minimize directional bias
			order := make([]P, 0, (y1-y0)*n)
			for y := y0; y < y1; y++ {
				for x := 0; x < n; x++ {
					order = append(order, P{X: x, Y: y})
				}
			}
			rr.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

			// Helper function: Thread-safe placement into the 'next' grid
			place := func(x, y int, c Cell) bool {
				rowLocks[y].Lock()
				ok := next.Grid[y][x].Type == Empty
				if ok {
					next.Grid[y][x] = c
				}
				rowLocks[y].Unlock()
				return ok
			}

			// Helper function: Thread-safe check if a cell in 'next' is empty
			emptyNext := func(x, y int) bool {
				rowLocks[y].Lock()
				empty := next.Grid[y][x].Type == Empty
				rowLocks[y].Unlock()
				return empty
			}

			for _, p := range order {
				x, y := p.X, p.Y
				cell := curr.Grid[y][x]

				switch cell.Type {
				case Fish:
					f := *cell.Fish
					f.BreedTimer++
					ns := neigh4(x, y, n)

					// Find candidates: Empty in 'curr' AND not yet occupied in 'next'
					cands := make([]P, 0, 4)
					for _, q := range ns {
						if curr.Grid[q.Y][q.X].Type == Empty && emptyNext(q.X, q.Y) {
							cands = append(cands, q)
						}
					}
					target := P{X: x, Y: y}
					if len(cands) > 0 {
						target = cands[rr.Intn(len(cands))]
					}

					if f.BreedTimer >= FishBreedSteps && (target.X != x || target.Y != y) {
						// Breed: Move child to target, leave parent at original position (reset timer)
						_ = place(target.X, target.Y, Cell{Type: Fish, Fish: &FishState{BreedTimer: 0}})
						_ = place(x, y, Cell{Type: Fish, Fish: &FishState{BreedTimer: 0}})
					} else {
						// Attempt to move. If preempted (race condition), stay at original position.
						if !place(target.X, target.Y, Cell{Type: Fish, Fish: &FishState{BreedTimer: f.BreedTimer}}) {
							_ = place(x, y, Cell{Type: Fish, Fish: &FishState{BreedTimer: f.BreedTimer}})
						}
					}

				case Shark:
					s := *cell.Shark
					s.BreedTimer++
					s.Energy -= SharkEnergyLoss
					if s.Energy <= 0 {
						continue // Shark dies (starvation)
					}

					ns := neigh4(x, y, n)
					fishC := make([]P, 0, 4)
					for _, q := range ns {
						// Look for fish in 'curr' that haven't been claimed in 'next'
						if curr.Grid[q.Y][q.X].Type == Fish && emptyNext(q.X, q.Y) {
							fishC = append(fishC, q)
						}
					}
					moved := false
					tx, ty := x, y

					if len(fishC) > 0 {
						// Eat fish: Move to fish location, gain energy
						t := fishC[rr.Intn(len(fishC))]
						tx, ty = t.X, t.Y
						s.Energy += SharkEnergyGain
						if place(tx, ty, Cell{Type: Shark, Shark: &SharkState{BreedTimer: s.BreedTimer, Energy: s.Energy}}) {
							moved = true
						}
					} else {
						// No fish found: Try to move to an empty adjacent square
						emptyC := make([]P, 0, 4)
						for _, q := range ns {
							if curr.Grid[q.Y][q.X].Type == Empty && emptyNext(q.X, q.Y) {
								emptyC = append(emptyC, q)
							}
						}
						if len(emptyC) > 0 {
							t := emptyC[rr.Intn(len(emptyC))]
							tx, ty = t.X, t.Y
						}
						if !place(tx, ty, Cell{Type: Shark, Shark: &SharkState{BreedTimer: s.BreedTimer, Energy: s.Energy}}) {
							// If move failed (blocked), stay put
							_ = place(x, y, Cell{Type: Shark, Shark: &SharkState{BreedTimer: s.BreedTimer, Energy: s.Energy}})
						}
					}

					// Breed: Only reproduce if the shark successfully moved
					if s.BreedTimer >= SharkBreedSteps && moved {
						_ = place(x, y, Cell{Type: Shark, Shark: &SharkState{BreedTimer: 0, Energy: SharkEnergyInit}})
					}
				}
			}
		}(y0, y1, rng)
	}

	wg.Wait()
	return next
}
