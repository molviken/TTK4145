package ElevAssigner

import (
	"../elevio"
	"fmt"

	"../network/peers"
)

var num_received = 0

type UDPmsg struct {
	MsgID  int
	ElevID int
	Message   int //Message is used to transmitt the cost or winner ID over UDP
	Order  elevio.ButtonEvent
}


var costMap map[int]UDPmsg

var shouldInit = true //For initiliazing the cost map


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
		var highestCostMsg = findHighestCost(costMap)
		num_received = 0
		highestCostMsg.Message = highestCostMsg.ElevID //Transmitt winner ID
		fmt.Println("Vinner ID: ", highestCostMsg.Message)
		highestCostMsg.MsgID = 3
		transmitt <- highestCostMsg
	}
}

func findHighestCost(costMap map[int]UDPmsg) UDPmsg {
	highestCost := 0
	var highestMsg UDPmsg
	for _, costMsg := range costMap {
		if costMsg.Message > highestCost {
			highestCost = costMsg.Message
			highestMsg = costMsg
		} else if costMsg.Message == highestCost { //Dersom alle har samme kost
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

func TransmittUDP(msgID int, elevID int, message int, order elevio.ButtonEvent, transmitt chan UDPmsg) {
	//var msg UDPmsg
	msg := UDPmsg{msgID, elevID, message, order}
	transmitt <- msg
}



