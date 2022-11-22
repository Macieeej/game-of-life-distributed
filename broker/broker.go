package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/stubs"
)

// 1 : Pass the report from the distributor. (Ticker request (broker -> dis))
// 2 : Connect with the distributor (Register (broker -> dis))
// 3 : Pass the keypress arguments from the distributor to the worker. (Keypress (dis -> broker) (broker -> worker))

// Channels that are used to communicate with broker and worker
var channels []chan [][]uint8
var workers []Worker
var nextId = 0

type Worker struct {
	id           int
	stateSwitch  int
	worker       *rpc.Client
	address      *string
	worldChannel chan [][]uint8
}

// Connect the worker in a loop
func register_loop(startY, endY, startX, endX int, world [][]uint8, out chan<- [][]uint8, turns int, w Worker) {
	for {
		response := new(stubs.Response)
		//registerReq := stubs.RegisterRequest{WorkerAddress: *w.address}
		workerReq := stubs.WorkerRequest{StartY: startY, EndY: endY, StartX: startX, EndX: endX, World: world, Turns: turns}
		err := w.worker.Call(stubs.PauseHandler, workerReq, &response)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
	}
}

// Initialise connecting worker, and if no error occurs, invoke register_loop.
func register(world stubs.Request, workerAddress string, callback string) (err error) {

	client, err := rpc.Dial("tcp", workerAddress)
	worker := Worker{
		id:           nextId,
		worker:       client,
		address:      &workerAddress,
		worldChannel: channels[nextId],
	}

	//var worldFragment [][]uint8

	if world.Threads == 1 && err == nil {
		channels := make([]chan [][]uint8, world.Threads)
		go register_loop(0, world.ImageHeight, 0, world.ImageWidth, world.World, channels[0], world.Turns, worker)
		/*worldPart := <-channels[0]
		    worldFragment = append(worldFragment, worldPart...)
			for j := range worldFragment {
				copy(world.World[j], worldFragment[j])
			}*/

	} else if err == nil {
		channels := make([]chan [][]uint8, world.Threads)
		unit := int(world.ImageHeight / world.Threads)
		for i := 0; i < world.Threads; i++ {
			channels[i] = make(chan [][]uint8)
			if i == world.Threads-1 {
				// Handling with problems if threads division goes with remainders
				//go worker(p, i*unit, p.ImageHeight, 0, p.ImageWidth, world, channels[i], c, turn)
				go register_loop(i*unit, world.ImageHeight, 0, world.ImageWidth, world.World, channels[i], world.Turns, worker)
			} else {
				//go worker(p, i*unit, (i+1)*unit, 0, p.ImageWidth, world, channels[i], c, turn)
				go register_loop(i*unit, (i+1)*unit, 0, world.ImageWidth, world.World, channels[i], world.Turns, worker)
			}
		}
		/*for i := 0; i < world.Threads; i++ {
			worldPart := <-channels[i]
			worldFragment = append(worldFragment, worldPart...)
		}
		for j := range worldFragment {
			copy(world.World[j], worldFragment[j])
		}*/
	} else {
		fmt.Println(err)
		return err
	}

	workers = append(workers, worker)
	nextId++

	return
}

// Send the work via the channel
func connectDistributor(work string, res *stubs.Response) error {

	return nil
}

func makeChannel(threads int) {
	channels = make([]chan [][]uint8, threads)
	for i := range channels {
		channels[i] = make(chan [][]uint8)
		fmt.Println("Created channel #", i)
	}
}

type Broker struct{}

// (CreateChannel)
func (b *Broker) MakeChannel(req stubs.ChannelRequest, res *stubs.StatusReport) (err error) {
	makeChannel(req.Threads)
	return
}

// Calls and connects to the worker (Subscribe)
func (b *Broker) ConnectWorker(req stubs.RegisterRequest, res *stubs.StatusReport) (err error) {
	world := stubs.Request{World: req.World, ImageHeight: req.ImageHeight, ImageWidth: req.ImageWidth, Turns: req.Turns, Threads: req.Threads}
	err = register(world, req.WorkerAddress, req.Callback)
	return
}

// (Publish)
func (b *Broker) ConnectDistributor(req stubs.StateRequest, res *stubs.ScreenShotResponse) (err error) {
	err = connectDistributor(req.State, res)
	return
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
