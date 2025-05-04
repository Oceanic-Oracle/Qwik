# Qwik

## Оглавление
- [Qwik](#qwik)
  - [Оглавление](#оглавление)
  - [Общее описание](#общее-описание)
  - [Стек технологий](#стек-технологий)
    - [Frontend](#frontend)
    - [Backend](#backend)
  - [Запуск](#запуск)
  - [API документация](#api-документация)

## Общее описание

Проект представляет собой агрегатор наподобии Amazon/Ozon/Wildberries.
Состоит из двух частей:

- Мобильного приложения под Android на C#
- Сервера на Golang

Микросервисную схему можно посмотреть в отдельном файле Mic_schema.jpg

## Стек технологий

### Frontend

- C#
- .NET MAUI
- REST API

### Backend

- Golang
- Postgres
- Redis
- gRPC
- Nginx
- FlyWay
- REST API
- Docker
- Swagger

## Запуск

Для запуска потребуется прописать эту команду из папки Backend при открытом Docker

```bash
docker compose up --build
```

## API документация

Документация доступна по адресам: