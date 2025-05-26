FROM golang:1.24-bookworm AS builder
ENV GOOS=linux \
    CGO_ENABLED=0
WORKDIR /go/src/app
ADD . /go/src/app

RUN go mod download && go build -o /go/bin/app

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /go/bin/app /
ENTRYPOINT ["/app"]