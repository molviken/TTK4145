package elevFunc

import (
	"container/list"
	elevio "../elevio"
	//queue ".././queue"
	"fmt"
	"time"
)
const(
	N = 4-1 // Number of floors -1
)
func PrintList(l *list.Element){
	fmt.Println("Kø: \n")
	for k := l; k != nil; k = k.Next(){
		fmt.Println(k.Value)
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
	const doorTime = 3 * time.Second
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

func GetDirection(floor int, order int)elevio.MotorDirection{
	dir :=  floor - order
	if (dir<0){return elevio.MD_Up}else if(dir>0){return elevio.MD_Down}else{return elevio.MD_Stop}
}


func ExecuteOrder(sensor int, order int){
	if (order < sensor){
		elevio.SetMotorDirection(elevio.MD_Down)
	}else if (order > sensor){
		elevio.SetMotorDirection(elevio.MD_Up)
	}else{
		elevio.SetMotorDirection(0)
	}
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
		FS = (N+1) - d
	} else{
		FS = (N+2) - d
	}
	fmt.Println("FS: ", FS)
	//return FS
}
// ((d<0) && (c_dir>0) && (button.Button == 1)) || ((d>0) && (c_dir>0) && (button.Button == 0))
// ((d<0) && (c_dir>0) && (button.Button == 0)) || ((d>0) && (c_dir>0) && (button.Button == 1))
