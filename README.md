# kinesis-demo
 example apps showcasing various capabilities provided by [kinesis](https://github.com/srcfoundry/kinesis)

<br>

## kinesis-app1

Kinesis-app1 is an example to showcase the flexibility of dynamically adding and shutting down components within an application. The example shows that the http endpoint (/kinesis-app1/comp1) for "comp1" component get dynamically added even though we initialize "httpServer" component prior to that. Similarly, when "comp1" is shutdown, the http routes to it get removed and a curl to the endpoint gives "{"status":"not found"}"


### Running
- run ```while sleep 1; do curl http://127.0.0.1:8080/kinesis-app1/comp1; done``` to repeatedly curl "kinesis-app1/comp1" endpoint.
- on another console run ```go run cmd/kinesis-app1/kinesis-app1.go <shutdown-delay>```, whereby "comp1" component would be shutdown after <b>"shutdown-delay"</b> seconds have elapsed.
  