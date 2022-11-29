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
var unit int

// var theWorld World

type World struct {
	world [][]uint8
	turns int
}

type WorkerParams struct {
	StartY int
	EndY   int
	StartX int
	EndX   int
}

type Worker struct {
	id           int
	stateSwitch  int
	worker       *rpc.Client
	address      *string
	params       WorkerParams
	worldChannel chan World
}

// Store the world
var p stubs.Params
var world [][]uint8
var completedTurns int

// Connect the worker in a loop
func subscribe_loop(w Worker, startGame chan bool) {
	fmt.Println("Loooping")
	response := new(stubs.Response)
	workerReq := stubs.WorkerRequest{WorkerId: w.id, StartY: w.params.StartY, EndY: w.params.EndY, StartX: w.params.StartX, EndX: w.params.EndX, World: world, Turns: p.Turns, Params: p}
	<-startGame
	go func() {
		for {
			wt := <-w.worldChannel
			updateResponse := new(stubs.StatusReport)
			updateRequest := stubs.UpdateRequest{World: wt.world, Turns: wt.turns}
			err := w.worker.Call(stubs.UpdateWorker, updateRequest, updateResponse)
			if err != nil {
				fmt.Println("Error calling UpdateWorker")
				//fmt.Println(err)
				fmt.Println("Closing subscriber thread.")
				//Place the unfulfilled job back on the topic channel.
				w.worldChannel <- wt
				break
			}
			fmt.Println("Updated worker:", w.id, "turns:", completedTurns)
		}
	}()
	err := w.worker.Call(stubs.ProcessTurnsHandler, workerReq, response)
	//go handleWorkers()
	if err != nil {
		fmt.Println("Error calling ProcessTurnsHandler")
		//fmt.Println(err)
		fmt.Println("Closing subscriber thread.")
		//worldChanS <- worldS
	}
}

// Initialise connecting worker, and if no error occurs, invoke register_loop.
func subscribe(workerAddress string) (err error) {
	fmt.Println("Subscription request")
	client, err := rpc.Dial("tcp", workerAddress)
	var newWorker Worker
	if nextId != p.Threads-1 {
		newWorker = Worker{
			id:           nextId,
			stateSwitch:  -1,
			worker:       client,
			address:      &workerAddress,
			worldChannel: worldChan[nextId],
			params: WorkerParams{
				StartX: 0,
				StartY: nextId * unit,
				EndX:   p.ImageWidth,
				EndY:   nextId * (unit + 1),
			},
		}
	} else {
		newWorker = Worker{
			id:           nextId,
			stateSwitch:  -1,
			worker:       client,
			address:      &workerAddress,
			worldChannel: worldChan[nextId],
			params: WorkerParams{
				StartX: 0,
				StartY: nextId * unit,
				EndX:   p.ImageWidth,
				EndY:   p.ImageHeight,
			},
		}
	}
	workers = append(workers, newWorker)
	nextId++
	startGame := make(chan bool)
	go func() {
		for {
			if p.Threads == len(workers) {
				startGame <- true
			}
		}
	}()
	if err == nil {
		fmt.Println("Looooop")
		go subscribe_loop(newWorker, startGame)
	} else {
		fmt.Println("Error subscribing ", workerAddress)
		fmt.Println(err)
		return err
	}

	return
}

