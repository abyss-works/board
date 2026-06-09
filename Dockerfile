FROM node:22-alpine AS frontend
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
COPY --from=frontend /frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o board .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/board .
EXPOSE 8080
CMD ["./board"]
