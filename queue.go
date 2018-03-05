package queue

import (
)


/* Define queues, we need a local queue and a remote queue. The remote queue should contain all
"outside" orders (of all elevators) in case of one of them dying. The local queue should only containt
"inside" orders of that elevator, and the outside orders assigned to that elevator.

The local queue needs to be saved on the disk due to the elevator dying. */

/*Run init to spawn the backup from disk in case of elevator coming back from dying
Also start the go routine for saving all local orders to disk
Initialize the linked list*/
func initQueue(){

}

/*Probably need more functions here, maybe put shouldStop in here?*/

func updateLocalQueue(order Order){

}


func updateRemoteQueue(order Order){
}

func removeLocalOrder(order Order){

}

func removeRemoteOrder(order Order){

}
/*This function finds the cost of adding an order to the queue, not sure what arguments
it needs to figure it out*/
func Cost(button *elevio.ButtonEvent, floor int, c_dir elevio.MotorDirection) int {
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
