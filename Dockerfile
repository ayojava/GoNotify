FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /gonotify .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /gonotify /gonotify

USER nonroot:nonroot

ENTRYPOINT ["/gonotify"]