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

// Handing out the world again to the worker.
/*func handleWorkers() {
	for {
		updateWorld()
		getReport()
		mergeWorld()
	}
}*/

/*// Broker -> Server
func updateWorld() {
	for _, w := range workers {
		fmt.Println("Updating world")
		w.worker.Call(stubs.UpdateWorld, stubs.UpdateRequest{World: world, Turns: completedTurns}, new(stubs.StatusReport))
		fmt.Println("Updated world")
	}
	//getReport()
}

// Server -> Broker
func getReport() {
	report := new(stubs.Response)
	for _, w := range workers {
		w.worker.Call(stubs.Report, stubs.ActionRequest{Action: stubs.NoAction}, report)
		completedTurns = report.TurnsDone
		w.worldChannel <- World{
			world: report.World,
			turns: report.TurnsDone,
		}
	}
	//updateWorld()
}*/

// Connect the worker in a loop
func subscribe_loop(w Worker) {
	fmt.Println("Loooping")
	response := new(stubs.Response)
	workerReq := stubs.WorkerRequest{WorkerId: w.id, StartY: w.params.StartY, EndY: w.params.EndY, StartX: w.params.StartX, EndX: w.params.EndX, World: world, Turns: p.Turns, Params: p}
	err := w.worker.Go(stubs.ProcessTurnsHandler, workerReq, response, nil)
	//go handleWorkers()
	if err != nil {
		fmt.Println("Error calling ProcessTurnsHandler")
		//fmt.Println(err)
		fmt.Println("Closing subscriber thread.")
		//worldChanS <- worldS
	}
	//time.Sleep(5 * time.Second)
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
		fmt.Println(updateResponse.Status)
	}
	//fmt.Println("Worker done", response.TurnsDone, response.World)
	/*for {
		err := w.worker.Call(stubs.PauseHandler, workerReq, &response)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
	}*/
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

	if err == nil {
		fmt.Println("Looooop")
		go subscribe_loop(newWorker)
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

	//nextId++

	return
}

// func register_loop() {

// }

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
	unit = p.Threads / p.ImageWidth
	completedTurns = 0
	// DONE: Make a channel for the world
	// if p.Threads == 1 && err == nil {
	// 	//go subscribe_loop(0, p.ImageHeight, 0, p.ImageWidth, world, worldChan[0], req.Turns, worker)
	// 	worldChan[0] <- World{world: world, turns: req.Turns}
	// 	fmt.Println("Created channel #", 0)
	// } else if err == nil {
	// 	if len(workers) != p.Threads-1 {

	// 		for i := 0; i < p.Threads; i++ {
	// 			//worldChan[i] = make(chan [][]uint8)
	// 			if i == p.Threads-1 {

	// 				worldChan[i] <- World{world: world, turns: req.Turns}
	// 				fmt.Println("Assigned world slice to the worldChan #", i)
	// 			} else {
	// 				worldChan[i] <- World{world: world, turns: req.Turns}
	// 				fmt.Println("Assigned world slice to the worldChan #", i)
	// 			}
	// 		}
	// 	} else {
	// 		return
	// 	}

	// }
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

// Accepting every information about world progress and merging into a world.
func mergeWorld(mwworld [][]uint8) {

	var newWorld [][]uint8
	for _, w := range workers {
		fmt.Println("mergeworld #", w.id)
		prevWorld := mwworld
		newWorld = append(newWorld, prevWorld...)
	}

	for i := range newWorld {
		copy(world[i], newWorld[i])
	}

}
func updateBroker(ubturns int, ubworld [][]uint8) error {
	topicmx.RLock()
	defer topicmx.RUnlock()
	for _, w := range workers {
		fmt.Println("Sending update to worker #", w.id)
		w.worldChannel <- World{
			world: ubworld,
			turns: ubturns,
		}
		//fmt.Println("Turn update Broker:", ubturns)
	}
	completedTurns = ubturns
	//fmt.Println("mergeWorld")
	mergeWorld(ubworld)
	//return errors.New("Broker did not update.")
	return nil
}

type Broker struct{}

// func (b *Broker) ReportStatus(req stubs.StateRequest, req *stubs.Response) (err error) {
// 	return err
// }

func (b *Broker) UpdateBroker(req stubs.UpdateRequest, res *stubs.StatusReport) (err error) {
	err = updateBroker(req.Turns, req.World)
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
