package eventHandler

import(
	"fmt"
	elevFunc "../elevFunc"
	//bcast ".././network/bcast"
	conn "../network/conn"
	queue "../queue"
	elevio "../elevio"
	"container/list"
		peers "../network/peers"
	assigner "../ElevAssigner"
)

/*
States:
idle - At a floor with closed door, awaiting orders
moving - Moving and can be between floors or going past a floor
doorOpen - At a floor with the door open
 */
var localL = list.New()
var remoteL = list.New()
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
var numElevsAlive int
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


func HandleEvents(button chan elevio.ButtonEvent, floorSensor chan int, obstr chan bool, stop chan bool, timeOut chan bool, timerReset chan bool,
	 receive chan assigner.OrderMsg, transmitt chan assigner.OrderMsg, peerUpdateCh chan peers.PeerUpdate, id int){
	select{
	case button_pressed:= <- button:
			if (button_pressed.Button == elevio.BT_Cab && !elevFunc.DuplicateOrder(button_pressed)){
				EventNewLocalOrder(button_pressed, timerReset)
				elevFunc.PrintList(localL.Front())

			}else if(button_pressed.Button != elevio.BT_Cab){
				EventNewRemoteOrder(remoteL, button_pressed, transmitt, id)
			}

	case floor := <- floorSensor:
			EventFloorReached(floor, timerReset)

		//case obstr := <- channel.obstr:

		//case stop := <- channel.stop:
	case <- timeOut:
		EventDoorTimeOut( timerReset)

	case a := <- receive:
		EventReceived(a, p, elevator1.curr_floor, elevator1.curr_dir, id)

	case p := <-peerUpdateCh:
		EventPeerUpdate(p)
	}

}

func GetDirection(floor int, order int) elevio.MotorDirection{
	dir :=  floor - order
	if (dir<0){return elevio.MD_Up}else if(dir>0){return elevio.MD_Down}else{return elevio.MD_Stop}
}

func EventReceived(a assigner.UDPMsg, p peers.PeerUpdate, floor int, c_dir elevio.MotorDirection, transmitt chan assigner.UDPMsg, id int){
	switch a.MsgID{
	case 1:
		assigner.ChooseElevator(a, p)
	case 2:
		cost := queue.Cost(a.Order, floor, elevator1.curr_dir)
		var msg assigner.UDPMsg
		msg.id = 1
		msg.Cost = cost
		msg.ElevID = id
		transmitt <- msg
	case 3:
		if (a.WinnerID == id){
			EventNewLocalOrder( a.Order, )
		}
	}
}

func EventPeerUpdate(p peers.PeerUpdate){
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)
	numElevsAlive = len(p.Peers)



}

func EventNewLocalOrder(button_pressed elevio.ButtonEvent, timerReset chan bool){
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event New Local Order in state: %s \n", stateString)

	queue.UpdateLocalQueue(localL, button_pressed)

	switch elevator1.state {
	case idle:
		elevator1.curr_dir = GetDirection(elevator1.curr_floor, button_pressed.Floor)
		elevio.SetMotorDirection(elevator1.curr_dir)

		if shouldStop(elevator1.curr_floor, elevator1.curr_dir) {
			elevator1.curr_dir = elevio.MD_Stop
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.state = doorOpen
			timerReset <- true
		}else{
			elevator1.state = moving
		}
	case doorOpen:
		if(shouldStop(elevator1.curr_floor, elevator1.curr_dir)){
			fmt.Println("I shouldStop Door open")
			timerReset <- true
		}

	case moving:
		fmt.Println("Moving\n")
	}

}


func EventNewRemoteOrder( button_pressed elevio.ButtonEvent, transmitt chan UDPmsg,elevMap ,id int){
	var msg assigner.UDPMsg
	msg.MsgID = 2
	msg.ElevID = id
	msg.Order = button_pressed
	transmitt <- msg

}

func EventFloorReached(floor int, timerReset chan bool){
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Floor Reached in state %s \n",stateString)
	elevator1.curr_floor = floor
	//Lights

	switch elevator1.state {
	case moving:
		if(shouldStop(floor, elevator1.curr_dir)){
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



func EventDoorTimeOut(timerReset chan bool){
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

func shouldStop(floorSensor int, dir elevio.MotorDirection) bool{
	for k:= localL.Front(); k != nil; k = k.Next(){

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

func StartBroadcast(port string){
	elevio.Init("localhost:"+port, 4)
	conn.DialBroadcastUDP(15675)
}
