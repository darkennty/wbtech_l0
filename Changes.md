# История оптимизаций производительности

## Описание

В рамках задания по оптимизации API-сервиса была проведена работа по профилированию и улучшению производительности существующего сервиса. Основная цель - оптимизировать сервис по CPU и памяти. Работа выполнена в рамках задания техношколы WB L4.5.

---

## Что было сделано

### 1. Нагрузочное тестирование

Для создания реалистичной нагрузки на API использовался инструмент **Yandex Pandora**.

Для создания нагрузки был сгенерирован файл ammo.txt, описывающий GET-запросы к сервису.

**Запуск нагрузочного тестирования:**
```bash
pandora load_test_config.yaml
```

Конфигурация нагрузки (`load_test_config.yaml`):
- Тип: HTTP запросы к `/order/{id}`
- Количество потоков: 4
- Количество соединений: 100
- Длительность: 40 секунд
- Профиль нагрузки: динамическая (сначала нарастание, потом убывание RPS)

---

### 2. Профилирование CPU


Для анализа использования процессора применялся встроенный инструмент **pprof**. Чтобы её использовать, необходимо было добавить следующие эндпоинты в обработчик (файл `internal/api/handler/handler.go`):

```go
debug := router.Group("/debug/pprof")
{
    debug.GET("/", gin.WrapF(pprof.Index))
    debug.GET("/profile", gin.WrapF(pprof.Profile))
    debug.GET("/trace", gin.WrapF(pprof.Trace))
}
```

**Сбор CPU профиля во время нагрузки:**
```bash
go tool pprof -http :8082 -seconds 40 http://localhost:8000/debug/pprof/profile
```

Используя команду, получили два профиля:
- `internal/pprof/profile_old.pb.gz` - до оптимизаций
- `internal/pprof/profile_new.pb.gz` - после оптимизаций

**Анализ профиля:**
```bash
# Просмотр top функций
go tool pprof -top internal/pprof/profile_old.pb.gz

# Веб-интерфейс с flame graph
go tool pprof -http=:8080 internal/pprof/profile_new.pb.gz

# Сравнение двух профилей
go tool pprof -base=internal/pprof/profile_old.pb.gz internal/pprof/profile_new.pb.gz
```

---

### 3. Трассировка выполнения

Для детального анализа работы goroutines и планировщика использовался **trace**. Трассировка также выполнялась под нагрузкой.

**Сбор trace во время нагрузки:**
```bash
# Запуск сбора trace (в отдельном терминале)
curl http://localhost:8000/debug/pprof/trace?seconds=40 > trace.out

# Анализ trace
go tool trace -http ":8083" trace.out
```

**Собранные traces:**
- `internal/pprof/trace_old.out` - до оптимизаций
- `internal/pprof/trace_new.out` - после оптимизаций

---

## Анализ профилирования до изменений

**Обнаружено в профиле:**
```
      flat  flat%   sum%        cum   cum%
    26.77s 70.67% 70.67%     26.96s 71.17%  runtime.cgocall
     1.55s  4.09% 74.76%      1.55s  4.09%  runtime.stdcall0
     1.40s  3.70% 78.46%      1.40s  3.70%  crypto/internal/fips140/sha256.blockSHANI
     0.91s  2.40% 80.86%      0.91s  2.40%  runtime.stdcall2
     0.75s  1.98% 82.84%      0.75s  1.98%  runtime.memmove
     0.67s  1.77% 84.61%      0.67s  1.77%  runtime.asyncPreempt
     0.32s  0.84% 85.45%      2.31s  6.10%  crypto/internal/fips140/sha256.(*Digest).checkSum
     0.27s  0.71% 86.17%      0.27s  0.71%  runtime.stdcall1
     0.26s  0.69% 86.85%      3.96s 10.45%  github.com/lib/pq/scram.(*Client).saltPassword
     0.26s  0.69% 87.54%      0.26s  0.69%  runtime.nextFreeFast
```

**Обнаружено по команде `peek logrus`:**

```
------------------------------------------------------+-------------
  flat  flat%   sum%        cum   cum%   calls calls% + context       
------------------------------------------------------+-------------
                                         6.08s   100% |   github.com/sirupsen/logrus.(*Entry).log
     0     0%     0%      6.08s 16.05%                | github.com/sirupsen/logrus.(*Entry).write
                                         5.99s 98.52% |   os.(*File).Write
                                         0.09s  1.48% |   github.com/sirupsen/logrus.(*JSONFormatter).Format
------------------------------------------------------+-------------

```

**Выявленные причины плохой производительности:**
- Отсутствовали настройки connection pool
- Каждый запрос создавал новое TCP соединение к PostgreSQL
- Полная аутентификация при каждом подключении
- Множественные запросы к БД (4 запроса к БД для `GetOrderByID`)
- Форматирование строки при каждом логе через fmt.Sprintf
- Неструктурированные логи

---

## Оптимизация

### Оптимизация №1: Настройка Connection Pool


