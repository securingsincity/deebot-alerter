package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	ifttt "github.com/lorenzobenvenuti/ifttt"
	vacbot "github.com/skburgart/go-vacbot"
)

const (
	CleanModeAuto       = "auto"
	CleanModeEdge       = "border"
	CleanModeSpot       = "spot"
	CleanModeSingleRoom = "singleRoom"
	CleanModeStop       = "stop"
)

var iftttKey = os.Getenv("IFTTT_KEY")
var iftttEvent = os.Getenv("IFTTT_EVENT")
var RunningModes = []string{CleanModeAuto, CleanModeEdge, CleanModeSpot, CleanModeSingleRoom}

var client = vacbot.NewFromConfigFile("./config")

type Clean struct {
	CleanType string `xml:"type,attr"`
}
type Control struct {
	Ret   string `xml:"ret,attr"`
	Clean Clean  `xml:"clean"`
	TD    string `xml:"td,attr"`
}

type XmlResult struct {
	XMLName xml.Name `xml:"query"`
	Ctl     Control  `xml:"ctl"`
}

var previousStatus = "stop"

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func main() {
	iftttClient := ifttt.NewIftttClient(iftttKey)
	callback := func(result interface{}, err error) {
		if err != nil {
			log.Printf("FAIL - %s", err)
		}
		if result != nil {
			resultValue := reflect.ValueOf(result)
			_, ok := resultValue.Type().FieldByName("Query")
			if ok {
				battery := reflect.ValueOf(result).FieldByName("Query").Bytes()
				var xmlResult = XmlResult{}
				batteryString := string(battery[:])
				if batteryString != "" {
					// fmt.Println(batteryString)
					err := xml.Unmarshal([]byte(battery), &xmlResult)
					if err != nil {
						fmt.Printf("error: %v", err)
						return
					}
					if xmlResult.Ctl.TD != "" {
						return
					}
					if xmlResult.Ctl.Clean.CleanType == CleanModeStop && previousStatus != CleanModeStop {
						previousStatus = CleanModeStop
						message := "Stopping"
						fmt.Println(message)
						iftttClient.Trigger(iftttEvent, []string{message})
					} else if contains(RunningModes, xmlResult.Ctl.Clean.CleanType) && !contains(RunningModes, previousStatus) {
						fmt.Println(batteryString)
						previousStatus = xmlResult.Ctl.Clean.CleanType
						message := fmt.Sprintf("Started Running in Mode %s", previousStatus)
						fmt.Println(message)
						iftttClient.Trigger(iftttEvent, []string{message})
					} else {
					}
				}
			}
		}
	}
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
