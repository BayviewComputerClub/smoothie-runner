module github.com/BayviewComputerClub/smoothie-runner/cli/main

go 1.13

require (
	github.com/BayviewComputerClub/smoothie-runner/protocol v0.0.0-20191005014351-73e6f012bacd
	google.golang.org/grpc v1.24.0
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol => ../../protocol
