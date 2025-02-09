package options

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func Show(flake, option string) (string, error) {
	buf := bytes.Buffer{}
	errBuf := bytes.Buffer{}

	cmd := exec.Command("nixos-option", "--flake", flake, option)

	cmd.Stdout = &buf
	cmd.Stderr = &errBuf

	err := cmd.Run()
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

	return out.String(), nil //render.RenderMD(out.String())
}
