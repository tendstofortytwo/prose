FROM golang:bookworm as builder

WORKDIR /
ENV SASS_VERSION=1.72.0
ENV SASS_TARFILE="dart-sass-${SASS_VERSION}-linux-x64.tar.gz"
ADD "https://github.com/sass/dart-sass/releases/download/$SASS_VERSION/$SASS_TARFILE" .
RUN mv $SASS_TARFILE sass.tgz

WORKDIR /prose
COPY --link go.mod go.sum ./
COPY --link cmd/prose cmd/prose
RUN go build ./cmd/prose

FROM bitnami/minideb:bookworm

WORKDIR /bin
COPY --link --from=builder /sass.tgz /
RUN tar xf /sass.tgz
RUN mv dart-sass/* .
COPY --link --from=builder /prose/prose .

WORKDIR /srv
COPY --link posts posts
COPY --link static static
COPY --link styles styles
COPY --link templates templates
ENTRYPOINT ["prose"]
