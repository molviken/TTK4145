package eventHandler

import (
	"fmt"
	"strconv"

	elevFunc "../elevFunc"
	//bcast ".././network/bcast"
	"container/list"

	assigner "../ElevAssigner"
	elevio "../elevio"
	conn "../network/conn"
	peers "../network/peers"
	queue "../queue"
)

/*
States:
idle - At a floor with closed door, awaiting orders
moving - Moving and can be between floors or going past a floor
doorOpen - At a floor with the door open
*/
var peersOnline peers.PeerUpdate
var localL = list.New()
var isCab = false

const (
	idle = iota
	moving
	doorOpen
)

//Elevator Struct to keep track on current floor, current direction, and state
type elevator struct {
	curr_floor int
	curr_dir   elevio.MotorDirection
	state      int
}

var elevator1 elevator
var tempButton elevio.ButtonEvent

/*
NewOrder: trggered when there is a new order added to the queue
FloorReached: when a floor is reached
DoorTimout: When the door has been open for three seconds in state openDoor, handle what happens next
ShouldStop: At each floor we check if there is an order to take if it is convinient
*/



func EventHandlerInit(startFloor int, elevId int) {
	elevator1.curr_floor = startFloor
	elevator1.curr_dir = elevio.MD_Stop
	elevator1.state = idle
}

func HandleEvents(button chan elevio.ButtonEvent, floorSensor chan int, obstr chan bool, stop chan bool, timeOut chan bool, timerReset chan bool,
	receive chan assigner.UDPmsg, transmitt chan assigner.UDPmsg, peerUpdateCh chan peers.PeerUpdate, id int, updateRemote chan elevio.ButtonEvent) {
	select {
	case button_pressed := <-button:

		if button_pressed.Button == elevio.BT_Cab && !elevFunc.DuplicateOrder(button_pressed, localL) {
			EventNewLocalOrder(button_pressed, timerReset, id, transmitt)
			elevFunc.PrintList(localL.Front())

		} else if button_pressed.Button != elevio.BT_Cab && !queue.DuplicateOrderRemote(button_pressed) {
			EventNewRemoteOrder(button_pressed, transmitt, id, 2)
		}

	case floor := <-floorSensor:
		EventFloorReached(floor, timerReset, id, transmitt)

		//case obstr := <- channel.obstr:

		//case stop := <- channel.stop:
	case <-timeOut:
		EventDoorTimeOut(timerReset, id, transmitt)

	case peers := <-peerUpdateCh:
		EventPeerUpdate(peers, transmitt)
		if (len(peers.Lost)!=0) {EventLostPeer(transmitt, id, peers, timerReset)}
		//if (peers.New != "") {}

	case msg := <-receive:
		EventReceivedMessage(msg, peersOnline, elevator1.curr_floor, id, transmitt, timerReset)
	}
}

func EventReceivedMessage(msg assigner.UDPmsg, peers peers.PeerUpdate, floor int, id int, transmitt chan assigner.UDPmsg, timerReset chan bool) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Received Message in state: %s \n", stateString)
	switch msg.MsgID {
	case 1:
		assigner.ChooseElevator(msg, peers, transmitt)
	case 2:
		cost := queue.Cost(msg.Order, floor, elevator1.curr_dir)
		/*Checking if the new order is in the same direction (buttonType) as the current order. If not the cost is decreased by 2*/
		if localL.Front() != nil && localL.Front().Value.(*elevio.ButtonEvent).Button != elevio.BT_Cab && localL.Front().Value.(*elevio.ButtonEvent).Button != msg.Order.Button {
			cost -= 2
		}
		assigner.TransmittUDP(1, id, cost, msg.Order, transmitt)
	case 3:
		if msg.Message == id && !queue.DuplicateOrderRemote(msg.Order) {
			if msg.Order.Floor == floor {
				timerReset <- true
				assigner.TransmittUDP(4, id, 0, msg.Order, transmitt)
			} else {
				EventNewLocalOrder(msg.Order, timerReset, id, transmitt)
			}
		}
		queue.AddRemoteOrder(msg.ElevID, msg.Order)
	case 4:
		queue.RemoveRemoteOrder(msg.Order)
	}
}

func EventPeerUpdate(p peers.PeerUpdate, transmitt chan assigner.UDPmsg) {
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New peer:      %q\n", p.New)
	fmt.Printf("  Lost peer:     %q\n", p.Lost)
	peersOnline = p
	/*
	if len(p.Lost) != 0 {
		id, _ := strconv.Atoi(p.Lost[0])
		fmt.Println("Peer id: ", id)
		assigner.LostPeer(id, transmitt)
	}*/

}

func EventNewLocalOrder(button_pressed elevio.ButtonEvent, timerReset chan bool, elevId int, transmitt chan assigner.UDPmsg) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event New Local Order in state: %s \n", stateString)

	queue.AddLocalOrder(localL, button_pressed)

	switch elevator1.state {
	case idle:
		elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, button_pressed.Floor)
		elevio.SetMotorDirection(elevator1.curr_dir)

		if shouldStop(elevator1.curr_floor, elevator1.curr_dir, elevId, transmitt) {
			elevator1.curr_dir = elevio.MD_Stop
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.state = doorOpen
			timerReset <- true
		} else {
			elevator1.state = moving
		}
	case doorOpen:
		if shouldStop(elevator1.curr_floor, elevator1.curr_dir, elevId, transmitt) {
			timerReset <- true
		}

	case moving:
		fmt.Println("Moving\n")
	}

}

