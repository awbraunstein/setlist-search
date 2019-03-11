# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang v1.12 base image
FROM golang:1.12

# Enable Go Modules.
ENV GO111MODULE=on

# Add Maintainer Info
LABEL maintainer="Andrew Braunstein <awbraunstein@gmail.com>"

ENV DIRPATH=$GOPATH/src/github.com/awbraunstein/setlist-search

# Set the Current Working Directory inside the container
WORKDIR $DIRPATH

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

ENV SETSEARCHERINDEX=$DIRPATH/.setsearcherindex

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["setlist-search", "-http=:8080"]
