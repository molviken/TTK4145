package queue

import (
	"container/list"
	"fmt"
	"io/ioutil"
    "encoding/json"
	assigner "../ElevAssigner"
	"../elevio"
)

const (
	N = 4 - 1 // num floors - 1
)
var shouldInit = true
var RemoteOrders map[elevio.ButtonEvent]int



func UpdateBackup(l *list.List){
	var backupList []elevio.ButtonEvent
	if (l.Front() != nil) {
		for k := l.Front(); k != nil; k = k.Next() {
			if (k.Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab){
            	temp := elevio.ButtonEvent{k.Value.(*elevio.ButtonEvent).Floor, k.Value.(*elevio.ButtonEvent).Button}
				backupList = append(backupList,temp)				
			}
		}
	}
    b1, _ := json.Marshal(backupList)
    ioutil.WriteFile("Backup", b1, 0644)
}
func ReadBackup(localL *list.List){
	//i := 0
	var backup []elevio.ButtonEvent
	c, err := ioutil.ReadFile("Backup")

	if err != nil {
		fmt.Println("No file exists yet, please continue")
		
	}else{

	json.Unmarshal(c, &backup)
	fmt.Println("Orders saved from backup: ")
	for _, order := range backup {
		AddLocalOrder(localL, order)
		}
	}	
}
func IsLocalOrder(floor int, buttonType elevio.ButtonType, localL *list.List) bool {
	if localL.Front() != nil {
		for j := localL.Front(); j != nil; j = j.Next() {
			if j.Value.(*elevio.ButtonEvent).Button == buttonType && j.Value.(*elevio.ButtonEvent).Floor == floor {
				return true
			}
		}
	}
	return false
}

func IsRemoteOrder(floor int, buttonType elevio.ButtonType) bool {
	var temp elevio.ButtonEvent
	temp.Button = buttonType
	temp.Floor = floor
	if(!shouldInit){
		if val, ok := RemoteOrders[temp]; ok {
			if val != 0 {
				return true
			}
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
// Highest Cost wins the order.
func Cost(button elevio.ButtonEvent, floor int, c_dir elevio.MotorDirection) int {
	var FS int
	d := button.Floor - floor
	if d < 0 {
		d = -d
	}
	if ((c_dir < 0) && (d > 0)) || ((c_dir > 0) && (d < 0)) {
		FS = 1
	} else if ((d < 0) && (c_dir > 0) && (button.Button == 1)) || ((d > 0) && (c_dir > 0) && (button.Button == 0)) {
		FS = (N + 1) - d
	}else if d == 0 {
		FS = (N + 3) - d
	}else {
		FS = (N + 2) - d
	}
	return FS
}
func DuplicateOrderRemote(order elevio.ButtonEvent) bool {
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
	var btEvent elevio.ButtonEvent

	if localL.Front() != nil {
		if localL.Front().Value.(*elevio.ButtonEvent).Floor == floor{ //isCab == true && 
			btEvent.Button = localL.Front().Value.(*elevio.ButtonEvent).Button
			btEvent.Floor = localL.Front().Value.(*elevio.ButtonEvent).Floor
			localL.Remove(localL.Front())
			if(btEvent.Button != elevio.BT_Cab){
				assigner.TransmittUDP(4, elevId, 0, btEvent, transmitt)
				RemoveRemoteOrder(btEvent)
		}
	}
		for k := localL.Front(); k != nil; k = k.Next() {
			if k.Value.(*elevio.ButtonEvent).Button == elevio.BT_Cab && k.Value.(*elevio.ButtonEvent).Floor == floor {
				localL.Remove(k)
			} else if k.Value.(*elevio.ButtonEvent).Button == elevio.BT_HallUp && k.Value.(*elevio.ButtonEvent).Floor == floor && dir == 1 {
				localL.Remove(k)
				btEvent.Button = k.Value.(*elevio.ButtonEvent).Button
				btEvent.Floor = k.Value.(*elevio.ButtonEvent).Floor
				assigner.TransmittUDP(4, elevId, 0, btEvent, transmitt)
				RemoveRemoteOrder(btEvent)
			} else if k.Value.(*elevio.ButtonEvent).Button == elevio.BT_HallDown && k.Value.(*elevio.ButtonEvent).Floor == floor && dir == -1 {
				localL.Remove(k)
				btEvent.Button = k.Value.(*elevio.ButtonEvent).Button
				btEvent.Floor = k.Value.(*elevio.ButtonEvent).Floor
				assigner.TransmittUDP(4, elevId, 0, btEvent, transmitt)
				RemoveRemoteOrder(btEvent)
			}
		}
	}

}
