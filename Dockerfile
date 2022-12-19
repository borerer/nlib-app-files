FROM golang AS builder
WORKDIR /nlib-app-files
COPY go.mod /nlib-app-files/go.mod
COPY go.sum /nlib-app-files/go.sum
RUN go mod download
COPY . /nlib-app-files
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

FROM alpine
WORKDIR /nlib-app-files
COPY --from=builder /nlib-app-files/nlib-app-files /nlib-app-files/nlib-app-files
ENTRYPOINT ["/nlib-app-files/nlib-app-files"]
