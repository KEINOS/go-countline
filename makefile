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
#      make bench_vs_wc ... Compare the countline CLI against `wc -l` on the
#                           1 GiB test file. `hyperfine` is required.
#      make kill ... Kill orphaned benchmark processes left behind by an
#                    aborted `make bench` run.
#      make test_docker ... Run unit tests with different versions of Go in a
#                           Docker container.
#  Note:
#    These tests will generate test data under ./countline/testdata directory. It
#    contains GiB size of data, so don't forget to remove them after finish the
#    test/dev.
# =============================================================================

.SILENT:

# Recipes use bash features such as `set -o pipefail`.
SHELL := bash

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
	set -o pipefail; go tool cover -func=coverage.out | tail -n 1 | grep 100.0% || (echo "Total coverage is not 100.0%"; exit 1)

# -----------------------------------------------------------------------------
#  Benchmarks
# -----------------------------------------------------------------------------
# `make bench` compares all implementations over the light and heavy data
# sets, then prints a benchstat summary. `benchstat` is required:
#
#   $ go install golang.org/x/perf/cmd/benchstat@latest
#
# Progress UX: each suite is repeated BENCH_COUNT times so benchstat can
# aggregate the samples, and every output line is timestamped. A result
# line appears every few seconds (light) or every ~10-30 s (heavy), so a
# moving timestamp means the run is alive, not hung.
#
# Override the knobs for a quick smoke run:
#
#   $ make bench BENCH_COUNT=1 BENCH_TIME_LIGHT=0.1s BENCH_TIME_HEAVY=0.1s
BENCH_COUNT ?= 6
BENCH_TIME_LIGHT ?= 1s
BENCH_TIME_HEAVY ?= 10s
BENCH_OUT ?= bench.txt

# run_bench_suite runs one benchmark suite BENCH_COUNT times, appending each run
# to BENCH_OUT and timestamping every line so progress stays visible (a moving
# timestamp means the run is alive, not hung).
#   $(1) = label   $(2) = extra `go test` flags   $(3) = -bench pattern
define run_bench_suite
	for i in $$(seq 1 $(BENCH_COUNT)); do \
		echo "--- $(1) run $$i/$(BENCH_COUNT) ---" \
			| while IFS= read -r l; do printf '[%s] %s\n' "$$(date +%H:%M:%S)" "$$l"; done; \
		set -o pipefail; \
		go test -run '^$$' -benchmem -count 1 $(2) -bench $(3) ./countline \
			| tee -a $(BENCH_OUT) \
			| while IFS= read -r l; do printf '[%s] %s\n' "$$(date +%H:%M:%S)" "$$l"; done; \
	done
endef

.PHONY: bench
bench: check_install_benchstat gen_data bench_lightweight bench_heavyweight
	echo ""
	echo "Benchmark summary ($(BENCH_COUNT) samples per benchmark):"
	benchstat $(BENCH_OUT)

.PHONY: bench_lightweight
bench_lightweight:
	echo "=== Lightweight benchmarks: Tiny/Small/Medium x 7 implementations ==="
	echo "    $(BENCH_COUNT) runs x 21 benchmarks x $(BENCH_TIME_LIGHT) each (about 2 min with defaults)."
	rm -f $(BENCH_OUT)
	$(call run_bench_suite,light,-benchtime $(BENCH_TIME_LIGHT),Benchmark_light)
	echo "bench_lightweight: done"

.PHONY: bench_heavyweight
bench_heavyweight:
	echo "=== Heavyweight benchmarks: Large/Huge x 7 implementations ==="
	echo "    $(BENCH_COUNT) runs x 14 benchmarks x $(BENCH_TIME_HEAVY) each (about 15 min with defaults)."
	echo "    Pauses of 10-30 s between lines are normal; the timestamps show it is still running."
	$(call run_bench_suite,heavy,-benchtime $(BENCH_TIME_HEAVY),Benchmark_heavy)
	echo "bench_heavyweight: done"

# bench_giant_size is optional and not part of `make bench`: one pass over
# the 1 GiB file takes minutes per implementation.
.PHONY: bench_giant_size
bench_giant_size:
	echo "=== Giant benchmark: 1 GiB x 7 implementations ==="
	echo "    $(BENCH_COUNT) runs; expect minutes of silence per line. Timestamps show progress."
	$(call run_bench_suite,giant,,Benchmark_giant)
	echo "bench_giant_size: done"

# bench_vs_wc compares the countline CLI against the system `wc -l` on the
# 1 GiB test file, using hyperfine. Both warm the OS page cache first, so this
# measures the line-counting work, not cold-disk reads.
#
# Note: requires hyperfine (https://github.com/sharkdp/hyperfine).
GIANT_FILE := countline/testdata/data_Giant.txt
.PHONY: bench_vs_wc
bench_vs_wc: gen_data build
	type hyperfine >/dev/null 2>&1 || { echo "hyperfine is required: https://github.com/sharkdp/hyperfine"; exit 1; }
	hyperfine --warmup 3 --runs 10 \
		--command-name 'countline' './dist/countline $(GIANT_FILE)' \
		--command-name 'wc -l' 'wc -l < $(GIANT_FILE)'

# kill terminates orphaned benchmark processes. Aborting `make bench` (e.g.
# Ctrl-C) can leave the compiled `countline.test` binary running in the
# background, pegging a CPU core. This finds those binaries and kills them.
#
# BENCH_PATTERN matches this project's benchmark test binary only, so it will
# not touch unrelated `go test` runs of other projects.
BENCH_PATTERN := countline\.test.*-test\.bench
.PHONY: kill
kill:
	echo "Searching for orphaned benchmark processes ..."; \
	pids=$$(pgrep -f '$(BENCH_PATTERN)' || true); \
	[ -z "$$pids" ] && { echo "OK ... none found"; exit 0; }; \
	echo "killing: $$pids"; kill $$pids 2>/dev/null || true; sleep 1; \
	pids=$$(pgrep -f '$(BENCH_PATTERN)' || true); [ -n "$$pids" ] && kill -9 $$pids 2>/dev/null || true; echo "OK ... killed"

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
