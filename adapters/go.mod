module github.com/BayviewComputerClub/smoothie-runner/adapters

go 1.14

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared

replace github.com/BayviewComputerClub/smoothie-runner/sandbox => ../sandbox

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

require (
	github.com/BayviewComputerClub/smoothie-runner/sandbox v0.0.0-20200315011139-7b50a3d3e486 // indirect
	github.com/BayviewComputerClub/smoothie-runner/shared v0.0.0-20200315011139-7b50a3d3e486
	github.com/BayviewComputerClub/smoothie-runner/util v0.0.0-20200315011139-7b50a3d3e486 // indirect
)
