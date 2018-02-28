package main

import (
	//"flag"
	//"time"
	"container/list"
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
	
	l := list.New()//var global elevio.ButtonEvent
	var elevator1 elevator
	//var indexxx int = 0
	for{
		select{
		case a:= <- drv_buttons:
			e := new(elevio.ButtonEvent)
			e = &a

			l.PushFront(e)
			//fmt.Println("Next: ", l.Front().Next())
			fmt.Println("list.Floor: ", l.Front().Value.(*elevio.ButtonEvent).Floor)
			fmt.Println("list.Button: ", l.Front().Value.(*elevio.ButtonEvent).Button)
		case a := <- drv_floors:
			elevator1.curr_floor = a
			if (l.Front() != nil){
				elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor)
			}
			
			fmt.Println("Dir:", elevator1.curr_dir)
			fmt.Println("curr floor: ", elevator1.curr_floor)
		case a := <- drv_stop:
			elevFunc.Fsm_Stop(a)
		}
		if (l.Front().Next() != nil && l.Front() != nil){
			elevFunc.CalculateCost(l.Front().Value.(elevio.ButtonEvent), elevator1.curr_floor, elevator1.curr_dir)
		}
		list = elevFunc.GoToOrder(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor, l.Front().Value.(elevio.ButtonEvent))
	}	

}
