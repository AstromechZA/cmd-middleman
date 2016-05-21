#!/usr/bin/env bash

set -e

function buildbinary {
    goos=$1
    goarch=$2

    echo "Building $goos $goarch binary"

    outputfolder="build/${goos}_${goarch}"
    echo "Output Folder $outputfolder"
    mkdir -pv $outputfolder

    export GOOS=$goos
    export GOARCH=$goarch

    go build -i -v -o "$outputfolder/rpc-cmd-client" github.com/AstromechZA/cmd-middleman/client
    go build -i -v -o "$outputfolder/rpc-cmd-server" github.com/AstromechZA/cmd-middleman/server

    echo "Done"
    ls -l "$outputfolder"
    echo
}

# build for mac
buildbinary darwin amd64

# build for linux
buildbinary linux amd64
