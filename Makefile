test:
	go test ./... -v -race -count=1 -cover -coverprofile=coverage.txt && go tool cover -func=coverage.txt

format:
	goimports -local "gitlab.diskarte.net" -w ./
	# We need to run `gofmt` with `-s` flag as well (best practice, linters require it).
	# `goimports` doesn't support `-s` flag just yet.
	# For details see https://github.com/golang/go/issues/21476
	gofmt -w -s ./

lint:
	golangci-lint run --deadline=5m -v

# Cgo DNS resolver is used by default because of the issue with Kubernetes
# see the discussion here https://github.com/coinsph/go-skeleton/issues/6
# If you need to use pure Go DNS resolver you can remove `--tags netcgo` from the build
#
# Application must be statically-linked to run in `FROM scratch` container
build:
	go build -o ./bin/utils ./cmd/

build_docker:
	docker build --tag=coinsph/bender:latest --file=docker/local.Dockerfile .
