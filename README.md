# Golang Binary Wrapper

[![Go Reference](https://pkg.go.dev/badge/github.com/nickalie/go-binwrapper.svg)](https://pkg.go.dev/github.com/nickalie/go-binwrapper)
[![CI](https://github.com/nickalie/go-binwrapper/actions/workflows/ci.yml/badge.svg)](https://github.com/nickalie/go-binwrapper/actions/workflows/ci.yml)

Provides a lightweight wrapper around command line executables, offering argument management,
stdin/stdout/stderr handling, timeout support, and environment configuration.

Inspired by npm package [bin-wrapper](https://github.com/kevva/bin-wrapper)

## Install

```
go get github.com/nickalie/go-binwrapper
```

## Example of usage

Run a command and capture output:

```go
package main

import (
	"fmt"
	"github.com/nickalie/go-binwrapper"
)

func main() {
	bin := binwrapper.NewBinWrapper().
		Dest("/usr/local/bin").
		ExecPath("echo")

	err := bin.Run("hello", "world")

	fmt.Printf("stdout: %s\n", string(bin.StdOut()))
	fmt.Printf("stderr: %s\n", string(bin.StdErr()))
	fmt.Printf("err: %v\n", err)
}
```

Use `Dest` to specify directory with binary:

```go
bin := binwrapper.NewBinWrapper().
    Dest("/path/to/directory").
    ExecPath("mytool")
```

If `Dest` is omitted, the executable is looked up in `PATH`:

```go
bin := binwrapper.NewBinWrapper().
    ExecPath("mytool")
```
