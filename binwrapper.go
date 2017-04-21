// Binary wrapper that makes command line tools seamlessly available as local golang dependencies.
// Inspired by and partially ported from npm package bin-wrapper: https://github.com/kevva/bin-wrapper
package binwrapper

import (
	"path/filepath"
	"os"
	"errors"
	"os/exec"
	"net/url"
	"strings"
	"net/http"
	"io"
	"github.com/mholt/archiver"
	"fmt"
	"io/ioutil"
	"runtime"
)

type Src struct {
	url      string
	os       string
	arch     string
	execPath string
}

type BinWrapper struct {
	src      []*Src
	dest     string
	execPath string
	strip    int
	output   []byte
	autoExe  bool

	stdErr       []byte
	stdOut       []byte
	stdIn        io.Reader
	stdOutWriter io.Writer

	args  []string
	env   []string
	debug bool
}

//Creates new Src instance
func NewSrc() *Src {
	return &Src{}
}

//Sets a URL pointing to a file to download.
func (s *Src) Url(value string) *Src {
	s.url = value
	return s
}

//Tie the source to a specific OS. Possible values are same as runtime.GOOS
func (s *Src) Os(value string) *Src {
	s.os = value
	return s
}

//Tie the source to a specific arch. Possible values are same as runtime.GOARCH
func (s *Src) Arch(value string) *Src {
	s.arch = value
	return s
}

//Tie the src to a specific binary file
func (s *Src) ExecPath(value string) *Src {
	s.execPath = value
	return s
}

//Creates ready to use BinWrapper instance
func NewBinWrapper() *BinWrapper {
	return &BinWrapper{}
}

//Adds a source to download
func (b *BinWrapper) Src(src *Src) *BinWrapper {
	b.src = append(b.src, src)
	return b
}

//Accepts a path which the files will be downloaded to
func (b *BinWrapper) Dest(dest string) *BinWrapper {
	b.dest = dest
	return b
}

//Define which file to use as the binary
func (b *BinWrapper) ExecPath(execPath string) *BinWrapper {

	if b.autoExe && runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(execPath))

		if ext != ".exe" {
			execPath += ".exe"
		}
	}

	b.execPath = execPath
	return b
}

//Adds .exe extension for windows executable path
func (b *BinWrapper) AutoExe() *BinWrapper {
	b.autoExe = true
	return b.ExecPath(b.execPath)
}

//Skips downloading a file
func (b *BinWrapper) SkipDownload() *BinWrapper {
	b.src = nil
	return b
}

//Strips a number of leading paths from file names on extraction.
func (b *BinWrapper) Strip(value int) *BinWrapper {
	b.strip = value
	return b
}

//Adds command line argument to run the binary with.
func (b *BinWrapper) Arg(name string, values ...string) *BinWrapper {
	values = append([]string{name}, values...)
	b.args = append(b.args, values...)
	return b
}

//Enabled debug output
func (b *BinWrapper) Debug() *BinWrapper {
	b.debug = true
	return b
}

//Returns arguments were added with Arg method
func (b *BinWrapper) Args() []string {
	return b.args
}

//Returns the full path to the binary
func (b *BinWrapper) Path() string {
	src := osFilterObj(b.src)

	if src != nil && src.execPath != "" {
		b.ExecPath(src.execPath)
	}

	if b.dest == "." {
		return b.dest + string(filepath.Separator) + b.execPath
	} else {
		return filepath.Join(b.dest, b.execPath)
	}

}

func (b *BinWrapper) StdIn(reader io.Reader) *BinWrapper {
	b.stdIn = reader
	return b
}

//Returns the binary's standard output after Run was called
func (b *BinWrapper) StdOut() []byte {
	return b.stdOut
}

func (b *BinWrapper) CombinedOutput() []byte {
	return append(b.stdOut, b.stdErr...)
}

func (b *BinWrapper) SetStdOut(writer io.Writer) *BinWrapper {
	b.stdOutWriter = writer
	return b
}

