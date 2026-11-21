#!/usr/bin/env bash
set -euo pipefail

ORIGIN=$(pwd)
TARGET=$ORIGIN/$1
VERSION=${GITHUB_REF#refs/tags/v*}
ldflags="-s -w -X main.version=$VERSION"

mkdir -p $TARGET
rm -rf $TARGET/*

echo "=> One-line installer"
cat .github/scripts/setup.sh.in | sed -e "s/@VERSION@/$VERSION/g" > $TARGET/setup-${VERSION}.sh
ls -al $TARGET/setup-${VERSION}.sh

echo
echo "=> Linux (amd64)"
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$ldflags" -o xp
tar cvfz "$TARGET/xp-linux-amd64-$VERSION.tar.gz" xp class-main.php web-main.php
ls -al "$TARGET/xp-linux-amd64-$VERSION.tar.gz"

echo
echo "=> Darwin (arm64)"
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "$ldflags" -o xp
tar cvfz "$TARGET/xp-darwin-arm64-$VERSION.tar.gz" xp class-main.php web-main.php
ls -al "$TARGET/xp-darwin-arm64-$VERSION.tar.gz"

echo
echo "=> Windows (amd64)"
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$ldflags" -o "xp.exe" .
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$ldflags" -o "tput.exe" ./cmd/tput
zip -j "$TARGET/xp-windows-amd64-$VERSION" xp.exe tput.exe class-main.php web-main.php
ls -al "$TARGET/xp-windows-amd64-$VERSION"