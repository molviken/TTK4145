package eventHandler

import (
	"fmt"
	"strconv"

	elevFunc "../elevFunc"
	//bcast ".././network/bcast"
	"container/list"

	assigner "../ElevAssigner"
	elevio "../elevio"
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
var LocalL = list.New()
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
	obstr 	   bool 
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
	elevator1.obstr = false
	queue.ReadBackup(LocalL)
	if LocalL.Front() != nil {
		elevio.SetMotorDirection(elevFunc.GetDirection(elevator1.curr_floor, LocalL.Front().Value.(*elevio.ButtonEvent).Floor))
		elevator1.state = moving
	}else{
		elevator1.state = idle
	}

}
func GetElevatorState() int{
	return elevator1.state
}

func HandleEvents(button chan elevio.ButtonEvent, floorSensor chan int, obstr chan bool, stop chan bool, timeOut chan bool, timerReset chan bool,
	receive chan assigner.UDPmsg, transmitt chan assigner.UDPmsg, peerUpdateCh chan peers.PeerUpdate, id int, updateRemote chan elevio.ButtonEvent, obstrTimerReset chan bool) {
	select {
	case button_pressed := <-button:

		if button_pressed.Button == elevio.BT_Cab && !elevFunc.DuplicateOrder(button_pressed, LocalL) {
			EventNewLocalOrder(button_pressed, timerReset, id, transmitt, obstrTimerReset)
			elevFunc.PrintList(LocalL.Front())

		} else if button_pressed.Button != elevio.BT_Cab && !queue.DuplicateOrderRemote(button_pressed) {
			EventNewRemoteOrder(button_pressed, transmitt, id, 2)
		}

	case floor := <-floorSensor:
		elevator1.obstr = false
		EventFloorReached(floor, timerReset, id, transmitt, obstrTimerReset)
		if(LocalL.Front()!=nil){
			obstrTimerReset <- true
		}
	case <-timeOut:
		EventDoorTimeOut(timerReset, id, transmitt, obstrTimerReset)

	case peers := <-peerUpdateCh:
		EventPeerUpdate(peers, transmitt)
		if (len(peers.Lost)!=0) {EventLostPeer(transmitt, id, peers, timerReset, obstrTimerReset)}
		//if (peers.New != "") {}

	case msg := <-receive:
		EventReceivedMessage(msg, peersOnline, elevator1.curr_floor, id, transmitt, timerReset, obstrTimerReset)

	case obstruction := <- obstr:
		fmt.Println("obstruction", obstruction)
		EventObstruction(obstruction, id, transmitt, timerReset, obstrTimerReset)

	case stopBt := <- stop:
		EventStop(stopBt, transmitt, timerReset, id)
	}
}

func EventStop(stop bool, transmitt chan assigner.UDPmsg, timerReset chan bool, elevId int){
	if (stop == true){
		elevator1.curr_dir = elevio.MD_Stop
		for k, v := range queue.RemoteOrders{
			if (v == elevId && v != 0){
				assigner.TransmittUDP(5, elevId, 0, k, transmitt)
			}
		}
	}
}

func EventObstruction(obstr bool, elevId int, transmitt chan assigner.UDPmsg, timerReset chan bool, obstrTimerReset chan bool){
	elevator1.obstr = true //Obstruction/stop state, all cost=0 in this state 
	prevDir := elevator1.curr_dir
	switch obstr{
	case true:
		
		if(LocalL.Front() != nil){
		//Transmitt all remote orders to other elevators
		for k, v := range queue.RemoteOrders{
			if (v == elevId && v != 0){
				assigner.TransmittUDP(5, elevId, 0, k, transmitt)
			}
		}
		//Remove these orders from the queue
		for k := LocalL.Front(); k != nil; k = k.Next(){
			if (k.Value.(*elevio.ButtonEvent).Button != elevio.BT_Cab){
				LocalL.Remove(k)
			}
		}
	}
		//Give the elevator a default local order 
		if(LocalL.Front() == nil){
			var defaultOrder elevio.ButtonEvent
			defaultOrder.Button = elevio.BT_Cab
			defaultOrder.Floor = 0
			EventNewLocalOrder(defaultOrder, timerReset, elevId, transmitt, obstrTimerReset)
			
		}

	case false:
		elevator1.curr_dir = prevDir
		elevio.SetMotorDirection(prevDir)
		elevator1.obstr = false

	}

	//fmt.Println("Dir etter obstr: ", elevator1.curr_dir)

}

