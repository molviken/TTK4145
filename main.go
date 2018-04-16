package main

import (
	elevio "./elevio"
	peers "./network/peers"
	task "./eventHandler"
	elevFunc "./elevFunc"
	bcast "./network/bcast"
	assigner "./ElevAssigner"
	"flag"
	"fmt"
	"strconv"
	//"os"
)


func main(){

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	//port := os.Args[2] // FOR USE WITH SIMULATOR
	port := "15657" // FOR USE WITH HARDWARE
	real_id, _ := strconv.Atoi(id)

	button := make(chan elevio.ButtonEvent)
	floorSensor := make(chan int)
	obstr := make(chan bool)
	timeOut := make(chan bool)
	timerReset := make(chan bool)
	UDPTransmit := make(chan assigner.UDPmsg)
	UDPReceive := make(chan assigner.UDPmsg)
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	obstrTimerReset := make(chan bool)
	
	elevio.Init("localhost:"+port, 4)
	startFloor := elevio.InitElevator()

	task.EventHandlerInit(startFloor, real_id)

	go elevio.PollButtons(button)
	go elevio.PollFloorSensor(floorSensor)
	go elevFunc.OpenDoor(timeOut, timerReset)
	go bcast.Transmitter(20015, UDPTransmit)
	go bcast.Receiver(20015, UDPReceive)
	go elevFunc.ObstructionTimeOut(obstr,obstrTimerReset, task.LocalL)
	go peers.Transmitter(21015, string(id), peerTxEnable)
	go peers.Receiver(21015, peerUpdateCh)

	fmt.Println(" ")
	fmt.Println("Elev ID: ",real_id)

	for{
		task.HandleEvents(button, floorSensor, obstr, timeOut, timerReset, UDPReceive, UDPTransmit, peerUpdateCh, real_id, obstrTimerReset)
	}

}
