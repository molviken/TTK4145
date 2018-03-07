package main

import (
	"container/list"
	elevio "./elevio"
	queue "./queue"
	task "./eventHandler"
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
	task.StartBroadcast()
	queue.InitQueue()


	button := make(chan elevio.ButtonEvent)
	floorSensor := make(chan int)
	obstr := make(chan bool)
	stop := make(chan bool)
	//transmitt := make(chan interface{})
	//receive := make(chan interface{})
	go elevio.PollButtons(button)
	go elevio.PollFloorSensor(floorSensor)
	go elevio.PollObstructionSwitch(obstr)
	go elevio.PollStopButton(stop)

	var localL = list.New()
	var remoteL = list.New()
	var init bool = false
	elevio.SetMotorDirection(elevio.MD_Down)

	for{
		task.HandleEvents(button, floorSensor, obstr, stop, localL, remoteL)
	}	

}
