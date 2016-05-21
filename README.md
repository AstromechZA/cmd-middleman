# RPC command middleman

This repository contains the Golang code for a RPC system for executing
whitelisted commands across a unix domain socket. The goal of this project is
to demonstrate a technique of running specific commands on a Docker host from
a container by mounting a socket file in the container. It can also be used to
escalate privileges for specific commands.

The components of the system are 2 standalone binaries, a server and a client.
The server listens on the socket for incoming commands, processes them and sends
the results back to the client.

Commands are checked against a whitelist file that contains a series of regular
expressions, one of which must match the call for it to be executed.

## Usage of server

```
rpc-cmd-server is the server side of an RPC system for executing commands
across a unix domain socket. This can be used for escalating priviledges of the
caller for specific commands or calling commands outside of a Docker container.

This has a high chance of being used maliciously if it results in privilege
escalation, but is certainly better than some alternatives.

To protect against running any old command this application requires a file
containing a number of regex patterns for whitelisted commands. At least one
of these patterns MUST match a command and its arguments in order to be executed.
  -socket string
        Path to a unix socket file being listened to by the server.
  -whitelist string
        Path to a file containing regex patterns. A regex pattern MUST match an incoming command for it to be run.
```

### Example call:

```
./rpc-cmd-server --socket /tmp/sock.sock --whitelist ./cmd-whitelist
```

## Usage of client

```
rpc-cmd-client is the client side of an RPC system for executing commands
across a unix domain socket. This can be used for escalating priviledges of the
caller for specific commands or calling commands outside of a Docker container.

This has a high chance of being used maliciously if it results in privilege
escalation, but is certainly better than some alternatives.

The exit code will be 120 if an error regarding the local client occurs. Any
other exit codes will be those of the RPC call.

Arguments will be passed through to the server side for execution. For example:
$ rpc-cmd-client --socket /tmp/s.sock ls -al
Will execute 'ls' with arguments '-al' on the server side.

If the command being executed does not match a whitelist pattern on the server,
it will exit with code 127 'Command not found.'.

  -socket string
        Path to a unix socket file being listened to by the server.
```

### Example call:

```
./rpc-cmd-client --socket /tmp/sock.sock ls -a -l
```

## Security

This application literally acts as a backdoor, so security is quite important.
A number of design decisions help make this harder to maliciously exploit:

- the server listens on a unix file socket that can only be read by users on the local system, it is not exposed as a network port.
- the server does not require root, it can be run as any user and will execute commands as that user
- the whitelist regex MUST match the entire command and its arguments, substring matches don't match
- the socket file is chmodded to 0600 on server launch and write access must be explicitely granted for other users

This is still a pretty bad idea to run unless you have VERY specific use cases.

## Docker workflow

1. create a directory for the socket file/s

    ```
    $ mkdir /home/user/sockets
    ```

2. start the server

    ```
    $ ./rpc-cmd-server --socket /home/user/sockets/cmd-rpc.sock --whitelist /home/user/my-cmd-rpc-whitelist
    ...
    ```

3. start the docker container with the socket mounted inside

    ```
    $ docker run --rm -ti -v /home/user/sockets:/sockets example-container /rpc-cmd-client --socket /sockets/cmd-rpc.sock touch /home/user/markerfile
    ```

This should result in the file `/home/user/markerfile` being created.
