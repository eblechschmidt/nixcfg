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
	List() ([]*Option, error)
	Path() string
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
	if t.tree == nil {
		var err error
		t.tree, err = t.parseItem(t.optRoot(), "")
		if err != nil {
			return nil, err
		}
	}
	if path == "" {
		return t.tree, nil
	}
	p := strings.Split(path, ".")
	log.Debug().Strs("path", p).Msg("Get item")
	return t.tree.Get(p)
}

func (t *Tree) List(path string) ([]*Option, error) {
	i, err := t.Get(path)
	if err != nil {
		log.Debug().Str("item", fmt.Sprintf("%+v", i)).Err(err).Msg("Tree.List() error")
		return nil, err
	}
	log.Debug().Str("Item", fmt.Sprintf("%+v", i)).Msg("Tree.List()")
	return i.List()
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
