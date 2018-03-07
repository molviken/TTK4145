package eventHandler

import(
	"fmt"
	elevFunc ".././elevFunc"
	//bcast ".././network/bcast"
	conn ".././network/conn"
	queue ".././queue"
	elevio ".././elevio"
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
func HandleEvents(button chan elevio.ButtonEvent, floorSensor chan int, obstr chan bool, stop chan bool, localL *list.List, remoteL *list.List){
	select{
	
	
		case button_pressed:= <- button:
		
			if (localL.Front() == nil){
				EventNewLocalOrder(localL, button_pressed)
			}else if(button_pressed.Button == elevio.BT_Cab){
				
				EventNewLocalOrder(localL, button_pressed)
			}
		
			if (remoteL.Front() == nil ){
				EventNewRemoteOrder(remoteL, button_pressed)
			}else if(button_pressed.Button != elevio.BT_Cab && !queue.DuplicateOrderRemote(remoteL, button_pressed)){
				EventNewRemoteOrder(remoteL, button_pressed)
			}

			if (localL.Front() != nil){
				
				fmt.Println("klar til execute")
				elevFunc.ExecuteOrder(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor, localL.Front())
			}
			elevFunc.PrintList(localL.Front())
			
		case floor := <- floorSensor:
			//fmt.Println("floorSensor: ", floor)

			elevFunc.ElevInit(floor, initElev)
			elevator1.curr_floor = floor
			if(localL.Front() != nil){
				elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor)
				shouldStop(floor, elevator1.curr_dir, localL)
				elevFunc.ExecuteOrder(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor, localL.Front())
			}
			if (localL.Front() != nil && localL.Front().Value.(*elevio.ButtonEvent).Floor == elevator1.curr_floor){
				localL.Remove(localL.Front())
				
			}
			if (localL.Front() != nil){
				elevFunc.ExecuteOrder(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor, localL.Front())
			}
			

		//case obstr := <- channel.obstr:


		//case stop := <- channel.stop:
			
	}
}

func GetDirection(sensor int, order int)elevio.MotorDirection{
	dir := sensor - order
	if (dir<0){return elevio.MD_Up}else if(dir>0){return elevio.MD_Down}else{return elevio.MD_Stop}
}


/*Events
The events that can occur
*/
//func EventFloorReached(channels Channels){}

func EventNewLocalOrder(localL *list.List, button_pressed elevio.ButtonEvent){
	queue.UpdateLocalQueue(localL,button_pressed)
}
func EventNewRemoteOrder(remoteL *list.List, button_pressed elevio.ButtonEvent){
	queue.UpdateRemoteQueue(remoteL, button_pressed)
}

//func EventDoorTimeOut(channels Channels){}

/* Vet ikke om dette trenger å være en event, men kanskje? Må hvertfall ha en slik funksjon */
func shouldStop(floorSensor int, dir elevio.MotorDirection, l *list.List){
	for k:= l.Front(); k != nil; k = k.Next(){
		if (k.Value.(*elevio.ButtonEvent).Floor == floorSensor){
			switch{
			case dir == elevio.MD_Up:
				if (k.Value.(*elevio.ButtonEvent).Button != 1){
					elevFunc.OpenDoor()
					l.Remove(k)
				}
			case dir == elevio.MD_Down:
				if (k.Value.(*elevio.ButtonEvent).Button != 0){
					elevFunc.OpenDoor()
					l.Remove(k)
				}
			}
		}
	}
}

func StartBroadcast(){
	elevio.Init("localhost:15657", 4)
	conn.DialBroadcastUDP(15657)
}
