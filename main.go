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
	//"os"
	"fmt"
	//"time"
	"strconv"
)


var costMap map[int] assigner.UDPmsg

func main(){
	//ip := elevio.GetMyIP()


	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	//port := os.Args[2]
	real_id, _ := strconv.Atoi(id)

	button := make(chan elevio.ButtonEvent)
	floorSensor := make(chan int)
	obstr := make(chan bool)
	stop := make(chan bool)
	timeOut := make(chan bool)
	timerReset := make(chan bool)
	UDPTransmit := make(chan assigner.UDPmsg)
	UDPReceive := make(chan assigner.UDPmsg)
	UpdateRemote := make(chan elevio.ButtonEvent)
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	obstrTimerReset := make(chan bool)
	

	//lights := make(chan int)

	task.StartBroadcast("15657")
	//Init(4)
	elevio.Init("localhost:15657", 4)
	startFloor := elevio.InitElevator()

	task.EventHandlerInit(startFloor, real_id)

	//queue.InitQueue()
	go elevio.PollButtons(button)
	go elevio.PollFloorSensor(floorSensor)
	go elevio.PollObstructionSwitch(obstr)
	go elevio.PollStopButton(stop)
	go elevFunc.OpenDoor(timeOut, timerReset)
	go bcast.Transmitter(15657, UDPTransmit)
	go bcast.Receiver(15657, UDPReceive)
	go elevFunc.ObstructionTimeOut(obstr,obstrTimerReset, task.LocalL)
	//go bcast.Transmitter(15657, costTransmit)
	//go bcast.Receiver(15657, costReceive)
	//go assigner.ChooseElevator(UDPReceive, peerUpdateCh)
	//go elevFunc.HandleLights(lights)
	go peers.Transmitter(15659, string(id), peerTxEnable)
	go peers.Receiver(15659, peerUpdateCh)
	fmt.Println(" ")
	fmt.Println("REAL ELEV ID  ",real_id)
	for{
		task.HandleEvents(button, floorSensor, obstr, stop, timeOut, timerReset, UDPReceive, UDPTransmit, peerUpdateCh, real_id, UpdateRemote, obstrTimerReset)
	}

}
