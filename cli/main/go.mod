module github.com/BayviewComputerClub/smoothie-runner/cli/main

go 1.14

require (
	github.com/BayviewComputerClub/smoothie-runner/protocol/runner v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/protocol/test-data v0.0.0-20200201204513-82f95cf7ffdf
	github.com/golang/protobuf v1.3.2
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	google.golang.org/grpc v1.27.0
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol/runner => ../../protocol/runner

replace github.com/BayviewComputerClub/smoothie-runner/protocol/test-data => ../../protocol/test-data
