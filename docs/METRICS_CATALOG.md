# METRICS_CATALOG.md — Полный перечень метрик и семантика

## Назначение

Единый каталог метрик AnyGate для Prometheus: имена, типы, единицы, лейблы, семантика, рекомендации по алертам и рекординг-рулам. Организация по подсистемам исключает дубли и обеспечивает предсказуемость для дашбордов и SLO.

---

## Соглашения

### Имена, типы, единицы

* Префикс: `anygate_`
* Стиль: `snake_case`
* Типы: `counter`, `gauge`, `histogram`
* Суффиксы единиц: `_seconds`, `_bytes`, `_total`

### Лейблы (низкая кардинальность)

* `route_id` — стабильный идентификатор правила (хэш шаблона и набора методов)
* `action` — `fixed|echo|static|proxy`
* `method` — буквальный метод запроса
* `status_class` — `1xx|2xx|3xx|4xx|5xx` (рекординг-рулом из `status`)
* `phase` — `parse|match|pre|action|post|write|upstream_connect|upstream_io|compress|sendfile`
* `upstream` — логическое имя/`host:port`
* `result` — конечное состояние операции (`ok|timeout_connect|timeout_read|reset|proto_error|tls_error|fail`)
* `reason` — причина (`idempotent|connect_error|read_timeout|policy`)
* `state` — состояние (`open|idle|in_use` и т. п.)
* `algo` — алгоритм компрессии (`gzip|brotli|zstd`)
* `kind` — вид ответа статики (`hit|miss|spa_fallback|range_single|range_multi`)

> Избегать `path`, `trace_id`, `upstream_ip` в лейблах. Для корреляции использовать экземплары и логи.

---

## Подсистема: HTTP сервер и пайплайн

### anygate_http_requests_total — counter

* **Лейблы:** `route_id`, `action`, `method`, `status_class`
* **Семантика:** количество завершённых HTTP-запросов с классификацией кода ответа
* **Пример:** рост `4xx/5xx` по маршруту

### anygate_http_request_duration_seconds — histogram

* **Лейблы:** `route_id`, `action`, `method`
* **Семантика:** полная длительность обработки запроса (end-to-end)
* **Рекоменд. buckets:** миллисекундные ступени для L7 (задаются глобально)

### anygate_http_request_bytes_total — counter

* **Лейблы:** `route_id`, `action`, `method`
* **Семантика:** суммарные байты тел запросов (RX)

### anygate_http_response_bytes_total — counter

* **Лейблы:** `route_id`, `action`, `method`, `status_class`
* **Семантика:** суммарные байты тел ответов (TX)

### anygate_router_match_duration_seconds — histogram

* **Лейблы:** нет
* **Семантика:** время сопоставления маршрута (trie-match) для выборки по всему инстансу

### anygate_pipeline_phase_duration_seconds — histogram

* **Лейблы:** `phase`
* **Семантика:** длительность фаз конвейера (`parse|match|pre|action|post|write|upstream_connect|upstream_io|compress|sendfile`)

---

## Подсистема: Прокси и апстрим

### anygate_proxy_upstream_requests_total — counter

* **Лейблы:** `upstream`, `result`
* **Семантика:** исход запроса к апстриму (`ok|timeout_connect|timeout_read|reset|proto_error|tls_error|fail`)

### anygate_proxy_upstream_connect_duration_seconds — histogram

* **Лейблы:** `upstream`
* **Семантика:** длительность установления соединения TCP/TLS к апстриму

### anygate_proxy_upstream_connections — gauge

* **Лейблы:** `upstream`, `state` (`open|idle|in_use`)
* **Семантика:** текущее состояние пула соединений к апстриму

### anygate_proxy_retries_total — counter

* **Лейблы:** `upstream`, `reason`
* **Семантика:** количество повторов (ретраев) по причинам

### anygate_proxy_circuit_open_total — counter

* **Лейблы:** `upstream`
* **Семантика:** число открытий circuit-breaker

### anygate_proxy_circuit_state — gauge

