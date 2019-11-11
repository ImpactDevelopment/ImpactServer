package util

import (
	"fmt"
	"log"
	"github.com/logrusorgru/aurora"
)

func LogError(msg interface{}) {
	log.Println(aurora.Red("[ERROR] " + fmt.Sprintf("%v", msg)).Bold()) // Sprintf for nil printing
}

func LogWarn(msg interface{}) {
	log.Println(aurora.Yellow("[WARNING] " + fmt.Sprintf("%v", msg)))
}

func LogInfo(msg interface{}) {
	log.Println(aurora.White("[INFO] " + fmt.Sprintf("%v", msg)))
}

func LogSuccess(msg interface{}) {
	log.Println(aurora.Green(fmt.Sprintf("%v", msg)).Bold())
}