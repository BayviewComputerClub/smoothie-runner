FROM bsscc/smoothie-runner-base-image:latest

ENV PORT=6821
ENV MAX_THREADS=1
ENV TESTING_DIR=testing-sessions
ENV CACHE_DIR=cache
ENV DEBUG=false
ENV SANDBOX=true
ENV CLEANUP_SESSIONS=true
ENV RLIMITS=true
ENV JAVA_SANDBOX_AGENT=java_sandbox.jar

EXPOSE $PORT

COPY . /usr/src/server
WORKDIR /usr/src/server

RUN chmod +x protocol/generate.sh && \
    bash protocol/generate.sh && \
    cd main && \
    go build ./... && \
    mv ./main /bin/smoothie-runner && \
    cd ../ && \
    mv util/java_sandbox.jar ./ && \
    mkdir -p testing-sessions

ENTRYPOINT smoothie-runner
