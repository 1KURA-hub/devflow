FROM golang:1.24-alpine AS builder

WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/seed ./cmd/seed

FROM busybox:1.36

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY --from=builder /out/seed /app/seed

EXPOSE 8080

CMD ["/app/server"]
