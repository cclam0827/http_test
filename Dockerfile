FROM golang:1.22.5-alpine3.20 AS builder

WORKDIR /app
COPY . .
# RUN apk --update add build-base
RUN apk add --no-cache git make build-base
RUN go build -o main main.go


FROM alpine:3.20.2

ENV APP_ENV=production
EXPOSE 4000

RUN apk --no-cache add tzdata
WORKDIR /app
COPY ./etc/app /etc/app
COPY --from=builder /app/main /app/

ENTRYPOINT ["/app/main", "start"]
