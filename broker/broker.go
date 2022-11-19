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

type Broker struct{}

// Connect the worker in a loop
func register_loop(client *rpc.Client) {
	for {
		// response := new(stubs.JobReport)
		// err := client.Call(callback, job, response)
		if err != nil {
			fmt.Println("Error")
			fmt.Println(err)
			fmt.Println("Closing subscriber thread.")
			//Place the unfulfilled job back on the topic channel.
			// topic <- job
			break
		}

	}
}

// Initialise connecting worker, and if no error occurs, invoke register_loop.
func register(workerAddress string, callback string) {

	client, err := rpc.Dial("tcp", workerAddress)
	if err == nil {
		go register_loop(ch, client, callback)
	} else {
		fmt.Println("Error subscribing ", factoryAddress)
		fmt.Println(err)
		return err
	}
	return
}

// Send the work via the channel
func connectDistributor() {

}

func makeChannel(threads int) {
	channels = make([]chan [][]uint8, threads)
	for i := range channels {
		channels[i] = make(chan [][]uint8)
	}
}

// Calls and connects to the worker (Subscribe)
func (b *Broker) ConnectWorker(req stubs.Request, res *stubs.Response) {
	register("", "")
	return
}

// (Publish)
func (b *Broker) ConnectDistributor() {
	connectDistributor()
	return
}

func (b *Broker) MakeChannel(req stubs.ChannelRequest, res *stubs.StatusReport) {
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
