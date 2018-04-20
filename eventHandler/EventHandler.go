package eventHandler

import (
	"fmt"
	"strconv"
	elevFunc "../elevFunc"
	"container/list"
	assigner "../ElevAssigner"
	elevio "../elevio"
	peers "../network/peers"
	queue "../queue"
)

var peersOnline peers.PeerUpdate
var LocalL = list.New()
var isCab = false
var elevator1 elevator
var tempButton elevio.ButtonEvent

const (
	idle = iota
	moving
	doorOpen
)

type elevator struct {
	curr_floor int
	curr_dir   elevio.MotorDirection
	state      int
	obstr 	   bool 
}



func EventHandlerInit(startFloor int, elevId int) {
	elevator1.curr_floor = startFloor
	elevator1.curr_dir = elevio.MD_Stop
	elevator1.obstr = false
	//If power shortage, all Cab orders are saved on backup file and will be read on startup.
	queue.ReadBackup(LocalL)
	if LocalL.Front() != nil {
		elevio.SetMotorDirection(elevFunc.GetDirection(elevator1.curr_floor, LocalL.Front().Value.(*elevio.ButtonEvent).Floor))
		elevator1.state = moving
	}else{
		elevator1.state = idle
	}
	elevFunc.SyncButtonLights(LocalL)

}
func HandleEvents(button chan elevio.ButtonEvent, floorSensor chan int, obstr chan bool, timeOut chan bool, timerReset chan bool,
	receive chan assigner.UDPmsg, transmitt chan assigner.UDPmsg, peerUpdateCh chan peers.PeerUpdate, id int, obstrTimerReset chan bool) {
	select {
	case button_pressed := <-button:
		if button_pressed.Button == elevio.BT_Cab && !elevFunc.DuplicateOrder(button_pressed, LocalL) {
			eventNewLocalOrder(button_pressed, timerReset, id, transmitt, obstrTimerReset)
			elevFunc.PrintList(LocalL.Front())

		} else if button_pressed.Button != elevio.BT_Cab && !queue.DuplicateOrderRemote(button_pressed) {
			eventNewRemoteOrder(button_pressed, transmitt, id, 2,  timerReset, obstrTimerReset)
		}

	case floor := <-floorSensor:
		elevio.SetFloorIndicator(floor)
		eventFloorReached(floor, timerReset, id, transmitt, obstrTimerReset)
		if(LocalL.Front()!=nil){
			obstrTimerReset <- true
		}

	case <-timeOut:
		eventDoorTimeOut(timerReset, id, transmitt, obstrTimerReset)

	case peers := <-peerUpdateCh:
		eventPeerUpdate(peers, transmitt)
		if (len(peers.Lost)!=0) {eventLostPeer(transmitt, id, peers, timerReset, obstrTimerReset)}

	case msg := <-receive:
		eventReceivedMessage(msg, peersOnline, elevator1.curr_floor, id, transmitt, timerReset, obstrTimerReset)

	case obstruction := <- obstr:
		fmt.Println("Obstruction: ", obstruction)
		eventObstruction(obstruction, id, transmitt, timerReset, obstrTimerReset)
	}
}
func eventObstruction(obstr bool, elevId int, transmitt chan assigner.UDPmsg, timerReset chan bool, obstrTimerReset chan bool){
	elevator1.obstr = true 
		
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
}

func eventReceivedMessage(msg assigner.UDPmsg, peers peers.PeerUpdate, floor int, id int, transmitt chan assigner.UDPmsg, timerReset chan bool, obstrTimerReset chan bool) {
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
		if (LocalL.Front() != nil && LocalL.Back().Value.(*elevio.ButtonEvent).Button != elevio.BT_Cab && LocalL.Back().Value.(*elevio.ButtonEvent).Button != msg.Order.Button && LocalL.Back().Value.(*elevio.ButtonEvent).Floor == msg.Order.Floor && len(peersOnline.Peers) != 1){
			cost = 1
		}
		assigner.TransmittUDP(1, id, cost, msg.Order, transmitt)
	case 3:
		if msg.Message == id{
			if (msg.Order.Floor == floor && elevator1.curr_dir == elevio.MD_Stop ){
				timerReset <- true
				assigner.TransmittUDP(4, id, 0, msg.Order, transmitt)
			} else {
				if(!elevFunc.DuplicateOrder(msg.Order, LocalL)){
					eventNewLocalOrder(msg.Order, timerReset, id, transmitt, obstrTimerReset)
				}

			}
		}
		queue.AddRemoteOrder(msg.ElevID, msg.Order)
		elevFunc.SyncButtonLights(LocalL)
	case 4:
		queue.RemoveRemoteOrder(msg.Order)
		elevFunc.SyncButtonLights(LocalL)

	case 5:
		var cost int
		if msg.ElevID == id {
			cost = 0
		}else{
			cost = queue.Cost(msg.Order, floor, elevator1.curr_dir)
		}
		assigner.TransmittUDP(1, id, cost, msg.Order, transmitt)
	}
}

func eventPeerUpdate(p peers.PeerUpdate, transmitt chan assigner.UDPmsg) {
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New peer:      %q\n", p.New)
	fmt.Printf("  Lost peer:     %q\n", p.Lost)
	peersOnline = p
}