// Make an connection with the Distributor
// And initialise the params.
func registerDistributor(req stubs.Request, res *stubs.StatusReport) (err error) {
	topicmx.RLock()
	defer topicmx.RUnlock()
	world = req.World
	p.Turns = req.Turns
	p.Threads = req.Threads
	p.ImageHeight = req.ImageHeight
	p.ImageWidth = req.ImageWidth
	unit = int(p.ImageHeight / p.Threads)
	completedTurns = 0
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

// Accepting every information about world progress and merging into a world.
// func mergeWorld(mwworld [][]uint8) {

// 	var newWorld [][]uint8

// 	for _, w := range workers {
// 		fmt.Println("mergeworld #", w.id)
// 		prevWorld := mwworld
// 		newWorld = append(newWorld, prevWorld...)
// 	}

// 	for i := range world {
// 		copy(world[i], newWorld[i])
// 	}
// }

var incr int = 0

func merge(ubworldSlice [][]uint8, w Worker) {
	for i := range ubworldSlice {
		copy(world[w.params.StartY+i], ubworldSlice[i])
	}
	// return
}

func matchWorker(id int) Worker {
	for _, w := range workers {
		if w.id == id {
			return w
		}
	}
	panic("No such worker")
}

func updateBroker(ubturns int, ubworldSlice [][]uint8, workerId int) error {
	topicmx.RLock()
	defer topicmx.RUnlock()
	merge(ubworldSlice, matchWorker(workerId))
	// worldCommunication[workerId] <- ubworldSlice
	incr++
	if incr == p.Threads {
		for _, w := range workers {
			fmt.Println("Sending update to worker #", w.id)
			w.worldChannel <- World{
				world: world,
				turns: ubturns,
			}
			//fmt.Println("Turn update Broker:", ubturns)
		}
		completedTurns = ubturns
		incr = 0

	}
	//fmt.Println("mergeWorld")
	//return errors.New("Broker did not update.")
	return nil
}

type Broker struct{}

// func (b *Broker) ReportStatus(req stubs.StateRequest, req *stubs.Response) (err error) {
// 	return err
// }

func (b *Broker) UpdateBroker(req stubs.UpdateRequest, res *stubs.StatusReport) (err error) {
	err = updateBroker(req.Turns, req.World, req.WorkerId)
	return err
}

/*func (b *Broker) UpdateWorker(req stubs.TickerRequest, res *stubs.UpdateRequest) (err error) {
	res.World = world
	res.Turns = completedTurns
	return nil
}*/

func (b *Broker) MakeChannel(req stubs.ChannelRequest, res *stubs.StatusReport) (err error) {
	makeChannel(req.Threads)
	return
}

/*func (b *Broker) MakeChannelFromWorker(req stubs.ChannelRequest, res *stubs.StatusReport) (err error) {
	makeChannel(req.Threads)
	return
}*/

// Calls and connects to the worker (Subscribe)
func (b *Broker) ConnectWorker(req stubs.SubscribeRequest, res *stubs.StatusReport) (err error) {
	//request := stubs.WorkerRequest{StartY: 0, EndY: p.ImageHeight, StartX: 0, EndX: p.ImageWidth, WorldChan: worldChan[nextId], Turns: completedTurns, Params: p}
	err = subscribe(req.WorkerAddress)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func (b *Broker) ConnectDistributor(req stubs.Request, res *stubs.Response) (err error) {
	err = registerDistributor(req, new(stubs.StatusReport))
	// Checks if the connection and the worker is still on
	if len(workers) == p.Threads {
		for _, w := range workers {
			startGame := make(chan bool)

			if w.id != p.Threads-1 {
				w.params = WorkerParams{
					StartX: 0,
					StartY: w.id * unit,
					EndX:   p.ImageWidth,
					EndY:   w.id * (unit + 1),
				}
			} else {
				w.params = WorkerParams{
					StartX: 0,
					StartY: w.id * unit,
					EndX:   p.ImageWidth,
					EndY:   p.ImageHeight,
				}
			}

			/*w.params = WorkerParams{StartX: 0,
			StartY: 0,
			EndX:   p.ImageWidth,
			EndY:   p.ImageHeight}*/
			go subscribe_loop(w, startGame)
			go func() {
				startGame <- true
			}()
		}
	} else if len(workers) < p.Threads {
		for _, w := range workers {
			w.params = WorkerParams{
				StartX: 0,
				StartY: w.id * unit,
				EndX:   p.ImageWidth,
				EndY:   w.id * (unit + 1),
			}
			startGame := make(chan bool)
			go subscribe_loop(w, startGame)
			go func() {
				startGame <- true
			}()
		}
	}
	for {
		if p.Turns == completedTurns {
			res.World = world
			res.TurnsDone = completedTurns
			return
		}
	}
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
