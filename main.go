package main

import (
	"container/list"
	elevio "./elevio"
	queue "./queue"
	task "./eventHandler"
	elevFunc "./elevFunc"
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

	button := make(chan elevio.ButtonEvent)
	floorSensor := make(chan int)
	obstr := make(chan bool)
	stop := make(chan bool)
	timeOut := make(chan bool)
	timerReset := make(chan bool)


	//transmitt := make(chan interface{})
	//receive := make(chan interface{})

	task.StartBroadcast()
	queue.InitQueue()
	elevio.SetMotorDirection(elevio.MD_Down)
	go elevio.PollButtons(button)
	go elevio.PollFloorSensor(floorSensor)
	go elevio.PollObstructionSwitch(obstr)
	go elevio.PollStopButton(stop)
	go elevFunc.OpenDoor(timeOut, timerReset)

	for{
		task.HandleEvents(button, floorSensor, obstr, stop, localL, remoteL, timeOut, timerReset)
	}

}
