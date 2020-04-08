# smoothie-runner
smoothie-runner allows for running, and verifying programs for competitive programming in a sandboxed environment.

## Environment Variables
smoothie-runner is configured through environment variables, because (¬‿¬)
* PORT=6821 - Port that API runs on
* MAX_THREADS=1 - Maximum concurrent workers that smoothie-runner can utilize for judging
* TESTING_DIR=testing-sessions - Directory used for judging sessions
* CACHE_DIR=cache - Directory used for caching test data 
* DEBUG=false - Whether or not to enable debug messages
* SANDBOX=true - Whether or not to enable sandboxing with ptrace & seccomp
* CLEANUP_SESSIONS=true - Whether or not to cleanup sessions after they are finished.
* RLIMITS=true - Whether or not resource limits should be applied.
* JAVA_SANDBOX_AGENT=java_sandbox.jar - Location of the [java-sandbox-agent](https://github.com/DMOJ/java-sandbox-agent/tree/d73cc65b7454250d7a7aac81edbb0c1d8fa64c62) jar file.

## Running
It is highly recommended to run smoothie-runner in Docker. You can also run it outside of a container, provided that it is on a Linux based operating system (kernel 3.19 or later).

You can pull the image here (for now):
```shell script
$ docker pull espidev/smoothie-runner
```

When running the container, you'll want to to add the `--cap-add=SYS_PTRACE` option. 
This capability needs to be permitted for the process in order to use sandboxing.

Read more here: [https://docs.docker.com/engine/security/seccomp/](https://docs.docker.com/engine/security/seccomp/)

Example:
```shell script
$ docker run -p 6821:6821 --name=runner --cap-add=SYS_PTRACE espidev/smoothie-runner
```

## Building (Docker)
A Dockerfile is provided in the repository.

## Building (Binary)
If you want to compile the program, make sure you have Go 1.13 installed, and that you are compiling for Linux-x64. You can simply run the build script:
```shell script
$ ./build.sh
```
 
 ## Requirements
 * Linux kernel 3.19+
 * https://github.com/DMOJ/java-sandbox-agent/
 
 ## Sandbox
 The sandbox loads a seccomp filter, which filters syscalls, and lets restricted calls be checked by the parent process with ptrace. 
 No libseccomp dependency because of https://github.com/elastic/go-seccomp-bpf! (ﾉ◕ヮ◕)ﾉ*:･ﾟ✧