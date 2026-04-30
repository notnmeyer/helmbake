# helmbake

helmbake packages a Helm chart with a known set of values merged into the chart's default `values.yaml`. Its `--set` and `--values` flags work like their counterparts on `helm upgrade` — install- and upgrade-time overrides still apply on top of the baked defaults.

## Why

Helm's built-in values merging (`helm install -f a.yaml -f b.yaml`) happens at install time. That works, but it means your deploy tooling needs to know which files to pass and in what order. helmbake shifts that merge to build time, producing a chart whose defaults already reflect the values you know about — env, region, customer, etc.

This is useful when:
- You want to publish pre-configured chart variants (one per environment, region, customer, etc.)
- Your deploy pipeline shouldn't need to know about values layering
- You want to inspect the final merged defaults before deploying

## Install

As a Helm plugin (recommended):

```
helm plugin install https://github.com/notnmeyer/helmbake
```

Pin to a specific version with `--version`:

```
helm plugin install https://github.com/notnmeyer/helmbake --version v0.2.0
```

Then invoke as `helm bake`:

```
helm bake -c ./mychart -f base.yaml -f prod.yaml
```

As a standalone binary:

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
| `--app-version` | | Override the appVersion in Chart.yaml |
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
