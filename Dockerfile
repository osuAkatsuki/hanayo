FROM golang:1.20

RUN apt-get update && apt-get install git -y

WORKDIR /srv/root

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . /srv/root

RUN git submodule init && git submodule update --remote --recursive --merge

RUN go build

EXPOSE 80

CMD ["./scripts/start.sh"]
