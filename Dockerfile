# syntax=docker/dockerfile:1

FROM golang:1.19 as base

FROM base as build

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build
RUN go build .

FROM base as prod

WORKDIR /app

# Install ffmpeg
RUN apt-get -y update && apt-get -y upgrade && apt-get install -y --no-install-recommends ffmpeg

COPY --from=build ./app/hStream .

COPY .env .