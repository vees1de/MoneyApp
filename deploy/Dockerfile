FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build


FROM scratch AS frontend-dist

COPY --from=frontend-builder /app/frontend/dist/ /


FROM golang:1.26-alpine AS backend-builder

WORKDIR /app/backend

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server ./cmd/api


FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=backend-builder /bin/server ./server
COPY --from=backend-builder /app/backend/migrations ./migrations
COPY --from=frontend-dist / ./frontend-dist

ENV APP_ENV=production
ENV HTTP_ADDR=:8080
ENV FRONTEND_DIST_DIR=/app/frontend-dist
ENV AUTH_ALLOW_INSECURE_DEV_AUTH=false

EXPOSE 8080

CMD ["./server"]
