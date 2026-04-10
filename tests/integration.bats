setup() {
  rm -rf tests/output/example
}

teardown() {
  rm -rf tests/output/example
}

@test "staging bake merges base and staging values" {
  run go run . -c tests/chart -f tests/values/base.yaml -f tests/values/staging.yaml -o tests/output
  [ "$status" -eq 0 ]

  run cat tests/output/example/values.yaml
  [[ "$output" == *"name: myapp"* ]]
  [[ "$output" == *"registry: registry-staging"* ]]
  [[ "$output" == *"tag: v1.0.0-staging"* ]]
}

@test "prod bake merges base and prod values" {
  run go run . -c tests/chart -f tests/values/base.yaml -f tests/values/prod.yaml -o tests/output
  [ "$status" -eq 0 ]

  run cat tests/output/example/values.yaml
  [[ "$output" == *"name: myapp"* ]]
  [[ "$output" == *"registry: registry-prod"* ]]
  [[ "$output" == *"tag: v1.0.0-prod"* ]]
}

@test "--set overrides values from files" {
  run go run . -c tests/chart -f tests/values/base.yaml -f tests/values/staging.yaml --set docker.image.tag=override -o tests/output
  [ "$status" -eq 0 ]

  run cat tests/output/example/values.yaml
  [[ "$output" == *"tag: override"* ]]
  [[ "$output" == *"registry: registry-staging"* ]]
}

@test "--version overrides Chart.yaml version" {
  run go run . -c tests/chart -f tests/values/base.yaml -o tests/output --version 2.0.0
  [ "$status" -eq 0 ]

  run cat tests/output/example/Chart.yaml
  [[ "$output" == *"version: 2.0.0"* ]]
}

@test "fails on missing chart" {
  run go run . -c nonexistent -f tests/values/base.yaml -o tests/output
  [ "$status" -ne 0 ]
}
