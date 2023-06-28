FROM golang:1.20
WORKDIR /Dimage
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ginchat
EXPOSE 5000
CMD [ "./ginchat" ]