func EventNewRemoteOrder(button_pressed elevio.ButtonEvent, transmitt chan assigner.UDPmsg, elevId int, msgId int) {

	//stateString := elevFunc.StateToString(elevator1.state)
	//fmt.Printf("Event New Remote Order in state: %s \n", stateString)

	var msg assigner.UDPmsg
	msg.MsgID = msgId
	msg.ElevID = elevId
	msg.Order = button_pressed
	transmitt <- msg

}

func EventFloorReached(floor int, timerReset chan bool, id int, transmitt chan assigner.UDPmsg) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Floor Reached in state %s \n", stateString)
	elevator1.curr_floor = floor

	/*Must do this to be able to remove later orders at the same floor in ScanForDouble*/
	if localL.Front() != nil && localL.Front().Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab {
		isCab = true
	} else {
		isCab = false
	}
	/**/
	
	switch elevator1.state {

	//Eneste problem nå er at dersom en hallup og halldown første og neste i lista blir den stuck, burde ikke få begge deler i lista egentlig men det skjer..
	case moving:
		if localL.Front().Value.(*elevio.ButtonEvent).Floor == floor {
			if !isCab {
				tempButton.Button = localL.Front().Value.(*elevio.ButtonEvent).Button
				tempButton.Floor = floor
				queue.RemoveRemoteOrder(tempButton)
				assigner.TransmittUDP(4, id, 0, tempButton, transmitt)
			}
			localL.Remove(localL.Front())
			if localL.Front() != nil {
				queue.ScanForDouble(elevator1.curr_dir, elevator1.curr_floor, localL, id, transmitt, isCab)
			}
			
			elevator1.curr_dir = elevio.MD_Stop
			elevio.SetMotorDirection(elevio.MD_Stop)

			elevator1.state = doorOpen
			timerReset <- true

		} else if shouldStop(floor, elevator1.curr_dir, id, transmitt) {
			if localL.Front() != nil {
				queue.ScanForDouble(elevator1.curr_dir, elevator1.curr_floor, localL, id, transmitt, isCab)
			}
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.curr_dir = elevio.MD_Stop
			elevator1.state = doorOpen
			timerReset <- true
		}
	}
}

func EventDoorTimeOut(timerReset chan bool, elevId int, transmitt chan assigner.UDPmsg) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Door TimeOut in state %s \n", stateString)
	switch elevator1.state {
	case doorOpen:
		elevio.SetDoorOpenLamp(false) //turn of light
		elevFunc.PrintList(localL.Front())
		//Ready for new direction

		if localL.Front() != nil {
			elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, localL.Front().Value.(*elevio.ButtonEvent).Floor)
			elevio.SetMotorDirection(elevator1.curr_dir)
			elevator1.state = moving

			/*Check for pressing button several times while door is open*/
		} else {
			elevator1.state = idle
		}
	}
	stateString = elevFunc.StateToString(elevator1.state)
	fmt.Printf("Elevator state after timeout is : %s \n", stateString)
	queue.PrintMap()


}

func shouldStop(floorSensor int, dir elevio.MotorDirection, elevId int, transmitt chan assigner.UDPmsg) bool {
	fmt.Println("I shouldStop")

	for k := localL.Front(); k != nil; k = k.Next() {

		if k.Value.(*elevio.ButtonEvent).Floor == floorSensor {
			switch {
			case k.Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab:
				localL.Remove(k)
				isCab = true
				return true
			case dir == elevio.MD_Up:

				if k.Value.(*elevio.ButtonEvent).Button != 1 {
					tempButton.Button = k.Value.(*elevio.ButtonEvent).Button
					tempButton.Floor = floorSensor
					isCab = true
					assigner.TransmittUDP(4, elevId, 0, tempButton, transmitt) //telling peers to delete order

					localL.Remove(k) //remove from local queue

					
					queue.RemoveRemoteOrder(tempButton)

					return true
				}
			case dir == elevio.MD_Down:
				if k.Value.(*elevio.ButtonEvent).Button != 0 {

					localL.Remove(k) //remove from local queue
					tempButton.Button = k.Value.(*elevio.ButtonEvent).Button
					tempButton.Floor = floorSensor

					assigner.TransmittUDP(4, elevId, 0, tempButton, transmitt)
					queue.RemoveRemoteOrder(tempButton) //Remove from remote queue
					return true

				}
			}
		}
	}
	return false
}

func EventLostPeer( transmitt chan assigner.UDPmsg, elevId int, peerStatus peers.PeerUpdate, timerReset chan bool) {
	fmt.Println("Peer: lost! Remote orders are being divided")
	
	peersAlive := len(peerStatus.Peers)
	switch {
	case peersAlive == 1:
		fmt.Println("All other peers lost, have to take all remote orders")
		for k, v := range queue.RemoteOrders{
			if (v != elevId && v != 0){
				EventNewLocalOrder(k, timerReset, elevId, transmitt)
			}

		}
	case peersAlive != 1:
		remainingId1, _ := strconv.Atoi(peerStatus.Peers[0])
		remainingId2, _ := strconv.Atoi(peerStatus.Peers[1])
		lostId, _ 		:= strconv.Atoi(peerStatus.Lost[0])
		if( (remainingId1 == elevId && elevId > remainingId2) || (remainingId2 == elevId && elevId > remainingId1) ){

			fmt.Println("One peer lost: ", lostId)
			for k, v := range queue.RemoteOrders{
				if (v == lostId){
					EventNewRemoteOrder(k, transmitt, elevId, 2)
					fmt.Println("sender du ut?")
				}
			}
		}
	}
}


func EventPeerBackOnline(){

}

/*Burde ha denne et annet sted*/
func StartBroadcast(port string) {
	elevio.Init("localhost:"+port, 4)
	conn.DialBroadcastUDP(15675)
}
