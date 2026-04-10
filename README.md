# helmbake

helmbake produces deployment-ready Helm charts by merging multiple values files into a single `values.yaml`. Instead of passing `-f base.yaml -f prod.yaml` at deploy time, you bake those overrides into the chart itself.

## Why

Helm's built-in values merging (`helm install -f a.yaml -f b.yaml`) happens at install time. This works, but it means your deploy tooling needs to know which files to pass and in what order. helmbake shifts that merge to build time, producing a self-contained chart that can be installed with no extra flags.

This is useful when:
- You want to publish pre-configured chart variants (one per environment, region, customer, etc.)
- Your deploy pipeline shouldn't need to know about values layering
- You want to inspect the final merged values before deploying

## Install

```
go install github.com/notnmeyer/helmbake@latest
```

## Usage

```
helmbake -c ./mychart -f base.yaml -f prod.yaml
```

This copies `mychart` to the current directory with a merged `values.yaml`. The merge behavior is the same as `helm install`'s.

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--chart` | `-c` | Path to the source Helm chart (required) |
| `--values` | `-f` | Values files to merge, in order (required, repeatable) |
| `--output` | `-o` | Output directory (default: current directory) |
| `--set` | | Set individual values (`key=value`, supports dotted paths like `image.tag=v2`) |
| `--version` | | Override the chart version in Chart.yaml |
| `--package` | | Package the baked chart into a `.tgz` archive |

### Examples

Bake and output an unpacked chart:

```
helmbake -c ./mychart -f base.yaml -f staging.yaml -o ./output
```

Bake, override a value, and package into a `.tgz`:

```
helmbake -c ./mychart -f base.yaml -f prod.yaml --set image.tag=v1.2.3 --version 1.0.0 --package
```

Inspect a packaged chart without extracting it:

```
tar tzf mychart-1.0.0.tgz                          # list files
tar xzf mychart-1.0.0.tgz -O mychart/values.yaml   # view a file
```
