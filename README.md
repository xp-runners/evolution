XP Runners
==========
[![Build status](https://github.com/xp-runners/evolution/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/xp-runners/evolution/actions/workflows/test.yml)
[![BSD License](https://raw.githubusercontent.com/xp-framework/web/master/static/licence-bsd.png)](https://github.com/xp-runners/evolution/blob/master/LICENSE.md)


This is the evolution of the [XP runners written in C#](https://github.com/xp-runners/reference), written in Go language.

Getting started
---------------
To install the XP runners, you can choose between the generic installer:

```sh
curl -sSL https://github.com/xp-runners/evolution/releases/download/v10.0.0/setup-10.0.0.sh | sh
```

Then, install the framework and you're all set to go:

```sh
$ composer global require xp-framework/core
# ...

$ xp version
XP 13.0.0-dev { PHP/8.6.0 & Zend/4.6.0 } @ Windows NT SURFACE 10.0 build 26300 (Windows 11) AMD64
Copyright (c) 2001-2026 the XP group
FileSystemCL<$APPDATA/Composer/vendor/xp-framework/core/src/main/php>
FileSystemCL<$APPDATA/Composer/vendor/xp-framework/core/src/test/php>
FileSystemCL<$APPDATA/Composer/vendor/xp-framework/core/src/main/resources>
FileSystemCL<$APPDATA/Composer/vendor/xp-framework/core/src/test/resources>
FileSystemCL<.>
```

Docker
------
Use the following inside your *Dockerfile*. This installs the `xp` runner binary and the entry point PHP files to `/usr/bin`:

```Dockerfile
RUN curl -sSL https://github.com/xp-runners/evolution/releases/download/v10.0.0/xp-linux-amd64-10.0.0.tar.gz | tar xvz -C /usr/bin
```

Commands
--------
The following commands are built in:

* **version** - Displays version and exits - *also `-v`*
* **eval {code}** - Evaluates code - *also `-e {code}`*
* **write {code}** - Evaluates code, writes result to Console - *also `-w {code}`*
* **dump {code}** - Evaluates code, var_dump()s result - *also `-d {code}`*
* **help [{command}]** - Displays help - *also `-?`*
* **run {class}** - Runs a class, class file or xar
* **ar {operation} [{sources}]** - Works with XAR archives
* **list** - Lists available subcommands - built-in, globally installed and locally available

If no command line arguments are given, the help command is run. If command line arguments are given, but no command is passed, the command defaults to *run*.

Local commands
--------------
You can define comands in *composer.json* by using its `scripts` section:

```json
{
  "require": { ... },
  "scripts": {
    "serve": "xp -supervise web -m async -c src/main/etc/dev -a 0.0.0.0:80 com.example.App"
  }
}
```

Typing `xp serve` in the console will run the above command. Any extra arguments will be concatenated to the command line in scripts.

Plugin architecture
-------------------
Libraries may provide commands by adding a vendor binary to their composer.json named xp.{vendor}.{name}[.{command}] - for the [test](https://github.com/xp-framework/unittest/blob/master/bin/xp.xp-framework.unittest.test) command, the entry is:

```json
{
  "bin": ["bin/xp.xp-framework.unittest.test"]
}
```

The `xp.{module}.{ucfirst(command)}Runner` class serves as the entry point (if the command name is omitted, it will simply be `xp.{module}.Runner`).

By installing the package globally, it becomes available in any directory.

```sh
$ composer global require xp-framework/unittest
# ...

$ xp test -e '$this->assertTrue(true)'
[.]

♥: 1/1 run (0 skipped), 1 succeeded, 0 failed
Memory used: 1265.94 kB (1416.26 kB peak)
Time taken: 0.000 seconds
```

*XP Runners know about Composer's filesystem locations and its dependency management. Thus, the above is the short form of `./xp.exe -m /path/to/unittest [-m /path/to/dependencies/of/unittest ...] xp.unittest.Runner ...`*

Execution
---------
There are four execution models:

1. Run once, exit (*this is the default*)
2. Watch a directory, spawn process every time a change occurs (*via `-watch {directory}`*)
3. Start the process, keep it running until exits with exitcode 0 or the user presses enter (*via `-supervise`*)
4. Repeated runs, cron-like (*via `-repeat {schedule}`*)

The second model is useful e.g. to implement continuous unittest running, see [here](https://github.com/xp-framework/xp-runners/pull/24).

* * *

Development
-----------

### Build and test

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

### Terminal utility for Windows

Build the standalone `tput` with `go build -o tput.exe ./cmd/tput`. It supports `lines`, `cols`, and `bel`, reads one capability per line with `-S`, and prints the runner version with `-V`. When no console size is available, `lines` and `cols` print `0`.
