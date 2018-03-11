package eventHandler

import(
	"fmt"
	elevFunc "../elevFunc"
	//bcast ".././network/bcast"
	conn "../network/conn"
	queue "../queue"
	elevio "../elevio"
	"container/list"
)

/*
States:
idle - At a floor with closed door, awaiting orders
moving - Moving and can be between floors or going past a floor
doorOpen - At a floor with the door open
 */

const (
	idle = iota
	moving
	doorOpen
)
//Elevator Struct to keep track on current floor, current direction, and state
type elevator struct{
	curr_floor int
	curr_dir elevio.MotorDirection
	state int
}

var elevator1 elevator

/*
NewOrder: trggered when there is a new order added to the queue
FloorReached: when a floor is reached
DoorTimout: When the door has been open for three seconds in state openDoor, handle what happens next
ShouldStop: At each floor we check if there is an order to take if it is convinient
*/
func EventHandlerInit(startFloor int){
	elevator1.curr_floor = startFloor
	elevator1.curr_dir = elevio.MD_Stop
	elevator1.state = idle

}


func HandleEvents(button chan elevio.ButtonEvent, floorSensor chan int, obstr chan bool, stop chan bool,
	localL *list.List, remoteL *list.List, timeOut chan bool, timerReset chan bool){
	select{
	case button_pressed:= <- button:
			if (button_pressed.Button == elevio.BT_Cab && !elevFunc.DuplicateOrder(button_pressed, localL)){
				EventNewLocalOrder(localL, button_pressed, timerReset)
				elevFunc.PrintList(localL.Front())
			}else if(button_pressed.Button != elevio.BT_Cab){
				EventNewRemoteOrder(remoteL, button_pressed)
			}

	case floor := <- floorSensor:
			EventFloorReached(floor, localL, timerReset)

		//case obstr := <- channel.obstr:

		//case stop := <- channel.stop:
	case <- timeOut:
		EventDoorTimeOut(localL, timerReset)
	}
}

func GetDirection(floor int, order int) elevio.MotorDirection{
	dir :=  floor - order
	if (dir<0){return elevio.MD_Up}else if(dir>0){return elevio.MD_Down}else{return elevio.MD_Stop}
}

func EventNewLocalOrder(localL *list.List, button_pressed elevio.ButtonEvent, timerReset chan bool){
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event New Local Order in state: %s \n", stateString)

	queue.UpdateLocalQueue(localL, button_pressed)

	switch elevator1.state {
	case idle:
		elevator1.curr_dir = GetDirection(elevator1.curr_floor, button_pressed.Floor)
		elevio.SetMotorDirection(elevator1.curr_dir)

		if shouldStop(elevator1.curr_floor, elevator1.curr_dir, localL) {
			elevator1.curr_dir = elevio.MD_Stop
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.state = doorOpen
			timerReset <- true
		}else{
			elevator1.state = moving
		}
	case doorOpen:
		if(shouldStop(elevator1.curr_floor, elevator1.curr_dir, localL)){
			fmt.Println("I shouldStop Door open")
			timerReset <- true
		}

	case moving:
		fmt.Println("Moving\n")
	}

}


func EventNewRemoteOrder(remoteL *list.List, button_pressed elevio.ButtonEvent){
	queue.UpdateRemoteQueue(remoteL, button_pressed)
}

func EventFloorReached(floor int, localL *list.List, timerReset chan bool){
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Floor Reached in state %s \n",stateString)
	elevator1.curr_floor = floor
	//Lights

	switch elevator1.state {
	case moving:
		if(shouldStop(floor, elevator1.curr_dir, localL)){
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.curr_dir = elevio.MD_Stop
			elevator1.state = doorOpen
			timerReset <- true

		}else if(localL.Front().Value.(*elevio.ButtonEvent).Floor == floor){
			localL.Remove(localL.Front())

			elevFunc.PrintList(localL.Front())

			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.curr_dir = elevio.MD_Stop
			elevator1.state = doorOpen
			timerReset <- true
		}

	}
}



func EventDoorTimeOut(localL *list.List, timerReset chan bool){
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Door TimeOut in state %s \n",stateString)
	//elevFunc.PrintList(localL.Front())
	switch elevator1.state{
	case doorOpen:
			elevio.SetDoorOpenLamp(false) //turn of light
			//Ready for new direction
			if(localL.Front() != nil){
				elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor)
				elevio.SetMotorDirection(elevator1.curr_dir)
				elevator1.state = moving

				/*Check for pressing button several times while door is open*/
			}else{
			elevator1.state = idle
			}
		}
	stateString = elevFunc.StateToString(elevator1.state)
	fmt.Printf("Elevator state after timeout is : %s \n", stateString)
}

func shouldStop(floorSensor int, dir elevio.MotorDirection, l *list.List) bool{
	for k:= l.Front(); k != nil; k = k.Next(){

		if (k.Value.(*elevio.ButtonEvent).Floor == floorSensor){
			switch{
			case k.Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab:
				l.Remove(k)
				return true
			case dir == elevio.MD_Up:
				if (k.Value.(*elevio.ButtonEvent).Button != 1){
					l.Remove(k)
					return true
				}
			case dir == elevio.MD_Down:
				if (k.Value.(*elevio.ButtonEvent).Button != 0){
					l.Remove(k)
					return true

				}
			}
		}
	}
	return false
}

func StartBroadcast(){
	elevio.Init("localhost:15657", 4)
	conn.DialBroadcastUDP(15678)
}
