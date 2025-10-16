# AnyGate — Обзор архитектуры

## Назначение

Этот документ описывает устройство AnyGate на уровне архитектуры: рантайм, жизненный цикл запроса, маршрутизацию, действия (actions), плагины (filters), наблюдаемость и безопасность. Тerminология унифицирована, диаграммы иллюстрируют потоки и компоненты.

---

## Терминология

* **Hot-path** — участки кода, задействованные для типичных запросов без редких ветвлений.
* **Action** — конечный обработчик маршрута: `Fixed`, `Echo`, `Static`, `Proxy`.
* **Filter (плагин)** — этапный фильтр, выполняемый до/после Action: `pre`/`post`.
* **Route** — правило DSL: методы + путь + действие.
* **Router** — матчер, предкомпилированный в компактное дерево (radix-trie).
* **Upstream** — целевой сервер/сервис, куда проксируется трафик.
* **Pool** — пул постоянных соединений к апстриму.
* **RCU-swap** — атомарная замена структур без блокировки читателей (для hot-reload).
* **Observability** — логи, метрики, трейсинг.

---

## Общая схема компонентов

```
                  ┌────────────────────────────────────────────────┐
                  │                    AnyGate                      │
                  ├──────────────────────────┬──────────────────────┤
Client           ▶ Listener (TCP/TLS/ALPN)   │  Control Plane       │
(HTTP/1.1,h2)    │  └─ Protocol Codec        │  ┌─────────────────┐ │
                  │     (h1 parser / h2)     │  │ Config Loader   │ │
                  │                           │  │  + ENV Override│ │
                  │   Runtime                 │  ├─────────────────┤ │
                  │   └─ Reactor (epoll)      │  │ RCU Router Swap │ │
                  │   └─ Scheduler (workers)  │  └─────────────────┘ │
                  │   └─ Buffers Pools        │                      │
                  ├──────────────────────────┴──────────────────────┤
                  │                 Data Plane                      │
                  │  ┌───────────┐   ┌───────────┐   ┌───────────┐ │
                  │  │ Router    ├──▶│ Filters   ├──▶│  Action   │ │
                  │  └───────────┘   └───────────┘   └───────────┘ │
                  │                                 ┌─────────────┐ │
                  │                                 │  Static FS  │ │
                  │                                 └─────────────┘ │
                  │                                 ┌─────────────┐ │
                  │                                 │ Upstreams   │ │
                  │                                 │ + Pools     │ │
                  │                                 └─────────────┘ │
                  ├────────────────────────────────────────────────┤
                  │               Observability                    │
                  │  Logs (JSON/pretty), Metrics, Tracing          │
                  └────────────────────────────────────────────────┘
```

---

## Рантайм

### Реактор и планировщик

* **Reactor (epoll, edge-triggered)**: неблокирующие сокеты, единый цикл событий, таймауты I/O, backpressure.
* **Scheduler (workers)**: фиксированный пул, work-stealing, парковка/разморозка, pinning по ядрам.
* **Timers**: таймер-колесо для connect/read/write таймаутов, отменяемые будильники.
* **Буферы**: пулы аренированных буферов для сетевых операций и компрессии (reuse, без фрагментации).

### Протоколы

* **HTTP/1.1**: парсинг стартовой строки, заголовков, chunked, streaming тела.
* **HTTP/2**: управление окнами, мультиплексирование потоков, приоритеты, h2c и TLS-вариант.
* **TLS (rustls)**: ALPN, session cache, OCSP stapling, профили шифрования, опциональный mTLS.

---

## Жизненный цикл запроса (pipeline)

```
ACCEPT → TLS/ALPN → PARSE → MATCH → PRE(filters) → ACTION → POST(filters) → WRITE
```

### Диаграмма потока

```
        ┌───────────┐     ┌───────────┐     ┌───────────┐
Client ─┤  ACCEPT   ├────▶│  PARSE    ├────▶│  MATCH    │
        └───────────┘     └───────────┘     └─────┬─────┘
                                                   │ RouteHit
                                              ┌────▼────┐
                                              │  PRE    │  (filters, may short-circuit)
                                              └────┬────┘
                                                   │
                                          ┌────────▼────────┐
                                          │     ACTION      │  (Fixed/Echo/Static/Proxy)
                                          └────────┬────────┘
                                                   │
                                              ┌────▼────┐
                                              │  POST   │  (filters, finalize)
                                              └────┬────┘
                                                   │
                                               ┌───▼───┐
                                               │ WRITE │
                                               └───────┘
```

