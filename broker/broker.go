package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"sync"

	"uk.ac.bris.cs/gameoflife/stubs"
)

// 1 : Pass the report from the distributor. (Ticker request (broker -> dis))
// 2 : Connect with the distributor (Register (broker -> dis))
// 3 : Pass the keypress arguments from the distributor to the worker. (Keypress (dis -> broker) (broker -> worker))

// Channels that are used to communicate with broker and worker
var channels []chan [][]uint8
var workers []Worker
var nextId = 0
var topicmx sync.RWMutex

// var theWorld World

// type World struct {
// 	world       [][]uint8
// 	imageHeight int
// 	imageWidth  int
// 	turns       int
// 	threads     int
// }

type Worker struct {
	id           int
	stateSwitch  int
	worker       *rpc.Client
	address      *string
	worldChannel chan [][]uint8
}

type Params struct {
	threads     int
	imageHeight int
	imageWidth  int
}

var p stubs.Params
var world [][]uint8
var completedTurns int

// Handing out the world again to the worker.
func handleWorkers(unit int) {
	mergeWorld()
	for i := 0; i < p.Threads; i++ {
		workers[i].worldChannel <- world
	}
}

// Accepting every information about world progress and merging into a world.
func mergeWorld() {
	var newWorld [][]uint8
	for _, w := range workers {
		newWorld = append(newWorld, <-w.worldChannel...)
	}

	for i := range newWorld {
		copy(world[i], newWorld[i])
	}
}

// Connect the worker in a loop
func subscribe_loop(startY, endY, startX, endX int, world [][]uint8, out chan<- [][]uint8, turns int, w Worker) {
	response := new(stubs.Response)
	workerReq := stubs.WorkerRequest{StartY: startY, EndY: endY, StartX: startX, EndX: endX, WorldChan: w.worldChannel, Turns: turns, Params: p}
	err := w.worker.Call(stubs.ProcessTurnsHandler, workerReq, &response)
	if err != nil {
		fmt.Println(err)
	}

	for {
		err := w.worker.Call(stubs.PauseHandler, workerReq, &response)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
	}
}

// Initialise connecting worker, and if no error occurs, invoke register_loop.
func subscribe(req stubs.WorkerRequest, workerAddress string) (err error) {

	client, err := rpc.Dial("tcp", workerAddress)
	worker := Worker{
		id:           nextId,
		worker:       client,
		address:      &workerAddress,
		worldChannel: channels[nextId],
	}

	if p.Threads == 1 && err == nil {
		go subscribe_loop(0, p.ImageHeight, 0, p.ImageWidth, world, channels[0], req.Turns, worker)

	} else if err == nil {

		if len(workers) != p.Threads-1 {
			unit := int(p.ImageHeight / p.Threads)
			for i := 0; i < p.Threads; i++ {
				channels[i] = make(chan [][]uint8)
				if i == p.Threads-1 {
					go subscribe_loop(i*unit, p.ImageHeight, 0, p.ImageWidth, world, workers[i].worldChannel, completedTurns, workers[i])
				} else {
					go subscribe_loop(i*unit, (i+1)*unit, 0, p.ImageWidth, world, workers[i].worldChannel, completedTurns, workers[i])
				}
			}
			go handleWorkers(unit)
		} else {
			return
		}

	} else {
		fmt.Println(err)
		return err
	}

	workers = append(workers, worker)
	nextId++

	return
}

func register_loop() {

}

// Make an connection with the Distributor
// And initialise the params.
func registerDistributor(req stubs.Request, res *stubs.StatusReport) (err error) {
	world = req.World
	completedTurns = req.Turns
	p.Threads = req.Threads
	p.ImageHeight = req.ImageHeight
	p.ImageWidth = req.ImageWidth
	//channels = make([]chan [][]uint8, p.Threads)
	go register_loop()
	return err
}

func makeChannel(threads int) {
	channels = make([]chan [][]uint8, threads)
	for i := range channels {
		channels[i] = make(chan [][]uint8)
		fmt.Println("Created channel #", i)
	}
}

func publish(req stubs.StateRequest) (err error) {
	/*topicmx.RLock()
	defer topicmx.RUnlock()
	if ch, ok := topics[topic]; ok {
		ch <- pair
	} else {
		return errors.New("No such topic.")
	}*/
	return
}

type Broker struct{}

// func (b *Broker) ReportStatus(req stubs.StateRequest, req *stubs.Response) (err error) {
// 	return err
// }

func (b *Broker) MakeChannel(req stubs.ChannelRequest, res *stubs.StatusReport) (err error) {
	makeChannel(req.Threads)
	return
}

// Calls and connects to the worker (Subscribe)
func (b *Broker) ConnectWorker(req stubs.SubscribeRequest, res *stubs.StatusReport) (err error) {
	request := stubs.WorkerRequest{StartY: 0, EndY: p.ImageHeight, StartX: 0, EndX: p.ImageWidth, WorldChan: channels[nextId], Turns: completedTurns, Params: p}
	err = subscribe(request, req.WorkerAddress)
	return
}

// (Publish)
func (b *Broker) ConnectDistributor(req stubs.Request, res *stubs.StatusReport) (err error) {
	err = registerDistributor(req, res)
	return
}

func (b *Broker) Publish(req stubs.StateRequest, res *stubs.Response) (err error) {
	res.World = world
	res.TurnsDone = completedTurns
	err = publish(stubs.StateRequest{State: req.State})
	return err
}

func (b *Broker) Pause(req stubs.PauseRequest, res *stubs.StatusReport) (err error) {
	return
}

func main() {
	// Listens to the distributor
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&Broker{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
