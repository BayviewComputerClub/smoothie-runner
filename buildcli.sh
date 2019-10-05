#!/bin/bash

cd cli/main
go build ./...
mv ./main ../../test/smoothie-runner-cli