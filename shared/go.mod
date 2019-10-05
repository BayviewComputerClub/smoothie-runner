module github.com/BayviewComputerClub/smoothie-runner/shared

go 1.13

replace github.com/BayviewComputerClub/smoothie-runner/protocol => ../protocol

require (
	github.com/BayviewComputerClub/smoothie-runner/protocol v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20191003171128-d98b1b443823 // indirect
	google.golang.org/grpc v1.24.0 // indirect
)
