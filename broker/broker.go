package gol

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/stubs"
)

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

type Broker struct{}

// Connect the worker in a loop
func register_loop(w Worker) {
	for {

	}
}

// Initialise connecting worker, and if no error occurs, invoke register_loop.
func register(workerAddress string, callback string) (err error) {

	client, err := rpc.Dial("tcp", workerAddress)
	worker := Worker{
		id:           nextId,
		worker:       client,
		address:      &workerAddress,
		worldChannel: channels[nextId],
	}
	workers = append(workers, worker)
	nextId++

	if err == nil {
		go register_loop(worker)
	} else {
		fmt.Println(err)
		return err
	}
	return
}

// Send the work via the channel
func connectDistributor(work string, res *stubs.Response) {

}

func makeChannel(threads int) {
	channels = make([]chan [][]uint8, threads)
	for i := range channels {
		channels[i] = make(chan [][]uint8)
	}
}

// Calls and connects to the worker (Subscribe)
func (b *Broker) ConnectWorker(req stubs.RegisterRequest, res *stubs.StatusReport) (err error) {
	err = register(req.WorkerAddress, "")
	return
}

// (Publish)
func (b *Broker) ConnectDistributor(req stubs.StateRequest, res *stubs.ScreenShotResponse) (err error) {
	connectDistributor(req.State, res)
	return
}

func (b *Broker) Pause(req stubs.PauseRequest, res *stubs.StatusReport) (err error) {
	return
}

func (b *Broker) MakeChannel(req stubs.ChannelRequest, res *stubs.StatusReport) (err error) {
	makeChannel(req.Threads)
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