func eventNewLocalOrder(button_pressed elevio.ButtonEvent, timerReset chan bool, elevId int, transmitt chan assigner.UDPmsg, obstrTimerReset chan bool) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event New Local Order in state: %s \n", stateString)
	queue.AddLocalOrder(LocalL, button_pressed)
	elevFunc.SyncButtonLights(LocalL)
	queue.UpdateBackup(LocalL)

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

func eventNewRemoteOrder(button_pressed elevio.ButtonEvent, transmitt chan assigner.UDPmsg, elevId int, msgId int, timerReset chan bool, obstrTimerReset chan bool) {
	if(len(peersOnline.Peers) == 1){
		if (button_pressed.Floor == elevator1.curr_floor && elevator1.curr_dir == elevio.MD_Stop){
			timerReset <- true
			}else{
				queue.AddRemoteOrder(elevId, button_pressed)
				eventNewLocalOrder(button_pressed, timerReset, elevId, transmitt, obstrTimerReset)
			}
	}else{
		stateString := elevFunc.StateToString(elevator1.state)
		fmt.Printf("Event New Remote Order in state: %s \n", stateString)
		assigner.TransmittUDP(msgId, elevId, 0, button_pressed, transmitt)
	}
}

func eventFloorReached(floor int, timerReset chan bool, id int, transmitt chan assigner.UDPmsg, obstrTimerReset chan bool) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Printf("Event Floor Reached in state %s ", stateString)
	elevator1.curr_floor = floor

	if (elevator1.obstr == true){
		elevator1.obstr = false
		if (LocalL.Front() != nil){
			elevio.SetMotorDirection(elevFunc.GetDirection(elevator1.curr_floor, LocalL.Front().Value.(*elevio.ButtonEvent).Floor))	
			elevator1.state = moving
		}else{
			elevio.SetMotorDirection(elevio.MD_Stop)
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
		obstrTimerReset <- true
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

func eventDoorTimeOut(timerReset chan bool, elevId int, transmitt chan assigner.UDPmsg, obstrTimerReset chan bool) {
	stateString := elevFunc.StateToString(elevator1.state)
	fmt.Println("Event Door TimeOut in state: ", stateString)
	switch elevator1.state {
	case doorOpen:
		elevio.SetDoorOpenLamp(false)
		elevFunc.PrintList(LocalL.Front())
		if LocalL.Front() != nil {
			elevator1.curr_dir = elevFunc.GetDirection(elevator1.curr_floor, LocalL.Front().Value.(*elevio.ButtonEvent).Floor)
			elevio.SetMotorDirection(elevator1.curr_dir)
			elevator1.state = moving
			obstrTimerReset <- true
		} else{
			elevator1.state = idle
		}
	}
	stateString = elevFunc.StateToString(elevator1.state)
	fmt.Println("Elevator state after timeout is : ", stateString)
}


func eventLostPeer( transmitt chan assigner.UDPmsg, elevId int, peerStatus peers.PeerUpdate, timerReset chan bool, obstrTimerReset chan bool) {
	fmt.Println("Peer lost! Remote orders are being divided")
	peersAlive := len(peerStatus.Peers)

	switch {
	case peersAlive == 1:
		fmt.Println("All other peers lost, have to take all remote orders")
		for k, v := range queue.RemoteOrders{
			if (v != elevId && v != 0){
				queue.RemoteOrders[k] = elevId
				eventNewLocalOrder(k, timerReset, elevId, transmitt, obstrTimerReset)
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
			for k, v := range queue.RemoteOrders{
				if (v == lostId){
					queue.RemoteOrders[k] = 0
					eventNewRemoteOrder(k, transmitt, elevId, 2,  timerReset, obstrTimerReset)
				}
			}
		}
	}
}

func shouldStop(floorSensor int, dir elevio.MotorDirection, elevId int, transmitt chan assigner.UDPmsg) bool {
	for k := LocalL.Front(); k != nil; k = k.Next() {
		if k.Value.(*elevio.ButtonEvent).Floor == floorSensor {
			switch {
			case k.Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab:
				LocalL.Remove(k)
				queue.UpdateBackup(LocalL)
				isCab = true
				return true
			case dir == elevio.MD_Up:
				if k.Value.(*elevio.ButtonEvent).Button != 1 {
					tempButton.Button = k.Value.(*elevio.ButtonEvent).Button
					tempButton.Floor = floorSensor
					assigner.TransmittUDP(4, elevId, 0, tempButton, transmitt)
					LocalL.Remove(k)
					queue.UpdateBackup(LocalL)
					queue.RemoveRemoteOrder(tempButton)
					return true
				}
			case dir == elevio.MD_Down:
				if k.Value.(*elevio.ButtonEvent).Button != 0 {
					tempButton.Button = k.Value.(*elevio.ButtonEvent).Button
					tempButton.Floor = floorSensor
					LocalL.Remove(k)
					assigner.TransmittUDP(4, elevId, 0, tempButton, transmitt)
					queue.RemoveRemoteOrder(tempButton)
					return true

				}
			}
		}
	}
	return false
}



