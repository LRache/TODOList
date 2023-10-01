FROM golang:latest
WORKDIR ./app
COPY . .
RUN go build main.go
EXPOSE 8080
ENTRYPOINT ["./main"]