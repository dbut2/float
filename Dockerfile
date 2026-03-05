FROM node:22-alpine AS frontend-builder

WORKDIR /app

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS backend-builder

WORKDIR /bank
COPY go.mod go.sum ./
RUN go mod download

COPY go/ ./go/
COPY db/ ./db/
COPY --from=frontend-builder /app/dist/ ./web/dist/
COPY web/embed.go ./web/embed.go

RUN go build -o /bin/float ./go/cmd/float

FROM alpine AS final

COPY --from=backend-builder /bin/float /float
#COPY float-key-sandbox.p12 /float-key-sandbox.p12

EXPOSE 8080

CMD ["/float"]
