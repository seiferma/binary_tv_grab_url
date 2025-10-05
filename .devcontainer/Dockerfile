ARG VERSION
FROM golang:${VERSION}-alpine

RUN adduser --disabled-password vscode

RUN apk add --no-cache git openssh-client make ca-certificates gcc musl-dev
RUN go install golang.org/x/tools/gopls@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest
RUN go install golang.org/x/tools/cmd/goimports@latest
RUN go install github.com/jandelgado/gcov2lcov@latest
