package main

import "math/rand"

func StepSeq(w *World) *World {
	n := w.Size
	cur := w.Grid

	// Double buffering for next state
	next := make([][]Cell, n)
	for i := 0; i < n; i++ {
		next[i] = make([]Cell, n)
	}

	//  Process Sharks First 
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			cell := cur[y][x]
			if cell.Type != Shark || cell.Shark == nil {
				continue
			}

			s := *cell.Shark
			// Decrease energy every step; if <= 0, the shark dies (is not written to next state)
			s.Energy -= SharkEnergyLoss
			if s.Energy <= 0 {
				continue
			}
			s.BreedTimer++

			// Look for fish in neighboring cells
			var fishNbrs []P
			for _, p := range neigh4(x, y, n) {
				if cur[p.Y][p.X].Type == Fish {
					fishNbrs = append(fishNbrs, p)
				}
			}

			moved := false

			if len(fishNbrs) > 0 {
				// Found fish: eat it and move to that position
				dst := fishNbrs[rand.Intn(len(fishNbrs))]

				// FIXED: Add energy gain instead of resetting to Init
				newEnergy := s.Energy + SharkEnergyGain

				mover := Cell{Type: Shark, Shark: &SharkState{
					BreedTimer: s.BreedTimer,
					Energy:     newEnergy,
				}}

				if s.BreedTimer >= SharkBreedSteps {
					// Breed: leave a new shark at the original position (Init energy), parent resets timer and moves
					if next[y][x].Type == Empty {
						next[y][x] = Cell{Type: Shark, Shark: &SharkState{BreedTimer: 0, Energy: SharkEnergyInit}}
					}
					mover.Shark.BreedTimer = 0
				}
				next[dst.Y][dst.X] = mover // Overwrite the fish (eat it)
				moved = true
			} else {
				// No fish: look for empty neighbors to move into
				var empties []P
				for _, p := range neigh4(x, y, n) {
					// Check if empty in current grid AND empty in next grid (to avoid collisions)
					if cur[p.Y][p.X].Type == Empty && next[p.Y][p.X].Type == Empty {
						empties = append(empties, p)
					}
				}
				if len(empties) > 0 {
					dst := empties[rand.Intn(len(empties))]
					mover := Cell{Type: Shark, Shark: &SharkState{
						BreedTimer: s.BreedTimer,
						Energy:     s.Energy,
					}}
					if s.BreedTimer >= SharkBreedSteps {
						// Breed
						if next[y][x].Type == Empty {
							next[y][x] = Cell{Type: Shark, Shark: &SharkState{BreedTimer: 0, Energy: SharkEnergyInit}}
						}
						mover.Shark.BreedTimer = 0
					}
					next[dst.Y][dst.X] = mover
					moved = true
				}
			}

			if !moved {
				// Stay in place
				if s.BreedTimer >= SharkBreedSteps && next[y][x].Type == Empty {
					next[y][x] = Cell{Type: Shark, Shark: &SharkState{BreedTimer: 0, Energy: SharkEnergyInit}}
					s.BreedTimer = 0
				}
				if next[y][x].Type == Empty {
					next[y][x] = Cell{Type: Shark, Shark: &SharkState{BreedTimer: s.BreedTimer, Energy: s.Energy}}
				}
			}
		}
	}

	//  Process Fish Next
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			cell := cur[y][x]
			if cell.Type != Fish || cell.Fish == nil {
				continue
			}

			// If this position in 'next' is already occupied by a shark (eaten), do not process this fish
			if next[y][x].Type == Shark {
				continue
			}

			f := *cell.Fish
			f.BreedTimer++

			var empties []P
			for _, p := range neigh4(x, y, n) {
				if cur[p.Y][p.X].Type == Empty && next[p.Y][p.X].Type == Empty {
					empties = append(empties, p)
				}
			}

			if len(empties) > 0 {
				dst := empties[rand.Intn(len(empties))]
				mover := Cell{Type: Fish, Fish: &FishState{BreedTimer: f.BreedTimer}}

				if f.BreedTimer >= FishBreedSteps {
					if next[y][x].Type == Empty {
						next[y][x] = Cell{Type: Fish, Fish: &FishState{BreedTimer: 0}}
					}
					mover.Fish.BreedTimer = 0
				}

				// If the destination is occupied by a shark in the 'next' grid (rare edge case), do not overwrite
				if next[dst.Y][dst.X].Type == Empty {
					next[dst.Y][dst.X] = mover
				}
			} else {
				// Stay in place
				if f.BreedTimer >= FishBreedSteps && next[y][x].Type == Empty {
					next[y][x] = Cell{Type: Fish, Fish: &FishState{BreedTimer: 0}}
					f.BreedTimer = 0
				}
				if next[y][x].Type == Empty {
					next[y][x] = Cell{Type: Fish, Fish: &FishState{BreedTimer: f.BreedTimer}}
				}
			}
		}
	}

	// Commit the new generation
	w.Grid = next
	return w
}
