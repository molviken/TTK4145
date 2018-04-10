package ElevAssigner

import(
  "../elevio"
  "../queue"
  "../network/peers"
//  "fmt"
)

var num_received  = 0

type UDPmsg struct{
  MsgID int
  ElevID int
  Cost int
  Order elevio.ButtonEvent
}

type CostReply struct {
  Id int
  Cost int
  Order elevio.ButtonEvent
}

type OrderMsg struct {
  Id int
  Order elevio.ButtonEvent
}



var costMap map[int] UDPmsg    //Used to collect cost, if a lift is dead its cost is set to inf

func ChooseElevator(msg UDPmsg, elevMap peers.PeerUpdate){
  var numOnline = len(elevMap.Peers)

  costMap[msg.ElevID] = msg
  num_received += 1

  if num_received == numOnline {


          var highestCostMsg = findHighestCost(costMap)

          queue.UpdateRemoteQueue(highestCostMsg.ElevID, highestCostMsg.Order)
          num_received = 0

        }
}


func findHighestCost(costMap map[int]UDPmsg)UDPmsg{
  highestCost := 0
  var highestMsg UDPmsg
  for _, costMsg := range(costMap) {
    if costMsg.Cost > highestCost {
        highestCost = costMsg.Cost
        highestMsg = costMsg
    }else if costMsg.Cost == highestCost{ //Dersom alle har samme kost
      if costMsg.ElevID < highestMsg.ElevID{
        highestMsg = costMsg
      }
    }
  }
  return highestMsg
}
