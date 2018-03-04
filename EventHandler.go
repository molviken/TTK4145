package taskHandler

import(
elevio "./elevio"
)

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


/*Channels
Bruke en struct til channels? Sett mange som gjør det siden vi har så mange, lettere å holde styr på.
*/

type Channels struct {
	newOrderChan chan bool

//Alle channels definert her
}

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
func HandleEvents(channels Channels){
	for{
		select{
		case <-channels.newOrderChan:
			EventNewOrder(channels);
		}
	}
}




/*Events
The events that can occur
*/
func EventFloorReached(channels Channels){
}

func EventNewOrder(channels Channels){
}

func EventDoorTimeOut(channels Channels){
}

/* Vet ikke om dette trenger å være en event, men kanskje? Må hvertfall ha en slik funksjon */
func EventShouldStop(floor int){

}
