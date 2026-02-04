FROM golang:1.25.6-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache bash
COPY --from=build /app/app .
COPY --from=build /app/public ./public
COPY --from=build /app/scripts ./scripts
RUN chmod +x ./scripts/*.sh

EXPOSE 8080
CMD ["./app"]
