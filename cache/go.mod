module github.com/BayviewComputerClub/smoothie-runner/judging

go 1.13

require (
	github.com/BayviewComputerClub/smoothie-runner/protocol/test-data v0.0.0-20191005014351-73e6f012bacd
	github.com/BayviewComputerClub/smoothie-runner/shared v0.0.0-20191005014351-73e6f012bacd
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared

replace github.com/BayviewComputerClub/smoothie-runner/protocol/test-data => ../protocol/test-data
