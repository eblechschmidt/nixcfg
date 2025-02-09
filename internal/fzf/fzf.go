package fzf

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var ErrCancelled = errors.New("cancelled")

// fzf exit codes
var (
	fzfExitInterrupted = 130
	fzfExitNoMatch     = 1
)

type Pipe func(w io.WriteCloser)

type Fzf struct {
	o *opts

	err       error
	selection [][]string

	pipe      io.WriteCloser
	done      chan bool
	cmd       *exec.Cmd
	closeOnce sync.Once
}

func New(opts ...Option) (*Fzf, error) {
	o := defaultOpts()
	for _, opt := range opts {
		o = opt(o)
	}

	bin, err := exec.LookPath("fzf")
	if err != nil {
		return nil, fmt.Errorf(
			"could not locate fzf: please install " +
				"fzf from https://github.com/junegunn/fzf")
	}
	if runtime.GOOS == "windows" && os.Getenv("TERM") != "" {
		bin = filepath.ToSlash(bin) // this is needed for git bash on windows
	}

	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell, err = exec.LookPath("sh")
		if err != nil {
			return nil, fmt.Errorf(
				"Needs sh installed on the machine to work and " +
					"set as SHELL in evironment variables")
		}
	}

	args := argsFromOpt(o)
	fzfCmd := fmt.Sprintf("%s %s", bin, strings.Join(args, " "))

	cmd := exec.Command(shell, "-c", fzfCmd)
	cmd.Stderr = os.Stderr

	if o.cwd != "" {
		cmd.Dir = o.cwd
	}

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	done := make(chan bool)

	f := Fzf{
		o:         o,
		cmd:       cmd,
		pipe:      pipe,
		closeOnce: sync.Once{},
		done:      done,
		selection: [][]string{},
	}

	go func() {
		defer func() {
			close(done)
			f.close()
		}()

		output, err := cmd.Output()

		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			switch {
			case ok && exitErr.ExitCode() == fzfExitInterrupted:
				f.err = fmt.Errorf("canceled")
			case ok && exitErr.ExitCode() == fzfExitNoMatch:
				break
			default:
				f.err = fmt.Errorf("failed run fzf: %v", err)
			}
		}
		f.parseSelection(output)
	}()

	return &f, nil
}

// parseSelection extracts the fields from fzf's output.
func (f *Fzf) parseSelection(out []byte) {
	lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		row := strings.Split(line, f.o.delim)
		// Trim padding
		for i, field := range row {
			row[i] = strings.TrimSpace(field)
		}
		f.selection = append(f.selection, row)
	}
}

// Add appends a new line of fields to fzf input.
func (f *Fzf) Add(fields []string) error {
	line := ""
	for i, field := range fields {
		if i > 0 {
			line += f.o.delim

			if field != "" && f.o.padding > 0 {
				line += strings.Repeat(" ", f.o.padding)
			}
		}
		line += field
	}
	if line == "" {
		return nil
	}

	_, err := fmt.Fprintln(f.pipe, line)
	return err
}

// Selection returns the field lines selected by the user through fzf.
func (f *Fzf) Selection() ([][]string, error) {
	f.close()
	<-f.done
	return f.selection, f.err
}

func (f *Fzf) close() error {
	var err error
	f.closeOnce.Do(func() {
		err = f.pipe.Close()
	})
	return err
}