* **Лейблы:** `upstream`
* **Семантика:** состояние брейкера (`0` закрыт, `1` открыт)

---

## Подсистема: Статика и ФС

### anygate_static_responses_total — counter

* **Лейблы:** `kind` (`hit|miss|spa_fallback|range_single|range_multi`)
* **Семантика:** классификация ответов файлового сервера

### anygate_static_sendfile_bytes_total — counter

* **Лейблы:** нет
* **Семантика:** байты, отданные через `sendfile`

### anygate_static_uncompressed_bytes_total — counter

* **Лейблы:** нет
* **Семантика:** объём исходных данных до компрессии

### anygate_static_compressed_bytes_total — counter

* **Лейблы:** `algo`
* **Семантика:** объём, отданный после on-the-fly компрессии

### anygate_static_precompressed_served_total — counter

* **Лейблы:** `algo`
* **Семантика:** количество выдач предсжатых ассетов

---

## Подсистема: Плагины и политики

### anygate_rate_limiter_dropped_total — counter

* **Лейблы:** `scope` (`ip|route|ip_route`)
* **Семантика:** отброшенные запросы лимитером

### anygate_security_denied_total — counter

* **Лейблы:** `reason` (`ip_deny|auth_fail|cors_reject`)
* **Семантика:** запреты политики безопасности

---

## Подсистема: TLS вход

### anygate_tls_handshakes_total — counter

* **Лейблы:** `result` (`ok|fail`)
* **Семантика:** итоги TLS-рукопожатий

### anygate_tls_sessions_reused_total — counter

* **Лейблы:** нет
* **Семантика:** повторно использованные TLS-сессии

---

## Подсистема: Рантайм и администрирование

### anygate_runtime_connections — gauge

* **Лейблы:** `state` (`open|closing`)
* **Семантика:** активные и закрывающиеся входные соединения

### anygate_runtime_workers — gauge

* **Лейблы:** `state` (`parked|running`)
* **Семантика:** состояние воркеров рантайма

### anygate_admin_config_reloads_total — counter

* **Лейблы:** `result` (`success|failure`)
* **Семантика:** загрузки конфигурации (горячие/при старте)

### anygate_build_info — gauge

* **Лейблы:** `version`, `revision`, `runtime`
* **Семантика:** константная «единица», идентифицирующая сборку

---

## Рекординг-рулы и производные

### status_class

```yaml
groups:
- name: anygate-recording
  rules:
  - record: job:anygate_http_requests_total:status_class
    expr: |
      sum by (route_id, action, method, status_class) (
        label_replace(anygate_http_requests_total, "status_class", "$1xx", "status", "(\\d)")
      )
```

### p95/p99 latency

```yaml
- record: job:anygate_http_request_duration_seconds:p95
  expr: |
    histogram_quantile(0.95,
      sum by (le) (rate(anygate_http_request_duration_seconds_bucket[5m])))
- record: job:anygate_http_request_duration_seconds:p99
  expr: |
    histogram_quantile(0.99,
      sum by (le) (rate(anygate_http_request_duration_seconds_bucket[5m])))
```

### error-rate per route

```yaml
- record: route:anygate_error_rate_5m
  expr: |
    sum by (route_id) (increase(anygate_http_requests_total{status_class="5xx"}[5m]))
    /
    sum by (route_id) (increase(anygate_http_requests_total[5m]))
```

---

## Алерты (примеры порогов)

### Spike 5xx (общий)

```yaml
- alert: AnyGateHighErrorRate
  expr: |
    sum(increase(anygate_http_requests_total{status_class="5xx"}[5m]))
    /
    sum(increase(anygate_http_requests_total[5m])) > 0.02
  for: 10m
  labels: { severity: page }
  annotations:
    summary: "Высокая доля 5xx в AnyGate"
    description: "Доля 5xx > 2% за 10м. Проверить апстримы и конфигурации."
```

### Degradation p95

