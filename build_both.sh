#!/usr/bin/env bash

set -e

# unofficial
VERSION="<unofficial build>"

if [[ "$@" == "official" ]]; then
    echo "Building official version."

    # first build the version string
    VERSION=1.0

    # add the git commit id and date
    VERSION="$VERSION (commit $(git rev-parse --short HEAD) @ $(git log -1 --date=short --pretty=format:%cd))"
fi

function buildbinary {
    goos=$1
    goarch=$2

    echo "Building $goos $goarch binary with version $VERSION"

    outputfolder="build/${goos}_${goarch}"
    echo "Output Folder $outputfolder"
    mkdir -pv $outputfolder

    export GOOS=$goos
    export GOARCH=$goarch

    go build -i -v -o "$outputfolder/rpc-cmd-client" -ldflags "-X \"main.Version=$VERSION\"" github.com/AstromechZA/cmd-middleman/client
    go build -i -v -o "$outputfolder/rpc-cmd-server" -ldflags "-X \"main.Version=$VERSION\"" github.com/AstromechZA/cmd-middleman/server

    echo "Done"
    ls -l "$outputfolder"
    echo
}

# build for mac
buildbinary darwin amd64

# build for linux
buildbinary linux amd64
