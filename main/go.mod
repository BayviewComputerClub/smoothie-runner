module github.com/BayviewComputerClub/smoothie-runner/main

go 1.13

require (
	github.com/BayviewComputerClub/smoothie-runner/judging v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/protocol/runner v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/protocol/test-data v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/shared v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/util v0.0.0-20200201204513-82f95cf7ffdf
	google.golang.org/grpc v1.24.0
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol/runner => ../protocol/runner

replace github.com/BayviewComputerClub/smoothie-runner/protocol/test-data => ../protocol/test-data

replace github.com/BayviewComputerClub/smoothie-runner/judging => ../judging

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared

replace github.com/BayviewComputerClub/smoothie-runner/adapters => ../adapters
