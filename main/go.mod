module github.com/BayviewComputerClub/smoothie-runner/main

go 1.14

require (
	github.com/BayviewComputerClub/smoothie-runner/cache v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/judging v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/protocol/runner v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/protocol/test-data v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/shared v0.0.0-20200315011139-7b50a3d3e486
	github.com/BayviewComputerClub/smoothie-runner/util v0.0.0-20200315011139-7b50a3d3e486
	google.golang.org/grpc v1.27.0
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol/runner => ../protocol/runner

replace github.com/BayviewComputerClub/smoothie-runner/protocol/test-data => ../protocol/test-data

replace github.com/BayviewComputerClub/smoothie-runner/judging => ../judging

replace github.com/BayviewComputerClub/smoothie-runner/sandbox => ../sandbox

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared

replace github.com/BayviewComputerClub/smoothie-runner/adapters => ../adapters

replace github.com/BayviewComputerClub/smoothie-runner/cache => ../cache
