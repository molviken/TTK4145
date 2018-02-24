package elevFunc

import (
	elevio ".././elevio"
	"fmt"
)
const(
	N = 4-1 // Number of floors -1
)

func GoToOrder(sensor int, order int){
	if (order < sensor){
		elevio.SetMotorDirection(elevio.MD_Down)
	}else if (order > sensor){
		elevio.SetMotorDirection(elevio.MD_Up)
	}else{
		elevio.SetMotorDirection(0)
	}
	fmt.Println("Ordered floor: %v\n",order)

}
// ALGORITHM:	FS = Suitability score		N = Floors -1		d = distance
// (1) Towards the call, same direction
//		FS = (N+2) -d
// (2) Towards the call, opposite direction
//		FS = (N+1) -d
// (3) Away from the call
//		FS = 1




func CalculateCost(button elevio.ButtonEvent, floor int, c_dir elevio.MotorDirection){
	var FS int
	d := button.Floor - floor
	if (d<0){d=-d}
	if (((d<0) && (c_dir>0) && (button.Button == 0)) || ((d>0) && (c_dir>0) && (button.Button == 1))){
		FS = (N+2) - d
	}else if (((d<0) && (c_dir>0) && (button.Button == 1)) || ((d>0) && (c_dir>0) && (button.Button == 0))){
		FS = (N+1) - d
	}else{
		FS = 1
	}
	fmt.Println(FS)
	//return FS
}

func Fsm_Stop(stop_button bool){
	if (stop_button){
		elevio.SetMotorDirection(0)
	}
}
// ((d<0) && (c_dir>0) && (button.Button == 1)) || ((d>0) && (c_dir>0) && (button.Button == 0))   
// ((d<0) && (c_dir>0) && (button.Button == 0)) || ((d>0) && (c_dir>0) && (button.Button == 1))      