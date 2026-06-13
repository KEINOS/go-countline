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
#    These tests will generate test data under ./countline/testdata directory. It
#    contains GiB size of data, so don't forget to remove them after finish the
#    test/dev.
# =============================================================================

.SILENT:

.PHONY: check
check: test

.PHONY: all
all: build

.PHONY: clean
clean: prune
	echo "removing dist/ ... "
	rm -rf ./dist/ && echo "OK"
	echo "removing test artifacts ... "
	rm -rf bench.txt coverage.out && echo "OK"

# -----------------------------------------------------------------------------
#  Build application
# -----------------------------------------------------------------------------

.PHONY: build
build:
	mkdir -p ./dist/
	go build -o ./dist/countline ./cmd/countline

# -----------------------------------------------------------------------------
#  Tests for local run
# -----------------------------------------------------------------------------
.PHONY: test
test: gen_data unit_test lint coverage

# gen_data generates test data under ./countline/testdata directory. It contains GiB size
# of data, so don't forget to remove them after finish the test/dev.
.PHONY: gen_data
gen_data:
	go mod download
	go generate ./...

# unit_test will run unit tests with race detector and coverage check.
.PHONY: unit_test
unit_test: gen_data
	go test -cover -race -coverprofile=coverage.out \
		./... \
		github.com/KEINOS/go-countline/countline/_alt \
		github.com/KEINOS/go-countline/countline/_gen

# lint will run lint check and static analysis with golangci-lint.
# For the configuration see: ../.golangci.yml
.PHONY: lint
lint:
	golangci-lint run --fix || exit 1
	golangci-lint run --fix ./countline/_alt || exit 1
	golangci-lint run --fix ./countline/_gen || exit 1

# coverage will fail if the total coverage is not 100%.
.PHONY: coverage
coverage: unit_test
	set -euo pipefail
	go tool cover -func=coverage.out | tail -n 1 | grep 100.0% || (echo "Total coverage is not 100.0%"; exit 1)

# bench will benchmark with various size of data.
#
# Note: `benchstat` is required to run this.
#   $ go install golang.org/x/perf/cmd/benchstat@latest
.PHONY: bench
bench: gen_data bench_lightweight bench_heavyweight
	@echo "Benchmark results:"
	@benchstat -filter ".name:/giant/" bench.txt > bench_giant.txt

.PHONY: bench_lightweight
bench_lightweight:
	@echo "Benchmarking with light weight datas (Tiny/Small/Medium) ..."
	@echo "(21 sub-benchmarks x 6 counts x 1s each = ~2 min. Each result line appears every ~1s)"
	go test -benchmem -count 6 -benchtime 1s -bench Benchmark_light ./... | tee bench.txt
	@echo "bench_lightweight: done"

.PHONY: bench_heavyweight
bench_heavyweight:
	@echo "Benchmarking with heavy sized datas (Large/Huge) ..."
	@echo "(14 sub-benchmarks x 6 counts x 10s each = ~14 min)"
	go test -benchmem -count 6 -benchtime 10s -bench Benchmark_heavy ./... | tee -a bench.txt
	@echo "bench_heavyweight: done"

.PHONY: bench_giant_size
bench_giant_size:
	@echo "Benchmarking with a giant size data (1 GiB) ..."
	go test -benchmem -count 6 -bench Benchmark_giant ./... | tee -a bench.txt
	@echo "bench_giant_size: done"

# -----------------------------------------------------------------------------
#  Docker installed only tests for various Go versions
# -----------------------------------------------------------------------------
.PHONY: test_docker
test_docker: build_docker go_min go_latest

# build_docker will build docker images for testing. It will pre-pull the base
# images for consistency.
.PHONY: build_docker
build_docker: pull_image
	printf "building images ... "
	docker compose --file ./.github/docker-compose.yml build --progress quiet
	echo "OK"

.PHONY: pull_image
pull_image:
	echo "[Building docker images]:"
	printf "pulling ... "
	docker pull --quiet golang:1.26-alpine
	printf "pulling ... "
	docker pull --quiet golang:alpine

.PHONY: go_min
go_min: build_docker
	echo "[Unit testing in Go v1.26(min)]:"
	docker compose --file ./.github/docker-compose.yml run min || exit 1
	echo "ok ... Go v1.26"

.PHONY: go_latest
go_latest: build_docker
	echo "[Unit testing in Go latest version]:"
	docker compose --file ./.github/docker-compose.yml run latest || exit 1
	echo "ok ... Go latest version"

# prune will remove all pruned containers, images and volumes of Docker.
#
# Note: This is for maintenance purpose only. Do not use this unless you know
#       what you are doing.
.PHONY: prune
prune: prune_container prune_image prune_volume

.PHONY: prune_container
prune_container:
	printf "prune container ... "
	docker container prune -f
	echo "OK"

.PHONY: prune_image
prune_image:
	printf "prune image ... "
	docker image prune -f
	echo "OK"

.PHONY: prune_volume
prune_volume:
	printf "prune volumes ... "
	docker volume prune -f
	echo "OK"

# -----------------------------------------------------------------------------
#  Prequisites
# -----------------------------------------------------------------------------

.PHONY: check_install_benchstat
check_install_benchstat:
	# go install "golang.org/x/perf/cmd/benchstat@latest"
	type benchstat >/dev/null 2>&1 || (echo "benchstat is not installed. Please install it first."; exit 1)
