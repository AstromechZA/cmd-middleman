FROM alpine:3.3

COPY rpc-cmd-client /rpc-cmd-client

CMD ["/rpc-cmd-client", "--socket", "/sockets/rpc.sock", "ls", "-al", "/"]
