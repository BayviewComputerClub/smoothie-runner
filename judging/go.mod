module github.com/BayviewComputerClub/smoothie-runner/judging

go 1.13

require (
	github.com/BayviewComputerClub/smoothie-runner/adapters v0.0.0-20191004210318-697ce8368920
	github.com/BayviewComputerClub/smoothie-runner/protocol v0.0.0-20191005014351-73e6f012bacd
	github.com/BayviewComputerClub/smoothie-runner/shared v0.0.0-20191005014351-73e6f012bacd
	github.com/BayviewComputerClub/smoothie-runner/util v0.0.0-20191005014351-73e6f012bacd
	github.com/rs/xid v1.2.1
	golang.org/x/sys v0.0.0-20191003212358-c178f38b412c
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol => ../protocol

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared

replace github.com/BayviewComputerClub/smoothie-runner/adapters => ../adapters
