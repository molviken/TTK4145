package main

import (
	//"flag"
	//"time"
	"fmt"
	elevFunc "./elevFunc"
	//bcast "./Network/network/bcast"
	//localip "./network/localip"
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
var simHost string = "15657"

type elevator struct{
	curr_floor int
	curr_dir elevio.MotorDirection
}
func main(){
	elevio.Init("localhost:15657", 4)
	fmt.Println("Testing testing")
	conn.DialBroadcastUDP(15657)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors  := make(chan int)
	drv_obstr   := make(chan bool)
	drv_stop    := make(chan bool) 
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	
	list := []elevio.ButtonEvent{}
	var global elevio.ButtonEvent
	var elevator1 elevator
	var index int
	for{
		select{
		case a:= <- drv_buttons:
			list[index] = a
			global = a
			index  += 1
		case a := <- drv_floors:
			elevator1.curr_floor = a
			//fmt.Println("Floor sensor:%d", a)
		case a := <- drv_stop:
			elevFunc.Fsm_Stop(a)
		}
		if (index != 0){
			elevFunc.CalculateCost(list[index1], elevator1.curr_floor, list[index-1])
		}

		//elevFunc.GoToOrder(elevator1.curr_floor, list[index].OrderFloor)
	}	

}
