package bake

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func setupChart(t *testing.T) (chartDir string, baseValues string, envValues string) {
	t.Helper()
	dir := t.TempDir()

	chartDir = filepath.Join(dir, "mychart")
	os.MkdirAll(filepath.Join(chartDir, "templates"), 0755)
	os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: mychart\nversion: 0.1.0\n"), 0644)
	os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte("original: true\n"), 0644)
	os.WriteFile(filepath.Join(chartDir, "templates", "deployment.yaml"), []byte("kind: Deployment\n"), 0644)

	baseValues = filepath.Join(dir, "base.yaml")
	os.WriteFile(baseValues, []byte("image:\n  repo: nginx\n  tag: latest\nreplicas: 1\n"), 0644)

	envValues = filepath.Join(dir, "prod.yaml")
	os.WriteFile(envValues, []byte("image:\n  tag: v1.2.3\nreplicas: 3\n"), 0644)

	return
}

func TestRun(t *testing.T) {
	chartDir, baseValues, envValues := setupChart(t)
	outputDir := t.TempDir()

	err := Run(Options{
		ChartPath:  chartDir,
		ValueFiles: []string{baseValues, envValues},
		OutputDir:  outputDir,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	outputChart := filepath.Join(outputDir, "mychart")

	// chart.yaml should be copied
	if _, err := os.Stat(filepath.Join(outputChart, "Chart.yaml")); err != nil {
		t.Error("Chart.yaml not found in output")
	}

	// templates should be copied
	if _, err := os.Stat(filepath.Join(outputChart, "templates", "deployment.yaml")); err != nil {
		t.Error("templates/deployment.yaml not found in output")
	}

	// values.yaml should contain merged values, not the original
	data, err := os.ReadFile(filepath.Join(outputChart, "values.yaml"))
	if err != nil {
		t.Fatalf("reading output values.yaml: %v", err)
	}

	var vals map[string]any
	if err := yaml.Unmarshal(data, &vals); err != nil {
		t.Fatalf("parsing output values.yaml: %v", err)
	}

	image, ok := vals["image"].(map[string]any)
	if !ok {
		t.Fatal("expected image to be a map")
	}
	if image["repo"] != "nginx" {
		t.Errorf("image.repo = %v, want nginx", image["repo"])
	}
	if image["tag"] != "v1.2.3" {
		t.Errorf("image.tag = %v, want v1.2.3", image["tag"])
	}
	if vals["replicas"] != 3 {
		t.Errorf("replicas = %v, want 3", vals["replicas"])
	}
}

func TestRunWithSetValues(t *testing.T) {
	chartDir, baseValues, _ := setupChart(t)
	outputDir := t.TempDir()

	err := Run(Options{
		ChartPath:  chartDir,
		ValueFiles: []string{baseValues},
		OutputDir:  outputDir,
		SetValues:  map[string]string{"image.tag": "override", "replicas": "5"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "mychart", "values.yaml"))
	if err != nil {
		t.Fatalf("reading output values.yaml: %v", err)
	}

	var vals map[string]any
	if err := yaml.Unmarshal(data, &vals); err != nil {
		t.Fatalf("parsing output values.yaml: %v", err)
	}

	image := vals["image"].(map[string]any)
	if image["tag"] != "override" {
		t.Errorf("image.tag = %v, want override", image["tag"])
	}
	if vals["replicas"] != "5" {
		t.Errorf("replicas = %v, want 5", vals["replicas"])
	}
}

func TestRunInvalidChart(t *testing.T) {
	err := Run(Options{
		ChartPath:  "/nonexistent",
		ValueFiles: []string{"whatever.yaml"},
		OutputDir:  t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error for invalid chart path")
	}
}
