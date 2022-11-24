package stubs

var ProcessTurnsHandler = "GolOperations.Process"

//var OperationsHandler = "GolOperations.Operations"
var JobHandler = "GolOperations.ListenToWork"
var PauseHandler = "GolOperations.ListenToPause"
var BrokerAndWorker = "Broker.ConnectWorker"
var BrokerChannel = "Broker.MakeChannel"
var MakeWorld = "Broker.MakeWorld"
var ConnectDistributor = "Broker.ConnectDistributor"
var ConnectWorker = "Broker.ConnectWorker"
var MakeChannel = "Broker.MakeChannel"
var Publish = "Broker.Publish"

const Save int = 0
const Quit int = 1
const Pause int = 2
const UnPause int = 3
const Kill int = 4
const Ticker int = 5

// REGISTER : DISTRIBUTOR
// SUBSCRIBE : WORKER

// (Broker -> Distributor)
// Applies for Save, Kill, Ticker

type Params struct {
	Threads     int
	ImageHeight int
	ImageWidth  int
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
	StartY int
	EndY   int
	StartX int
	EndX   int
	World  [][]uint8
	Turns  int
	Params Params
}

type PauseRequest struct {
	Pause bool
}

// (Distributor -> Broker)
// (Worker -> Broker)
type ChannelRequest struct {
	Threads int
}

// Connect to the broker from the first tiem and initialise world
type RegisterRequest struct {
	World       [][]uint8
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// ----------------- Keypresses --------------------

// (Broker -> Distributor)
// Applies for Save, Kill, Ticker

// (Worker -> Broker)
type SubscribeRequest struct {
	WorkerAddress string
}

// (Broker -> Worker)
type StateRequest struct {
	State int
}

// ----------------- Ticker -----------------------
// (Distributor -> Broker)
type TickerRequest struct {
}

// (Broker -> Distributor)
type TickerResponse struct {
	CompletedTurns  int
	AliveCellsCount int
}

// Response that doesn't require any additional data
type StatusReport struct {
	Status int
}

// 1. The distributor initialises the board, gets the input from the IO.
// 2. The distributor passes the value to the broker how many threads there would be.
// 3. The broker receives a request by 'ChannelRequest' and makes a channel, which communicates between workers and the broker.
// 4. The broker invokes a rpc call (GOLOperation) to the worker. 																=> rpc.GO
// 5. The worker accepts the rpc client, and iterates all the GOLOperation.
// 5-1. The distributor needs a report about calculating the number of the alive cell every 2 seconds.							=> goroutine for a (function)_loop
// (The counting alive function can be implemented either in the distributor, or in the worker.)
// 5-2. The distributor listens to a keypress, and if any action(keypress) has occured,											=> goroutine for a (function)_loop
//       we need to pass the action to the broker, and the broker sends the action to the worker.
