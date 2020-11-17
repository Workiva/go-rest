FROM golang:1.15-alpine

RUN apk add --update build-base git openssh

## github credentials
ARG GIT_SSH_KEY
RUN git config --global url.git@github.com:.insteadOf https://github.com/
RUN mkdir ~/.ssh; ssh-keyscan -t rsa github.com > ~/.ssh/known_hosts
RUN chmod -R 700 ~/.ssh; echo "${GIT_SSH_KEY}" > ~/.ssh/id_rsa; chmod 600 ~/.ssh/id_rsa
RUN eval "$(ssh-agent -s)" && \
    ssh-add ~/.ssh/id_rsa

ENV GOPATH=/go
ENV BUILD_PATH=$GOPATH/src/github.com/Workiva/go-rest
ENV V2=$BUILD_PATH/v2

WORKDIR $BUILD_PATH

# v1
COPY go.mod $BUILD_PATH/
COPY go.sum $BUILD_PATH/
COPY rest $BUILD_PATH/rest

# v2
COPY v2/go.mod $V2/
COPY v2/go.sum $V2/
COPY v2/rest $V2/rest

RUN test -z $(go fmt ./...)

# v1 - check formatting, build, test
RUN go mod download
RUN test -z $(go mod tidy -v)
RUN go mod verify
RUN go build ./...
RUN go test ./...

# v2 - check formatting, build, test
WORKDIR $V2
RUN go mod download
RUN test -z $(go mod tidy -v)
RUN go mod verify
RUN go build ./...
RUN go test ./...

# artifacts
ARG BUILD_ARTIFACTS_AUDIT=/go/src/github.com/Workiva/go-rest/go.sum:/go/src/github.com/Workiva/go-rest/v2/go.sum

# no-op container
FROM scratch
