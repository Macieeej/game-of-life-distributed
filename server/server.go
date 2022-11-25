package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

// analogue to updateWorld function
/** Super-Secret `reversing a string' method we can't allow clients to see. **/
/*func ReverseString(s string, i int) string {
time.Sleep(time.DurationCall runes[j], runes[i]
}
return string(runes)
}*/

var listener net.Listener
var pause bool
var waitToUnpause chan bool

var turnChan chan int
var worldChan chan [][]uint8

var internalTurn chan int
var internalWorld chan [][]uint8

var globalWorld [][]uint8
var completedTurns int

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

type GolOperations struct{}

func receiveFromBroker(t int, world [][]uint8) {
	internalTurn <- t
	internalWorld <- world

}
func sendToBroker() (int, [][]uint8) {
	turn := <-turnChan
	world := <-worldChan

	return turn, world
}
func (s *GolOperations) Report(req stubs.ActionRequest, res *stubs.Response) (err error) {
	res.TurnsDone, res.World = sendToBroker()
	return
}
func (s *GolOperations) UpdateWorld(req stubs.UpdateRequest, res *stubs.StatusReport) (err error) {
	receiveFromBroker(req.Turns, req.World)
	return
}

// func (s *GolOperations) ListenToQuit(req stubs.KillRequest, res *stubs.Response) (err error) {
// 	listener.Close()
// 	os.Exit(0)
// 	return
// }

// func (s *GolOperations) ListenToPause(req stubs.PauseRequest, res *stubs.Response) (err error) {
// 	pause = req.Pause
// 	if !pause {
// 		waitToUnpause <- true
// 	}
// 	return
// }

// func communicateBroker(t chan int) {
// 	turn := <-t
// 	Broker <- turn
// }

func (s *GolOperations) Process(req stubs.WorkerRequest, res *stubs.Response) (err error) {

	fmt.Println("Processing")
	//worldChan <- req.World
	var newWorld [][]uint8
	pause = false
	//threads := 1
	turn := 0
	for t := 0; t < req.Turns; t++ {
		fmt.Println("A")
		turn = <-internalTurn
		fmt.Println("B")
		globalWorld = <-internalWorld
		fmt.Println("Calculating turn ")
		newWorld, _ = CalculateNextState(req.Params.ImageHeight, req.Params.ImageWidth, req.StartY, req.EndY, globalWorld)
		//newWorld, _ = CalculateNextState(req.Params.ImageHeight, req.Params.ImageWidth, 0, req.Params.ImageHeight, <-worldChan)
		//worldChan <- newWorld
		fmt.Println("Turn done on a server: ", turn)
		turn++
		for i := range newWorld {
			copy(globalWorld[i], newWorld[i])
		}
		completedTurns = turn
		turnChan <- turn
		worldChan <- globalWorld
		//if pause {
		//	<-waitToUnpause
		//}
		//if !pause /*&& !quit*/ {
		//turn = <-turnChan
		//if threads == 1 {
		//newWorld, _ = CalculateNextState(req.Params.ImageHeight, req.Params.ImageWidth, 0, req.Params.ImageHeight, <-worldChan)
		//worldChan <- newWorld
		//}
		/*} /*else {
			if quit {
				break
			} else {
				continue
			}
		}*/
	}
	fmt.Println("Turn done on a server: ", turn)
	res.World = newWorld
	res.TurnsDone = turn
	return
}

// kill := make(chan bool)

func main() {
	//pAddr := flag.String("port", "8050", "Port to listen on")
	//brokerAddr := flag.String("broker", "127.0.0.1:8030", "Address of broker instance")
	flag.Parse()
	//client, err := rpc.Dial("tcp", *brokerAddr)
	client, err := rpc.Dial("tcp", "127.0.0.1:8030")
	if err != nil {
		fmt.Println(err)
	}
	rpc.Register(&GolOperations{})
	//fmt.Println(*pAddr)
	//fmt.Println(getOutboundIP() + ":" + *pAddr)
	//listener, err := net.Listen("tcp", ":"+*pAddr)
	fmt.Println("8050")
	fmt.Println(getOutboundIP() + ":" + "8050")
	listenerr, err := net.Listen("tcp", ":"+"8050")
	if err != nil {
		fmt.Println(err)
	}
	subscribe := stubs.SubscribeRequest{
		//WorkerAddress: getOutboundIP() + ":" + *pAddr,
		WorkerAddress: getOutboundIP() + ":" + "8050",
	}
	client.Call(stubs.ConnectWorker, subscribe, new(stubs.StatusReport))
	turnChan = make(chan int)
	defer listenerr.Close()
	rpc.Accept(listenerr)

}
