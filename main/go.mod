module github.com/BayviewComputerClub/smoothie-runner/main

go 1.13

require (
	github.com/BayviewComputerClub/smoothie-runner/protocol v0.0.0-20191003014411-2beccbe92862
	google.golang.org/grpc v1.24.0
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol => ../protocol

replace github.com/BayviewComputerClub/smoothie-runner/judging => ../judging

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util
