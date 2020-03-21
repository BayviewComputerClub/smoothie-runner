# smoothie-runner
`
The smooothest way to do competitive programming judging! sweeet
`

smoothie-runner allows for running, and verifying programs for competitive programming in a sandboxed environment.

## Environment Variables
smoothie-runner is configured through environment variables, because (¬‿¬)
* PORT - Port that API runs on
* MAX_THREADS - Maximum number of threads that smoothie-runner can utilize for judging
* TESTING_DIR - The directory that is used for judging sessions
* DEBUG - Whether or not to enable debug messages
* SANDBOX - Whether or not to enable sandboxing with ptrace & seccomp
* CLEANUP_SESSIONS - Whether or not to cleanup sessions after they are finished.
* RLIMITS - Whether or not resource limits should be applied.

## Running
It is highly recommended to run smoothie-runner in Docker. You can also run it outside of a container, provided that it is on a Linux based operating system (kernel 3.19 or later).

You can pull the image here (for now):
```
$ docker pull espidev/smoothie-runner
```

## Building (Docker)
A Dockerfile is provided in the repository.

## Building (Binary)
If you want to compile the program, make sure you have Go 1.13 installed, and that you are compiling for Linux-x64. You can simply run the build script:
```
$ ./build.sh
```
 
 ## Sandbox
 The sandbox loads a seccomp filter, which filters syscalls, and lets restricted calls be checked by the parent process with ptrace. 
 No libseccomp dependency because of https://github.com/elastic/go-seccomp-bpf! (ﾉ◕ヮ◕)ﾉ*:･ﾟ✧