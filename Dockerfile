FROM golang:1.14.1-alpine3.11 AS build

WORKDIR /app
ADD . .

RUN go build -ldflags="-s -w" -o ea-gaming-review

FROM alpine:3.11.5 AS final

RUN apk add --no-cache curl

WORKDIR /app
COPY --from=build /app/ea-gaming-review ./

EXPOSE 8080
HEALTHCHECK CMD curl --fail http://localhost:8080/health || exit 1
ENTRYPOINT [ "/app/ea-gaming-review" ]