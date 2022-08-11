package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/srcfoundry/kinesis"
	"github.com/srcfoundry/kinesis/common"
	"github.com/srcfoundry/kinesis/component"
)

func main() {
	app := new(kinesis.App)
	app.Name = "kinesis-app1"
	app.RWMutex = &sync.RWMutex{}
	err := app.Add(app)
	if err != nil {
		log.Printf("failed to start %s, due to %s", app.GetName(), err)
		os.Exit(1)
	}

	httpServer := new(common.HttpServer)
	httpServer.Name = "httpserver"
	httpServer.RWMutex = &sync.RWMutex{}
	err = app.Add(httpServer)
	if err != nil {
		log.Printf("failed to start %s, due to %s", httpServer.GetName(), err)
		os.Exit(1)
	}

	comp1 := new(component.SimpleComponent)
	comp1.Name = "comp1"
	comp1.RWMutex = &sync.RWMutex{}
	err = app.Add(comp1)
	if err != nil {
		log.Printf("failed to start %s, due to %s", comp1.GetName(), err)
		os.Exit(1)
	}

	var sleepTime int

	go func() {
		// check if delay is passed as argument
		if len(os.Args) == 2 {
			sleepTime, _ = strconv.Atoi(os.Args[1])
		}
		if sleepTime <= 0 {
			sleepTime = 5
		}
		time.Sleep(time.Second * time.Duration(sleepTime))
		log.Println("sending", component.Shutdown, "signal to", comp1.GetName())
		comp1.Notify(5*time.Second, component.ControlMsgId, map[component.MsgClassifierId]interface{}{component.ControlMsgId: component.Shutdown}, nil)
	}()

	subscribe := make(chan interface{}, 1)
	defer close(subscribe)

	app.Subscribe("main.subscriber", subscribe)

	// raise interrupt 5 seconds after comp1 has shutdown
	// go func() {
	// 	time.Sleep(time.Second * time.Duration(sleepTime+5))
	// 	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	// }()

	for notification := range subscribe {
		if notification == component.Stopped {
			log.Println("Exiting")
			os.Exit(0)
		}
	}
}
