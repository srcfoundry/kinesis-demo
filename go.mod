module github.com/srcfoundry/kinesis-demo

go 1.17

require (
	github.com/gorilla/mux v1.8.0
	github.com/srcfoundry/kinesis v0.0.11
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
)

replace github.com/srcfoundry/kinesis => ../kinesis
