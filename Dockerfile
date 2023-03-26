FROM golang:alpine AS builder

COPY ./ /app

WORKDIR /app

RUN go build -o yurl .

FROM scratch

COPY --from=builder /app/yurl /bin/yurl

WORKDIR /app

ENTRYPOINT ["/bin/yurl"]