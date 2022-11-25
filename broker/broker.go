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
var worldChan []chan World
var workers []Worker
var nextId = 0
var topicmx sync.RWMutex
var workerId int

// var theWorld World

type World struct {
	world       [][]uint8
	StartY      int
	EndY        int
	StartX      int
	EndX        int
	imageHeight int
	imageWidth  int
	turns       int
}

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

// Store the world
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
func subscribe_loop(worldChanS chan World, worker *rpc.Client) {
	for {
		fmt.Println("Loooping")
		worldS := <-worldChanS
		response := new(stubs.Response)
		workerReq := stubs.WorkerRequest{StartY: worldS.StartY, EndY: worldS.EndY, StartX: worldS.StartX, EndX: worldS.EndX, World: worldS.world, Turns: worldS.turns, Params: p}
		err := worker.Call(stubs.ProcessTurnsHandler, workerReq, response)
		if err != nil {
			fmt.Println("Error")
			fmt.Println(err)
			fmt.Println("Closing subscriber thread.")
			//worldChanS <- worldS
		}
		fmt.Println("Worker done", response.World, response.TurnsDone)
	}
	/*for {
		err := w.worker.Call(stubs.PauseHandler, workerReq, &response)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
	}*/
}

// Initialise connecting worker, and if no error occurs, invoke register_loop.
func subscribe(workerIdS int, workerAddress string) (err error) {
	fmt.Println("Subscription request")
	topicmx.RLock()
	worldChanS := worldChan[workerIdS]
	topicmx.RUnlock()
	client, err := rpc.Dial("tcp", workerAddress)
	if err == nil {
		fmt.Println("Looooop")
		go subscribe_loop(worldChanS, client)
		workerId++
	} else {
		fmt.Println("Error subscribing ", workerAddress)
		fmt.Println(err)
		return err
	}
	/*if p.Threads == 1 && err == nil {
		go subscribe_loop(0, p.ImageHeight, 0, p.ImageWidth, world, worldChan[0], req.Turns, worker)

	} else if err == nil {

		if len(workers) != p.Threads-1 {
			unit := int(p.ImageHeight / p.Threads)
			for i := 0; i < p.Threads; i++ {
				worldChan[i] = make(chan [][]uint8)
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
	}*/

	//workers = append(workers, worker)
	//nextId++

	return
}

func register_loop() {

}

// Make an connection with the Distributor
// And initialise the params.
func registerDistributor(req stubs.Request, res *stubs.StatusReport) (err error) {
	topicmx.RLock()
	defer topicmx.RUnlock()
	world = req.World
	//completedTurns = req.Turns
	completedTurns = 0
	p.Threads = req.Threads
	p.ImageHeight = req.ImageHeight
	p.ImageWidth = req.ImageWidth
	workerId = 0
	// DONE: Make a channel for the world
	if p.Threads == 1 && err == nil {
		//go subscribe_loop(0, p.ImageHeight, 0, p.ImageWidth, world, worldChan[0], req.Turns, worker)
		worldChan[0] <- World{world: world, StartY: 0, EndY: p.ImageHeight, StartX: 0, EndX: p.ImageWidth, imageHeight: p.ImageHeight, imageWidth: p.ImageWidth, turns: req.Turns}
		fmt.Println("Created channel #", 0)
	} else if err == nil {
		if len(workers) != p.Threads-1 {
			unit := int(p.ImageHeight / p.Threads)
			for i := 0; i < p.Threads; i++ {
				//worldChan[i] = make(chan [][]uint8)
				if i == p.Threads-1 {
					worldChan[i] <- World{world: world, StartY: i * unit, EndY: p.ImageHeight, StartX: 0, EndX: p.ImageWidth, imageHeight: p.ImageHeight, imageWidth: p.ImageWidth, turns: req.Turns}
					fmt.Println("Assigned world slice to the worldChan #", i)
				} else {
					worldChan[i] <- World{world: world, StartY: i * unit, EndY: (i + 1) * unit, StartX: 0, EndX: p.ImageWidth, imageHeight: p.ImageHeight, imageWidth: p.ImageWidth, turns: req.Turns}
					fmt.Println("Assigned world slice to the worldChan #", i)
				}
			}
		} else {
			return
		}

	}
	return err
}

func makeChannel(threads int) {
	topicmx.Lock()
	defer topicmx.Unlock()
	worldChan = make([]chan World, threads)
	for i := range worldChan {
		worldChan[i] = make(chan World)
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
	//request := stubs.WorkerRequest{StartY: 0, EndY: p.ImageHeight, StartX: 0, EndX: p.ImageWidth, WorldChan: worldChan[nextId], Turns: completedTurns, Params: p}
	err = subscribe(workerId, req.WorkerAddress)
	if err != nil {
		fmt.Println(err)
	}
	return
}

// (Publish)
func (b *Broker) ConnectDistributor(req stubs.Request, res *stubs.StatusReport) (err error) {
	err = registerDistributor(req, res)
	return
}

func (b *Broker) Publish(req stubs.TickerRequest, res *stubs.Response) (err error) {
	res.World = world
	res.TurnsDone = completedTurns
	//err = publish(stubs.StateRequest{State: req.State})
	return err
}

func (b *Broker) Pause(req stubs.PauseRequest, res *stubs.StatusReport) (err error) {
	return
}

func main() {
	// Listens to the distributor
	//pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&Broker{})
	listener, _ := net.Listen("tcp", ":"+"8030")
	defer listener.Close()
	rpc.Accept(listener)
}
