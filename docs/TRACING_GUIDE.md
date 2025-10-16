# TRACING_GUIDE.md — спаны, атрибуты и пропагация

## Назначение

Единые правила трейсинга AnyGate на базе OpenTelemetry: иерархия спанов, атрибуты, статусы, события и пропагация контекста. Требование — трасса одного запроса читается сквозь все фазы: приём → роутер/фильтры → действие (static/proxy/…) → запись ответа → апстримы.

---

## Ресурсные атрибуты (Resource)

* `service.name = "anygate"`
* `service.version = "<semver>"`
* `deployment.environment = "<env>"`
* `service.instance.id = "<hostname>_<pid>"`
* `cloud.region` / `cloud.zone` (если применимо)

---

## Иерархия спанов (серверный запрос)

```
server.request ──────────────────────────────────────────────────────────────┐
  http.parse                                                                 │
  router.match                                                               │
  filters.pre                                                                │
  action.fixed   | action.echo   | action.static   | action.proxy ─┬─────────┤
                                                       upstream.connect      │
                                                       upstream.request      │
                                                       upstream.response     │
  filters.post                                                               │
  http.write                                                                 │
```

* Корневой спан: `server.request`
* Внутренние спаны моделируют фазы пайплайна (см. ARCHITECTURE.md)
* Для `action.proxy` создаются дочерние спаны апстрима

---

## Имена спанов

* `server.request`
* `http.parse`
* `router.match`
* `filters.pre`
* `action.fixed` / `action.echo` / `action.static` / `action.proxy`
* `upstream.connect`
* `upstream.request`
* `upstream.response`
* `filters.post`
* `http.write`

---

## Базовые атрибуты спанов

### Общие

* `trace_id`, `span_id` — как в W3C Trace-Context
* `anygate.route_id` — стабильный идентификатор правила
* `anygate.action` — `fixed|echo|static|proxy`
* `enduser.id` — если определён политикой/плагином (опционально)

### HTTP (вход)

* `http.method`
* `http.target` (путь + query, с маскировкой чувствительных параметров)
* `http.route` (шаблон маршрута, например `/users/:id`)
* `http.scheme` (`http|https`)
* `server.address` (хост AnyGate)
* `server.port`
* `client.address` (IP клиента, с маскированием при политике приватности)
* `user_agent.original` (если есть)

### HTTP (апстрим)

* `server.address` (хост апстрима)
* `server.port`
* `network.protocol.name` (`http`)
* `network.protocol.version` (`1.1|2`)
* `anygate.upstream` (логическое имя/`host:port`)

### Производительность

* `anygate.req.size` (байты тела запроса)
* `anygate.resp.size` (байты тела ответа)

---

## Статусы спанов (Status)

* `server.request`:

  * `ERROR` при внутренних сбоях и кодах `5xx`
  * `UNSET` при успешной обработке или `4xx` по вине клиента
  * `ERROR` для `429` по политике лимита — по договорённости (рекомендуется `UNSET`, чтобы не красить всю трассу)
* `action.proxy`:

  * `ERROR` при `502/504/503` и иных ошибках апстрима
* `action.static`, `action.fixed`, `action.echo`:

  * `UNSET` при успехе, `ERROR` при аварийных ситуациях (`500` и т. п.)
* Дочерние `upstream.*`:

  * `ERROR` при таймаутах/сбросах/ошибках TLS/протокола

---

## События (Span Events)

### Повторы и устойчивость

* `retry`
  Атрибуты: `attempt`, `reason` (`connect_error|read_timeout|idempotent`)

### Брейкер

* `circuit_open`
  Атрибуты: `upstream`, `window_s`

### Статика/SPA

* `spa_fallback`
  Атрибуты: `index`, `path`

### Валидация/конфиг

* `config_reload`
  Атрибуты: `result` (`success|failure`), `revision`

---

## Пропагация контекста

### Приём входящего запроса

* Поддержка заголовков:

  * `traceparent`, `tracestate` (W3C Trace-Context)
  * `baggage` (W3C Baggage) — переносится без модификации
* Если заголовки отсутствуют — создаётся новый trace

### Исходящий запрос к апстриму

