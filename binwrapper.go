// Package binwrapper provides an executable wrapper with convenient methods
// for running command line tools and capturing their output.
// Inspired by and partially ported from npm package bin-wrapper: https://github.com/kevva/bin-wrapper
package binwrapper

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// BinWrapper wraps executable and provides convenient methods to interact with
type BinWrapper struct {
	dest     string
	execPath string
	autoExe  bool

	stdErr       []byte
	stdOut       []byte
	stdIn        io.Reader
	stdOutWriter io.Writer

	args    []string
	env     []string
	debug   bool
	cmd     *exec.Cmd
	mu      sync.Mutex
	timeout time.Duration
}

// NewBinWrapper creates BinWrapper instance
func NewBinWrapper() *BinWrapper {
	return &BinWrapper{}
}

// Timeout sets timeout for the command. By default it's 0 (binary will run till end).
func (b *BinWrapper) Timeout(timeout time.Duration) *BinWrapper {
	b.timeout = timeout
	return b
}

// Dest accepts a path which the binary is located in
func (b *BinWrapper) Dest(dest string) *BinWrapper {
	b.dest = dest
	return b
}

// ExecPath define a file to use as the binary
func (b *BinWrapper) ExecPath(execPath string) *BinWrapper {

	if b.autoExe && runtime.GOOS == "windows" && execPath != "" {
		ext := strings.ToLower(filepath.Ext(execPath))

		if ext != ".exe" {
			execPath += ".exe"
		}
	}

	b.execPath = execPath
	return b
}

// AutoExe adds .exe extension for windows executable path
func (b *BinWrapper) AutoExe() *BinWrapper {
	b.autoExe = true
	if b.execPath != "" {
		return b.ExecPath(b.execPath)
	}
	return b
}

// Arg adds command line argument to run the binary with.
func (b *BinWrapper) Arg(name string, values ...string) *BinWrapper {
	values = append([]string{name}, values...)
	b.args = append(b.args, values...)
	return b
}

// Debug enables debug output
func (b *BinWrapper) Debug() *BinWrapper {
	b.debug = true
	return b
}

// Args returns arguments were added with Arg method
func (b *BinWrapper) Args() []string {
	return b.args
}

// Path returns the full path to the binary
func (b *BinWrapper) Path() string {
	if b.dest == "." {
		return b.dest + string(filepath.Separator) + b.execPath
	}

	return filepath.Join(b.dest, b.execPath)
}

// StdIn sets reader to read executable's stdin from
func (b *BinWrapper) StdIn(reader io.Reader) *BinWrapper {
	b.stdIn = reader
	return b
}

// StdOut returns the binary's stdout after Run was called
func (b *BinWrapper) StdOut() []byte {
	return b.stdOut
}

// CombinedOutput returns combined executable's stdout and stderr
func (b *BinWrapper) CombinedOutput() []byte {
	out := make([]byte, len(b.stdOut), len(b.stdOut)+len(b.stdErr))
	copy(out, b.stdOut)
	return append(out, b.stdErr...)
}

// SetStdOut set writer to write executable's stdout
func (b *BinWrapper) SetStdOut(writer io.Writer) *BinWrapper {
	b.stdOutWriter = writer
	return b
}

// Env specifies the environment of the executable.
// If Env is nil, Run uses the current process's environment.
// Elements of env should be in form: "ENV_VARIABLE_NAME=value"
func (b *BinWrapper) Env(env []string) *BinWrapper {
	b.env = env
	return b
}

// StdErr returns the executable's stderr after Run was called
func (b *BinWrapper) StdErr() []byte {
	return b.stdErr
}

// Reset removes arguments, stdin, stdout writer, env, and clears captured stdout/stderr
func (b *BinWrapper) Reset() *BinWrapper {
	b.args = nil
	b.stdOut = nil
	b.stdErr = nil
	b.stdIn = nil
	b.stdOutWriter = nil
	b.env = nil
	b.mu.Lock()
	b.cmd = nil
	b.mu.Unlock()
	return b
}

// Run runs the binary with provided arg list.
// Arg list is appended to args set through Arg method
// Returns context.DeadlineExceeded in case of timeout
func (b *BinWrapper) Run(arg ...string) error {
	arg = append(b.args, arg...)

	if b.debug {
		fmt.Println("BinWrapper.Run: " + b.Path() + " " + strings.Join(arg, " "))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if b.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, b.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, b.Path(), arg...)

	b.mu.Lock()
	b.cmd = cmd
	b.mu.Unlock()

	if b.env != nil {
		cmd.Env = b.env
	}

	if b.stdIn != nil {
		cmd.Stdin = b.stdIn
	}

	var stdout io.Reader

	if b.stdOutWriter != nil {
		cmd.Stdout = b.stdOutWriter
	} else {
		var err error
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	err = cmd.Start()

	if err != nil {
		return err
	}

	if stdout != nil {
		b.stdOut, err = io.ReadAll(stdout)
		if err != nil {
			return fmt.Errorf("failed to read stdout: %w", err)
		}
	}

	b.stdErr, err = io.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("failed to read stderr: %w", err)
	}

	err = cmd.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		return context.DeadlineExceeded
	}

	return err
}

// Kill terminates the process
func (b *BinWrapper) Kill() error {
	b.mu.Lock()
	cmd := b.cmd
	b.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Kill()
	}

	return nil
}
