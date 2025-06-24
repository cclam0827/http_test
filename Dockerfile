FROM golang:1.24.4-alpine3.22 AS builder

WORKDIR /app
COPY . .
# RUN apk --update add build-base
RUN apk add --no-cache git make build-base
RUN go build -o main main.go


FROM alpine:3.22

ENV APP_ENV=production
EXPOSE 4000

RUN apk --no-cache add tzdata
WORKDIR /app
COPY ./etc/app /etc/app
COPY --from=builder /app/main /app/

ENTRYPOINT ["/app/main", "start"]