func (b *BinWrapper) Env(env []string) *BinWrapper {
	b.env = env
	return b
}

//Returns the binary's standard error after Run was called
func (b *BinWrapper) StdErr() []byte {
	return b.stdErr
}

//Removes all arguments set with Arg method, cleans StdOut and StdErr
func (b *BinWrapper) Reset() *BinWrapper {
	b.args = []string{}
	b.stdOut = nil
	b.stdErr = nil
	b.stdIn = nil
	b.stdOutWriter = nil
	b.env = nil
	return b
}

//Runs the binary with provided arg list.
//Arg list is appended to args set through Arg method
func (b *BinWrapper) Run(arg ...string) error {

	if b.src != nil && len(b.src) > 0 {
		err := b.findExisting()

		if err != nil {
			return err
		}
	}

	arg = append(b.args, arg...)

	if b.debug {
		fmt.Println("BinWrapper.Run: " + b.Path() + " " + strings.Join(arg, " "))
	}

	cmd := exec.Command(b.Path(), arg...)

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
		stdout, _ = cmd.StdoutPipe()
	}

	stderr, _ := cmd.StderrPipe()

	err := cmd.Start()

	if err != nil {
		return err
	}

	cmd.CombinedOutput()

	if stdout != nil {
		b.stdOut, _ = ioutil.ReadAll(stdout)
	}

	b.stdErr, _ = ioutil.ReadAll(stderr)
	return cmd.Wait()
}

func (b *BinWrapper) findExisting() error {
	_, err := os.Stat(b.Path())

	if os.IsNotExist(err) {
		fmt.Printf("%s not found. Downloading...\n", b.Path())
		return b.download()
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

func (b *BinWrapper) download() error {
	src := osFilterObj(b.src)

	if src == nil {
		return errors.New("No binary found matching your system. It's probably not supported.")
	}

	file, err := b.downloadFile(src.url)

	if err != nil {
		return err
	}

	fmt.Printf("%s downloaded. Trying to extract...\n", file)

	err = b.extractFile(file)

	if err != nil {
		return err
	}

	if src.execPath != "" {
		b.ExecPath(src.execPath)
	}

	return nil
}

func (b *BinWrapper) extractFile(file string) error {
	var arc archiver.Archiver

	for _, v := range archiver.SupportedFormats {
		if v.Match(file) {
			arc = v
			break
		}
	}

	if arc == nil {
		fmt.Printf("%s is not an archive or have unsupported archive format\n", file)
		return nil
	}

	defer os.Remove(file)
	err := arc.Open(file, b.dest)

	if err != nil {
		return err
	}

	if b.strip == 0 {
		return nil
	} else {
		return b.stripDir()
	}
}

func (b *BinWrapper) stripDir() error {
	dir := b.dest

	var dirsToRemove []string

	for i := 0; i < b.strip; i++ {
		files, err := ioutil.ReadDir(dir)

		if err != nil {
			return err
		}

		for _, v := range files {
			if v.IsDir() {

				if dir != b.dest {
					dirsToRemove = append(dirsToRemove, dir)
				}

				dir = filepath.Join(dir, v.Name())
				break
			}
		}
	}

	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return err
	}

	for _, v := range files {
		err := os.Rename(filepath.Join(dir, v.Name()), filepath.Join(b.dest, v.Name()))

		if err != nil {
			return err
		}
	}

	for _, v := range dirsToRemove {
		os.RemoveAll(v)
	}

	return nil
}

func (b *BinWrapper) downloadFile(value string) (string, error) {

	if b.dest == "" {
		b.dest = "."
	}

	err := os.MkdirAll(b.dest, 0755)

	if err != nil {
		return "", err
	}

	fileURL, err := url.Parse(value)

	if err != nil {
		return "", err
	}

	path := fileURL.Path

	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]
	fileName = filepath.Join(b.dest, fileName)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

	if err != nil {
		return "", err
	}

	defer file.Close()

	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := check.Get(value)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)

	return fileName, err
}
