package parser

import (
	"bufio"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/eblechschmidt/nixcfg/internal/repl"
	"github.com/rs/zerolog/log"
)

var skip = []string{
	"nixosConfigurations.nixserve.config.home-manager.extraSpecialArgs",
}

type Parser struct {
	repl *repl.Repl
}

func NewWithFlake(flake string) (*Parser, error) {
	r, err := repl.NewWithFlake(flake)
	if err != nil {
		return nil, err
	}
	return &Parser{repl: r}, nil
}

func (p *Parser) Parse(path string, recursive bool) (any, error) {
	if slices.Contains(skip, path) {
		return "skipped", nil
	}

	res, err := p.repl.Eval(path)
	if err != nil {
		return res, err
	}

	s := bufio.NewScanner(strings.NewReader(res))
	if !s.Scan() {
		// empty result
		return nil, nil
	}

	// parse multiline attr set
	if s.Text() == "{" {
		attrSet := make(map[string]any)
		for s.Scan() {
			if strings.HasPrefix(s.Text(), "}") {
				continue
			}
			parts := strings.Split(s.Text(), "=")
			if len(parts) != 2 {
				return "", fmt.Errorf("unexpected number of equal signs in line: %s", s.Text())
			}

			varName := strings.Trim(parts[0], " ;")
			v := strings.Trim(parts[1], " ;")

			var val any
			if recursive && ((v == "{ ... }") || (v == "[ ... ]")) {
				val, err = p.Parse(path+"."+varName, recursive)
				if err != nil {
					val = err.Error()
				}
			} else {
				val = parseVal(v)
			}

			attrSet[varName] = val

			log.Debug().
				Str("path", path).
				Str("name", varName).
				Any("val", val).
				Msg("Parsed")
		}
		return attrSet, nil
	}

	// attr set in one line
	if strings.HasPrefix(s.Text(), "{") {
		text := strings.Trim(s.Text(), "{}")

		attrSet := make(map[string]any)

		parts := strings.Split(text, "=")
		if len(parts) != 2 {
			return "", fmt.Errorf("unexpected number of equal signs in line: %s", text)
		}

		varName := strings.Trim(parts[0], " ;")
		v := strings.Trim(parts[1], " ;")

		var val any
		if recursive && ((v == "{ ... }") || (v == "[ ... ]")) {
			val, err = p.Parse(path+"."+varName, recursive)
			if err != nil {
				val = err.Error()
			}
		} else {
			val = parseVal(v)
		}

		attrSet[varName] = val

		log.Debug().
			Str("path", path).
			Str("name", varName).
			Any("val", val).
			Msg("Parsed")

		return attrSet, nil
	}

	// parse list
	if strings.HasPrefix(s.Text(), "[") {
		str := "builtins.length " + path
		log.Debug().Str("eval", str).Msg("Get length of list")
		length, err := p.repl.Eval(str)
		if err != nil {
			log.Err(err).Msg("Error retrieving length of list")
			return nil, fmt.Errorf("error retrieveing length of list: %w", err)
		}
		n, err := strconv.Atoi(length)
		if err != nil {
			log.Err(err).Msg("Error retrieving length of list")
			return nil, fmt.Errorf("error retrieveing length of list: %w", err)
		}

		var list = make([]any, n)
		for i := 0; i < n; i++ {
			val, err := p.Parse(fmt.Sprintf("(builtins.elemAt %s %d)", path, i), recursive)
			if err != nil {
				val = err.Error()
			}
			list[i] = val
		}

		return list, nil
	}

	// any value
	return parseVal(s.Text()), nil
}

func (p *Parser) Close() error {
	return p.repl.Close()
}

func parseVal(val string) any {
	switch val {
	case "[ ]":
		return []string{}
	case "null":
		return nil
	case "true":
		return true
	case "false":
		return false
	case "{ }":
		var v struct{}
		return v
	}
	if strings.HasPrefix(val, "«") {
		return val
	}
	if strings.HasPrefix(val, "\"") {
		return strings.Trim(val, "\"")
	}
	if v, err := strconv.Atoi(val); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(val, 64); err == nil {
		return v
	}

	if strings.HasPrefix(val, "/") || strings.HasPrefix(val, "./") ||
		strings.HasPrefix(val, "../") {
		return val
	}

	log.Error().Str("val", val).Msg("Not implemented")

	return val
}
