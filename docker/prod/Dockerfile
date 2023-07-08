# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-buster AS build

ARG user
ARG uid

RUN useradd -u $uid -d /home/$user $user

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN  CGO_ENABLED=0 go build -o /users cmd/main.go

## Deploy
FROM golang:1.19-buster

ARG app_port
ARG user
ARG uid

WORKDIR /app

COPY --from=build /users /app/users
COPY --from=build /app/database /app/database

RUN apt-get update && \
    apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent

RUN curl -sSL https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
RUN echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ bionic main" > /etc/apt/sources.list.d/migrate.list
RUN apt-get update && \
    apt-get install -y migrate
    
# RUN addgroup $user
# RUN adduser -S $user -u $uid -G $user
RUN useradd -u $uid -d /home/$user $user

RUN chown -R $user:$user /app
RUN chmod +x /app/users

USER $user
RUN mkdir -p /app/logs
RUN touch /app/logs/app-log.json
RUN touch /app/logs/query-log.json
RUN chmod -R 777 /app/logs

EXPOSE $app_port

CMD ["/app/users"]