* **PRE** может прервать выполнение (deny/redirect/fixed) через `ActionOverride`.
* **POST** выполняется после Action (логирование, метрики, заголовки).

---

## Маршрутизация

### DSL → IR → Trie

* **DSL**: строка с методами, путем и действием (`proxy|static|fixed|echo`).
* **IR**: нормализованная запись (метод-битмаска, сегменты пути, параметры `:name`, хвостовой `*`).
* **Trie (radix)**: компактное дерево по сегментам, узлы содержат:

  * таблицу дочерних по литералам,
  * слот параметра `:name`,
  * слот маски `*` (splat),
  * список действий по методам.

### Правила сопоставления

* приоритет по методу → точный путь → `:param` → `*` → порядок объявления.
* параметры пути и `splat` передаются в контекст фильтров/действий.
* префиксные прокси-правила корректно собирают целевой URI без дублирования префикса.

### Диаграмма trie

```
root
 ├── "api"
 │    └── "v1"
 │         ├── ":id"  (param)
 │         │     └── "*" (splat) → Action(PROXY)
 │         └── "ping" → Action(FIXED/ECHO)
 └── "" (/) → Action(STATIC "/dist/")
```

---

## Actions (действия)

### Fixed

* Предсобранный ответ со статусом и необязательным телом.
* Контент-тайп: автоопределение JSON/текст, переопределяемый заголовками.

### Echo

* Диагностика: метод, путь, заголовки, первые байты тела (лимит), опциональный статус.

### Static

* **Путь**: нормализация, защита от выхода за корень, index-имена, SPA-fallback.
* **Отдача**: fast-path через `sendfile`; fallback: `mmap` для мелких файлов, `writev` для склейки.
* **Кеш-семантика**: ETag (слабый/сильный), Last-Modified, conditional requests.
* **Диапазоны**: single/multi-range.
* **Компрессия**: gzip/brotli/zstd on-the-fly (пулы), автоотдача precompressed.

### Proxy

* **Клиент**: неблокирующие соединения, keep-alive, пулы per-upstream.
* **Протоколы**: h1/h2 к апстриму (ALPN/конфиг), gRPC passthrough.
* **Заголовки**: `X-Forwarded-For/Proto/Host`, `Forwarded`, перепись `Host`.
* **Надёжность**: retries для идемпотентных, circuit-breaker, passive/active health-checks.
* **Балансировка**: round-robin, least-conn; fair-очереди ожиданий.

### Последовательности (sequence)

**Static fast-path**

```
Client → Listener → Parser → Router → PRE → Static(sendfile) → POST → Client
```

**Proxy passthrough**

```
Client → Parser → Router → PRE → Proxy(get conn from pool)
      → Upstream(req) → Upstream(resp) → POST → Client
```

**PRE short-circuit (deny)**

```
... → Router → PRE(deny→Fixed 403) ──┐
                                     └→ WRITE → Client
```

---

## Плагины (filters)

### Контракт

* **pre(ctx) → Result<(), ActionOverride>**: может модифицировать запрос/заголовки или коротко замкнуть пайплайн предопределённым ответом.
* **post(ctx) → Result<(), ()>**: доступ к статусу/размеру и времени, может править заголовки ответа.

### Встроенные фильтры

* **logger**: структурные логи, маскирование секретов.
* **rate-limiter**: token/leaky-bucket, области `ip`, `route`, `ip_route`.
* **headers**: add/remove/set, запрет опасных заголовков.
* **cors**: источники, методы, заголовки, preflight.
* **ip allow/deny**: сети/маски.
* **auth**: проверка заголовка/токена.

### Цепочка фильтров

```
[PRE] logger → ip-allow/deny → rate-limit → headers → auth → … → [ACTION] → [POST] logger → headers → …
```

* Формируется на этапе компиляции конфигурации, хранится как массив вариантов enum; в hot-path отсутствуют виртуальные вызовы.

---

## Наблюдаемость

### Логи

* Форматы: JSON и человекочитаемый.
* Поля: trace-id/span-id, маршрут, action, код ответа, длительности фаз, размеры тел, ключевые заголовки (с маскированием).

### Метрики (Prometheus)

* Счётчики кодов по маршрутам/группам.
* Гистограммы задержек: total, parse, route, action, upstream.
* Состояние пулов апстрима, промахи/попадания кеш-метаданных статики, компрессия.

### Трейсинг (OpenTelemetry)

* Спаны на фазы pipeline и внешние вызовы к апстриму.
* Атрибуты: маршрут, метод, upstream-endpoint, попытки retries, причина circuit-breaker.

