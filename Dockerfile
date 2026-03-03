FROM nginx:alpine AS frontend

COPY web/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY web/public/ /usr/share/nginx/html/

ENV BACKEND_HOST=backend:8080

EXPOSE 8080

FROM golang:1.26-alpine AS backend-builder

WORKDIR /bank
COPY go.mod go.sum ./
RUN go mod download

COPY go/ ./go/
COPY db/ ./db/

RUN go build -o /bin/float ./go/cmd/float

FROM alpine AS backend

COPY --from=backend-builder /bin/float /float

EXPOSE 8080

CMD ["/float"]
