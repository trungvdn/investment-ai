FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server && CGO_ENABLED=0 go build -o /out/ingest ./cmd/ingest
FROM alpine:3.22
RUN apk add --no-cache ca-certificates poppler-utils && adduser -D macro
COPY --from=build /out/server /out/ingest /usr/local/bin/
USER macro
EXPOSE 3000
ENTRYPOINT ["server"]
