package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/srcfoundry/kinesis"
	"github.com/srcfoundry/kinesis/common"
	"github.com/srcfoundry/kinesis/component"
)

func main() {
	app := new(kinesis.App)
	app.Name = "kinesis-app1"
	app.RWMutex = &sync.RWMutex{}
	app.Add(app)

	httpServer := new(common.HttpServer)
	httpServer.Name = "httpserver"
	httpServer.RWMutex = &sync.RWMutex{}
	app.Add(httpServer)

	comp1 := new(component.SimpleComponent)
	comp1.Name = "comp1"
	comp1.RWMutex = &sync.RWMutex{}
	app.Add(comp1)

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
		errCh := make(chan error)
		comp1.Notify(func() (context.Context, interface{}, chan<- error) {
			return context.TODO(), component.Shutdown, errCh
		})
		<-errCh
	}()

	subscribe := make(chan interface{}, 1)
	defer close(subscribe)

	app.Subscribe("main.subscriber", subscribe)

	// pass ctrl-c after 5 seconds after comp1 has shutdown
	go func() {
		time.Sleep(time.Second * time.Duration(sleepTime+5))
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	for notification := range subscribe {
		if notification == component.Stopped {
			log.Println("Exiting")
			os.Exit(0)
		}
	}
}
