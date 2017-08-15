FROM golang:1.8.3-alpine3.6

WORKDIR /app
COPY main.go main.go
RUN go build

FROM alpine:3.6
RUN apk add ca-certificates --update
COPY --from=0 /app/app /bin/app

EXPOSE 80 

ENTRYPOINT ["app"]

