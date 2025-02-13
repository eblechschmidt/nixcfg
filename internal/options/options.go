package options

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/eblechschmidt/nixcfg/internal/repl"
	"github.com/rs/zerolog/log"
)

type Item interface {
	// ListOptions(r *repl.Repl) []string
	Get(path []string) (Item, error)
}

type Tree struct {
	tree Item
	repl *repl.Repl
	host string
}

func NewTree(r *repl.Repl, hostname string) *Tree {
	return &Tree{
		repl: r,
		tree: nil,
		host: hostname,
	}
}

func (t *Tree) optRoot() string {
	return fmt.Sprintf("nixosConfigurations.%s.options", t.host)
}
func (t *Tree) cfgRoot() string {
	return fmt.Sprintf("nixosConfigurations.%s.config", t.host)
}

func (t *Tree) Get(path string) (Item, error) {
	p := strings.Split(path, ".")
	log.Debug().Strs("path", p).Msg("Get item")
	if t.tree == nil {
		var err error
		t.tree, err = t.parseItem(t.optRoot(), "")
		if err != nil {
			return nil, err
		}
	}
	return t.tree.Get(p)
}

func (t *Tree) parseItem(root, path string) (Item, error) {
	expr := toPath(root, path)
	log.Debug().Str("root", root).Str("path", path).Str("expr", expr).Msg("parseItem")
	res, err := t.repl.Eval(expr)
	if err != nil {
		return nil, fmt.Errorf("could not parse item: %w", err)
	}

	s := bufio.NewScanner(strings.NewReader(res))
	if !s.Scan() {
		// empty result
		return nil, nil
	}

	// parse multiline attr set
	if s.Text() == "{" {
		attr := NewAttrSet(root, path, t)
		for s.Scan() {
			if strings.HasPrefix(s.Text(), "}") {
				continue
			}
			parts := strings.Split(s.Text(), "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("unexpected number of equal signs in line: %s", s.Text())
			}

			varName := strings.Trim(parts[0], " ;")
			valStr := strings.Trim(parts[1], " ;")

			if varName == "_type" && valStr == "\"option\"" {
				return NewOption(path, t), nil
			}

			if strings.HasPrefix(varName, "_") {
				// skip
				continue
			}

			attr.Set(varName, nil)

			log.Debug().
				Str("path", path).
				Str("name", varName).
				Msg("Parsed")
		}
		return attr, nil
	}

	return nil, nil
}

type AttrSet struct {
	path     string
	tree     *Tree
	children map[string]Item
	root     string
}

func (a *AttrSet) Set(attr string, val Item) {
	if _, ok := a.children[attr]; ok {
		if val != nil {
			a.children[attr] = val
		}
		return
	}
	a.children[attr] = val
}

func (a *AttrSet) Get(p []string) (Item, error) {
	if len(p) == 0 {
		return nil, nil
	}
	i, ok := a.children[p[0]]
	if !ok {
		if len(p) > 1 {
			return nil, fmt.Errorf("attribute %s.<<%s>>.%s missing", a.path, p[0], strings.Join(p[1:], "."))
		}
		return nil, fmt.Errorf("%s.<<%s>>", a.path, p[0])
	}
	if i == nil {
		log.Debug().Str("root", a.root).Str("path", a.path).Str("p[0]", p[0]).Str("toPath", toPath(a.path, p[0])).Msg("AttrSet.Get() > parseItem")
		val, err := a.tree.parseItem(a.root, toPath(a.path, p[0]))
		if err != nil {
			return nil, err
		}
		a.children[p[0]] = val
	}
	if len(p) > 1 {
		return a.children[p[0]].Get(p[1:])
	}
	return a.children[p[0]], nil
}

func NewAttrSet(root, path string, t *Tree) *AttrSet {
	return &AttrSet{
		path:     path,
		tree:     t,
		root:     root,
		children: make(map[string]Item),
	}
}

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

func toPath(elem ...string) (path string) {
	i := 0
	for _, e := range elem {
		if e != "" {
			if i > 0 {
				path = path + "."
			}
			path = path + e
			i++
		}
	}
	return
}

// type AttrSet struct {
// 	optInfo
// 	children map[string]Item
// }

// func (a *AttrSet) ListOptions(r *repl.Repl) []string {
// 	var list []string
// 	for k, v := range a.children {
// 		val := v
// 		if _, ok := v.(*ToEval); ok {
// 			val = parse(r, a.path+"."+k)
// 			a.children[k] = val
// 		}
// 		for _, o := range val.ListOptions(r) {
// 			list = append(list, k+"."+o)
// 		}
// 	}
// 	return list
// }

// func (a *AttrSet) Get(path []string) (Item, error) {
// 	return nil
// }

// type List struct {
// 	optInfo
// 	children []Item
// }

// func (l *List) ListOptions(r *repl.Repl) []string {
// 	var list []string
// 	for i, v := range l.children {
// 		val := v
// 		if _, ok := v.(*ToEval); ok {
// 			val = parse(r, fmt.Sprintf("builtins.elemAt %s %d", l.path, i))
// 			l.children[i] = val
// 		}
// 		for _, o := range val.ListOptions(r) {
// 			list = append(list, fmt.Sprintf(l.path+".[%d].%s", i, o))
// 		}

