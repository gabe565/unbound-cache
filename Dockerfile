#syntax=docker/dockerfile:1.13

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.6.1 AS xx

FROM --platform=$BUILDPLATFORM golang:1.24.0-alpine AS build
WORKDIR /app

COPY --from=xx / /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Set Golang build envs based on Docker platform string
ARG TARGETPLATFORM
RUN --mount=type=cache,target=/root/.cache \
  CGO_ENABLED=0 xx-go build -ldflags='-w -s' -trimpath -o unbound-cache


FROM mvance/unbound:1.22.0@sha256:76906da36d1806f3387338f15dcf8b357c51ce6897fb6450d6ce010460927e90
COPY --from=build /app/unbound-cache /usr/bin
CMD ["unbound-cache"]
