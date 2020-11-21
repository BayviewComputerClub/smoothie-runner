module github.com/BayviewComputerClub/smoothie-runner/judging

go 1.14

require (
	github.com/BayviewComputerClub/smoothie-runner/adapters v0.0.0-20191004210318-697ce8368920
	github.com/BayviewComputerClub/smoothie-runner/cache v0.0.0-20191005014351-73e6f012bacd
	github.com/BayviewComputerClub/smoothie-runner/protocol/runner v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/sandbox v0.0.0-20200315011139-7b50a3d3e486
	github.com/BayviewComputerClub/smoothie-runner/shared v0.0.0-20200315011139-7b50a3d3e486
	github.com/BayviewComputerClub/smoothie-runner/util v0.0.0-20200315011139-7b50a3d3e486
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.2.1
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol/runner => ../protocol/runner

replace github.com/BayviewComputerClub/smoothie-runner/protocol/test-data => ../protocol/test-data

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/sandbox => ../sandbox

replace github.com/BayviewComputerClub/smoothie-runner/cache => ../cache

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared

replace github.com/BayviewComputerClub/smoothie-runner/adapters => ../adapters
