package stubs

import "uk.ac.bris.cs/gameoflife/util"

var ActionHandler = "Broker.Action"
var ActionReport = "Broker.ActionWithReport"
var ConnectDistributor = "Broker.ConnectDistributor"
var ConnectWorker = "Broker.ConnectWorker"
var MakeChannel = "Broker.MakeChannel"
var Publish = "Broker.Publish"
var UpdateBroker = "Broker.UpdateBroker"
var HandleCellFlip = "Broker.HandleCellFlip"
var ProcessTurnsHandler = "GolOperations.Process"
var ActionHandlerWorker = "GolOperations.Action"
var ActionReportWorker = "GolOperations.ActionWithReport"
var UpdateWorker = "GolOperations.UpdateWorker"

const NoAction int = 0
const Save int = 1
const Quit int = 2
const Pause int = 3
const UnPause int = 4
const Kill int = 5

// REGISTER : DISTRIBUTOR
// SUBSCRIBE : WORKER

type Params struct {
	Threads     int
	ImageHeight int
	ImageWidth  int
	Turns       int
}

type Response struct {
	World     [][]uint8
	TurnsDone int
}

type Request struct {
	World       [][]uint8
	Threads     int
	Turns       int
	ImageWidth  int
	ImageHeight int
}

type WorkerRequest struct {
	WorkerId int
	StartY   int
	EndY     int
	StartX   int
	EndX     int
	World    [][]uint8
	Turns    int
	Params   Params
}

type PauseRequest struct {
	Pause bool
}

type ChannelRequest struct {
	Threads int
}

type RegisterRequest struct {
	World       [][]uint8
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type UpdateRequest struct {
	World    [][]uint8
	Turns    int
	WorkerId int
	CellFlip []util.Cell
}

type SubscribeRequest struct {
	WorkerAddress string
}

// (Broker -> Worker)
type StateRequest struct {
	State int
}

type TickerRequest struct {
}

// Response that doesn't require any additional data
type StatusReport struct {
	Status int
}

type CellFlipResponse struct {
	CompletedTurns int
	CellFlipped    []util.Cell
}
