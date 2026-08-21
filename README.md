# adaptive_cron

# Примеры команд для использования

## 1. Простая версия для малого сайта
./adaptive-cron --profile lite

## 2. Стандартная версия с логированием
./adaptive-cron --profile standard --config config/standard.yaml

## 3. Enterprise версия
./adaptive-cron --profile enterprise --config config/enterprise.yaml

## 4. Кастомизация - включить Prometheus в lite версии
./adaptive-cron --profile lite --enable metrics.prometheus

## 5. Отключить некоторые features
./adaptive-cron --profile standard --disable log.json,transport.grpc

## 6. Показать доступные профили
./adaptive-cron --list-profiles

## 7. Использование с Docker
docker-compose -f docker-compose.yml -f docker-compose.lite.yml up

## 8. Создание своего профиля
cp config/standard.yaml config/myprofile.yaml
## Отредактировать myprofile.yaml
./adaptive-cron --profile myprofile --config config/myprofile.yaml