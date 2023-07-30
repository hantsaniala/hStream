# hStream

A simple VOD server made with Go.

## Requirements

- Go >= 1.19
- Redis
- PostgreSQL
- Docker
- docker-compose

## Install

First clone the project with:

```sh
git clone git@github.com:hantsaniala/hStream.git
```

Create you own .env file from given .env.example with:

```sh
cp .env.example .env
```

Then edit it to match your existing credentials.

Finally you can install with:

```sh
go install .
```

Or directly install without cloning the project with:

```sh
go install github.com/hantsaniala/hStream
```

## Build

Or you can build the project first with:

```sh
go build .
```

## Run

After building the project, you can run it with:

```sh
hStream server run
```

Task broker must be run in parallel:

```sh
hStream broker run
```

## TODO

- [ ] Add gRPC support
- [ ] Add plugin type struct for microservice
- [ ] Add a CRUD page dashboard with vue
- [ ] Add support for other video format
- [ ] Add support for audio format

## Author

[Hantsaniala Eléo](https://t.me/hantsaniala3) 2023
