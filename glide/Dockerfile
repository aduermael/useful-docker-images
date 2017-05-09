FROM golang:1.8.1-alpine

RUN apk add --update curl git
RUN curl -s https://glide.sh/get | sh

CMD ["glide", "up"]