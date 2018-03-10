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
var initElev bool = false



/*
States:
idle - At a floor with closed door, awaiting orders
moving - Moving and can be between floors or going past a floor
doorOpen - At a floor with the door open

Tanken er å bruke switch(state) i hver event
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
/* Initialize the go routines needed to handle events, and initialize the elevetor*/



/*
NewOrder: trggered when there is a new order added to the queue
FloorReached: when the correct floor is reached, and we have a new objective
DoorTimout: When the door has been open for three seconds in state openDoor, handle what happens next
EventShouldStop: At each floor we check if there is an order to take if it is convinient
*/

//Denne funskjonen skal egentlig ikke gjøre så mye utenom select mellom alle eventsene
func HandleEvents(button chan elevio.ButtonEvent, floorSensor chan int, obstr chan bool, stop chan bool, localL *list.List, remoteL *list.List,
	timeOut chan bool, timerReset chan bool){
	select{
	case button_pressed:= <- button:
			if (button_pressed.Button == elevio.BT_Cab){
				EventNewLocalOrder(localL, button_pressed, timerReset)
			}else if(button_pressed.Button != elevio.BT_Cab){
				EventNewRemoteOrder(remoteL, button_pressed)
			}


	case floor := <- floorSensor:
			EventFloorReached(floor, localL, timerReset)
			//fmt.Println("floorSensor: ", floor)

		//case obstr := <- channel.obstr:


		//case stop := <- channel.stop:
	case <- timeOut:
		EventDoorTimeOut()
	}
}



func GetDirection(floor int, order int) elevio.MotorDirection{
	dir :=  floor - order
	if (dir<0){return elevio.MD_Up}else if(dir>0){return elevio.MD_Down}else{return elevio.MD_Stop}
}


/*Events
The events that can occur
*/
//func EventFloorReached(channels Channels){}

func EventNewLocalOrder(localL *list.List, button_pressed elevio.ButtonEvent, timerReset chan bool){
	queue.UpdateLocalQueue(localL, button_pressed)
	switch elevator1.state {
	case idle:
		elevator1.curr_dir = GetDirection(elevator1.curr_floor, button_pressed.Floor)
		elevio.SetButtonLamp(button_pressed.Button, button_pressed.Floor,true)//Sett lys med en gang

		if shouldStop(elevator1.curr_floor, elevator1.curr_dir, localL) {
			elevator1.curr_dir = elevio.MD_Stop
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.state = doorOpen
			timerReset <- true //Channel for door timer, when timer is out, EventDoorTimeOut happens
		}else{
			elevator1.state = moving
		}
	case moving:
	case doorOpen:
		if shouldStop(elevator1.curr_floor, elevator1.curr_dir, localL){

			//remove order here, not in shouldStop?
			//Need to find a way to find the correct order though
		}
	default:
		if(localL.Front() != nil){
		elevFunc.ExecuteOrder(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor)
		//Sets the new direction if there are remaining floors
	}
	}
}


func EventNewRemoteOrder(remoteL *list.List, button_pressed elevio.ButtonEvent){
	queue.UpdateRemoteQueue(remoteL, button_pressed)
}

func EventFloorReached(floor int, localL *list.List, timerReset chan bool){
	fmt.Println("Event Floor %d Reached. \n", floor)
	elevator1.curr_floor = floor
	//Lights

	switch elevator1.state {
	case moving:
		if(shouldStop(elevator1.curr_floor, elevator1.curr_dir, localL)){
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.curr_dir = elevio.MD_Stop
			elevator1.state = doorOpen
			timerReset <- true
			//Channel for door timer, when timer is out, EventDoorTimeOut happens
		}
	default:
		elevFunc.ExecuteOrder(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor)
	}
}


func EventDoorTimeOut(){

}

func shouldStop(floorSensor int, dir elevio.MotorDirection, l *list.List) bool{
	for k:= l.Front(); k != nil; k = k.Next(){
		if (k.Value.(*elevio.ButtonEvent).Floor == floorSensor){
			switch{
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
	conn.DialBroadcastUDP(15657)
}
