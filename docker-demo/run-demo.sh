#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Building binaries.."
export GOOS=linux
export GOARCH=amd64
go build -i -v -o "$DIR/client-container/rpc-cmd-client" github.com/AstromechZA/cmd-middleman/client
go build -i -v -o "$DIR/server-container/rpc-cmd-server" github.com/AstromechZA/cmd-middleman/server

echo "Building containers.."
docker build -t rpc-cmd-server-container "$DIR/server-container"
docker build -t rpc-cmd-client-container "$DIR/client-container"

echo ""
echo "Launching server.."
docker run -d --name rcs-container rpc-cmd-server-container
echo ""
echo "Launching client.."
docker run --rm --volumes-from rcs-container rpc-cmd-client-container || true
echo ""
echo "Stopping server"
docker stop rcs-container
echo ""
echo "Removing server"
docker rm rcs-container
