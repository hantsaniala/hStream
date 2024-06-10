# hStream

[![Go Reference](https://pkg.go.dev/badge/github.com/hantsaniala/hStream.svg)](https://pkg.go.dev/github.com/hantsaniala/hStream)

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

## Run

You can run the project using docker compose:

```sh
docker compose up -d --build
```

It will build necessary dependencies and run your app under http://localhost:5480

## Generating key

Before generating key, don't forget to change master key folder name. You can use random key from [here](https://randomkeygen.com/).

To generate key necessary for the project, run:

```sh
# To generate video encryption key
./hStream gen key

# To generate keypair
./hStream gen keypair
```

## TODO

- [ ] Add gRPC support
- [ ] Add plugin type struct for microservice
- [ ] Add a CRUD page dashboard with vue
- [ ] Add support for other video format
- [ ] Add support for audio format

## Author

[Hantsaniala Eléo](https://t.me/hantsaniala3) 2023
