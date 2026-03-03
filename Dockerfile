FROM golang:1.26-alpine AS backend-builder

WORKDIR /bank
COPY go.mod ./
RUN go mod download

COPY go/ ./go/
COPY db/ ./db/

RUN go build -o /app ./go/cmd/bank

FROM alpine AS backend

COPY --from=backend-builder /app /app

EXPOSE 8080

CMD ["/app"]
