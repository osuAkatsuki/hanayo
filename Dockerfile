FROM golang:1.21

RUN apt-get update && apt-get install git -y

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
RUN apt-get -y install nodejs

WORKDIR /srv/root

COPY go.mod go.sum ./
RUN go mod download && go mod verify

RUN apt install -y python3-pip
RUN pip install --break-system-packages git+https://github.com/osuAkatsuki/akatsuki-cli

COPY . /srv/root

RUN git submodule init && git submodule update --remote --recursive --merge

RUN go build

RUN cd web && npm install --global gulp-cli && npm install && gulp

EXPOSE 80

CMD ["./scripts/start.sh"]