func EventReceivedMessage(msg assigner.UDPmsg, peers peers.PeerUpdate, floor int, id int, transmitt chan assigner.UDPmsg, timerReset chan bool, obstrTimerReset chan bool) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Received Message in state: %s \n", stateString)
	switch msg.MsgID {
	case 1:
		assigner.ChooseElevator(msg, peers, transmitt)
	case 2:
		cost := queue.Cost(msg.Order, floor, elevator1.curr_dir)
		
		if elevator1.obstr {
			cost = 0
		}
		/*Checking if the new order is in the same direction (buttonType) as the current order. If not the cost is decreased by 2*/

		if (LocalL.Front() != nil && LocalL.Back().Value.(*elevio.ButtonEvent).Button != elevio.BT_Cab && LocalL.Back().Value.(*elevio.ButtonEvent).Button != msg.Order.Button && LocalL.Back().Value.(*elevio.ButtonEvent).Floor == msg.Order.Floor && len(peersOnline.Peers) != 1){
			cost = 0
		}
		assigner.TransmittUDP(1, id, cost, msg.Order, transmitt)
	case 3:
		if msg.Message == id{
			if (msg.Order.Floor == floor && elevator1.curr_dir == elevio.MD_Stop ){
				timerReset <- true
				assigner.TransmittUDP(4, id, 0, msg.Order, transmitt)
			} else {
				if(!elevFunc.DuplicateOrder(msg.Order, LocalL)){
					EventNewLocalOrder(msg.Order, timerReset, id, transmitt, obstrTimerReset)
				}

			}
		}

		queue.AddRemoteOrder(msg.ElevID, msg.Order)
		elevFunc.SyncButtonLights(LocalL)
		queue.PrintMap()
	case 4:
		queue.RemoveRemoteOrder(msg.Order)
		elevFunc.SyncButtonLights(LocalL)

	case 5:
		var cost int
		fmt.Println("Obstruction")
		if msg.ElevID == id {
			cost = 0
		}else{
			cost = queue.Cost(msg.Order, floor, elevator1.curr_dir)

		}
		assigner.TransmittUDP(1, id, cost, msg.Order, transmitt)
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

func EventNewLocalOrder(button_pressed elevio.ButtonEvent, timerReset chan bool, elevId int, transmitt chan assigner.UDPmsg, obstrTimerReset chan bool) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event New Local Order in state: %s \n", stateString)

	queue.AddLocalOrder(LocalL, button_pressed)
	elevFunc.SyncButtonLights(LocalL)
	queue.UpdateBackup(LocalL)
	//queue.ReadBackup()
	switch elevator1.state {
	case idle:
		obstrTimerReset <- true
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


	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event New Remote Order in state: %s \n", stateString)

	var msg assigner.UDPmsg
	msg.MsgID = msgId
	msg.ElevID = elevId
	msg.Order = button_pressed
	transmitt <- msg

}

func EventFloorReached(floor int, timerReset chan bool, id int, transmitt chan assigner.UDPmsg, obstrTimerReset chan bool) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Floor Reached in state %s \n", stateString)
	elevator1.curr_floor = floor

	if (elevator1.obstr == true){
		elevator1.obstr = false
		if (LocalL.Front() != nil){
			elevio.SetMotorDirection(elevFunc.GetDirection(elevator1.curr_floor, LocalL.Front().Value.(*elevio.ButtonEvent).Floor))	
			elevator1.state = moving
		}else{
			elevator1.state = idle
		}

	}

	
	if LocalL.Front() != nil && LocalL.Front().Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab {
		isCab = true
	} else {
		isCab = false
	}
	
	switch elevator1.state {
	case moving:
		obstrTimerReset <- true //reset obstruction timer
		if LocalL.Front().Value.(*elevio.ButtonEvent).Floor == floor {
			if !isCab {
				tempButton.Button = LocalL.Front().Value.(*elevio.ButtonEvent).Button
				tempButton.Floor = floor
				queue.RemoveRemoteOrder(tempButton)
				
				assigner.TransmittUDP(4, id, 0, tempButton, transmitt)
			}
			LocalL.Remove(LocalL.Front())
			if LocalL.Front() != nil {
				queue.ScanForDouble(elevator1.curr_dir, elevator1.curr_floor, LocalL, id, transmitt, isCab)
			}

			elevator1.curr_dir = elevio.MD_Stop
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevFunc.SyncButtonLights(LocalL)
			elevator1.state = doorOpen
			timerReset <- true

		} else if shouldStop(floor, elevator1.curr_dir, id, transmitt) {
			if LocalL.Front() != nil {
				queue.ScanForDouble(elevator1.curr_dir, elevator1.curr_floor, LocalL, id, transmitt, isCab)
			}
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator1.curr_dir = elevio.MD_Stop
			elevator1.state = doorOpen
			elevFunc.SyncButtonLights(LocalL)
			timerReset <- true
		}
	}
}

