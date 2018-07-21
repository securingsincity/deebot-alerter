package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	vacbot "github.com/securingsincity/go-vacbot"
)

const (
	CLEAN_MODE_AUTO        = "auto"
	CLEAN_MODE_EDGE        = "edge"
	CLEAN_MODE_SPOT        = "spot"
	CLEAN_MODE_SINGLE_ROOM = "single_room"
	CLEAN_MODE_STOP        = "stop"
)

var client = vacbot.NewFromConfigFile("./config")

func callback(result interface{}, err error) {
	if err != nil {
		log.Printf("FAIL - %s", err)
	}
	if result != nil {
		resultValue := reflect.ValueOf(result)
		_, ok := resultValue.Type().FieldByName("Query")
		if ok {
			battery := reflect.ValueOf(result).FieldByName("Query").Bytes()
			batteryString := string(battery[:])
			if batteryString != "" {
				fmt.Println(batteryString)
				if strings.Contains(batteryString, CLEAN_MODE_STOP) {
					fmt.Println("STOPPED")
				} else {
					fmt.Println("RUNNING")
				}
			}
		}
	}
}

func main() {
	log.Printf("Called battery %s", "foo")
	fiveMinuteTicker := time.NewTicker(20 * time.Second)
	go func() {
		for _ = range fiveMinuteTicker.C {
			client.FetchCleanState()
		}
	}()
	// call initial call
	client.FetchCleanState()
	go client.RecvHandler(callback)
	fmt.Scanln()
	fiveMinuteTicker.Stop()
}
