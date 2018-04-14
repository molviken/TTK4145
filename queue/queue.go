package queue

import (
	"container/list"
	"fmt"

	assigner "../ElevAssigner"
	"../elevio"
)

const (
	N = 4 - 1 // num floors - 1
)

/* Define queues, we need a local queue and a remote queue. The remote queue should contain all
"outside" orders (of all elevators) in case of one of them dying. The local queue should only containt
"inside" orders of that elevator, and the outside orders assigned to that elevator.

The local queue needs to be saved on the disk due to the elevator dying. */

/*Run init to spawn the backup from disk in case of elevator coming back from dying
Also start the go routine for saving all local orders to disk
Initialize the linked list*/
func InitQueue() {
	//localL := list.New()
	//remoteL := list.New()
}

var shouldInit = true
var RemoteOrders map[elevio.ButtonEvent]int

func IsLocalOrder(floor int, buttonType elevio.ButtonType, localL *list.List) bool {
	for j := localL.Front(); j != nil; j = j.Next() {
		if j.Value.(*elevio.ButtonEvent).Button == buttonType && j.Value.(*elevio.ButtonEvent).Floor == floor {
			return true
		}
	}
	return false

}

func IsRemoteOrder(floor int, buttonType elevio.ButtonType) bool {
	var temp elevio.ButtonEvent
	temp.Button = buttonType
	temp.Floor = floor
	if val, ok := RemoteOrders[temp]; ok {
		if val != 0 {
			return true
		}
	}
	return false
}

func AddLocalOrder(l *list.List, order elevio.ButtonEvent) {
	e := new(elevio.ButtonEvent)
	e = &order
	l.PushBack(e)
}

func AddRemoteOrder(ID int, remoteOrder elevio.ButtonEvent) {
	if shouldInit {
		RemoteOrders = make(map[elevio.ButtonEvent]int)
		shouldInit = false
		fmt.Println("RemoteOrder map initialized")
	}

	RemoteOrders[remoteOrder] = ID
}

func RemoveLocalOrder(ll *list.List, order *list.Element) {
	ll.Remove(order)
}

func RemoveRemoteOrder(remoteOrder elevio.ButtonEvent) {
	fmt.Println("Removing order ")
	RemoteOrders[remoteOrder] = 0 //When the ID is 0 there is no order for that button
}

/*This function finds the cost of adding an order to the queue*/

func Cost(button elevio.ButtonEvent, floor int, c_dir elevio.MotorDirection) int {
	var FS int
	d := button.Floor - floor
	if d < 0 {
		d = -d
	}
	if ((c_dir < 0) && (d > 0)) || ((c_dir > 0) && (d < 0)) {
		//fmt.Println("(3) Away from the call")
		FS = 1
	} else if d == 0 {
		FS = (N + 2) - d
	} else if ((d < 0) && (c_dir > 0) && (button.Button == 1)) || ((d > 0) && (c_dir > 0) && (button.Button == 0)) {
		//fmt.Println("(2) Towards the call, opposite direction")
		FS = (N + 1) - d
	} else {
		//fmt.Println("(1) Towards the call, same direction")
		FS = (N + 3) - d

	}

	//fmt.Println("FS: ", FS)
	return FS
}

func DuplicateOrderLocal(ll *list.List, order elevio.ButtonEvent) bool {
	if ll.Front() != nil {
		for k := ll.Front(); k != nil; k = k.Next() {
			if k.Value.(elevio.ButtonEvent) == order {
				return true
			}
		}
	}
	return false
}

func DuplicateOrderRemote(order elevio.ButtonEvent) bool {
	PrintMap()
	fmt.Println("RemoteOrders[order] = :",RemoteOrders[order])
	if val, ok := RemoteOrders[order]; ok {

		if val != 0 {
			return true
		}
	}

	return false
}
func PrintMap() {
	fmt.Println(" ")
	fmt.Println("Remote Order Map:")
	for key, val := range RemoteOrders {
		fmt.Printf("ButtonEvent:  %d  Elevator ID:   %d    \n", key, val)
	}
	fmt.Println(" ")
}

func ScanForDouble(dir elevio.MotorDirection, floor int, localL *list.List, elevId int, transmitt chan assigner.UDPmsg, isCab bool) {
	fmt.Println("Scanning for orders at floor")

	if localL.Front() != nil {
		if isCab == true && localL.Front().Value.(*elevio.ButtonEvent).Floor == floor{
			localL.Remove(localL.Front())
		}

		for k := localL.Front(); k != nil; k = k.Next() {
			if k.Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab && k.Value.(*elevio.ButtonEvent).Floor == floor {
				localL.Remove(k)
				fmt.Println("local cab fjerna")
			} else if k.Value.(*elevio.ButtonEvent).Button == elevio.BT_HallUp && k.Value.(*elevio.ButtonEvent).Floor == floor && dir == 1 {
				localL.Remove(k)
				var btEvent elevio.ButtonEvent
				btEvent.Button = k.Value.(*elevio.ButtonEvent).Button
				btEvent.Floor = k.Value.(*elevio.ButtonEvent).Floor
				assigner.TransmittUDP(4, elevId, 0, btEvent, transmitt)

				RemoveRemoteOrder(btEvent)
				fmt.Println("remote up fjerna")
			} else if k.Value.(*elevio.ButtonEvent).Button == elevio.BT_HallDown && k.Value.(*elevio.ButtonEvent).Floor == floor && dir == -1 {
				localL.Remove(k)
				var btEvent elevio.ButtonEvent
				btEvent.Button = k.Value.(*elevio.ButtonEvent).Button
				btEvent.Floor = k.Value.(*elevio.ButtonEvent).Floor
				assigner.TransmittUDP(4, elevId, 0, btEvent, transmitt)
				fmt.Println("remote ned fjerna")
				RemoveRemoteOrder(btEvent)
			}
		}
	}
}
