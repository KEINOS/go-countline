# =============================================================================
#  Test Container for Vaious Go Versions (Last Updated: 2022-08-05)
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
