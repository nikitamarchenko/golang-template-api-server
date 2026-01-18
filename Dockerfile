# https://github.com/GoogleContainerTools/distroless

# Start by building the application.
FROM golang:1.25.6 AS build

WORKDIR /go/src/app

# Download packages
COPY go.mod go.sum ./ 
RUN go mod download

# Build app
COPY . . 
RUN CGO_ENABLED=0 go build -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

ARG BINARY_NAME=app

COPY --from=build --chown=nonroot:nonroot /go/bin/app /usr/local/bin/${BINARY_NAME}

EXPOSE 8080

USER nonroot

ENTRYPOINT [ "app" ]

CMD ["server"]
