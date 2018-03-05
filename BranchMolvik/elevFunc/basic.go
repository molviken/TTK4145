package elevFunc

import (
	"time"
	"container/list"
	elevio ".././elevio"
	"fmt"
)
const(
	N = 4-1 // Number of floors -1
)

func OpenDoor(){
	elevio.SetMotorDirection(0)
	elevio.SetDoorOpenLamp(true)
	time.Sleep(1*time.Second)
	elevio.SetDoorOpenLamp(false)
}

func PrintList(l *list.Element){
	for k := l; k != nil; k = k.Next(){
		fmt.Println("List: ",k.Value)
	}
}

func CloneList(l *list.List)*list.List{
	temp := list.New()
	for k := l.Front(); k != nil; k = k.Next(){
		temp.PushBack(k)
	}
	return temp
}
func ElevInit(a int, init bool){
	if(init == false && a == 0){
		elevio.SetMotorDirection(elevio.MD_Stop)
		init = true
	}
}

func GetDirection(sensor int, order int)elevio.MotorDirection{
	dir := sensor - order
	if (dir<0){return elevio.MD_Up}else if(dir>0){return elevio.MD_Down}else{return elevio.MD_Stop}
}

func ScanFloor(floorSensor int, dir elevio.MotorDirection, l *list.List){
	for k:= l.Front(); k != nil; k = k.Next(){
		if (k.Value.(*elevio.ButtonEvent).Floor == floorSensor){
			switch{
			case dir == elevio.MD_Up:
				if (k.Value.(*elevio.ButtonEvent).Button != 1){
					OpenDoor()
					l.Remove(k)
				}
			case dir == elevio.MD_Down:
				if (k.Value.(*elevio.ButtonEvent).Button != 0){
					OpenDoor()
					l.Remove(k)
				}
			}
		}
	}
}

func GoToOrder(sensor int, order int, l *list.List){
	if (order < sensor){
		elevio.SetMotorDirection(elevio.MD_Down)
	}else if (order > sensor){
		elevio.SetMotorDirection(elevio.MD_Up)
	}else{
		elevio.SetMotorDirection(0)
		OpenDoor()
		l.Remove(l.Front())		
	}
}

// ALGORITHM:	FS = Suitability score		N = Floors -1		d = distance
// (1) Towards the call, same direction
//		FS = (N+2) - d
// (2) Towards the call, opposite direction
//		FS = (N+1) - d
// (3) Away from the call
//		FS = 1
func CalculateCost(button *elevio.ButtonEvent, floor int, c_dir elevio.MotorDirection)int{
	var FS int
	d := button.Floor - floor
	if (d<0){d=-d}
	//if (((d<0) && (c_dir>0) && (button.Button == 0)) || ((d>0) && (c_dir>0) && (button.Button == 1))){
	if( ((c_dir < 0) && (d > 0)) || ((c_dir > 0) && (d < 0)) ){
		//fmt.Println("(3) Away from the call")
		FS = 1
	} else if (((d<0) && (c_dir>0) && (button.Button == 1)) || ((d>0) && (c_dir>0) && (button.Button == 0))){
		//fmt.Println("(2) Towards the call, opposite direction")
		FS = (N+1) - d
	} else{
		//fmt.Println("(1) Towards the call, same direction")
		FS = (N+2) - d
	}
	//fmt.Println("FS: ", FS)
	return FS
}

func Fsm_Stop(stop_button bool){
	if (stop_button){
		elevio.SetMotorDirection(0)
	}
}
// ((d<0) && (c_dir>0) && (button.Button == 1)) || ((d>0) && (c_dir>0) && (button.Button == 0))   
// ((d<0) && (c_dir>0) && (button.Button == 0)) || ((d>0) && (c_dir>0) && (button.Button == 1))      


