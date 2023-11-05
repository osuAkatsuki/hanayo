FROM golang:1.20

RUN apt-get update && apt-get install git -y

WORKDIR /srv/root

COPY . .

RUN go mod download && go mod verify

RUN git submodule init && git submodule update --remote --recursive --merge

RUN apt install -y python3-pip

RUN go build

EXPOSE 80

CMD ["./scripts/start.sh"]
