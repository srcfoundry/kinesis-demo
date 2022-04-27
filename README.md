# kinesis-demo
 example apps showcasing various capabilities provided by [kinesis](https://github.com/srcfoundry/kinesis)

<br>

## kinesis-app1

Kinesis-app1 is an example to showcase the flexibility of dynamically adding and shutting down components within an application, in addition to its endpoints. 
The example shows that the http endpoint (/kinesis-app1/comp1) for "comp1" component getting dynamically added and GET query on the same returning details about comp1. 
Similarly, when "comp1" is shutdown after a few seconds, the http routes to it get removed and a GET on the endpoint yields "{"status":"not found"}"

<br>

![](images/kinesis-app1-demo.gif)

<br>

#### Running
- run ```while sleep 2; do curl http://127.0.0.1:8080/kinesis-app1/comp1; echo -e '\n'; done``` to repeatedly query "kinesis-app1/comp1" endpoint.
- on another console run ```go run cmd/kinesis-app1/kinesis-app1.go <shutdown-delay>```, whereby "comp1" component would be shutdown after <b>"shutdown-delay"</b> seconds have elapsed.
  