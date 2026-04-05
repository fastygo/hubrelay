package core

import (
	"fmt"
	"sort"
)

type PluginRegistry struct {
	factories map[string]PluginFactory
}

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		factories: make(map[string]PluginFactory),
	}
}

func (r *PluginRegistry) Register(name string, factory PluginFactory) {
	if r == nil || factory == nil || name == "" {
		return
	}
	r.factories[name] = factory
}

func (r *PluginRegistry) BuildAll(ctx PluginFactoryContext) ([]Plugin, error) {
	if r == nil {
		return nil, nil
	}

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)

	var plugins []Plugin
	for _, name := range names {
		factory := r.factories[name]
		items, err := factory(ctx)
		if err != nil {
			return nil, fmt.Errorf("build plugins for %s: %w", name, err)
		}
		plugins = append(plugins, items...)
	}
	return plugins, nil
}
