package binwrapper_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nickalie/go-binwrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSuccess(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("echo")

	err := bin.Run("hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello\n", string(bin.StdOut()))
}

func TestRunNonexistentBinary(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("nonexistent-binary-abc123xyz")

	err := bin.Run()
	assert.Error(t, err)
}

func TestRunWithArgs(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("echo").
		Arg("-n", "hello")

	err := bin.Run("world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(bin.StdOut()))
}

func TestArgs(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		Arg("--flag").
		Arg("-o", "value")

	assert.Equal(t, []string{"--flag", "-o", "value"}, bin.Args())
}

func TestDest(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		Dest("/usr/bin").
		ExecPath("echo")

	assert.Equal(t, filepath.Join("/usr/bin", "echo"), bin.Path())
}

func TestPathWithDotDest(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		Dest(".").
		ExecPath("mybinary")

	path := bin.Path()
	assert.True(t, strings.HasPrefix(path, "."))
	assert.Contains(t, path, "mybinary")
}

func TestPathNoDest(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("mybinary")

	assert.Equal(t, "mybinary", bin.Path())
}

func TestStdErr(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("sh").
		Arg("-c", "echo error >&2")

	err := bin.Run()
	assert.NoError(t, err)
	assert.Equal(t, "error\n", string(bin.StdErr()))
}

func TestCombinedOutput(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("sh").
		Arg("-c", "echo out && echo err >&2")

	err := bin.Run()
	assert.NoError(t, err)

	combined := string(bin.CombinedOutput())
	assert.Contains(t, combined, "out")
	assert.Contains(t, combined, "err")
}

func TestCombinedOutputDoesNotCorruptStdOut(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("sh").
		Arg("-c", "echo out && echo err >&2")

	err := bin.Run()
	require.NoError(t, err)

	stdOutBefore := string(bin.StdOut())
	_ = bin.CombinedOutput()
	stdOutAfter := string(bin.StdOut())

	assert.Equal(t, stdOutBefore, stdOutAfter, "CombinedOutput must not corrupt StdOut")
}

func TestSetStdOut(t *testing.T) {
	var buf bytes.Buffer
	bin := binwrapper.NewBinWrapper().
		ExecPath("echo").
		SetStdOut(&buf)

	err := bin.Run("hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello\n", buf.String())
}

func TestStdIn(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("cat").
		StdIn(strings.NewReader("input data"))

	err := bin.Run()
	assert.NoError(t, err)
	assert.Equal(t, "input data", string(bin.StdOut()))
}

func TestEnv(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("sh").
		Arg("-c", "echo $MY_TEST_VAR").
		Env([]string{"MY_TEST_VAR=test_value"})

	err := bin.Run()
	assert.NoError(t, err)
	assert.Equal(t, "test_value\n", string(bin.StdOut()))
}

func TestReset(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("echo").
		Arg("first")

	err := bin.Run()
	require.NoError(t, err)
	assert.Equal(t, "first\n", string(bin.StdOut()))

	bin.Reset()

	assert.Nil(t, bin.Args())
	assert.Nil(t, bin.StdOut())
	assert.Nil(t, bin.StdErr())

	bin.Arg("second")
	err = bin.Run()
	require.NoError(t, err)
	assert.Equal(t, "second\n", string(bin.StdOut()))
}

func TestTimeout(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("sleep").
		Timeout(50 * time.Millisecond)

	err := bin.Run("10")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestKillBeforeRun(t *testing.T) {
	bin := binwrapper.NewBinWrapper()
	err := bin.Kill()
	assert.NoError(t, err)
}

func TestNonZeroExitCode(t *testing.T) {
	bin := binwrapper.NewBinWrapper().
		ExecPath("sh").
		Arg("-c", "echo output && exit 1")

	err := bin.Run()
	assert.Error(t, err)
	assert.Equal(t, "output\n", string(bin.StdOut()))
}
