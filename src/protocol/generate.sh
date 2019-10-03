#!/bin/bash

protoc -I protocol protocol/smoothie-runner.proto --go_out=plugins=grpc:protocol