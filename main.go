// Package main is the entry point for the Wa-Tor ecological simulation.
//
// It handles command-line argument parsing, initializes the simulation world,
// and controls the main execution loop for both sequential and parallel modes.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

// parsePositionalArgsIntoFlags parses the 7 required positional arguments into configuration flags.
//
// It returns true if 7 positional arguments were provided and parsed successfully; false otherwise.
//
// The expected argument order is:
// [NumShark] [NumFish] [FishBreed] [SharkBreed] [Starve] [GridSize] [Threads]
func parsePositionalArgsIntoFlags(size, fish, sharks, steps, workers *int) bool {
	args := flag.Args()
	if len(args) != 7 {
		return false
	}

	toInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}

	*sharks = toInt(args[0])         // NumShark
	*fish = toInt(args[1])           // NumFish
	FishBreedSteps = toInt(args[2])  // FishBreed (Global var from world.go)
	SharkBreedSteps = toInt(args[3]) // SharkBreed (Global var from world.go)
	SharkEnergyInit = toInt(args[4]) // Starve -> SharkEnergyInit (Global var from world.go)
	*size = toInt(args[5])           // GridSize
	*workers = toInt(args[6])        // Threads
	return true
}

// main is the application entry point.
//
// It performs the following steps:
//  1. Defines and parses command-line flags.
//  2. Overrides flags with positional arguments if provided.
//  3. Validates input parameters to prevent invalid simulation states.
//  4. Initializes the World grid and seeds random agents.
//  5. Launches the simulation loop (GUI mode using Ebiten or CLI mode).
func main() {
	// Define Flag parameters
	size := flag.Int("size", 150, "grid size (N x N)")
	fish := flag.Int("fish", 800, "initial fish count")
	sharks := flag.Int("sharks", 150, "initial shark count")
	steps := flag.Int("steps", 100, "number of steps")
	mode := flag.String("mode", "seq", "seq or par")
	workers := flag.Int("workers", runtime.NumCPU(), "goroutines for par mode")

	gui := flag.Bool("gui", false, "show GUI window")
	seed := flag.Int64("seed", time.Now().UnixNano(), "random seed")
	statsEvery := flag.Int("statsEvery", 0, "print stats every N steps (0 = never)")
	quiet := flag.Bool("quiet", false, "suppress console prints")
	nogui := flag.Bool("nogui", false, "force disable GUI even if -gui is set")

	flag.Parse()

	// If positional arguments are present, override flag parameters
	_ = parsePositionalArgsIntoFlags(size, fish, sharks, steps, workers)

	// Set random seed
	rand.Seed(*seed)

	// Basic parameter validation
	if *workers < 1 {
		log.Fatalf("workers must be >= 1, got %d", *workers)
	}
	if *size <= 0 {
		log.Fatalf("size must be > 0, got %d", *size)
	}
	if *fish < 0 || *sharks < 0 {
		log.Fatalf("fish/sharks must be >= 0, got fish=%d sharks=%d", *fish, *sharks)
	}
	if *fish+*sharks > (*size)*(*size) {
		log.Fatalf("too many agents: fish + sharks (%d) > size*size (%d)", *fish+*sharks, (*size)*(*size))
	}
	if FishBreedSteps <= 0 || SharkBreedSteps <= 0 || SharkEnergyInit <= 0 {
		log.Fatalf("FB/SB/Starve must be > 0, got FB=%d SB=%d Starve=%d",
			FishBreedSteps, SharkBreedSteps, SharkEnergyInit)
	}

	// GUI toggle logic
	if *nogui {
		*gui = false
	}
	if *gui {
		*statsEvery = 0 // GUI mode: suppress step printing
		*quiet = true
	}

	// Concurrency setting
	runtime.GOMAXPROCS(*workers)

	// Initialize world
	w := NewWorld(*size)
	SeedRandom(w, *fish, *sharks)

	if !*quiet {
		fmt.Printf("CFG sharks=%d fish=%d FB=%d SB=%d Starve=%d size=%d threads=%d mode=%s gui=%t seed=%d\n",
			*sharks, *fish, FishBreedSteps, SharkBreedSteps, SharkEnergyInit, *size, *workers, *mode, *gui, *seed)
	}

	// GUI mode branch
	if *gui {
		runMode := *mode
		if runMode != "par" {
			runMode = "seq"
		}
		if err := runGUI(w, runMode, *workers); err != nil {
			panic(err)
		}
		return
	}

	// Terminal mode: run loop
	start := time.Now()
	for i := 0; i < *steps; i++ {
		if *mode == "par" {
			stepSeed := *seed + int64(i)
			w = StepPar(w, *workers, stepSeed)
		} else {
			w = StepSeq(w)
		}

		if !*quiet && *statsEvery > 0 && (i%*statsEvery == 0) {
			f, s := Count(w)
			fmt.Printf("step=%03d  fish=%5d  sharks=%5d\n", i, f, s)
		}
	}
	elapsed := time.Since(start)

	if !*quiet {
		fmt.Printf("mode=%s workers=%d size=%d steps=%d time=%v\n",
			*mode, *workers, *size, *steps, elapsed)
	}
}
