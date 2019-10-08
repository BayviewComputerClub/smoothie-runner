#!/bin/bash

./protocol/generate.sh
cd main
go build ./...
mv ./main ../test/smoothie-runner
