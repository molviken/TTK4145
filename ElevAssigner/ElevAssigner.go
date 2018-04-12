package ElevAssigner

import (
	"../elevio"
	//"../queue"
	"fmt"

	"../network/peers"
)

var num_received = 0

type UDPmsg struct {
	MsgID  int
	ElevID int
	Cost   int
	Order  elevio.ButtonEvent
}

type CostReply struct {
	Id    int
	Cost  int
	Order elevio.ButtonEvent
}

type OrderMsg struct {
	Id    int
	Order elevio.ButtonEvent
}

var costMap map[int]UDPmsg

var shouldInit = true

func ChooseElevator(msg UDPmsg, elevMap peers.PeerUpdate, transmitt chan UDPmsg) {

	if shouldInit {
		costMap = make(map[int]UDPmsg)
		shouldInit = false
		fmt.Println("Costmap initialized")
	}
	costMap[msg.ElevID] = msg
	var numOnline = len(elevMap.Peers)
	num_received += 1
	if num_received == numOnline {
		PrintCostMap()
		var highestCostMsg = findHighestCost(costMap)
		num_received = 0
		highestCostMsg.Cost = highestCostMsg.ElevID
		fmt.Println("Vinner ID: ", highestCostMsg.Cost)
		highestCostMsg.MsgID = 3
		transmitt <- highestCostMsg
	}
}

func findHighestCost(costMap map[int]UDPmsg) UDPmsg {
	highestCost := 0
	var highestMsg UDPmsg
	for _, costMsg := range costMap {
		if costMsg.Cost > highestCost {
			highestCost = costMsg.Cost
			highestMsg = costMsg
		} else if costMsg.Cost == highestCost { //Dersom alle har samme kost
			if costMsg.ElevID < highestMsg.ElevID {
				highestMsg = costMsg
			}
		}
	}
	return highestMsg
}
func PrintCostMap() {
	fmt.Println(" ")
	fmt.Println("Costmap:")
	for val, key := range costMap {
		fmt.Print(val)
		fmt.Println(key)
	}
	fmt.Println(" ")
}

func TransmittUDP(msgID int, elevID int, cost int, order elevio.ButtonEvent, transmitt chan UDPmsg) {
	//var msg UDPmsg
	msg := UDPmsg{msgID, elevID, cost, order}
	transmitt <- msg
}

func LostPeer(id int, transmitt chan UDPmsg) {
	var msg UDPmsg
	msg.ElevID = 0
	msg.MsgID = 2
	msg.Cost = 0
	for k, v := range costMap {
		if k == id {
			msg.Order = v.Order
			transmitt <- msg
			//fmt.Println("Hva blir transmitta:  ", msg)
		}
	}
}
