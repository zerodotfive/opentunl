FROM --platform=$BUILDPLATFORM node:24-trixie AS static

ADD ui /build
WORKDIR /build
RUN npm install; npm run build

FROM --platform=$BUILDPLATFORM golang:1.27-trixie AS build

ADD . /build
COPY --from=static /build/dist /build/ui/dist
WORKDIR /build

ARG TARGETOS
ARG TARGETARCH
ARG GITHUB_REF_NAME
RUN go mod tidy;  \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X github.com/zerodotfive/opentunl/internal/info.ApplicationVersion=${GITHUB_REF_NAME}" \
       -o bin/opentunl-server ./cmd/server

FROM scratch AS binary
COPY --from=build /build/bin/opentunl-server /opentunl-server

FROM debian:trixie
COPY --from=build /build/bin/opentunl-server /usr/local/bin/opentunl-server
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl openvpn openvpn-auth-ldap \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /var/cache/apt/archives/*

CMD ["/usr/local/bin/opentunl-server"]
