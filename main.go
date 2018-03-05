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
	IP_PLASS_3 = "129.241.187.150"
	PORT_ELEV = ""
	IP_WANGEN = "192.168.43.56"
)

type elevator struct{
	curr_floor int
	curr_dir elevio.MotorDirection
}
func main(){
	elevio.Init("localhost:15657", 4)
	fmt.Println("Testing testing")
	conn.DialBroadcastUDP(15657)

	var init bool = false
	elevio.SetMotorDirection(elevio.MD_Down)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors  := make(chan int)
	drv_obstr   := make(chan bool)
	drv_stop    := make(chan bool) 
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	l := list.New()
	var elevator1 elevator
	for{
		select{
		case a:= <- drv_buttons:
			e := new(elevio.ButtonEvent)
			e = &a
			l.PushBack(e)
			elevFunc.CalculateCost(e, elevator1.curr_floor, elevator1.curr_dir)
			if(l.Front() != nil){
				elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor)
				elevFunc.GoToOrder(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor, l)
			}
		case a := <- drv_floors:
			elevFunc.ElevInit(a, init)
			elevator1.curr_floor = a
			elevio.SetFloorIndicator(a)
			elevFunc.ScanFloor(a, elevator1.curr_dir, l)
			if(l.Front() != nil){
				elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor)
				elevFunc.GoToOrder(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor, l)
			}
		case a := <- drv_stop:
			elevFunc.Fsm_Stop(a)
		}
	}	

}
