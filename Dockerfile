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

WORKDIR $BUILD_PATH

COPY go.mod $BUILD_PATH/
COPY go.sum $BUILD_PATH/
COPY rest $BUILD_PATH/rest

RUN test -z $(go fmt ./...)
RUN go mod download
RUN test -z $(go mod tidy -v)
RUN go mod verify
RUN go build ./...
RUN go test ./...

# artifacts
ARG BUILD_ARTIFACTS_AUDIT=/go/src/github.com/Workiva/go-rest/go.sum

# no-op container
FROM scratch
