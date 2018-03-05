package eventHandler

import(
	elevFunc ".././elevFunc"
	//bcast ".././network/bcast"
	conn ".././network/conn"
	queue ".././queue"
	elevio ".././elevio"
	"container/list"
)
type Channels struct {
	newOrderChan chan bool
	buttons chan elevio.ButtonEvent
	floorSensor chan int
	obstr chan bool
	stop chan bool
	transmitt chan interface{}
	receive chan interface{}
}

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
func eventHandlerInit(channels Channels){
	var elevator elevator1

}


/*
NewOrder: trggered when there is a new order added to the queue
FloorReached: when the correct floor is reached, and we have a new objective
DoorTimout: When the door has been open for three seconds in state openDoor, handle what happens next
EventShouldStop: At each floor we check if there is an order to take if it is convinient
*/

//Denne funskjonen skal egentlig ikke gjøre så mye utenom select mellom alle eventsene
func HandleEvents(channel Channels, localL *list.List, remoteL *list.List){
	select{
		case button_pressed:= <- channel.button:
			if (button_pressed.Button == elevio.MD_Cab && !queue.DuplicateOrderLocal()){
				EventNewLocalOrder(queue.LocalL)
			}else if (button_pressed.Button != elevio.MD_Cab && !queue.DuplicateOrderRemote()){
				EventNewRemoteOrder()
			}


		case floor := <- channel.floorSensor:
			elevator1.curr_floor = floor
			elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor)
			if(l.Front() != nil){
				shouldStop(floor, elevator1.curr_dir, queue.LocalL)
			}


		case obstr := <- channel.obstr:


		case stop := <- channel.stop:
			
	}
}

func GetDirection(sensor int, order int)elevio.MotorDirection{
	dir := sensor - order
	if (dir<0){return elevio.MD_Up}else if(dir>0){return elevio.MD_Down}else{return elevio.MD_Stop}
}


/*Events
The events that can occur
*/
func EventFloorReached(channels Channels){
}

func EventNewLocalOrder(button_pressed elevio.ButtonEvent){
	queue.UpdateLocalQueue(button_pressed)
	if(l.Front() != nil){
		elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor)
		elevFunc.GoToOrder(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor, l)
	}
}
func EventNewRemoteOrder(button_pressed elevio.ButtonEvent){
	queue.UpdateRemoteQueue(button_pressed)
	if(l.Front() != nil){
		elevFunc.GoToOrder(elevator1.curr_floor, l.Front().Value.(*elevio.ButtonEvent).Floor, l)
	}
}

func EventDoorTimeOut(channels Channels){
}

/* Vet ikke om dette trenger å være en event, men kanskje? Må hvertfall ha en slik funksjon */
func shouldStop(floorSensor int, dir elevio.MotorDirection, l *list.List){
	for k:= l.Front(); k != nil; k = k.Next(){
		if (k.Value.(*elevio.ButtonEvent).Floor == floorSensor){
			switch{
			case dir == elevio.MD_Up:
				if (k.Value.(*elevio.ButtonEvent).Button != 1){
					OpenDoor()
					l.Remove(k)
				}
			case dir == elevio.MD_Down:
				if (k.Value.(*elevio.ButtonEvent).Button != 0){
					OpenDoor()
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