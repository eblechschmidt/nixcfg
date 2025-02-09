package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Repl struct {
	cmd     *exec.Cmd
	err     error
	eval    chan string
	result  chan string
	replErr chan string
	close   bool
	mtx     sync.RWMutex
	info    *bytes.Buffer
}

// NewWithFlake returns a new repl using the given flake for evaluation
func NewWithFlake(flake string) (*Repl, error) {
	var err error
	r := Repl{close: false, info: new(bytes.Buffer)}
	r.cmd = exec.Command("nix", "repl", "--expr", fmt.Sprintf("builtins.getFlake \"%s\"", flake))

	err = r.createStdin()
	if err != nil {
		return nil, err
	}

	err = r.createStdout()
	if err != nil {
		return nil, err
	}

	err = r.createStderr()
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Start nxi repl process")
	err = r.cmd.Start()
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Repl) createStdin() error {
	stdin, err := r.cmd.StdinPipe()
	if err != nil {
		return err
	}
	r.eval = make(chan string)
	go func() {
		defer stdin.Close()

		for o := range r.eval {
			log.Debug().Msgf("Evaluating: %s", o)
			_, r.err = io.WriteString(stdin, fmt.Sprintf("%s\n", o))
		}

		log.Debug().Msg("Sending go routine stoped")
	}()
	return nil
}

func (r *Repl) createStdout() error {
	stdout, err := r.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	r.result = make(chan string)
	go func() {
		var err error
		defer stdout.Close()

		s := bufio.NewScanner(stdout)
		var sb strings.Builder
		multiline := false
		for {
			for s.Scan() {
				if s.Text() == "" {
					continue
				}
				if !multiline {
					if s.Text() == "{" || s.Text() == "[" {
						multiline = true
						sb.Reset()
					}
				}
				if multiline {
					sb.WriteString(fmt.Sprintf("%s\n", s.Text()))
					if s.Text() == "}" || s.Text() == "]" {
						r.result <- sb.String()
						multiline = false
						break
					}
					continue
				}
				r.result <- s.Text()
				multiline = false
			}
			if s.Err() != nil {
				log.Err(err)
				break
			}
			if r.closeNow() {
				break
			}
		}

		log.Debug().Msg("Receiving go routine stoped")
	}()
	return nil
}
func (r *Repl) createStderr() error {
	stderr, err := r.cmd.StderrPipe()
	if err != nil {
		return err
	}

	r.replErr = make(chan string)
	go func() {
		// var err error
		defer stderr.Close()

		b := make([]byte, 1024)
		buf := new(bytes.Buffer)
		for {
			n, err := stderr.Read(b)
			if err != nil {
				log.Err(err)
			}
			_, err = buf.Write(b[:n])
			if err != nil {
				log.Err(err)
			}

			if n < 1024 {
				if strings.HasPrefix(buf.String(), "error:") {
					r.replErr <- buf.String()
				} else {
					r.info.WriteString(buf.String())
				}
				buf.Reset()
			}

			if r.closeNow() {
				break
			}
		}

		log.Debug().Msg("Receiving go routine stoped")
	}()
	return nil
}

// Eval returns the result of the evaluation in the nix repl of a given string
func (r *Repl) Eval(expr string) (string, error) {
	start := time.Now()
	r.err = nil
	r.info.Reset()
	r.eval <- expr
	select {
	case result := <-r.result:
		result = strip(result)
		elapsed := time.Since(start)
		if r.err != nil {
			return "", r.err
		}
		log.Debug().Str("Result", result).Msgf("Evaluation done after %s", elapsed)
		// clean non-printable runes
		return result, nil
	case e := <-r.replErr:
		e = strip(e)
		elapsed := time.Since(start)
		if r.err != nil {
			return "", r.err
		}
		log.Error().Str("Error", e).Msgf("Evaluation done after %s", elapsed)
		return e, fmt.Errorf("evaluation error")
	}
}

// closeNow returns true when close is requested to shutdwon receiving go routine
func (r *Repl) closeNow() bool {
	r.mtx.RLock()
	b := r.close
	r.mtx.RUnlock()
	return b
}

// Close stops all runing go routines for process comunication and safely exits
// the nxi repl process
func (r *Repl) Close() error {
	r.mtx.Lock()
	r.close = true
	r.mtx.Unlock()

	close(r.eval)
	err := r.cmd.Wait()
	if err != nil {
		return err
	}
	close(r.result)
	log.Debug().Msg("All buffers flushed. Process stoped.")
	return nil
}

// const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
const ansi = "(?:\\x1B[@-Z\\\\-_]|[\\x80-\\x9A\\x9C-\\x9F]|(?:\\x1B\\[|\\x9B)[0-?]*[ -/]*[@-~])"

var re = regexp.MustCompile(ansi)

// strip remvose ansi color codes from a string
func strip(str string) string {
	return re.ReplaceAllString(str, "")
}
