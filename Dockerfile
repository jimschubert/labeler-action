FROM golang:1.14.1-alpine as builder
ENV GOOS=linux \
    GOARCH=386 \
    CGO_ENABLED=0

#RUN apk --no-cache add gcc g++ make ca-certificates && apk add git

WORKDIR /go/src/app
ADD . /go/src/app

RUN go mod download && go build -o /go/bin/app

FROM gcr.io/distroless/base-debian10
COPY --from=builder /go/bin/app /
ENTRYPOINT ["/app"]