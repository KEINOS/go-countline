# =============================================================================
#  Makefile for testing
# =============================================================================
#  This Makefile is for convenience to run tests.
#
#  You need `make` command installed, then you can run the tests as below:
#
#      make test ... Run unit tests, lint check, static analysis and coverage
#                    check.
#      make bench ... Run benchmark. `benchstat` is required to run this. See
#                     the comment in the "bench:" section.
#      make test_docker ... Run unit tests with different versions of Go in a
#                           Docker container.
#  Note:
#    These tests will generate test data under ./cl/testdata directory. It
#    contains GiB size of data, so don't forget to remove them after finish the
#    test/dev.
# =============================================================================

.SILENT:

# -----------------------------------------------------------------------------
#  Tests for local run
# -----------------------------------------------------------------------------
test: gen_data unit_test lint coverage

# gen_data generates test data under ./cl/testdata directory. It contains GiB size
# of data, so don't forget to remove them after finish the test/dev.
gen_data:
	go mod download
	go generate ./...

# unit_test will run unit tests with race detector and coverage check.
unit_test: gen_data
	go test -cover -race -coverprofile=coverage.out \
		./... \
		github.com/KEINOS/go-countline/cl/_alt \
		github.com/KEINOS/go-countline/cl/_gen \
		github.com/KEINOS/go-countline/_example/countline

# lint will run lint check and static analysis with golangci-lint.
# For the configuration see: ../.golangci.yml
lint:
	golangci-lint run || exit 1
	golangci-lint run ./cl/_alt || exit 1
	golangci-lint run ./cl/_gen || exit 1
	golangci-lint run ./_example/countline || exit 1

# coverage will fail if the total coverage is not 100%.
coverage: unit_test
	set -euo pipefail
	go tool cover -func=coverage.out | tail -n 1 | grep 100.0% || (echo "Total coverage is not 100.0%"; exit 1)

# bench will benchmark with various size of data.
#
# Note: `benchstat` is required to run this.
#   $ go install golang.org/x/perf/cmd/benchstat@latest
bench: gen_data
    # go install "golang.org/x/perf/cmd/benchstat@latest"
	type benchstat >/dev/null 2>&1 || (echo "benchstat is not installed. Please install it first."; exit 1)

	set -eu -o pipefail

	printf "Benchmarking with light weight datas ... "
	go test -benchmem -count 6 -benchtime 10s -bench Benchmark_light ./... > bench.txt
	echo "OK"

	printf "Benchmarking with heavy sized datas ... "
	go test -benchmem -count 6 -bench Benchmark_heavy ./... >> bench.txt
	echo "OK"

	printf "Benchmarking with a giant size data ... "
	go test -benchmem -count 6 -bench Benchmark_giant ./... | tee -a bench.txt
	echo "OK"

	echo "Benchmark results:"
	benchstat -filter ".name:/giant/" bench.txt > bench_giant.txt

# -----------------------------------------------------------------------------
#  Docker installed only tests for various Go versions
# -----------------------------------------------------------------------------
test_docker: build_docker go_min go_latest

# build_docker will build docker images for testing. It will pre-pull the base
# images for consistency.
build_docker:
	set -eu -o pipefail
	echo "[Building docker images]:"
	printf "pulling ... "
	docker pull --quiet golang:1.22-alpine
	printf "pulling ... "
	docker pull --quiet golang:alpine
	printf "building images ... "
	docker compose --file ./.github/docker-compose.yml build --progress quiet
	echo "OK"

go_min: build_docker
	echo "[Unit testing in Go v1.22(min)]:"
	docker compose --file ./.github/docker-compose.yml run min || exit 1
	echo "ok ... Go v1.22"

go_latest: build_docker
	echo "[Unit testing in Go latest version]:"
	docker compose --file ./.github/docker-compose.yml run latest || exit 1
	echo "ok ... Go latest version"

# prune will remove all pruned containers, images and volumes of Docker.
#
# Note: This is for maintenance purpose only. Do not use this unless you know
#       what you are doing.
prune:
	printf "prune container ... "
	docker container prune -f
	printf "prune image ... "
	docker image prune -f
	printf "prune volumes ... "
	docker volume prune -f
	echo "OK"