Были добавлены настройки, ограничивающие  максимальное количество одновременно открытых, а также неиспользуемых соединений. Кроме того, было установлено максимальное время используемого и простаиваемого соединений.

**Изменения кода (файл `internal/repository/postgres.go`):**
```go
func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
    db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode))
    if err != nil {
        return nil, err
    }

+   db.SetMaxOpenConns(25)
+   db.SetMaxIdleConns(10)
+   db.SetConnMaxLifetime(5 * time.Minute)
+   db.SetConnMaxIdleTime(1 * time.Minute)

    if err = db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

**Итого:**
- Соединения переиспользуются между запросами
- Нет повторной аутентификации
- Уменьшена нагрузка на PostgreSQL

---

### Оптимизация №2: Объединение SQL запросов через JOIN

**Файл:** `internal/repository/order_postgres.go`



**Предыдущий код**:
```go
query := fmt.Sprintf("SELECT * FROM %s WHERE order_uid=$1;", orderTable)
if err = r.db.Get(&order, query, uuid); err != nil {
    // ...
}
query = fmt.Sprintf("SELECT d.name, d.phone, d.zip, d.city, d.address, d.region, d.email FROM %s d WHERE order_uid=$1;", deliveryTable)
if err = r.db.Get(&(order.Delivery), query, uuid); err != nil {
	// ...
}

query = fmt.Sprintf("SELECT p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee FROM %s p WHERE order_uid=$1;", paymentTable)
if err = r.db.Get(&(order.Payment), query, uuid); err != nil {
	// ...
}

query = fmt.Sprintf("SELECT i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status FROM %s i WHERE order_uid=$1;", itemTable)
if err = r.db.Select(&(order.Items), query, uuid); err != nil {
	// ...
}
```

**Новый код**:
```go
query := fmt.Sprintf(`
    SELECT 
        o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
        o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
        d.name as delivery_name, d.phone as delivery_phone, d.zip as delivery_zip, 
        d.city as delivery_city, d.address as delivery_address, d.region as delivery_region, d.email as delivery_email,
        p.transaction, p.request_id, p.currency, p.provider, p.amount, 
        p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
    FROM %s o
    LEFT JOIN %s d ON o.order_uid = d.order_uid
    LEFT JOIN %s p ON o.order_uid = p.order_uid
    WHERE o.order_uid = $1;
`, orderTable, deliveryTable, paymentTable)
var result model.DBResult
if err = r.db.Get(&result, query, orderUUID); err != nil {
    // ...
}

// ...
// Заполнение model.Order{}
// ...

itemQuery := fmt.Sprintf(`
		SELECT i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, 
		       i.size, i.total_price, i.nm_id, i.brand, i.status 
		FROM %s i 
		WHERE i.order_uid = $1;
	`, itemTable)

if err = r.db.Select(&order.Items, itemQuery, orderUUID); err != nil {
	// ...
}
```

**Итого:**
- Уменьшено количество обращений к БД с *четырёх* до *двух*
- Вследствие этого меньше нагрузки на connection pool

---

### Оптимизация №3: Структурированное логирование

Была изменена работа со строкой для лога.

**Изменения кода (файл `internal/api/handler/order.go`):**
```go
- logrus.Print(fmt.Sprintf("OrderUID: %s - using database data. Time: %d ms", 
    orderUID, queryTime.Milliseconds()))
+ logrus.WithFields(logrus.Fields{
    "order_uid": orderUID,
    "time_ms":   queryTime.Milliseconds(),
    "source":    "database",
}).Info("Order retrieved")
```


**Итого:**
- Нет аллокации временных строк
- Форматирование строки происходит только, если логирование включено

---

## Результаты оптимизации

### Сравнение CPU профилей

**До оптимизации** (`profile_old.pb.gz`):
```
Showing nodes accounting for 33.16s, 87.54% of 37.88s total
Dropped 581 nodes (cum <= 0.19s)
Showing top 10 nodes out of 209
      flat  flat%   sum%        cum   cum%
    26.77s 70.67% 70.67%     26.96s 71.17%  runtime.cgocall
     1.55s  4.09% 74.76%      1.55s  4.09%  runtime.stdcall0
     1.40s  3.70% 78.46%      1.40s  3.70%  crypto/internal/fips140/sha256.blockSHANI
     0.91s  2.40% 80.86%      0.91s  2.40%  runtime.stdcall2
     0.75s  1.98% 82.84%      0.75s  1.98%  runtime.memmove
     0.67s  1.77% 84.61%      0.67s  1.77%  runtime.asyncPreempt
     0.32s  0.84% 85.45%      2.31s  6.10%  crypto/internal/fips140/sha256.(*Digest).checkSum
     0.27s  0.71% 86.17%      0.27s  0.71%  runtime.stdcall1
     0.26s  0.69% 86.85%      3.96s 10.45%  github.com/lib/pq/scram.(*Client).saltPassword
     0.26s  0.69% 87.54%      0.26s  0.69%  runtime.nextFreeFast
