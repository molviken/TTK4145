package main

import (
	//"flag"
	//"time"
	"fmt"
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
	//ip,_ := localip.LocalIP();
	//fmt.Println(ip)
	elevio.Init("localhost:15657", 4)
	fmt.Println("Testing testing")
	conn.DialBroadcastUDP(15657)
	elevio.SetDoorOpenLamp(true)
	c_floor := make(chan int)
	elevio.SetMotorDirection(elevio.MD_Down)
	go elevio.PollFloorSensor(c_floor)
	for{
		select{
		case a := <- c_floor:
			fmt.Println("%+v\n", a)
			if a == 3{
				elevio.SetMotorDirection(elevio.MD_Down)	
			} else if a == 0{
				elevio.SetMotorDirection(elevio.MD_Up)
			}

		}
		
	}
	

}
