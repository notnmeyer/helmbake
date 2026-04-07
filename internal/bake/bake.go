package bake

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/notnmeyer/helmbake/internal/merge"
	"gopkg.in/yaml.v3"
)

type Options struct {
	ChartPath    string
	ValueFiles   []string
	OutputDir    string
	SetValues    map[string]string
	ChartVersion string
}

func Run(opts Options) error {
	chartYAML := filepath.Join(opts.ChartPath, "Chart.yaml")
	if _, err := os.Stat(chartYAML); err != nil {
		return fmt.Errorf("not a valid helm chart (missing Chart.yaml): %s", opts.ChartPath)
	}

	merged, err := merge.Files(opts.ValueFiles)
	if err != nil {
		return fmt.Errorf("merging values: %w", err)
	}

	for k, v := range opts.SetValues {
		merge.SetPath(merged, k, v)
	}

	outputChart, err := copyChart(opts.ChartPath, opts.OutputDir)
	if err != nil {
		return fmt.Errorf("copying chart: %w", err)
	}

	mergedYAML, err := yaml.Marshal(merged)
	if err != nil {
		return fmt.Errorf("marshaling merged values: %w", err)
	}

	valuesPath := filepath.Join(outputChart, "values.yaml")
	if err := os.WriteFile(valuesPath, mergedYAML, 0644); err != nil {
		return fmt.Errorf("writing values.yaml: %w", err)
	}

	if opts.ChartVersion != "" {
		if err := setChartVersion(filepath.Join(outputChart, "Chart.yaml"), opts.ChartVersion); err != nil {
			return fmt.Errorf("setting chart version: %w", err)
		}
	}

	fmt.Printf("baked chart written to %s\n", outputChart)
	return nil
}

func setChartVersion(chartYAMLPath, version string) error {
	data, err := os.ReadFile(chartYAMLPath)
	if err != nil {
		return err
	}

	var chart map[string]any
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return err
	}

	chart["version"] = version

	out, err := yaml.Marshal(chart)
	if err != nil {
		return err
	}

	return os.WriteFile(chartYAMLPath, out, 0644)
}

func copyChart(chartPath, outputDir string) (string, error) {
	chartName := filepath.Base(chartPath)
	dest := filepath.Join(outputDir, chartName)

	if err := os.RemoveAll(dest); err != nil {
		return "", err
	}

	if err := copyDir(chartPath, dest); err != nil {
		return "", err
	}

	return dest, nil
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}
