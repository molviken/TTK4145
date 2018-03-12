package ElevAssginer

import(
  "../elevio"
  //"../queue"
  "../network/peers"
//  "fmt"
)

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


  costMap[reply.Id] = reply
  num_received += 1
  if num_received == numOnline {
          

          for _, replymsg := range(costMap){
            if replymsg.Cost < lowestCost  {
              lowestCost = replymsg.Cost
              lowestID = replymsg.Id
            }else if replymsg.Cost == lowestCost{ //Dersom alle har samme kost
              if replymsg.Id < lowestID{
                lowestID = replymsg.Id
              }
            }
          }
          queue.UpdateRemoteQueue(reply.Order, lowestID)
          num_received = 0

        }*/
}


func findHighestCost(costMap map[int]UDPmsg)UDPmsg{
  lowestCost := 0
  var lowestMsg UDPmsg
  for _, costMsg := range costMap {
    if CostMsg.Cost < m {
        lowestCost = costMsg.Cost
        lowestMsg = costMsg
    }
  }
  return lowest
}
