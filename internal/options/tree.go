package options

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type Options map[string]any

func New() Options {
	return make(Options)
}

func (o Options) Add(path string, val any) error {
	log.Debug().Str("path", path).Any("val", val).Msg("Add value")
	parts := strings.Split(path, ".")
	parent := o
	for i := 0; i < len(parts)-1; i++ {
		var ok bool
		p := parts[i]
		log.Debug().Msgf("Part %d %s", i, p)
		if _, ok = parent[p]; !ok {
			parent[p] = make(Options)
			log.Debug().Any("tree", o).Msgf("Option %s does not exist -> create", p)
			parent = parent[p].(Options)
			continue
		}
		if _, ok := parent[p].(Options); !ok {
			return fmt.Errorf(
				"expected %s in path %s in options tree to be of type Options but received %t",
				p, path, parent[p],
			)
		}
		parent = parent[p].(Options)
	}
	parent[parts[len(parts)-1]] = val
	log.Debug().Any("tree", o).Any("val", val).Msg("Set value")
	return nil
}

func (o Options) JSON() (string, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