---

## Конфигурация и hot-reload

### Форматы и источники

* YAML/JSON/TOML; переопределение значений переменными окружения.

### Процесс загрузки

```
read files/env → validate schema → compile DSL→IR→trie → build filters/actions → RCU-swap
```

* При успешной сборке новая версия конфигурации атомарно подменяет текущую без обрыва активных keep-alive.

---

## Безопасность

* **TLS/mTLS** per-route/group; профили шифрования и ALPN.
* **Изоляция процесса**: запуск без привилегий, chroot/static-root.
* **Лимиты**: размеры заголовков/тел, таймауты, защита от медленных клиентов.
* **Политики заголовков**: HSTS/secure headers через `headers`/`cors`.
* **Supply chain**: подписанные релизы и SBOM (см. SECURITY.md).

---

## Потоки данных и управление памятью

### Потоки

* **RX**: сокет → parser → router → pre → action → post → encoder → сокет.
* **TX**: для Proxy двунаправленный стрим (relaying) с управлением backpressure.

### Память

* Пулы буферов RX/TX, предварительное выделение для small-write, reusing chunkов.
* Предсобранные ответы в `Fixed`/части `Echo`.
* LRU-кеш мелких статических файлов (конфигурируемый лимит).

---

## Ошибки и исключения

* Классы: `client_error`, `upstream_error`, `gateway_error`, `timeout`, `limit_exceeded`.
* Маппинг в коды HTTP и стабильные форматы ошибок (каталог ошибок).
* В логи уходит технический контекст, в ответ — безопасное, неразоблачающее сообщение (при `secure_error_log` включённой).

---

## Расширяемость

* **Filters**: добавляются как новые варианты enum с конфиг-структурами; регистрируются в фабрике плагинов.
* **Actions**: отдельные модули, расширяемые через конфиг DSL.
* **Фичи**: HTTP/3/QUIC, io_uring, динамические плагины (cdylib/WASM) — фича-флаги, выключены по умолчанию, вне hot-path.

---

## Диаграмма: взаимодействие Proxy с пулом и ретраями

```
        ┌──────────────────────────┐
        │   Proxy Action (ctx)     │
        └───────────┬──────────────┘
                    │ lookup route→ upstream set
                    ▼
            ┌───────────────┐   get/establish
            │  Conn Pool    ├─────────────────┐
            └──────┬────────┘                 │
                   │                          │
                   ▼                          ▼
           Upstream Req                Retry Policy (idem)
                   │                          │
                   ▼                          │
           Upstream Resp  ───ok───────────────┘
                   │
                   ▼
             POST filters → WRITE
```

---

## Диаграмма: Static c ETag/Range

```
Router → Static
   │        ┌────────────────────────────────────────────┐
   │        │ 1) Resolve path (safe join, index, SPA)    │
   │        │ 2) Stat + metadata cache                   │
   │        │ 3) If-None-Match/If-Modified-Since check   │
   │        │ 4) Range parse (single/multi)              │
   │        │ 5) sendfile | mmap+writev | compress       │
   │        └────────────────────────────────────────────┘
   └────────────────────────→ WRITE
```

---

## Стабильность и производительность

* Узкие контуры: router-match, Fixed/Echo fast-path, Static sendfile, Proxy relay — минимальные аллокации, данные по возможности не копируются повторно.
* Конфигурационные ветки и сложные проверки вынесены из hot-path (в compile-time шаг и подготовку структур).
* Метрики задержек фаз позволяют локализовать деградации.

---

## Приложение: карта модулей

```
crates/
  runtime/    ─ reactor, scheduler, timers, buffers
  http/       ─ codecs h1/h2, parsing, encoding
  tls/        ─ rustls integration, ALPN
  router/     ─ DSL parser, IR, trie, RCU
  actions/    ─ fixed, echo, static, proxy (+ upstream pools)
  plugins/    ─ filter contract, built-ins (logger, rate, headers, cors, ip, auth)
  obs/        ─ logs, metrics, tracing
  config/     ─ models, validation, env override
bin/anygate   ─ CLI (serve/check/pack-spa), signals, hot-reload
```

---

## Итог

Архитектура AnyGate — тонкий, предсказуемый L7-движок с предкомпилированной маршрутизацией, быстрыми действиями, встроенными фильтрами и прозрачной наблюдаемостью. Рантайм, пайплайн и конфигурация спроектированы так, чтобы изменять поведение вне горячего пути и сохранять стабильность под нагрузкой.
