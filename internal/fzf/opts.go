package fzf

import (
	"fmt"
)

type opts struct {
	cwd             string
	ansi            bool
	delim           string
	preview         string
	tabstop         int
	border          bool
	query           string
	reverse         bool
	color           string
	info            string
	padding         int
	withnth         []int
	previewWordWrap bool
	tac             bool
	header          string
	bind            string
}

type Info int

const (
	InfoInline = iota
	InfoHidden
)

func defaultOpts() *opts {
	return &opts{
		cwd:     "",
		ansi:    true,
		delim:   "\x01",
		preview: "",
		tabstop: 2,
		border:  true,
		query:   "",
		reverse: true,
		color: "fg:#D8DEE9,bg:#2E3440,hl:#A3BE8C,fg+:#D8DEE9,bg+:#434C5E," +
			"hl+:#A3BE8C,pointer:#BF616A,info:#4C566A,spinner:#4C566A," +
			"header:#4C566A,prompt:#81A1C1,marker:#EBCB8B",
		info:            "default",
		padding:         1,
		withnth:         []int{},
		previewWordWrap: false,
		tac:             true,
		header:          "",
		bind:            "",
	}
}

type Option = func(o *opts) *opts

// WithInfo changes the info style
func WithInfo(i Info) Option {
	return func(o *opts) *opts {
		switch i {
		case InfoHidden:
			o.info = "hidden"
		case InfoInline:
			o.info = "inline"
		default:
			o.info = "default"
		}
		return o
	}
}

// WithCwd set the current working dir in which fzf is executed
// This is especially needed when zk is not run in the root of th zettelkasten
func WithCwd(cwd string) Option {
	return func(o *opts) *opts {
		o.cwd = cwd
		return o
	}
}

// WithQuery adds a starting query to the options
func WithQuery(q string) Option {
	return func(o *opts) *opts {
		o.query = q
		return o
	}
}

// WithPreview adds a preview command to the options
func WithPreviewCmd(p string) Option {
	return func(o *opts) *opts {
		o.preview = p
		return o
	}
}

// WithShowFields limits the fields that are shown in the list
func WithShowFields(nth []int) Option {
	return func(o *opts) *opts {
		o.withnth = nth
		return o
	}
}

// WithPreviewWordWrap adds word wrapping for preview window
func WithPreviewWordWrap() Option {
	return func(o *opts) *opts {
		o.previewWordWrap = true
		return o
	}
}

// WithHeader adds a header to fzf (it is always set to first)
func WithHeader(h string) Option {
	return func(o *opts) *opts {
		o.header = h
		return o
	}
}

// WithBind binds keys and events to actions
func WithBind(b string) Option {
	return func(o *opts) *opts {
		o.bind = b
		return o
	}
}
// WithReverse revreses the sort order of the items
func WithReverse() Option {
	return func(o *opts) *opts {
		o.reverse = true
		return o
	}
}

func argsFromOpt(o *opts) []string {
	args := []string{}
	if o.ansi {
		args = append(args, "--ansi")
	}
	if o.delim != "" {
		args = append(args, fmt.Sprintf("--delimiter='%s'", o.delim))
	}
	if o.preview != "" {
		args = append(args, fmt.Sprintf("--preview='%s'", o.preview))
	}
	if o.tabstop > 0 {
		args = append(args, fmt.Sprintf("--tabstop=%d", o.tabstop))
	}
	if o.border {
		args = append(args, "--border")
	}
	if o.query != "" {
		args = append(args, fmt.Sprintf("--query='%s'", o.query))
	}
	if o.color != "" {
		args = append(args, fmt.Sprintf("--color='%s'", o.color))
	}
	if o.reverse {
		args = append(args, "--reverse")
	}
	if o.info != "" {
		args = append(args, fmt.Sprintf("--info=%s", o.info))
	}
	if len(o.withnth) > 0 {
		s := "--with-nth="
		for i, nth := range o.withnth {
			if i > 0 {
				s += ","
			}
			s += fmt.Sprintf("%d", nth)
		}

		args = append(args, s)
	}
	// preview window
	pw := ""
	if o.previewWordWrap {
		pw += ":wrap"
	}
	if pw != "" {
		args = append(args, fmt.Sprintf("--preview-window=%s", pw))
	}
	if o.tac {
		args = append(args, "--tac")
	}
	if o.header != "" {
		args = append(args, fmt.Sprintf("--header='%s'", o.header))
		args = append(args, "--header-first")
	}
	if o.bind != "" {
		args = append(args, fmt.Sprintf("--bind='%s'", o.bind))
	}
	return args
}
