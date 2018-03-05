package main

import (
	"container/list"
	"fmt"
	queue "./queue"
	task "./eventHandler"
	elevFunc "./elevFunc"
	elevio "./elevio"
	conn "./network/conn"
)
const (
	C_TYPE = "udp"
	C_HOST = "localhost"
	C_IP = "129.241.187.159"  // Localip pålogget Eduroam
	H_IP = "192.168.43.131" // Localip pålogget hotspot
	PORT_ELEV = ""
)


func main(){
	task.StartBroadcast()
	queue.InitQueue()
	var localL = list.New()
	var remoteL = list.New()
	var channel task.Channels
	var init bool = false
	elevio.SetMotorDirection(elevio.MD_Down)

	channel.buttons := make(chan elevio.ButtonEvent)
	channel.floorSensor := make(chan int)
	channel.obstr := make(chan bool)
	channel.stop := make(chan bool)
	channel.transmitt := make(chan interface{})
	channel.receive := make(chan interface{})
	go elevio.PollButtons(channel.buttons)
	go elevio.PollFloorSensor(channel.floorSensor)
	go elevio.PollObstructionSwitch(channel.obstr)
	go elevio.PollStopButton(channels.stop)

	var elevator1 elevator

	for{
		task.HandleEvents(channel)
	}	

}
