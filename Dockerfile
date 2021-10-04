
# using the latest Golang image for Alpine Linux

FROM golang:1.15-alpine

WORKDIR /app

# copy all files in the current directory to /app
COPY . ./

# Install GCC which is needed to run go binaries
RUN apk add build-base

# download modules
RUN go mod download

# build binary named go-note
RUN go build -o /go-note

# expose port 8081
EXPOSE 8081

# run the go-note binary
CMD [ "/go-note" ]
