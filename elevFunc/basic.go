package elevFunc

import (
	"container/list"
	elevio ".././elevio"
	"fmt"
)
const(
	N = 4-1 // Number of floors -1
)

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

func GoToOrder(sensor int, order int, l *list.Element){

	if (order < sensor){
		elevio.SetMotorDirection(elevio.MD_Down)
	}else if (order > sensor){
		elevio.SetMotorDirection(elevio.MD_Up)
	}else{
		elevio.SetMotorDirection(0)
		l = l.Next()
		if(l != nil){
			fmt.Println(l.Value)
		}else{
			fmt.Println("Listen er tomm")
		}
		
	}
	//fmt.Println("Ordered floor: ",order)

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

func Fsm_Stop(stop_button bool){
	if (stop_button){
		elevio.SetMotorDirection(0)
	}
}
// ((d<0) && (c_dir>0) && (button.Button == 1)) || ((d>0) && (c_dir>0) && (button.Button == 0))   
// ((d<0) && (c_dir>0) && (button.Button == 0)) || ((d>0) && (c_dir>0) && (button.Button == 1))      

