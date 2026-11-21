XP Runners
==========
[![BSD License](https://raw.githubusercontent.com/xp-framework/web/master/static/licence-bsd.png)](https://github.com/xp-runners/evolution/blob/master/LICENSE.md)


This is the evolution of the [XP runners written in C#](https://github.com/xp-runners/reference), written in Go language.

## Build and test

```sh
export CGO_ENABLED=0
go test ./...

# For the current OS
go build -o xp .

# Cross-compilation
GOOS=linux GOARCH=amd64 go build -o xp .
GOOS=darwin GOARCH=arm64 go build -o xp .
GOOS=windows GOARCH=amd64 go build -o xp.exe .
```