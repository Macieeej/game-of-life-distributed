package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

var listener net.Listener
var pause bool
var quit bool
var kill bool = false
var waitToUnpause chan bool

var brokerAddr string
var client *rpc.Client

// updateBroker
var turnChan chan int
var worldChan chan [][]uint8

// updateWorker
var workerTurnChan chan int
var workerWorldChan chan [][]uint8

var turnInternal chan int
var worldInternal chan [][]uint8

var workerId int
var nextAddr string
var globalWorld [][]uint8
var completedTurns int
var incr int
var resume chan bool
var done chan bool

func getOutboundIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr).IP.String()
	return localAddr
}

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

func CalculateNextState(height, width, startY, endY int, world [][]byte) ([][]byte, []util.Cell) {

	newWorld := make([][]byte, endY-startY)
	flipCell := make([]util.Cell, height, width)
	for i := 0; i < endY-startY; i++ {
		newWorld[i] = make([]byte, width)
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

type GolOperations struct{}

func UpdateBroker2(tchan chan int, wchan chan [][]uint8, client *rpc.Client) {
	for {
		t := <-tchan
		ws := <-wchan
		towork := stubs.UpdateRequest{Turns: t, World: ws, WorkerId: workerId}
		status := new(stubs.StatusReport)
		err := client.Call(stubs.UpdateBroker, towork, status)
		if err != nil {
			fmt.Println("RPC client returned error:")
			fmt.Println(err)
			fmt.Println("Dropping division.")
		}
	}
}

func (s *GolOperations) Action(req stubs.StateRequest, res *stubs.StatusReport) (err error) {
	switch req.State {
	case stubs.Pause:
		pause = true
	case stubs.UnPause:
		pause = false
	}
	return nil
}

func (s *GolOperations) ActionWithReport(req stubs.StateRequest, res *stubs.StatusReport) (err error) {
	switch req.State {
	case stubs.Quit:
		quit = true
		fmt.Println("pause")
	case stubs.Save:
	case stubs.Kill:
		kill = true
		defer os.Exit(0)
	}
	fmt.Println("deafault")
	return nil
}

func (s *GolOperations) UpdateWorker(req stubs.UpdateRequest, res *stubs.StatusReport) (err error) {
	fmt.Println("UpdateWorld called")
	fmt.Println("From:", req.Turns)
	globalWorld = req.World
	completedTurns = req.Turns
	res.Status = 7
	incr++
	return
}

func worker(p stubs.Params, startY, endY, startX, endX int, world [][]uint8, out chan<- [][]uint8, turn int) {
	newPart := make([][]uint8, endY-startY)
	for i := range newPart {
		newPart[i] = make([]uint8, endX)
	}
	newPart, _ = CalculateNextState(p.ImageHeight, p.ImageWidth, startY, endY, world)
	out <- newPart
}

func (s *GolOperations) Process(req stubs.WorkerRequest, res *stubs.Response) (err error) {
	fmt.Println("Processing")
	workerId = req.WorkerId
	var newWorldSlice [][]uint8
	globalWorld = req.World
	pause = false
	quit = false
	turn := 0
	incr = 0
	// HARDCODE NO OF THREADS ON THE --SERVER SIDE'S WORKER--
	distThreads := 2
	for t := 0; t < req.Turns; t++ {
		if incr == t && !pause && !quit {
			if pause {
				fmt.Println("Paused")
			}
			if !kill {
				if distThreads == 1 {
					newWorldSlice, _ = CalculateNextState(req.Params.ImageHeight, req.Params.ImageWidth, req.StartY, req.EndY, globalWorld)
					turn++
					turnChan <- turn
					worldChan <- newWorldSlice
				} else {
					var worldFragment [][]uint8
					channels := make([]chan [][]uint8, distThreads)
					unit := int((req.EndY - req.StartY) / distThreads)
					for i := 0; i < distThreads; i++ {
						channels[i] = make(chan [][]uint8)
						if i == distThreads-1 {
							// Handling with problems if threads division goes with remainders
							go worker(req.Params, req.StartY+(i*unit), req.EndY, 0, req.Params.ImageWidth, globalWorld, channels[i], turn)
						} else {
							go worker(req.Params, req.StartY+(i*unit), req.StartY+((i+1)*unit), 0, req.Params.ImageWidth, globalWorld, channels[i], turn)
						}
					}
					for i := 0; i < distThreads; i++ {
						worldPart := <-channels[i]
						worldFragment = append(worldFragment, worldPart...)
					}
					turn++
					turnChan <- turn
					worldChan <- worldFragment
				}

				fmt.Println(turn)
			} else {
				if kill {
					break
				} else {
					continue
				}
			}
		} else {
			t--
		}
	}
	res.World = newWorldSlice
	res.TurnsDone = turn
	return
}

func handleConnection() {
	client, err := rpc.Dial("tcp", brokerAddr)
	if err != nil {
		fmt.Println(err)
	}
	go UpdateBroker2(turnChan, worldChan, client)
}

func (*GolOperations) ListenToDistributor(req stubs.AddressRequest, res *stubs.StatusReport) (err error) {
	brokerAddr = req.Address
	handleConnection()
	return
}

func main() {
	pAddr := flag.String("port", "8050", "Port to listen on")
	// brokerAddr := flag.String("broker", "127.0.0.1:8030", "Address of broker instance")
	flag.Parse()
	fmt.Println("@")
	listenerr, err := net.Listen("tcp", ":"+*pAddr)
	if err != nil {
		fmt.Println(err)
	}
	rpc.Register(&GolOperations{})
	rpc.Accept(listenerr)

	//client, err := rpc.Dial("tcp", "127.0.0.1:8030")
	//fmt.Println(*pAddr)
	fmt.Println(getOutboundIP() + ":" + *pAddr)

	//fmt.Println(getOutboundIP() + ":" + "8050")
	// client, err := rpc.Dial("tcp", brokerAddr)
	//listenerr, err := net.Listen("tcp", ":"+"8050")

	// subscribe := stubs.SubscribeRequest{
	// 	WorkerAddress: getOutboundIP() + ":" + *pAddr,
	// 	//WorkerAddress: getOutboundIP() + ":" + "8050",
	// }
	turnChan = make(chan int)
	turnInternal = make(chan int)
	worldChan = make(chan [][]uint8)
	worldInternal = make(chan [][]uint8)
	waitToUnpause = make(chan bool)

	//go receive()
	//go send()
	// client.Call(stubs.ConnectWorker, subscribe, new(stubs.StatusReport))

	//client.Call(stubs.ConnectWorker, subscribe, new(stubs.StatusReport))
	defer listenerr.Close()
	// go UpdateBroker2(turnChan, worldChan, client)
	//go UpdateWorker2(client)
}
