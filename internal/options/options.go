package options

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"

	"github.com/eblechschmidt/nixcfg/internal/render"
	"github.com/rs/zerolog/log"
)

type Option struct {
	Path  string
	Value string
}

func List(flake, option string) (<-chan Option, error) {
	result := make(chan Option)
	cmd := []string{"nixos-option", "-r", "--flake", flake, option}
	log.Debug().Msgf("Run command '%s'", strings.Join(cmd, " "))
	c := exec.Command(cmd[0], cmd[1:]...)

	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = createStdout(stdout, result)
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("Start command")
	err = c.Start()
	if err != nil {
		return nil, err
	}
	go func() {
		err := c.Wait()
		if err != nil {
			log.Err(err)
		}
		stdout.Close()
	}()

	return result, nil
}

func createStdout(stdout io.Reader, result chan Option) error {

	go func() {
		log.Debug().Msg("Start receiving go rountine")
		var err error
		defer close(result)

		s := bufio.NewScanner(stdout)

		s.Split(
			func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				if atEOF && len(data) == 0 {
					return 0, nil, nil
				}
				if i := bytes.IndexByte(data, '\n'); i >= 0 {
					// We have a full newline-terminated line.
					return i + 1, data[0:i], nil
				}
				// If we're at EOF, we have a final, non-terminated line. Return it.
				if atEOF {
					return len(data), data, nil
				}
				// Request more data.
				return 0, nil, nil
			},
		)

		for {
			for s.Scan() {
				if pos := strings.Index(s.Text(), "="); pos >= 0 {
					o := Option{
						Path:  strings.Trim(s.Text()[:pos-1], " "),
						Value: strings.Trim(s.Text()[pos+1:], " ;"),
					}
					result <- o
					// log.Debug().
					// 	Str("path", o.Path).
					// 	Str("value", o.Value).
					// 	Msg("Retrieved option")
				}
			}
			if s.Err() != nil {
				log.Err(err)
				break
			}
		}

		log.Debug().Msg("Receiving go routine stoped")
	}()
	return nil
}

func Show(flake, option string) (string, error) {
	buf := bytes.Buffer{}
	errBuf := bytes.Buffer{}

	cmd := []string{"nixos-option", "--flake", flake, option}
	log.Debug().Msgf("Run command '%s'", strings.Join(cmd, " "))
	c := exec.Command(cmd[0], cmd[1:]...)

	c.Stdout = &buf
	c.Stderr = &errBuf

	err := c.Run()
	if err != nil {
		log.Err(err).Msg(errBuf.String())
		return "", err
	}

	s := bufio.NewScanner(&buf)
	out := strings.Builder{}
	code := false
	for s.Scan() {
		// not an option rather an atribute set of options
		if strings.HasPrefix(s.Text(), "This attribute set contains") {
			// don't show anything
			break
		}
		// any other head line after example
		if code && (!strings.HasPrefix(s.Text(), "  ") || s.Text() == "") {
			out.WriteString("```\n")
			code = false
		}
		if !strings.HasPrefix(s.Text(), "  ") && s.Text() != "" {
			hl := "# " + strings.Trim(s.Text(), ":") + "\n\n"
			out.WriteString(hl)
			// Example headline contains code
			if strings.HasPrefix(s.Text(), "Example:") {
				code = true
				out.WriteString("```nix\n")
			}
			continue
		}
		out.WriteString(s.Text() + "\n")
	}

	if code {
		out.WriteString("```\n")
	}

	return render.RenderMD(out.String())
}
