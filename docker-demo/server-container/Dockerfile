FROM alpine:3.3

COPY rpc-cmd-server /rpc-cmd-server
COPY whitelist /whitelist

RUN mkdir /sockets

VOLUME /sockets

CMD ["/rpc-cmd-server", "--whitelist", "/whitelist", "--socket", "/sockets/rpc.sock"]
