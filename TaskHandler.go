package taskHandler

import(
elevio "./elevio"
)

const (
	C_TYPE = "udp"
	C_HOST = "localhost"
	C_IP = "129.241.187.159"  // Localip pålogget Eduroam
	H_IP = "192.168.43.131" // Localip pålogget hotspot
	PORT_ELEV = ""
)

const(
	idle = iota
	moving
	doorOpen

)


type elevator struct{
	curr_floor int
	curr_dir elevio.MotorDirection
	state int
}
	
floorChan = make(chan int)
obstrChan = make(chan bool)
stopButtonChan = make(chan bool)
//LocalOrderChan = make(chan elevio.ButtonEvent))
buttonChan = make(chan elevio.ButtenEvent)

func taskHandlerInit{
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstrChan)
	go elevio.PollStopButton(stopButtonChan)
	go elevio.PollButtons(buttonChan)
	
	var elevator elevator1 


}
func runGoRoutines(){	

	select{
	//Handle button pressed
	case a:= <- buttonChan:
			e := new(elevio.ButtonEvent)
			e = &a

			queue.PushBack(e)
			
			
	//Handle lights 
		case a := <- floorChan:
			elevFunc.ElevInit(a, init)
			
			elevator1.curr_floor = a
			
			elevio.SetFloorIndicator(a)
			
			
			
			
			
		case a := <- drv_stop:
			elevFunc.Fsm_Stop(a)
		}
		if ( l.Front() != nil && l.Front().Next() != nil){
			//elevFunc.CalculateCost(l.Front().Value.(elevio.ButtonEvent), elevator1.curr_floor, elevator1.curr_dir)
		}



		if(l.Front() != nil){
			elevFunc.GoToOrder(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor, l.Front())
			fmt.Println(l.Front().Value)
		}
	}	

}


}

func EventFloorReached(){
	

	



}

func EventNewOrder(){


}

func DoorTimeOut(){

}

func shouldStop(floor int){
}


