package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

// analogue to updateWorld function
/** Super-Secret `reversing a string' method we can't allow clients to see. **/
/*func ReverseString(s string, i int) string {
	time.Sleep(time.Duration(rand.Intn(i)) * time.Second)
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}*/

type GolOperations struct{}

func (s *GolOperations) Process(req stubs.Request, res *stubs.Response) (err error) {
	if req.Turns == 0 {
		err = errors.New("Number of turns must be bigger than 0")
		return
	}

	pause := false
	quit := false
	turn := 0
	threads := 1

	for t := 0; t < req.Turns; t++ {
		cellFlip := make([]util.Cell, req.ImageHeight*req.ImageWidth)
		//if pause {
		//	<-waitToUnpause
		//}
		if !pause && !quit {
			turn = t
			for j := range req.World {
				copy(req.PrevWorld[j], req.World[j])
			}
			if threads == 1 {
				req.World, cellFlip = CalculateNextState(req.ImageHeight, req.ImageWidth, 0, req.ImageHeight, req.World)
			}

			/*for _, cell := range cellFlip {
				// defer wg.Done()
				c.events <- CellFlipped{
					CompletedTurns: turn,
					Cell:           cell,
				}
			}

			c.events <- TurnComplete{
				CompletedTurns: turn,
			}*/

		} else {
			if quit {
				break
			} else {
				continue
			}
		}
		fmt.Println(cellFlip)
	}

	res.World = req.World
	res.TurnsDone = turn
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GolOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
