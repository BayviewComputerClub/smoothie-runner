FROM golang:1.13.4-buster

ENV PORT=6821
ENV MAXTHREADS=4
ENV TESTING_DIR=testing-sessions
ENV DEBUG=false
ENV SANDBOX=true

EXPOSE $PORT

COPY . /usr/src/server
WORKDIR /usr/src/server

RUN apt update -y && \
    apt install build-essential openjdk-11-jdk-headless golang-goprotobuf-dev -y && \
    chmod +x protocol/generate.sh && \
    bash protocol/generate.sh && \
    cd main && \
    go build ./... && \
    mv ./main /bin/smoothie-runner && \
    cd ../ && \
    mkdir -p testing-sessions

ENTRYPOINT smoothie-runner