// 	}
// 	return list
// }

// type Option struct {
// 	optInfo
// 	Value       any
// 	Default     any
// 	Type        string
// 	Description string
// 	DeclaredBy  []string
// 	DefinedBy   []string
// }

// func (o *Option) ListOptions(r *repl.Repl) []string {
// 	return []string{""}
// }

// func (e *ToEval) ListOptions(r *repl.Repl) []string {
// 	panic("ListOptions should not be called on ToEval")
// }

// func (t OptionTree) ListOptions() []string {
// 	if _, ok := t.tree.(*ToEval); ok {
// 		t.tree = parse(t.repl, t.root)
// 	}
// 	opts := t.tree.ListOptions(t.repl)
// 	for i := range opts {
// 		opts[i] = strings.Trim(opts[i], ".")
// 	}
// 	return opts
// }

// func (t OptionTree) getOptions(parent any) {}

// func parse(r *repl.Repl, path string) Item {
// 	res, err := r.Eval(path)
// 	if err != nil {
// 		log.Err(err).Str("res", res)
// 		return nil
// 	}

// 	log.Debug().Str("res", res).Msg("Evaluation done")

// 	s := bufio.NewScanner(strings.NewReader(res))
// 	if !s.Scan() {
// 		// empty result
// 		return nil
// 	}

// 	// parse multiline attr set
// 	if s.Text() == "{" {
// 		attrSet := make(map[string]Item)
// 		for s.Scan() {
// 			if strings.HasPrefix(s.Text(), "}") {
// 				continue
// 			}
// 			parts := strings.Split(s.Text(), "=")
// 			if len(parts) != 2 {
// 				log.Err(
// 					fmt.Errorf("unexpected number of equal signs in line: %s", s.Text()),
// 				)
// 				return nil
// 			}

// 			varName := strings.Trim(parts[0], " ;")
// 			valStr := strings.Trim(parts[1], " ;")

// 			if varName == "_type" && valStr == "\"option\"" {
// 				return parseOption(r, path)
// 			}

// 			if strings.HasPrefix(varName, "_") {
// 				// skip
// 				continue
// 			}

// 			var val Item
// 			if (valStr == "{ ... }") || (valStr == "[ ... ]") {
// 				val = &ToEval{}
// 			}

// 			attrSet[varName] = val

// 			log.Debug().
// 				Str("path", path).
// 				Str("name", varName).
// 				Any("val", val).
// 				Msg("Parsed")
// 		}
// 		a := &AttrSet{children: attrSet}
// 		a.path = path
// 		return a
// 	}

// 	// any value
// 	return &ToEval{}
// }

// func parseOption(r *repl.Repl, path string) *Option {
// 	o := Option{
// 		Value:       "",
// 		Default:     getVal(r, path+".default"),
// 		Type:        getVal(r, path+".type.description").(string),
// 		Description: getVal(r, path+".description").(string),
// 		DeclaredBy:  getDeclaredBy(r, path+".declarations"),
// 		DefinedBy:   getDefinedBy(r, path+".definitionsWithLocations"),
// 	}
// 	o.path = path
// 	return &o
// }

// func getVal(r *repl.Repl, path string) any {
// 	res, err := r.Eval(path)
// 	if err != nil {
// 		log.Err(err).Str("res", res)
// 		return nil
// 	}
// 	return parseVal(res)
// }

// func getDeclaredBy(r *repl.Repl, path string) []string {
// 	str := "builtins.length " + path
// 	log.Debug().Str("eval", str).Msg("Get length of list")
// 	length, err := r.Eval(str)
// 	if err != nil {
// 		log.Err(err).Msg("Error retrieving length of list")
// 		return nil
// 	}
// 	n, err := strconv.Atoi(length)
// 	if err != nil {
// 		log.Err(err).Msg("Error retrieving length of list")
// 		return nil
// 	}

// 	var list = make([]string, n)
// 	for i := 0; i < n; i++ {
// 		val, ok := getVal(r, fmt.Sprintf("(builtins.elemAt %s %d)", path, i)).(string)
// 		if !ok {
// 			val = err.Error()
// 		}
// 		list[i] = val
// 	}
// 	return list
// }

// func getDefinedBy(r *repl.Repl, path string) []string {
// 	str := "builtins.length " + path
// 	log.Debug().Str("eval", str).Msg("Get length of list")
// 	length, err := r.Eval(str)
// 	if err != nil {
// 		log.Err(err).Msg("Error retrieving length of list")
// 		return nil
// 	}
// 	n, err := strconv.Atoi(length)
// 	if err != nil {
// 		log.Err(err).Msg("Error retrieving length of list")
// 		return nil
// 	}

// 	var list = make([]string, n)
// 	for i := 0; i < n; i++ {
// 		val := getVal(r, fmt.Sprintf("(builtins.elemAt %s %d).file", path, i)).(string)
// 		if err != nil {
// 			val = err.Error()
// 		}
// 		list[i] = val
// 	}
// 	return list

// }
