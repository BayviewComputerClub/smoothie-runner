#!/bin/bash
export PATH=$PATH:$GOPATH/bin
protoc -I protocol protocol/runner/smoothie-runner.proto --go_out=plugins=grpc:protocol
protoc -I protocol protocol/test-data/test-data.proto --go_out=plugins=grpc:protocol