```

**После оптимизации** (`profile_new.pb.gz`):
```
Showing nodes accounting for 2.75s, 71.80% of 3.83s total
Dropped 151 nodes (cum <= 0.02s)
Showing top 10 nodes out of 219
      flat  flat%   sum%        cum   cum%
     1.92s 50.13% 50.13%      1.95s 50.91%  runtime.cgocall
     0.27s  7.05% 57.18%      0.27s  7.05%  runtime.stdcall1
     0.15s  3.92% 61.10%      0.15s  3.92%  runtime.stdcall6
     0.09s  2.35% 63.45%      0.09s  2.35%  runtime.stdcall2
     0.08s  2.09% 65.54%      0.08s  2.09%  runtime.stdcall4
     0.06s  1.57% 67.10%      0.06s  1.57%  runtime.nextFreeFast
     0.06s  1.57% 68.67%      0.06s  1.57%  runtime.stdcall8
     0.05s  1.31% 69.97%      0.05s  1.31%  runtime.asyncPreempt
     0.04s  1.04% 71.02%      0.04s  1.04%  aeshashbody
     0.03s  0.78% 71.80%      0.36s  9.40%  bufio.(*Reader).Read
```

**Сравнение двух профилей:**
```
Showing nodes accounting for -30.95s, 81.71% of 37.88s total
Dropped 638 nodes (cum <= 0.19s)
Showing top 10 nodes out of 193
      flat  flat%   sum%        cum   cum%
   -24.85s 65.60% 65.60%    -25.01s 66.02%  runtime.cgocall
    -1.52s  4.01% 69.61%     -1.52s  4.01%  runtime.stdcall0
    -1.40s  3.70% 73.31%     -1.40s  3.70%  crypto/internal/fips140/sha256.blockSHANI
    -0.82s  2.16% 75.48%     -0.82s  2.16%  runtime.stdcall2
    -0.74s  1.95% 77.43%     -0.74s  1.95%  runtime.memmove
    -0.62s  1.64% 79.07%     -0.62s  1.64%  runtime.asyncPreempt
    -0.32s  0.84% 79.91%     -2.31s  6.10%  crypto/internal/fips140/sha256.(*Digest).checkSum
    -0.26s  0.69% 80.60%     -3.96s 10.45%  github.com/lib/pq/scram.(*Client).saltPassword
    -0.22s  0.58% 81.18%     -2.22s  5.86%  crypto/internal/fips140/sha256.(*Digest).Write
    -0.20s  0.53% 81.71%     -0.20s  0.53%  runtime.nextFreeFast
```

### Визуализация яерез Flame Graph

Визуализация CPU профиля доступна в файле `internal/pprof/pprof_diff.svg`.

---

### Анализ trace

В папке `internal/pprof` есть два файла с трейсингом до и после оптимизации: 

**Просмотр trace до оптимизации:**
```bash
go tool trace -http ":8083" internal/pprof/trace_old.out
```

**Просмотр trace после оптимизации:**
```bash
go tool trace -http ":8084" internal/pprof/trace_new.out
```

Пра сравнении трассирования были выявлены следующие отличия:
- Увеличилось количество активных горутин
- Уменьшилось время блокировок на операциях с БД
- Паузы GC стали короче из-за меньшего количества аллокаций

---

## Воспроизведение результатов

Все файлы для воспроизведения тестирования представлены в папке `internal/pprof`.

### 1. Запуск сервиса

```bash
# Запустить Docker контейнеры (PostgreSQL, Kafka)
docker-compose up -d

# Запустить сервис
go run ./cmd/app/main.go
```

### 2. Генерация тестовых данных

Чтобы сгенерировать тестовые данные, их нужно взять из БД. Чтобы заполнить БД заказами, можно использовать мой проект `github.com/darkennty/wbtech-l0-kafka-message-generator` или отправить POST-запросы вручную.

Для генерации файла `ammo.txt` был использован скрипт `generate_ammo.bat`, находящийся в папке `internal/pprof`:

```bash
# Создание файла ammo.txt для Pandora
.\internal\pprof\generate_ammo.bat
```

### 3. Запуск профилирования

**Терминал 1 - сервис:**
```bash
go run ./cmd/app/main.go
```

**Терминал 2 - нагрузочное тестирование:**
```bash
cd internal/pprof
pandora load_test_config.yaml
```

**Терминал 3 - сбор CPU профиля:**
```bash
go tool pprof -http :8082 -seconds 40 http://localhost:8000/debug/pprof/profile
```

**Терминал 4 - сбор trace:**
```bash
# С помощью скрипта
.\internal\pprof\traceprof.bat

# Вручную
curl http://localhost:8000/debug/pprof/trace?seconds=40 -o trace.out
go tool trace -http ":8083" trace.out
```

### 4. Анализ результатов

```bash
# Просмотр top функций
go tool pprof -top profile.pb.gz

# Веб-интерфейс с flame graph
go tool pprof -http=:8080 profile.pb.gz

# Сравнение с baseline
go tool pprof -base=internal/pprof/profile_old.pb.gz profile.pb.gz
```
