package gomochi

import (
	"encoding/json"
	"os"
	"strings"
)

var templates = &TemplateConfigSet{}

func normalise(name string) string {
	return strings.Trim(strings.ReplaceAll(strings.ToLower(name), " ", ""), "[]\t")
}

func LoadTemplateConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(templates); err != nil {
		return err
	}

	return nil
}

type TemplateConfigSet struct {
	Templates map[string]*TemplateConfig `json:"templates"`
}

// Get looks for the requested template. If it isn't found it will return nil.
func (set *TemplateConfigSet) Get(name string) *TemplateConfig {
	normal := normalise(name)

	for k, c := range set.Templates {
		if normalise(k) == normal || normalise(c.Name) == normal || normalise(c.ID) == normal {
			return c
		}
	}

	return nil
}

type TemplateConfig struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Fields []*FieldConfig `json:"fields"`
}

// Get will get the field with the requested name. Returns nil if not found.
//
// This is safe to call if TemplateConfig is nil, which allows easily chain
// requests without constantly nil check.
func (t *TemplateConfig) Get(name string) *FieldConfig {
	name = normalise(name)

	if t == nil {
		return nil
	}

	for _, f := range t.Fields {
		if normalise(f.Name) == name || normalise(f.ID) == name {
			return f
		}
	}

	return nil
}

type FieldConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