```yaml
- alert: AnyGateLatencyP95Degraded
  expr: job:anygate_http_request_duration_seconds:p95 > 0.250
  for: 15m
  labels: { severity: ticket }
  annotations:
    summary: "p95 латентности превышает целевой порог"
    description: "p95 > 250ms. Проверить апстримы, лимиты и брейкер."
```

### Upstream failures

```yaml
- alert: AnyGateUpstreamFailures
  expr: |
    sum by (upstream) (increase(anygate_proxy_upstream_requests_total{result=~"timeout_.*|reset|proto_error|tls_error"}[5m])) > 50
  for: 5m
  labels: { severity: page }
  annotations:
    summary: "Ошибки апстрима"
    description: "Серия upstream ошибок > 50 за 5м на {{ $labels.upstream }}."
```

### Circuit open

```yaml
- alert: AnyGateCircuitOpen
  expr: anygate_proxy_circuit_state == 1
  for: 2m
  labels: { severity: page }
  annotations:
    summary: "Открыт circuit-breaker"
    description: "Брейкер открыт на {{ $labels.upstream }}. Проверить здоровье апстрима."
```

### TLS handshake fail rate

```yaml
- alert: AnyGateTLSHandshakeFailures
  expr: |
    sum(increase(anygate_tls_handshakes_total{result="fail"}[5m]))
    /
    clamp_min(sum(increase(anygate_tls_handshakes_total[5m])), 1) > 0.05
  for: 10m
  labels: { severity: ticket }
  annotations:
    summary: "Рост отказов TLS рукопожатий"
    description: "Доля неуспешных TLS > 5%. Проверить сертификаты/алгоритмы."
```

### Config reload failures

```yaml
- alert: AnyGateConfigReloadFailed
  expr: increase(anygate_admin_config_reloads_total{result="failure"}[15m]) > 0
  for: 0m
  labels: { severity: ticket }
  annotations:
    summary: "Не применились изменения конфига"
    description: "Ошибка загрузки конфига (hot-reload). См. логи с кодами VLD-1xxx."
```

### Rate limiter drops

```yaml
- alert: AnyGateRateLimiterDrops
  expr: sum(increase(anygate_rate_limiter_dropped_total[5m])) > 100
  for: 5m
  labels: { severity: info }
  annotations:
    summary: "Лимитер отбрасывает запросы"
    description: "Снижение доступности на входе. Проверить профиль лимитера/трафик."
```

### Static errors

```yaml
- alert: AnyGateStaticAnomalies
  expr: |
    increase(anygate_static_responses_total{kind="miss"}[10m]) > 1000
    or
    increase(anygate_static_precompressed_served_total[10m]) == 0
  for: 10m
  labels: { severity: info }
  annotations:
    summary: "Аномалии раздачи статики"
    description: "Рост MISS или исчезли предсжатые ассеты. Проверить деплой SPA/кеши."
```

---

## Экземплары и корреляция

* На `*_duration_seconds` включены экземплары с `trace_id` для склейки с трейсами в UI.
* В алертах рекомендуется добавлять ссылку на LogQL-запрос по `trace_id` через дашборд.

---

## Мини-гайд по дашбордам

* **Общие:** RPS, p95/p99, доля `5xx`, топ-маршруты по ошибкам (`route_id`), распределение по `action`.
* **Прокси:** ошибки/результаты по `upstream`, соединения пула, ретраи, брейкер.
* **Статика:** hit/miss, объёмы sendfile и компрессии, доля `spa_fallback`.
* **Безопасность:** rate-drops, deny по причинам, TLS рукопожатия.
* **Рантайм:** соединения, воркеры, успешность reload.

---

## Проверки приёмки каталога

* Каждая метрика уникальна по имени и семантике; дубликатов нет.
* Лейблы соответствуют списку и не включают высококардинальные значения.
* Рекординг-рулы создают `status_class` и квантили без пересчёта имён исходных метрик.
* Алерты компилируются, выражения используют существующие метрики и лейблы.
* Сценарные дашборды собираются без изменения набора имён.
