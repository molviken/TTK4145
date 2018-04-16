package elevFunc

import (
	"container/list"
	elevio "../elevio"
	queue "../queue"
	"fmt"
	"time"

)
const(
	 // Number of floors
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
	const doorTime = 1* time.Second
	timer := time.NewTimer(0*time.Second)

	for {
		select {
		case <-timerReset:
			timer.Reset(doorTime)
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



// ALGORITHM:	FS = Suitability score		N = Floors -1		d = distance
// (1) Towards the call, same direction
//		FS = (N+2) - d
// (2) Towards the call, opposite direction
//		FS = (N+1) - d
// (3) Away from the call
//		FS = 1
func CalculateCost(button elevio.ButtonEvent, floor int, c_dir elevio.MotorDirection){
	var FS int
	d := button.Floor - floor
	fmt.Println("d = ", d)
	if (d<0){d=-d}
	//if (((d<0) && (c_dir>0) && (button.Button == 0)) || ((d>0) && (c_dir>0) && (button.Button == 1))){
	if( ((c_dir < 0) && (d > 0)) || ((c_dir > 0) && (d < 0)) ){
		FS = 1
	} else if (((d<0) && (c_dir>0) && (button.Button == 1)) || ((d>0) && (c_dir>0) && (button.Button == 0))){
		FS = (NumFloors) - d
	} else{
		FS = (NumFloors+1) - d
	}
	fmt.Println("FS: ", FS)
	//return FS
}
