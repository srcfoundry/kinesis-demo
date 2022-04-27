package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/srcfoundry/kinesis"
	"github.com/srcfoundry/kinesis/common"
	"github.com/srcfoundry/kinesis/component"
)

func main() {
	app := new(kinesis.App)
	app.Name = "kinesis-app1"
	app.Add(app)

	httpServer := new(common.HttpServer)
	httpServer.Name = "httpserver"
	app.Add(httpServer)

	comp1 := new(component.SimpleComponent)
	comp1.Name = "comp1"
	app.Add(comp1)

	go func() {
		var sleepTime int
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

	for notification := range subscribe {
		if notification == component.Stopped {
			log.Println("Exiting")
			os.Exit(0)
		}
	}
}
