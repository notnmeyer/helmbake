package merge

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// later files take precedence over earlier ones
func Files(paths []string) (map[string]any, error) {
	result := make(map[string]any)

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		var vals map[string]any
		if err := yaml.Unmarshal(data, &vals); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		result = DeepMerge(result, vals)
	}

	return result, nil
}

// when both values are maps they are merged recursively, otherwise src wins
func DeepMerge(dst, src map[string]any) map[string]any {
	out := make(map[string]any, len(dst))
	for k, v := range dst {
		out[k] = v
	}

	for k, srcVal := range src {
		dstVal, exists := out[k]
		if !exists {
			out[k] = srcVal
			continue
		}

		dstMap, dstOK := dstVal.(map[string]any)
		srcMap, srcOK := srcVal.(map[string]any)
		if dstOK && srcOK {
			out[k] = DeepMerge(dstMap, srcMap)
		} else {
			out[k] = srcVal
		}
	}

	return out
}

// sets a value at a dotted path (e.g. "image.tag" -> {image: {tag: value}})
func SetPath(m map[string]any, key string, value any) {
	parts := strings.Split(key, ".")
	current := m
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		current = next
	}
	current[parts[len(parts)-1]] = value
}
