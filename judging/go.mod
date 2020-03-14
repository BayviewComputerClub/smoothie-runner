module github.com/BayviewComputerClub/smoothie-runner/judging

go 1.13

require (
	github.com/BayviewComputerClub/smoothie-runner/adapters v0.0.0-20191004210318-697ce8368920
	github.com/BayviewComputerClub/smoothie-runner/cache v0.0.0-20191005014351-73e6f012bacd
	github.com/BayviewComputerClub/smoothie-runner/protocol/runner v0.0.0-20200201204513-82f95cf7ffdf
	github.com/BayviewComputerClub/smoothie-runner/sandbox v0.0.0-00010101000000-000000000000
	github.com/BayviewComputerClub/smoothie-runner/shared v0.0.0-20191005014351-73e6f012bacd
	github.com/BayviewComputerClub/smoothie-runner/util v0.0.0-20191005014351-73e6f012bacd
	github.com/kr/pretty v0.2.0 // indirect
	github.com/rs/xid v1.2.1
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol/runner => ../protocol/runner

replace github.com/BayviewComputerClub/smoothie-runner/protocol/test-data => ../protocol/test-data

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/sandbox => ../sandbox

replace github.com/BayviewComputerClub/smoothie-runner/cache => ../cache

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared

replace github.com/BayviewComputerClub/smoothie-runner/adapters => ../adapters
