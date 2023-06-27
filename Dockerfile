FROM golang:1.20.5
WORKDIR /Dimage
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-ginchat
CMD [ "/docker-ginchat" ]