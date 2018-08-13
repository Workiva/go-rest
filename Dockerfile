FROM golang:1.10.3-alpine3.8

## github credentials
RUN apk add --update bash curl git openssh py-pip
ARG GIT_SSH_KEY
RUN git config --global url.git@github.com:.insteadOf https://github.com/
RUN mkdir ~/.ssh; ssh-keyscan -t rsa github.com > ~/.ssh/known_hosts
RUN chmod -R 700 ~/.ssh; echo "${GIT_SSH_KEY}" > ~/.ssh/id_rsa; chmod 600 ~/.ssh/id_rsa
RUN eval "$(ssh-agent -s)" && ssh-add ~/.ssh/id_rsa

#install glide
RUN go get -u github.com/Masterminds/glide

WORKDIR /go/src/github.com/Workiva/go-rest
COPY . /go/src/github.com/Workiva/go-rest

# install dependencies
RUN glide install

# run tests
RUN go test $(glide novendor)

# artifacts
ARG BUILD_ARTIFACTS_AUDIT=/go/src/github.com/Workiva/go-rest/glide.lock

# no-op container
FROM scratch
