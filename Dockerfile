FROM node:22-alpine AS frontend-builder

WORKDIR /app

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM nginx:alpine AS frontend

COPY web/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=frontend-builder /app/dist/ /usr/share/nginx/html/

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
