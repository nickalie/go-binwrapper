# go-binwrapper

[![Go Reference](https://pkg.go.dev/badge/github.com/nickalie/go-binwrapper.svg)](https://pkg.go.dev/github.com/nickalie/go-binwrapper)
[![CI](https://github.com/nickalie/go-binwrapper/actions/workflows/ci.yml/badge.svg)](https://github.com/nickalie/go-binwrapper/actions/workflows/ci.yml)

A lightweight Go wrapper around command-line executables. Provides a fluent API for argument management, stdin/stdout/stderr handling, timeout support, environment configuration, and process lifecycle control.

Inspired by the npm package [bin-wrapper](https://github.com/kevva/bin-wrapper).

## Install

```
go get github.com/nickalie/go-binwrapper
```

## Usage

### Basic execution

```go
bin := binwrapper.NewBinWrapper().ExecPath("echo")

err := bin.Run("hello", "world")
fmt.Println(string(bin.StdOut())) // "hello world\n"
```

### Specifying binary location

Use `Dest` to set the directory containing the binary. If omitted, the executable is looked up in `PATH`.

```go
bin := binwrapper.NewBinWrapper().
    Dest("/usr/local/bin").
    ExecPath("mytool")
```

### Pre-configured arguments

Arguments added with `Arg` are prepended to every `Run` call.

```go
bin := binwrapper.NewBinWrapper().
    ExecPath("echo").
    Arg("-n", "hello")

bin.Run("world")
fmt.Println(string(bin.StdOut())) // "hello world"
```

### Capturing stderr and combined output

```go
bin := binwrapper.NewBinWrapper().
    ExecPath("sh").
    Arg("-c", "echo out && echo err >&2")

bin.Run()

fmt.Println(string(bin.StdOut()))         // "out\n"
fmt.Println(string(bin.StdErr()))         // "err\n"
fmt.Println(string(bin.CombinedOutput())) // "out\nerr\n"
```

### Providing stdin

```go
bin := binwrapper.NewBinWrapper().
    ExecPath("cat").
    StdIn(strings.NewReader("input data"))

bin.Run()
fmt.Println(string(bin.StdOut())) // "input data"
```

### Redirecting stdout to a writer

When `SetStdOut` is used, output goes directly to the writer instead of being captured in `StdOut()`.

```go
var buf bytes.Buffer
bin := binwrapper.NewBinWrapper().
    ExecPath("echo").
    SetStdOut(&buf)

bin.Run("hello")
fmt.Println(buf.String()) // "hello\n"
```

### Timeouts

Returns `context.DeadlineExceeded` if the command exceeds the timeout.

```go
bin := binwrapper.NewBinWrapper().
    ExecPath("sleep").
    Timeout(2 * time.Second)

err := bin.Run("60")
// err == context.DeadlineExceeded
```

### Environment variables

```go
bin := binwrapper.NewBinWrapper().
    ExecPath("sh").
    Arg("-c", "echo $MY_VAR").
    Env([]string{"MY_VAR=hello"})

bin.Run()
fmt.Println(string(bin.StdOut())) // "hello\n"
```

### Reuse with Reset

`Reset` clears arguments, stdio, environment, and captured output so the wrapper can be reused.

```go
bin := binwrapper.NewBinWrapper().ExecPath("echo")

bin.Run("first")
fmt.Println(string(bin.StdOut())) // "first\n"

bin.Reset()
bin.Run("second")
fmt.Println(string(bin.StdOut())) // "second\n"
```

### Process control

```go
// Kill a running process
bin.Kill()
```

### Windows support

`AutoExe` automatically appends `.exe` to the executable path on Windows.

```go
bin := binwrapper.NewBinWrapper().
    AutoExe().
    ExecPath("mytool") // becomes "mytool.exe" on Windows
```

### Debug mode

```go
bin := binwrapper.NewBinWrapper().
    ExecPath("mytool").
    Debug() // prints the full command before execution
```

## API

| Method | Description |
|---|---|
| `NewBinWrapper()` | Create a new instance |
| `Dest(path)` | Set the directory containing the binary |
| `ExecPath(name)` | Set the executable name |
| `AutoExe()` | Append `.exe` on Windows |
| `Arg(name, values...)` | Add command-line arguments |
| `Args()` | Get current arguments |
| `Path()` | Get the full path to the binary |
| `Run(args...)` | Execute the binary |
| `Kill()` | Terminate the running process |
| `Timeout(d)` | Set execution timeout |
| `StdIn(reader)` | Set stdin source |
| `StdOut()` | Get captured stdout |
| `StdErr()` | Get captured stderr |
| `CombinedOutput()` | Get stdout + stderr combined |
| `SetStdOut(writer)` | Redirect stdout to a writer |
| `Env(vars)` | Set environment variables |
| `Debug()` | Enable debug output |
| `Reset()` | Clear state for reuse |

All setter methods return `*BinWrapper` for chaining.

## License

[MIT](LICENSE)
