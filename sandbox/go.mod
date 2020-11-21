module github.com/BayviewComputerClub/smoothie-runner/sandbox

go 1.14

require (
	github.com/BayviewComputerClub/smoothie-runner/shared v0.0.0-20200315011139-7b50a3d3e486
	github.com/BayviewComputerClub/smoothie-runner/util v0.0.0-20191005014351-73e6f012bacd
	github.com/elastic/go-seccomp-bpf v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/tklauser/go-sysconf v0.0.0-20200125124152-4f5f1f2b970a
	golang.org/x/net v0.0.0-20191003171128-d98b1b443823
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9
)

replace github.com/BayviewComputerClub/smoothie-runner/protocol/runner => ../protocol/runner

replace github.com/BayviewComputerClub/smoothie-runner/protocol/test-data => ../protocol/test-data

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/cache => ../cache

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared

replace github.com/BayviewComputerClub/smoothie-runner/adapters => ../adapters