func EventDoorTimeOut(timerReset chan bool, elevId int, transmitt chan assigner.UDPmsg, obstrTimerReset chan bool) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Door TimeOut in state %s \n", stateString)
	switch elevator1.state {
	case doorOpen:
		elevio.SetDoorOpenLamp(false) //turn of light
		elevFunc.PrintList(LocalL.Front())
		//Ready for new direction

		if LocalL.Front() != nil {
			elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, LocalL.Front().Value.(*elevio.ButtonEvent).Floor)
			elevio.SetMotorDirection(elevator1.curr_dir)
			elevator1.state = moving
			obstrTimerReset <- true
			 //start obstructiontimer after door has been open

			/*Check for pressing button several times while door is open*/
		} else{
			elevator1.state = idle

		}
	}
	stateString = elevFunc.StateToString(elevator1.state)
	fmt.Printf("Elevator state after timeout is : %s \n", stateString)

	queue.PrintMap()
}


func EventLostPeer( transmitt chan assigner.UDPmsg, elevId int, peerStatus peers.PeerUpdate, timerReset chan bool, obstrTimerReset chan bool) {
	fmt.Println("Peer: lost! Remote orders are being divided")


	peersAlive := len(peerStatus.Peers)
	switch {
	case peersAlive == 1:
		fmt.Println("All other peers lost, have to take all remote orders")
		for k, v := range queue.RemoteOrders{
			if (v != elevId && v != 0){
				queue.RemoteOrders[k] = elevId
				EventNewLocalOrder(k, timerReset, elevId, transmitt, obstrTimerReset)
			}

		}
	case peersAlive != 1:
		var lostUDP assigner.UDPmsg
		lostUDP.Message = 0
		remainingId1, _ := strconv.Atoi(peerStatus.Peers[0])
		remainingId2, _ := strconv.Atoi(peerStatus.Peers[1])
		fmt.Println("rem1: ", remainingId1)
		fmt.Println("rem2: ", remainingId2)
		lostId, _ 		:= strconv.Atoi(peerStatus.Lost[0])
		assigner.CostMap[lostId] = lostUDP
		if( (remainingId1 == elevId && elevId > remainingId2) || (remainingId2 == elevId && elevId > remainingId1) ){

			fmt.Println("One peer lost: ", lostId)
			for k, v := range queue.RemoteOrders{
				if (v == lostId){
					queue.RemoteOrders[k] = 0
					EventNewRemoteOrder(k, transmitt, elevId, 2)
					fmt.Println("sender du ut?")
				}
			}
		}
	}
}

func shouldStop(floorSensor int, dir elevio.MotorDirection, elevId int, transmitt chan assigner.UDPmsg) bool {
	fmt.Println("SHouldStop = TRUE")

	for k := LocalL.Front(); k != nil; k = k.Next() {

		if k.Value.(*elevio.ButtonEvent).Floor == floorSensor {
			switch {
			case k.Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab:
				LocalL.Remove(k)
				queue.UpdateBackup(LocalL)
				isCab = true
				fmt.Println("SHouldStop = TRUE")
				return true
			case dir == elevio.MD_Up:

				if k.Value.(*elevio.ButtonEvent).Button != 1 {
					tempButton.Button = k.Value.(*elevio.ButtonEvent).Button
					tempButton.Floor = floorSensor
					isCab = true
					assigner.TransmittUDP(4, elevId, 0, tempButton, transmitt) //telling peers to delete order

					LocalL.Remove(k) //remove from local queue
					queue.UpdateBackup(LocalL)

					queue.RemoveRemoteOrder(tempButton)
					fmt.Println("SHouldStop = TRUE")
					return true
				}
			case dir == elevio.MD_Down:
				if k.Value.(*elevio.ButtonEvent).Button != 0 {

					LocalL.Remove(k) //remove from local queue
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



