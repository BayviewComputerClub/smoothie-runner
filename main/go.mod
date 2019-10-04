module github.com/BayviewComputerClub/smoothie-runner/main

go 1.13

require google.golang.org/grpc v1.24.0

replace github.com/BayviewComputerClub/smoothie-runner/protocol => ../protocol

replace github.com/BayviewComputerClub/smoothie-runner/judging => ../judging

replace github.com/BayviewComputerClub/smoothie-runner/util => ../util

replace github.com/BayviewComputerClub/smoothie-runner/shared => ../shared
