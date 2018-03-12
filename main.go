package main

import (
	"container/list"
	elevio "./elevio"
	//queue "./queue"
	task "./eventHandler"
	elevFunc "./elevFunc"
	"os"
	bcast "./network/bcast"
	"flag"
	//"time"
)
const (
	C_TYPE = "udp"
	C_HOST = "localhost"
	C_IP = "129.241.187.159"  // Localip pålogget Eduroam
	H_IP = "192.168.43.131" // Localip pålogget hotspot
	PORT_ELEV = ""
)
	type Channels struct {
	button chan elevio.ButtonEvent
	floorSensor chan int
	obstr chan bool
	stop chan bool
	transmitt chan interface{}
	receive chan interface{}
}




func main(){
	var localL = list.New()
	var remoteL = list.New()
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
	//lights := make(chan int)
	transmitt := make(chan task.UDPmsg)
	receive := make(chan task.UDPmsg)

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
	go bcast.Transmitter(15657, transmitt)
	go bcast.Receiver(15657, receive)
	//go elevFunc.HandleLights(lights)

	for{
		task.HandleEvents(button, floorSensor, obstr, stop, localL, remoteL, timeOut, timerReset, receive, transmitt)
	}

}
