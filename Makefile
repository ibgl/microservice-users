.PHONY: migrate-up migrate-create deploy

include .env

prod?=
user?=user

postgres_url="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${DB_HOST}:5432/${POSTGRES_DB}?sslmode=disable"
docker_compose_args=$(if $(prod), -f docker-compose.prod.yml, -f docker-compose.yml)

migrate-up:
	docker compose $(docker_compose_args) run migrate -path /migrations -database $(postgres_url) -verbose up

migrate-down:
	docker compose $(docker_compose_args) run migrate -path /migrations -database $(postgres_url) -verbose down

migrate-drop:
	docker compose $(docker_compose_args) run migrate -path /migrations -database $(postgres_url) -verbose drop

migrate-create:	
	docker compose $(docker_compose_args) run migrate create -dir /migrations -ext sql $(name)	

docker-stop:
	docker compose $(docker_compose_args) stop

docker-build:
	docker compose $(docker_compose_args) build users-app --build-arg user=$(user)

docker-up:
	docker compose $(docker_compose_args) up -d --remove-orphans

git-pull:
	git pull origin main 

deploy: docker-stop git-pull docker-build docker-up
