module github.com/BayviewComputerClub/smoothie-runner/cli/main

go 1.14

require (
	github.com/BayviewComputerClub/smoothie-runner/protocol/runner v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/protocol/test-data v0.0.0-20200201204513-82f95cf7ffdf
	github.com/golang/protobuf v1.4.1
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol/runner => ../../protocol/runner

replace github.com/BayviewComputerClub/smoothie-runner/protocol/test-data => ../../protocol/test-data
