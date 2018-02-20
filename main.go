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
	/*"os"
	"os/signal"
	"strconv"
	"time"
*/
)
const (
	C_TYPE = "udp"
	C_HOST = "localhost"
	C_IP = "129.241.187.159"  // Localip pålogget Eduroam
	H_IP = "192.168.43.131" // Localip pålogget hotspot
	PORT_ELEV = ""
)
var simHost string = "15657"


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
	


	for{
		
		select{
		case a:= <- drv_buttons:
			elevFunc.GetOrder(a)	
		case a := <- drv_floors:
			fmt.Println("Floor sensor:%v\n", a)
		}
		
	}	

}
