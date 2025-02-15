package options

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type AttrSet struct {
	path     string
	tree     *Tree
	children map[string]Item
	root     string
}

func NewAttrSet(root, path string, t *Tree) *AttrSet {
	return &AttrSet{
		path:     path,
		tree:     t,
		root:     root,
		children: make(map[string]Item),
	}
}

func (a *AttrSet) Path() string {
	return a.path
}

func (a *AttrSet) List() ([]*Option, error) {
	var o []*Option
	for k, v := range a.children {
		if v == nil {
			var err error
			v, err = a.tree.parseItem(a.root, toPath(a.path, k))
			if err != nil {
				return nil, err
			}
			a.children[k] = v
		}
		if v == nil {
			log.Fatal().Str("path", a.path+"."+k).Msg("Attribute stil nil")
		}
		opts, err := v.List()
		if err != nil {
			log.Err(err)
			continue
		}
		o = append(o, opts...)
	}
	return o, nil
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
