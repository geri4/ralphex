FROM umputun/baseimage:buildgo-latest as builder

ARG GIT_BRANCH
ARG GITHUB_SHA
ARG CI

ENV GOFLAGS="-mod=vendor"

ADD . /build
WORKDIR /build

RUN go build -o /build/ralphex -ldflags "-X main.revision=${GIT_BRANCH}-${GITHUB_SHA:0:7}-$(date +%Y%m%dT%H%M%S) -s -w" ./cmd/ralphex

FROM umputun/baseimage:scratch-latest

COPY --from=builder /build/ralphex /srv/ralphex

CMD ["/srv/ralphex"]
