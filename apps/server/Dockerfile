FROM golang:1.26.5-alpine3.23 AS build
WORKDIR /src
COPY go.work ./
COPY apps/server/go.mod apps/server/go.sum apps/server/
COPY contracts/generated/go /src/contracts/generated/go
RUN cd apps/server && go mod download
COPY apps/server apps/server
ARG VERSION=unknown
RUN cd apps/server && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.buildVersion=${VERSION}" -o /out/ninefold-api ./cmd/api

FROM alpine:3.23
ARG VERSION=unknown
LABEL org.opencontainers.image.title="Ninefold Sudoku server" \
      org.opencontainers.image.version="${VERSION}"
RUN apk upgrade --no-cache \
 && addgroup -S -g 10001 ninefold \
 && adduser -S -D -H -u 10001 -G ninefold ninefold
WORKDIR /app
COPY --from=build --chown=ninefold:ninefold /out/ninefold-api /app/ninefold-api
RUN mkdir /app/data && chown ninefold:ninefold /app/data
USER 10001:10001
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["/app/ninefold-api"]
