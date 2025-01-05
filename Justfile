b:
  make build-bacalhau

build:
	goreleaser build --single-target --clean -o bacalhau --snapshot