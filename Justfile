b:
  make build-bacalhau

build:
	goreleaser build --single-target --clean -o bin/darts1 --snapshot