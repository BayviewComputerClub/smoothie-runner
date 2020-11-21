#!/bin/bash
export PATH=$PATH:$GOPATH/bin
protoc -I protocol protocol/runner/smoothie-runner.proto --go-grpc_out=require_unimplemented_servers=false:protocol --go_out=protocol --go_opt=paths=source_relative
protoc -I protocol protocol/test-data/test-data.proto --go-grpc_out=require_unimplemented_servers=false:protocol --go_out=protocol --go_opt=paths=source_relative
