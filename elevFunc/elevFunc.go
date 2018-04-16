package elevFunc

import (
	"container/list"
	elevio "../elevio"
	queue "../queue"
	"fmt"
	"time"

)
const(
	NumFloors = 4
	NumButtons = 3
)


func PrintList(l *list.Element){
	fmt.Println("Kø: ")
	for k := l; k != nil; k = k.Next(){
		fmt.Println(k.Value)
	}
}

//Sets light for all orders in the remote and local queue
func SyncButtonLights(localL *list.List){

	for i := 0; i < NumFloors; i++ {
		for k := 0; k < NumButtons; k++ {
			if (k == 0 && i == NumFloors-1) || (k == 1 && i == 0) {
				continue
			} else {
				switch  k{
				case 2:

					elevio.SetButtonLamp(elevio.BT_Cab, i, queue.IsLocalOrder(i, elevio.BT_Cab, localL))
				case 0, 1:
					elevio.SetButtonLamp(elevio.BT_HallUp, i, queue.IsRemoteOrder(i, elevio.BT_HallUp))
					elevio.SetButtonLamp(elevio.BT_HallDown, i, queue.IsRemoteOrder(i, elevio.BT_HallDown))

				}
			}
		}
	}
}
func StateToString(state int) string{
	switch state {
	case 0:
		return "idle"
	case 1:
		return "moving"
	case 2:
		return "doorOpen"
	default:
		return "Not a valid state"

	}
}
func DuplicateOrder(button elevio.ButtonEvent, localL *list.List) bool{
	for k := localL.Front(); k != nil; k = k.Next(){
		if ((k.Value.(*elevio.ButtonEvent).Floor == button.Floor) && (k.Value.(*elevio.ButtonEvent).Button == button.Button)) {
			return true
		}
	}
	return false
}
func OpenDoor(timeOut chan<- bool, timerReset <-chan bool){
	const doorTime = 2* time.Second
	timer := time.NewTimer(0*time.Second)
	fmt.Println("ARE YOU EVEN ON DOOR?")
	for {
		select {
		case <-timerReset:
			timer.Reset(doorTime)
			elevio.SetDoorOpenLamp(true)
		case <-timer.C:
			timer.Stop()
			timeOut <- true
		}
	}
}
func ObstructionTimeOut(obstr chan<- bool, obstrTimerReset <-chan bool, l *list.List ){
	const obstrTime = 5* time.Second
	obstrTimer := time.NewTimer(0*time.Second)

	for {
		select {
		case <-obstrTimerReset:
			obstrTimer.Reset(obstrTime)
			fmt.Println("Timer reset")

		case <-obstrTimer.C:
			PrintList(l.Front())
			fmt.Println("Timer expired")
			//obstrTimer.Stop()
			if(l.Front() != nil){
				obstr <- true
			}

		}
	}
}
func GetDirection(floor int, order int)elevio.MotorDirection{
	dir :=  floor - order
	if (dir<0){return elevio.MD_Up}else if(dir>0){return elevio.MD_Down}else{return elevio.MD_Stop}
}
