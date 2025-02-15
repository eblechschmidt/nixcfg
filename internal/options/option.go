package options

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type Option struct {
	path        string
	tree        *Tree
	children    Item
	hasChildren bool
}

func NewOption(path string, t *Tree) *Option {
	return &Option{
		path:        path,
		tree:        t,
		hasChildren: false,
	}
}

func (o *Option) Path() string {
	return o.path
}

func (o *Option) List() ([]*Option, error) {
	if !o.hasChildren {
		return []*Option{o}, nil
	}

	if o.children == nil {
		log.Fatal().Msg("Parsing children not implemented yet")
	}

	return o.children.List()
}

// In case option has children (attribute set of (submodule), list of (submodule))
// we retrieve the chidlren from that Get
func (o *Option) Get(p []string) (Item, error) {
	if !o.hasChildren {
		return nil, nil
	}

	if o.children == nil {
		log.Fatal().Msg("Parsing children not implemented yet")
	}

	return o.children.Get(p)
}

func (o *Option) ValueStr() string {
	val, err := o.tree.repl.Eval(toPath(o.tree.cfgRoot(), o.path))
	log.Debug().Err(err).Str("path", o.path).Msg("Option.ValueStr()")
	if err != nil {
		return ""
	}
	return val
}

func (o *Option) Default() string {
	val, err := o.tree.repl.Eval(toPath(o.tree.optRoot(), o.path, "default"))
	if err != nil {
		log.Debug().Err(err).Str("path", o.path).Msg("Option.Default()")
		return ""
	}
	return val
}

func (o *Option) Description() string {
	val, err := o.tree.repl.Eval(toPath(o.tree.optRoot(), o.path, "description"))
	if err != nil {
		log.Debug().Err(err).Str("path", o.path).Msg("Option.Description()")
		return ""
	}
	return val
}

func (o *Option) Type() string {
	val, err := o.tree.repl.Value(toPath(o.tree.optRoot(), o.path, "type.description"))
	if err != nil {
		log.Debug().Err(err).Str("path", o.path).Msg("Option.Type()")
		return ""
	}
	return val.(string)
}

type Declaration struct {
	File         string
	Column, Line int
}

func (o *Option) DeclaredBy() []Declaration {
	expr := toPath(o.tree.optRoot(), o.path, "declarationPositions")
	n, err := o.tree.repl.Length(expr)
	if err != nil {
		return []Declaration{}
	}
	var dec []Declaration
	for i := 0; i < n; i++ {
		col, ok := o.getSubValInt(expr, i, "column")
		if !ok {
			continue
		}
		line, ok := o.getSubValInt(expr, i, "line")
		if !ok {
			continue
		}
		file, ok := o.getSubValStr(expr, i, "file")
		if !ok {
			continue
		}
		dec = append(dec, Declaration{
			Column: col, Line: line, File: file,
		})
	}
	return dec
}

type Definition struct {
	File  string
	Value any
}

func (o *Option) DefinedBy() []Definition {
	expr := toPath(o.tree.optRoot(), o.path, "definitionsWithLocations")
	n, err := o.tree.repl.Length(expr)
	if err != nil {
		return []Definition{}
	}
	var def []Definition
	for i := 0; i < n; i++ {
		val := o.getSubValAny(expr, i, "value")
		file, ok := o.getSubValStr(expr, i, "file")
		if !ok {
			continue
		}
		def = append(def, Definition{
			File: file, Value: val,
		})
	}
	return def
}

func (o *Option) getSubValAny(expr string, i int, sub string) any {
	expr = fmt.Sprintf("(builtins.elemAt (%s) %d).%s", expr, i, sub)
	val, err := o.tree.repl.Value(expr)
	if err != nil {
		log.Debug().
			Str("expr", expr).
			Err(err).
			Msg("could not get value")
		return nil
	}
	return val
}
func (o *Option) getSubValInt(expr string, i int, sub string) (int, bool) {
	expr = fmt.Sprintf("(builtins.elemAt (%s) %d).%s", expr, i, sub)
	val, err := o.tree.repl.Value(expr)
	if err != nil {
		log.Debug().
			Str("expr", expr).
			Err(err).
			Msg("could not get value")
		return -1, false
	}
	if i, ok := val.(int); ok {
		return i, true
	}
	log.Debug().
		Str("expr", expr).
		Msg("could not convert expr to int")
	return -1, false
}
func (o *Option) getSubValStr(expr string, i int, sub string) (string, bool) {
	expr = fmt.Sprintf("(builtins.elemAt (%s) %d).%s", expr, i, sub)
	val, err := o.tree.repl.Value(expr)
	if err != nil {
		log.Debug().
			Str("expr", expr).
			Err(err).
			Msg("could not get value")
		return "", false
	}
	if s, ok := val.(string); ok {

		return s, true
	}
	log.Debug().
		Str("expr", expr).
		Msg("could not convert expr to int")
	return "", false
}
