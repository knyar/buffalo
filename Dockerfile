FROM golang:1.22-alpine

COPY . /app
WORKDIR /app

RUN go build

EXPOSE 8000
USER nobody:nobody

ENTRYPOINT ["/app/buffalo"]
