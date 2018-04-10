package main

import (
	//"container/list"
	elevio "./elevio"
	//queue "./queue"
	peers "./network/peers"
	task "./eventHandler"
	elevFunc "./elevFunc"
	bcast "./network/bcast"
	assigner "./ElevAssigner"
	"flag"
	"os"

	//"time"
)

	type Channels struct {
	button chan elevio.ButtonEvent
	floorSensor chan int
	obstr chan bool
	stop chan bool
	//transmitt chan task.UDP
	//receive chan task.UDP
}




func main(){
	//ip := elevio.GetMyIP()


	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	port := os.Args[2]
	
	button := make(chan elevio.ButtonEvent)
	floorSensor := make(chan int)
	obstr := make(chan bool)
	stop := make(chan bool)
	timeOut := make(chan bool)
	timerReset := make(chan bool)
	UDPTransmit := make(chan assigner.UDPmsg)
	UDPReceive := make(chan assigner.UDPmsg)
	//costTransmit := make(chan assigner.CostReply)
	//costReceive := make(chan assigner.CostReply)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	//lights := make(chan int)

	task.StartBroadcast(port)
	//Init(4)
	startFloor := elevio.InitElevator()

	task.EventHandlerInit(startFloor)
	//queue.InitQueue()
	go elevio.PollButtons(button)
	go elevio.PollFloorSensor(floorSensor)
	go elevio.PollObstructionSwitch(obstr)
	go elevio.PollStopButton(stop)
	go elevFunc.OpenDoor(timeOut, timerReset)
	go bcast.Transmitter(15657, UDPTransmit)
	go bcast.Receiver(15657, UDPReceive)
	//go bcast.Transmitter(15657, costTransmit)
	//go bcast.Receiver(15657, costReceive)
	//go assigner.ChooseElevator(UDPReceive, peerUpdateCh)
	//go elevFunc.HandleLights(lights)
	go peers.Transmitter(15658, string(id), peerTxEnable)
	go peers.Receiver(15658, peerUpdateCh)

	for{
		task.HandleEvents(button, floorSensor, obstr, stop, timeOut, timerReset, UDPReceive, UDPTransmit, peerUpdateCh, id)
	}

}
