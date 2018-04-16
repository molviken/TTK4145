package ElevAssigner

import (
	"../elevio"
	"fmt"

	"../network/peers"
)

var num_received = 0
var CostMap map[int]UDPmsg
var shouldInit = true //For initiliazing the cost map

type UDPmsg struct {
	MsgID  int
	ElevID int
	Message   int //Message is used to transmitt the cost or winner ID over UDP
	Order  elevio.ButtonEvent
}



func ChooseElevator(msg UDPmsg, elevMap peers.PeerUpdate, transmitt chan UDPmsg) {
	num_received += 1
	if shouldInit {
		CostMap = make(map[int]UDPmsg)
		shouldInit = false
		fmt.Println("Costmap initialized")
	}
	CostMap[msg.ElevID] = msg
	var numOnline = len(elevMap.Peers)
	if num_received == numOnline {
		PrintCostMap()
		var highestCostMsg = findHighestCost(CostMap)
		num_received = 0
		highestCostMsg.Message = highestCostMsg.ElevID
		highestCostMsg.MsgID = 3
		transmitt <- highestCostMsg
	}
}
func findHighestCost(costMap map[int]UDPmsg) UDPmsg {
	highestCost := 0
	var highestMsg UDPmsg
	for _, CostMsg := range costMap {
		if CostMsg.Message > highestCost {
			highestCost = CostMsg.Message
			highestMsg = CostMsg
		} else if CostMsg.Message == highestCost {
			if CostMsg.ElevID < highestMsg.ElevID {
				highestMsg = CostMsg
			}
		}
	}
	return highestMsg
}
func PrintCostMap() {
	fmt.Println(" ")
	fmt.Println("Costmap:")
	for val, key := range CostMap {
		fmt.Print(val)
		fmt.Println(key)
	}
	fmt.Println(" ")
}
func TransmittUDP(msgID int, elevID int, message int, order elevio.ButtonEvent, transmitt chan UDPmsg) {
	msg := UDPmsg{msgID, elevID, message, order}
	transmitt <- msg
}



