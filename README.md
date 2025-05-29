# Person API

Сервис для хранения информации о людях с автоматическим обогащением данных (возраст, пол, национальность) по открытому API.

## Основные возможности

* При добавлении нового человека по фамилии, имени и (опционально) отчеству автоматически запрашивает и сохраняет:

  * возраст (Agify)
  * пол (Genderize)
  * национальность (Nationalize)
* REST API для CRUD-операций и поиска с фильтрами и пагинацией
* Логирование на уровнях `debug` и `info`
* Хранение данных в PostgreSQL с миграциями через Goose
* Генерация Swagger-документации
* Запуск через Docker / Docker Compose или локально с Makefile

## Стек технологий

* Go 1.24
* chi (HTTP Router)
* sqlx + lib/pq (PostgreSQL)
* Goose (миграции)
* Swagger (swaqqo/http-swagger)
* Docker / Docker Compose

## Предварительные требования

* Go 1.24
* Docker и Docker Compose (необязательно)
* PostgreSQL 15+ (локально или в контейнере)

## Переменные окружения

Положите файл `.env` в корень проекта (копируйте из `.env.example`):

```dotenv
# DSN для подключения к базе
DB_DSN=postgres://user:password@db:5432/persons?sslmode=disable
# Порт HTTP-сервера
SERVER_PORT=8080
# Уровень логирования: debug, info, error
LOG_LEVEL=info
```

## Запуск в Docker / Docker Compose

1. Убедитесь, что в корне есть `.env`.
2. Запустить через Docker Compose:
    ```bash
      docker-compose up --build
    ```
   или 
3.  ```bash
      make compose-up
    ```

3. Сервис и база будут подняты, миграции применятся автоматически.
4. Доступ к API: `http://localhost:${SERVER_PORT}`.

## Makefile

* `make build`       — собрать бинарник
* `make run`         — запустить локально
* `make migrate-up`  — применить миграции вверх
* `make migrate-down`— откатить миграции
* `make swagger`     — сгенерировать Swagger-документацию
* `make docker-build`— собрать Docker-образ
* `make docker-run`  — запустить контейнер
* `make compose-up`  — поднять через docker-compose
* `make compose-down`— остановить контейнеры
* `make clean`       — удалить бинарник

## API Endpoints

| Метод  | Путь            | Описание                                 |
| ------ | --------------- | ---------------------------------------- |
| GET    | `/persons`      | Получить список с фильтрами и пагинацией |
| GET    | `/persons/{id}` | Получить одного человека по ID           |
| POST   | `/persons`      | Создать нового (тело запроса ниже)       |
| PUT    | `/persons/{id}` | Обновить существующего                   |
| DELETE | `/persons/{id}` | Удалить по ID                            |

### Пример тела POST `/persons`

```json
{
  "name": "Dmitriy",
  "surname": "Ushakov",
  "patronymic": "Vasilevich"  // опционально
}
```

Все ответы возвращаются в формате JSON. В случае ошибок — структура `{ "error": "описание" }`.

## Swagger UI

После запуска сервиса доступен Swagger UI:

```
http://localhost:${SERVER_PORT}/swagger/index.html
```

## Логирование

* Уровень логов задаётся переменной `LOG_LEVEL`.
* В коде используются `info`- и `debug`-логи для отслеживания вызовов и ошибок.





## Тесты

Чтобы запустить тесты
```bash
make test
```

Или:

```bash
go test ./... -v
```

## Линтер

```bash
make lint
```
