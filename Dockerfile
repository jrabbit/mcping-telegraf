FROM golang
COPY go.mod go.sum /srv/
WORKDIR /srv
RUN go mod download
COPY . /srv/
RUN go build
ENTRYPOINT /srv/mcping-telegraf
