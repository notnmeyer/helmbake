package merge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeepMerge(t *testing.T) {
	tests := []struct {
		name string
		dst  map[string]any
		src  map[string]any
		want map[string]any
	}{
		{
			name: "simple overwrite",
			dst:  map[string]any{"a": "1", "b": "2"},
			src:  map[string]any{"b": "3"},
			want: map[string]any{"a": "1", "b": "3"},
		},
		{
			name: "nested merge",
			dst: map[string]any{
				"top": map[string]any{"a": "1", "b": "2"},
			},
			src: map[string]any{
				"top": map[string]any{"b": "3", "c": "4"},
			},
			want: map[string]any{
				"top": map[string]any{"a": "1", "b": "3", "c": "4"},
			},
		},
		{
			name: "src adds new keys",
			dst:  map[string]any{"a": "1"},
			src:  map[string]any{"b": "2"},
			want: map[string]any{"a": "1", "b": "2"},
		},
		{
			name: "src overwrites non-map with map",
			dst:  map[string]any{"a": "string"},
			src:  map[string]any{"a": map[string]any{"nested": "val"}},
			want: map[string]any{"a": map[string]any{"nested": "val"}},
		},
		{
			name: "empty src returns dst",
			dst:  map[string]any{"a": "1"},
			src:  map[string]any{},
			want: map[string]any{"a": "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeepMerge(tt.dst, tt.src)
			if !mapsEqual(got, tt.want) {
				t.Errorf("DeepMerge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFiles(t *testing.T) {
	dir := t.TempDir()

	base := filepath.Join(dir, "base.yaml")
	os.WriteFile(base, []byte("image:\n  repo: nginx\n  tag: latest\nreplicas: 1\n"), 0644)

	env := filepath.Join(dir, "prod.yaml")
	os.WriteFile(env, []byte("image:\n  tag: v1.2.3\nreplicas: 3\n"), 0644)

	got, err := Files([]string{base, env})
	if err != nil {
		t.Fatalf("Files() error: %v", err)
	}

	image, ok := got["image"].(map[string]any)
	if !ok {
		t.Fatal("expected image to be a map")
	}
	if image["repo"] != "nginx" {
		t.Errorf("image.repo = %v, want nginx", image["repo"])
	}
	if image["tag"] != "v1.2.3" {
		t.Errorf("image.tag = %v, want v1.2.3", image["tag"])
	}
	if got["replicas"] != 3 {
		t.Errorf("replicas = %v, want 3", got["replicas"])
	}
}

func TestSetPath(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		key   string
		value any
		want  map[string]any
	}{
		{
			name:  "top-level key",
			input: map[string]any{},
			key:   "replicas",
			value: "3",
			want:  map[string]any{"replicas": "3"},
		},
		{
			name:  "nested key",
			input: map[string]any{"image": map[string]any{"repo": "nginx"}},
			key:   "image.tag",
			value: "v1.0",
			want:  map[string]any{"image": map[string]any{"repo": "nginx", "tag": "v1.0"}},
		},
		{
			name:  "creates intermediate maps",
			input: map[string]any{},
			key:   "a.b.c",
			value: "deep",
			want:  map[string]any{"a": map[string]any{"b": map[string]any{"c": "deep"}}},
		},
		{
			name:  "overwrites non-map intermediate",
			input: map[string]any{"a": "scalar"},
			key:   "a.b",
			value: "val",
			want:  map[string]any{"a": map[string]any{"b": "val"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPath(tt.input, tt.key, tt.value)
			if !mapsEqual(tt.input, tt.want) {
				t.Errorf("SetPath() result = %v, want %v", tt.input, tt.want)
			}
		})
	}
}

func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		aMap, aOK := av.(map[string]any)
		bMap, bOK := bv.(map[string]any)
		if aOK && bOK {
			if !mapsEqual(aMap, bMap) {
				return false
			}
		} else if av != bv {
			return false
		}
	}
	return true
}