* Наследуемая семплинг-решение (parent-based)
* В каждый апстрим-запрос добавляются:

  * `traceparent`, `tracestate`
  * `baggage` (при включённой политике форварда)

### Совместимость

* Поддержка B3 заголовков доступна фичей (выкл. по умолчанию)

---

## Семплинг и экземплары

* Sempling: parent-based + ratio (глобальная доля)
* Ошибочные трассы принудительно записываются (always-on для `ERROR`)
* Экземплары метрик `anygate_http_request_duration_seconds` прикрепляют `trace_id` для связки «метрика → трейс»

---

## Маскирование в атрибутах

* Параметры URL и заголовки подчиняются тем же правилам, что и в LOG_FORMAT.md
* В атрибутах трейсинга **не** размещаются секреты (значения заменяются `***`), userinfo в URI удаляется

---

## Примеры атрибутов по спанам

### server.request

* `http.method=GET`
* `http.target=/v1/items?q=1&access_token=***`
* `http.route=/v1/*`
* `client.address=203.0.113.10`
* `anygate.route_id=r:7d2c9e4b`
* `anygate.action=proxy`
* Status: `UNSET` или `ERROR` по правилам выше

### router.match

* `anygate.route_id=r:7d2c9e4b`
* `anygate.route_depth` (глубина trie)
* `anygate.match.kind=literal|param|splat`

### action.proxy

* `anygate.upstream=auth:8000`
* `server.address=auth`
* `server.port=8000`
* `network.protocol.version=2` (если h2)
* События `retry`/`circuit_open` при необходимости

### upstream.connect

* `net.peer.name=auth`
* `net.peer.port=8000`
* Status `ERROR` при `timeout_connect|tls_error`

---

## Пример читаемой трассы end-to-end (дерево)

```
server.request [OK] http.method=GET http.target=/v1/items?q=1…
 ├─ http.parse [OK]
 ├─ router.match [OK] anygate.route_id=r:7d2c9e4b
 ├─ filters.pre [OK]
 ├─ action.proxy [OK] anygate.upstream=auth:8000
 │   ├─ upstream.connect [OK] network.protocol.version=2
 │   ├─ upstream.request [OK]
 │   └─ upstream.response [OK] http.status_code=200
 ├─ filters.post [OK]
 └─ http.write [OK] http.status_code=200
```

Ошибка апстрима читается так:

```
server.request [ERROR] http.status_code=504 code=AG-UPS-2003
 ├─ …
 └─ action.proxy [ERROR]
     ├─ upstream.connect [OK]
     ├─ upstream.request [OK]
     └─ upstream.response [ERROR] event=retry attempt=2 reason=read_timeout
```

---

## Экспорт и приёмники

* Экспортер: OTLP/HTTP или OTLP/gRPC
* Сборщики: Tempo/Jaeger/OTel Collector (поддерживаются ресурсные атрибуты, батчинг, сжатие)

---

## Требования к внедрению

* Все спаны и атрибуты из этого документа должны присутствовать в типовых сценариях (`fixed`, `echo`, `static`, `proxy`)
* Секреты редактируются до записи атрибутов
* События `retry`, `circuit_open`, `spa_fallback` протоколируются в соответствующих спанах
* `trace_id` отражается в логах и как экземплар метрик
* Имена спанов/атрибутов стабильны; изменения — только через RFC

---

## Диагностика и рекомендации

* При росте `5xx` ищите `server.request[ERROR]` и дочерние `action.proxy[ERROR]`
* При деградации p95 — проверяйте длительности `router.match`, `action.*`, `upstream.*`
* При ошибках маршрутизации — `router.match` содержит вид совпадения и `anygate.route_id`
* Для SPA — событие `spa_fallback` подтверждает возврат `index.html`

---

## Мини-чек-лист «трасса читается»

* Корневой `server.request` есть всегда и содержит HTTP-атрибуты
* Фазы пайплайна представлены своими спанами
* Для прокси присутствуют `upstream.*`
* Статусы проставлены по политике
* События добавлены при ретраях/брейкере/SPA
* Корреляция с логами по `trace_id` и с метриками через экземплар работает
