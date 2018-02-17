package main

import (
	//"flag"
	"fmt"
	hw "./hardware"
	channels "./hardware"
	/*"os"
	"os/signal"
	"strconv"
	"time"
*/
)

func main(){
	hw.SetButtonLamp(channels.BUTTON_DOWN2, 2, 1)
	fmt.Println("Testing testing")
}
