package main

import (
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

var listener net.Listener
var job chan int
var pause bool
var waitToUnpause chan bool
var savedWorld [][]uint8
var completedTurns int

func mod(a, b int) int {
	return (a%b + b) % b
}

func calculateNeighbours(height, width int, world [][]byte, y int, x int) int {

	h := height
	w := width
	noOfNeighbours := 0

	neighbour := []byte{world[mod(y+1, h)][mod(x, w)], world[mod(y+1, h)][mod(x+1, w)], world[mod(y, h)][mod(x+1, w)],
		world[mod(y-1, h)][mod(x+1, w)], world[mod(y-1, h)][mod(x, w)], world[mod(y-1, h)][mod(x-1, w)],
		world[mod(y, h)][mod(x-1, w)], world[mod(y+1, h)][mod(x-1, w)]}

	for i := 0; i < 8; i++ {
		if neighbour[i] == 255 {
			noOfNeighbours++
		}
	}

	return noOfNeighbours
}

func calculateNextState(height, width, startY, endY int, world [][]byte) ([][]byte, []util.Cell) {

	newWorld := make([][]byte, endY-startY)
	flipCell := make([]util.Cell, height, width)
	for i := 0; i < endY-startY; i++ {
		newWorld[i] = make([]byte, len(world[0]))
		// copy(newWorld[i], world[startY+i])
	}

	for y := 0; y < endY-startY; y++ {
		for x := 0; x < width; x++ {
			noOfNeighbours := calculateNeighbours(height, width, world, startY+y, x)
			if world[startY+y][x] == 255 {
				if noOfNeighbours < 2 {
					newWorld[y][x] = 0
					flipCell = append(flipCell, util.Cell{X: x, Y: startY + y})
				} else if noOfNeighbours == 2 || noOfNeighbours == 3 {
					newWorld[y][x] = 255
				} else if noOfNeighbours > 3 {
					newWorld[y][x] = 0
					flipCell = append(flipCell, util.Cell{X: x, Y: startY + y})
				}
			} else if world[startY+y][x] == 0 && noOfNeighbours == 3 {
				newWorld[y][x] = 255
				flipCell = append(flipCell, util.Cell{X: x, Y: startY + y})
			}
		}
	}

	return newWorld, flipCell
}

// Invokes function that are required more than just sending data back
func job_loop(job chan int) {
	for {
		action := <-job
		switch action {
		case stubs.Save:
		case stubs.UnPause:
		case stubs.Ticker:
		case stubs.Quit:
			pauseServer()
		case stubs.Pause:
			pauseServer()
		case stubs.Kill:
			killServer()
		}
	}
}

func pauseServer() {
	// pause = req.Pause
	if !pause {
		waitToUnpause <- true
	}
}

func killServer() {
	listener.Close()
	os.Exit(0)
}

type GolOperations struct{}

// Actions that doesn't require any response
func (s *GolOperations) ListenToPause(req stubs.StateRequest, res *stubs.StatusReport) (err error) {
	job <- req.State
	return
}

// Actions that does require response (ScreenShotResponse)
func (s *GolOperations) ListenForWork(req stubs.StateRequest, res *stubs.Response) (err error) {
	job <- req.State
	res.World = savedWorld
	res.TurnsDone = completedTurns
	return
}

func (s *GolOperations) Process(req stubs.Request, res *stubs.Response) (err error) {

	if req.Turns == 0 {
		res.World = req.World
		res.TurnsDone = 0
		return
	}
	pause = false
	for t := 0; t < req.Turns; t++ {
		if pause {
			<-waitToUnpause
		}
		if !pause {
			completedTurns = t
			savedWorld, _ = calculateNextState(req.ImageHeight, req.ImageWidth, 0, req.ImageHeight, req.World)
		} /*else {
			if quit {
				break
			} else {
				continue
			}
		}*/
	}

	res.World = savedWorld
	res.TurnsDone = completedTurns
	return
}

// kill := make(chan bool)

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GolOperations{})
	listener, _ = net.Listen("tcp", ":"+*pAddr)
	job = make(chan int)
	go job_loop(job)
	defer listener.Close()
	rpc.Accept(listener)

}
