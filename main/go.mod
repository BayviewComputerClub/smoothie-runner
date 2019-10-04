module github.com/BayviewComputerClub/smoothie-runner/main

go 1.13

require (
	github.com/BayviewComputerClub/smoothie-runner/judging v0.0.0-00010101000000-000000000000
	github.com/BayviewComputerClub/smoothie-runner/protocol v0.0.0-20191004151926-35d6b89e92c5
	github.com/BayviewComputerClub/smoothie-runner/util v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.24.0
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol => ../protocol

replace github.com/BayviewComputerClub/smoothie-runner/judging => ../judging

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared
