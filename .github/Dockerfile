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
    bash \
    git \
    gcc musl-dev \
    make

WORKDIR /workspaces

ENTRYPOINT [ "make", "unit_test" ]
