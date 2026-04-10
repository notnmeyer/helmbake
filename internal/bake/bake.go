package bake

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/notnmeyer/helmbake/internal/merge"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
)

type Options struct {
	ChartPath    string
	ValueFiles   []string
	OutputDir    string
	SetValues    map[string]string
	ChartVersion string
	AppVersion   string
	Package      bool
}

func Run(opts Options) error {
	chartMeta, err := readChartYAML(filepath.Join(opts.ChartPath, "Chart.yaml"))
	if err != nil {
		return err
	}

	merged, err := merge.Files(opts.ValueFiles)
	if err != nil {
		return fmt.Errorf("merging values: %w", err)
	}

	for k, v := range opts.SetValues {
		merge.SetPath(merged, k, v)
	}

	outputChart, err := copyChart(opts.ChartPath, opts.OutputDir, chartMeta.Name)
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
		if err := setChartField(filepath.Join(outputChart, "Chart.yaml"), "version", opts.ChartVersion); err != nil {
			return fmt.Errorf("setting chart version: %w", err)
		}
	}

	if opts.AppVersion != "" {
		if err := setChartField(filepath.Join(outputChart, "Chart.yaml"), "appVersion", opts.AppVersion); err != nil {
			return fmt.Errorf("setting app version: %w", err)
		}
	}

	if opts.Package {
		pkg := action.NewPackage()
		pkg.Destination = opts.OutputDir
		tgzPath, err := pkg.Run(outputChart, nil)
		if err != nil {
			return fmt.Errorf("packaging chart: %w", err)
		}

		if err := os.RemoveAll(outputChart); err != nil {
			return fmt.Errorf("cleaning up unpacked chart: %w", err)
		}

		fmt.Printf("packaged chart: %s\n", tgzPath)
		return nil
	}

	fmt.Printf("baked chart written to %s\n", outputChart)
	return nil
}

func setChartField(chartYAMLPath, field, value string) error {
	data, err := os.ReadFile(chartYAMLPath)
	if err != nil {
		return err
	}

	var chart map[string]any
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return err
	}

	chart[field] = value

	out, err := yaml.Marshal(chart)
	if err != nil {
		return err
	}

	return os.WriteFile(chartYAMLPath, out, 0644)
}

type chartYAML struct {
	Name string `yaml:"name"`
}

func readChartYAML(path string) (chartYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return chartYAML{}, fmt.Errorf("not a valid helm chart (missing Chart.yaml): %s", path)
	}
	var meta chartYAML
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return chartYAML{}, fmt.Errorf("parsing Chart.yaml: %w", err)
	}
	if meta.Name == "" {
		return chartYAML{}, fmt.Errorf("Chart.yaml missing required 'name' field")
	}
	return meta, nil
}

func copyChart(chartPath, outputDir, chartName string) (string, error) {
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
