#!/bin/bash
export PATH=$PATH:$GOPATH/bin
protoc -I protocol protocol/smoothie-runner.proto --go_out=plugins=grpc:protocol
