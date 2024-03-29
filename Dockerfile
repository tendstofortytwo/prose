FROM golang:bookworm

WORKDIR /
ENV SASS_VERSION=1.72.0
ENV SASS_TARFILE="dart-sass-${SASS_VERSION}-linux-x64.tar.gz"
ADD "https://github.com/sass/dart-sass/releases/download/$SASS_VERSION/$SASS_TARFILE" .
RUN tar -xf $SASS_TARFILE
RUN mv dart-sass/* bin/

WORKDIR /prose
COPY --link go.mod go.sum ./
COPY --link cmd/prose cmd/prose
RUN go build ./cmd/prose

WORKDIR /bin
RUN cp /prose/prose .

WORKDIR /srv
COPY --link posts posts
COPY --link static static
COPY --link styles styles
COPY --link templates templates
ENV PATH=/bin
ENTRYPOINT ["prose"]
