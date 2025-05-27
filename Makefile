# Makefile

# Переменные
SWAG          := swag
ENV_FILE      := .env

ifneq ("$(wildcard $(ENV_FILE))","")
include .env
export
endif


DB_DSN		?= postgres://user:password@db:5432/persons?sslmode=disable
SERVER_PORT ?= 8080
LOG_LEVEL 	?= info


.PHONY: all help swagger compose-up compose-down

all: help

help:
	@echo "Usage:"
	@echo "  make compose-up    — поднять через docker-compose"
	@echo "  make compose-down  — остановить docker-compose"


swagger:
	@echo "→ Генерация Swagger в internal/handler/docs"
	$(SWAG) init \
		--output internal/handler/docs \
		--generalInfo cmd/person-api/main.go \
		--parseDependency

compose-up:
	@echo "→ docker-compose up"
	docker-compose up --build

compose-down:
	@echo "→ docker-compose down"
	docker-compose down
