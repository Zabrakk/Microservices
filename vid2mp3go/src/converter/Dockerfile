FROM golang:1.22-bullseye

RUN apt-get update \
	&& apt-get install -y --no-install-recommends --no-install-suggests ffmpeg

WORKDIR /app
COPY go.mod go.sum ./
COPY . ./
RUN go mod download
RUN  CGO_ENABLED=0 GOOS=linux go build -o /go-converter

CMD ["/go-converter"]
