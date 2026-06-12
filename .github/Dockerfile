# =============================================================================
#  Test Container for Various Go Versions
# =============================================================================
# Default version: golang:alpine
ARG VARIANT="alpine"

# -----------------------------------------------------------------------------
#  Main Stage
# -----------------------------------------------------------------------------
FROM golang:${VARIANT}

RUN apk add --no-cache \
    git \
    gcc musl-dev \
    make \
    && \
    wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest

WORKDIR /workspaces

ENTRYPOINT [ "make", "unit_test" ]
