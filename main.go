package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/notnmeyer/helmbake/internal/bake"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "helmbake",
		Short: "bake helm charts by merging values files into a single default",
		RunE:  runBake,
	}

	rootCmd.Flags().StringP("chart", "c", "", "path to the helm chart")
	rootCmd.Flags().StringSliceP("values", "f", nil, "values files to merge (in order, last wins)")
	rootCmd.Flags().StringP("output", "o", "", "output directory for the baked chart (default: current directory)")
	rootCmd.Flags().StringSlice("set", nil, "set individual values (key=value)")
	rootCmd.Flags().String("version", "", "override the chart version in Chart.yaml")
	rootCmd.Flags().String("app-version", "", "override the appVersion in Chart.yaml")
	rootCmd.Flags().Bool("package", false, "package the baked chart into a .tgz archive")
	rootCmd.MarkFlagRequired("chart")
	rootCmd.MarkFlagRequired("values")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runBake(cmd *cobra.Command, args []string) error {
	chart, _ := cmd.Flags().GetString("chart")
	values, _ := cmd.Flags().GetStringSlice("values")
	output, _ := cmd.Flags().GetString("output")
	setVals, _ := cmd.Flags().GetStringSlice("set")
	chartVersion, _ := cmd.Flags().GetString("version")
	appVersion, _ := cmd.Flags().GetString("app-version")
	pkg, _ := cmd.Flags().GetBool("package")

	if output == "" {
		output = "."
	}

	sets, err := parseSetValues(setVals)
	if err != nil {
		return err
	}

	opts := bake.Options{
		ChartPath:    chart,
		ValueFiles:   values,
		OutputDir:    output,
		SetValues:    sets,
		ChartVersion: chartVersion,
		AppVersion:   appVersion,
		Package:      pkg,
	}

	return bake.Run(opts)
}

func parseSetValues(setVals []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, s := range setVals {
		k, v, ok := strings.Cut(s, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --set value %q: must be key=value", s)
		}
		result[k] = v
	}
	return result, nil
}
