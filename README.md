order-service/
├── cmd/app/main.go                 # Точка входа приложения
├── internal/                       # Внутренние пакеты
│   ├── config/config.go           # Конфигурация приложения
│   ├── domain/models.go           # Бизнес-сущности
│   ├── repository/                # Слой доступа к данным
│   │   ├── interface.go           # Интерфейсы репозиториев
│   │   ├── postgres/postgres.go   # PostgreSQL репозиторий
│   │   └── cache/cache.go         # In-memory кэш
│   ├── service/                   # Бизнес-логика
│   │   ├── interface.go           # Интерфейсы сервисов
│   │   └── order_service.go       # Сервис заказов
│   └── handler/                   # Обработчики
│       ├── http/handler.go        # HTTP хендлеры
│       └── kafka/consumer.go      # Kafka consumer
├── migrations/001_create_tables.sql # Миграции БД
├── web/static/                    # Веб-интерфейс
│   ├── index.html                 # HTML страница
│   └── script.js                  # JavaScript логика
├── configs/config.yaml            # Конфигурационный файл
├── go.mod                         # Go модули
├── go.sum                         # Go зависимости
└── README.md                      # Документация#   R a b o t a - L 0 
 
 
