# =============================================================================
#  Makefile for testing
# =============================================================================
#  If you have `make` command installed, you can run the tests as below:
#
# - To run the unit test/lint check/static analysis locally (golangci-lint required):
#     $ make test
# - To run benchmark:
#     $ make bench
# - To run the unit tests on vaious Go versions (docker required):
#     $ make test_docker
# =============================================================================

# -----------------------------------------------------------------------------
#  Tests for local run
# -----------------------------------------------------------------------------
test: gen_data unit_test lint coverage

.SILENT:

# gen_data generates test data under ./cl/testdata directory. It contains GiB size
# of data, so don't forget to remove them after finish the test/dev.
gen_data:
	go mod download
	go generate ./...

unit_test: gen_data
	go test -cover -race -coverprofile=coverage.out ./... \
		./... \
		github.com/KEINOS/go-countline/cl/_alt \
		github.com/KEINOS/go-countline/cl/_gen \
		github.com/KEINOS/go-countline/_example/countline

# Config for golangci-lint see: ../.golangci.yml
lint:
	golangci-lint run || exit 1
	golangci-lint run ./cl/_alt || exit 1
	golangci-lint run ./cl/_gen || exit 1
	golangci-lint run ./_example/countline || exit 1

# It will fail if the coverage is not 100%.
coverage: unit_test
	set -euo pipefail
	go tool cover -func=coverage.out | tail -n 1 | grep 100.0% || (echo "Total coverage is not 100.0%"; exit 1)

# Benchmark with various size of data.
#
# Note: `benchstat` is required to run this.
#   $ go install golang.org/x/perf/cmd/benchstat@latest
bench:
	set -eu -o pipefail

	printf "Benchmarking with light weight datas ... "
	go test -benchmem -count 5 -benchtime 10s -bench Benchmark_light ./... > bench.txt
	echo "OK"

	printf "Benchmarking with heavy sized datas ... "
	go test -benchmem -bench Benchmark_heavy ./... >> bench.txt
	echo "OK"

	printf "Benchmarking with a giant size data ... "
	go test -benchmem -count 5 -bench Benchmark_giant ./... | tee -a bench.txt
	echo "OK"

	echo "Benchmark results:"
	benchstat -sort delta bench.txt

# -----------------------------------------------------------------------------
#  Docker installed only tests for various Go versions
# -----------------------------------------------------------------------------
test_docker: build_docker go1_16 go1_17 go1_18 go1_19 go_latest

build_docker:
	set -eu -o pipefail
	echo "[Building docker images]:"
	printf "pulling ... "
	docker pull --quiet golang:1.16-alpine
	printf "pulling ... "
	docker pull --quiet golang:1.17-alpine
	printf "pulling ... "
	docker pull --quiet golang:1.18-alpine
	printf "pulling ... "
	docker pull --quiet golang:1.19-alpine
	printf "pulling ... "
	docker pull --quiet golang:alpine
	printf "building images ... "
	docker compose --file ./.github/docker-compose.yml build --progress quiet
	echo "OK"

go1_16: build_docker
	echo "[Unit testing in Go v1.16]:"
	docker compose --file ./.github/docker-compose.yml run v1_16 || exit 1
	echo "ok ... Go v1.16"

go1_17: build_docker
	echo "[Unit testing in Go v1.17]:"
	docker compose --file ./.github/docker-compose.yml run v1_17 || exit 1
	echo "ok ... Go v1.17"

go1_18: build_docker
	echo "[Unit testing in Go v1.18]:"
	docker compose --file ./.github/docker-compose.yml run v1_18 || exit 1
	echo "ok ... Go v1.18"

go1_19: build_docker
	echo "[Unit testing in Go v1.19]:"
	docker compose --file ./.github/docker-compose.yml run v1_19 || exit 1
	echo "ok ... Go v1.19"

go_latest: build_docker
	echo "[Unit testing in Go latest version]:"
	docker compose --file ./.github/docker-compose.yml run latest || exit 1
	echo "ok ... Go latest version"

# For maintenance
prune:
	printf "prune container ... "
	docker container prune -f
	printf "prune image ... "
	docker image prune -f
	printf "prune volumes ... "
	docker volume prune -f