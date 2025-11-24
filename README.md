# Wa-Tor Simulation Project

*A high-performance implementation of A. K. Dewdney’s predator–prey simulation on a toroidal world, with sequential and parallel algorithms plus a real-time GUI.*

---

## Features

- **Dual execution modes**
  - `seq` — single-threaded baseline for correctness.
  - `par` — multi-threaded (spatial decomposition + double buffering).
- **Real-time visualization** (Ebiten v2)  
  Blue = Ocean (Empty) · Yellow = Fish (Prey) · Red = Sharks (Predators)
- **Classic mechanics:** movement, breeding, starvation, predation.
- **Flexible configuration:** standard flags (e.g., `-gui`, `-mode`) **and** the required **“7 positional arguments”** interface.

---

## Prerequisites

- **OS:** Linux (Ubuntu 22.04) or **WSL2** (recommended).
- **Go:** 1.22+
- **Graphics dependencies** (for GUI on Linux/WSL2 X11):

```bash
sudo apt update
sudo apt install -y libasound2-dev libgl1-mesa-dev \
  libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev
````

> Using **WSL**/**WSLg** (Win11) is easiest. If you use an external X server (e.g., VcXsrv/X410), make sure it’s running before launching the GUI.
> If a GUI is not available, add `-nogui`.

---

## Build & Run

Run commands from the project root.

### Option A: Run without building (recommend)

```bash
# GUI + sequential (default)
go run . -gui

# GUI + parallel (use all CPU cores)
go run . -gui -mode=par -workers=$(nproc)

# Headless benchmark (no GUI)
go run . -mode=par -steps=200 -nogui -statsEvery=10
```

### Option B: Build a binary

```bash
go build -o wator .
```

Then:

* **Linux/macOS**

  ```bash
  ./wator -gui
  ./wator -gui -mode=par -workers=$(nproc)
  ```
* **Windows PowerShell (or WSL calling Windows binary)**

  ```powershell
  .\wator.exe -gui -mode=par -workers $env:NUMBER_OF_PROCESSORS
  ```

---

## Positional Arguments Interface

Required order:

```
[NumShark] [NumFish] [FishBreed] [SharkBreed] [Starve] [GridSize] [Threads]
```

Example (300 sharks, 1500 fish, 80×80 grid, 16 threads):

```bash
./wator 300 1500 3 8 5 80 16 -gui
```

---

## Project Structure

| File             | Description                                                 |
| ---------------- | ----------------------------------------------------------- |
| `main.go`        | Entry point; flag or positional-arg parsing; run-loop control. |
| `world.go`       | Core data structures (`World`, `Cell`) and shared rules.    |
| `step_seq.go`    | **Sequential** simulation logic (baseline).                 |
| `step_par.go`    | **Parallel** simulation logic (spatial decomposition).      |
| `view_ebiten.go` | GUI rendering with Ebiten v2.                               |
| `go.mod`         | Go module definition.                                       |

---

## Performance Results (example)

**Env:** Linux (WSL2) · Grid: 500×500 · Steps: 1000
*(You can include a plot as `results_graph.png` if desired.)*

### Speedup Table

code：

go run . -size 500 -steps 1000 -mode par -workers 1 -nogui

go run . -size 500 -steps 1000 -mode par -workers 2 -nogui

go run . -size 500 -steps 1000 -mode par -workers 4 -nogui

go run . -size 500 -steps 1000 -mode par -workers 8 -nogui

| Threads     | Execution Time (s) | Speedup   |
| ----------- | ------------------ | --------- |
| **1 (Seq)** | 32.83              | **1.00×** |
| **2**       | 15.95              | 2.06×     |
| **4**       | 9.56               | 3.43×     |
| **8**       | 6.31               | 5.20×     |

**Notes:** Near-linear scaling (1→4) shows effective load splitting (row striping). At higher thread counts, goroutine scheduling and synchronization overheads reduce efficiency.

---

## License

This project is licensed under the MIT License.
