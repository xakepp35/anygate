# OBSERVABILITY_TAXONOMY.md — единый словарь метрик, логов и трейсинга

## Назначение

Установить единый набор имён, лейблов и форматов для наблюдаемости AnyGate: метрики (Prometheus), логи (JSON/CBOR) и трейсы (OpenTelemetry). Требования: отсутствие дублей метрик, предсказуемые и стабильные имена, дружелюбие к Loki (низкая кардинальность лейблов) и детерминированная сериализация логов.

---

## Общие принципы

* Имена метрик — snake_case с префиксом `anygate_`, единицы в суффиксе (`_seconds`, `_bytes`, `_total`).
* Лейблы — только низкой кардинальности; динамику оставлять в полях события, а не в лейблах.
* Логи — JSON Lines и канонический CBOR с **фиксированным порядком полей** и стабильными ключами.
* Корреляция: `trace_id` и `span_id` в каждом логе; в метриках — экземплары (exemplars) с `trace_id`.
* Трейсы — OpenTelemetry, ресурсные атрибуты и имена спанов фиксированы.

---

## Метрики (Prometheus)

### Запросы и ответы

* `anygate_http_requests_total{route_id,action,method,status_class}` — счётчик запросов.
  `status_class ∈ {1xx,2xx,3xx,4xx,5xx}`.
* `anygate_http_request_duration_seconds{route_id,action,method}` — гистограмма полной длительности запроса.
* `anygate_http_request_bytes_total{route_id,action,method}` — суммарные байты тела запроса.
* `anygate_http_response_bytes_total{route_id,action,method,status_class}` — суммарные байты тела ответа.

### Роутер и пайплайн

* `anygate_router_match_duration_seconds` — гистограмма времени матчинга.
* `anygate_pipeline_phase_duration_seconds{phase}` — гистограмма фаз: `parse`, `match`, `pre`, `action`, `post`, `write`.

### Прокси и апстрим

* `anygate_proxy_upstream_requests_total{upstream, result}` — счётчик исходов: `ok`, `timeout_connect`, `timeout_read`, `reset`, `proto_error`, `tls_error`.
* `anygate_proxy_upstream_connect_duration_seconds{upstream}` — гистограмма времени установления соединения.
* `anygate_proxy_upstream_connections{upstream,state}` — gauge активных соединений: `open`, `idle`, `in_use`.
* `anygate_proxy_retries_total{upstream,reason}` — счётчик повторов: `idempotent`, `connect_error`, `read_timeout`.
* `anygate_proxy_circuit_open_total{upstream}` — счётчик открытий брейкера.
* `anygate_proxy_circuit_state{upstream}` — gauge состояния брейкера: `0|1`.

### Статика и компрессия

* `anygate_static_responses_total{kind}` — счётчик: `hit`, `miss`, `spa_fallback`, `range_single`, `range_multi`.
* `anygate_static_sendfile_bytes_total` — суммарно отдано через sendfile.
* `anygate_static_uncompressed_bytes_total` — суммарно несжатых байт (до компрессии).
* `anygate_static_compressed_bytes_total{algo}` — суммарно сжатых байт; `algo ∈ {gzip,brotli,zstd}`.
* `anygate_static_precompressed_served_total{algo}` — отдано предсжатых.

### Лимитирование и политика

* `anygate_rate_limiter_dropped_total{scope}` — сброшено лимитером; `scope ∈ {ip,route,ip_route}`.
* `anygate_security_denied_total{reason}` — запреты: `ip_deny`, `auth_fail`, `cors_reject`.

### TLS/вход

* `anygate_tls_handshakes_total{result}` — итоги рукопожатий: `ok`, `fail`.
* `anygate_tls_sessions_reused_total` — повторное использование сессий.

### Рантайм и управление

* `anygate_runtime_connections{state}` — gauge приёмных соединений: `open`, `closing`.
* `anygate_runtime_workers{state}` — gauge воркеров: `parked`, `running`.
* `anygate_admin_config_reloads_total{result}` — загрузки конфига: `success`, `failure`.

#### Замечания по лейблам

* `route_id` — стабильный идентификатор правила (хэш шаблона пути и набора методов), **не** оригинальный путь.
* `method` — буквальный метод запроса (ограниченная кардинальность гипотезой реальных методов).
* `upstream` — логическое имя/хост:порт; избегать динамических URL с параметрами.
* Не добавлять `path` или `trace_id` в лейблы метрик.

---

## Логи (Loki, JSON/CBOR)

### Каналы и уровни

* Канал — stdout/stderr, одна запись на строку (JSON Lines) или поток CBOR (для бинарной доставки).
* Уровни: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`.
  Дополнительно: `severity` — целое (`TRACE=10`, `DEBUG=20`, `INFO=30`, `WARN=40`, `ERROR=50`).

### Фиксированный порядок полей

Поля сериализуются **строго** в следующем порядке (и в JSON, и в CBOR-каноническом ключевом порядке):

`ts`, `level`, `severity`, `code`, `message`, `component`, `phase`, `trace_id`, `span_id`, `route_id`, `route_pattern`, `action`, `method`, `scheme`, `host`, `port`, `path`, `query`, `status`, `duration_ms`, `bytes_rx`, `bytes_tx`, `client_ip`, `client_port`, `user_agent`, `upstream`, `upstream_ip`, `attempt`, `retry`, `policy`, `secure`, `labels`, `extras`

#### Описания и типы

* `ts` — RFC3339Nano с наносекундами UTC (строка), напр. `2025-10-17T13:37:42.123456789Z`.
* `level` — строка из набора уровней.
* `severity` — целое.
* `code` — код каталога ошибок или `"AG-OK"` для успехов.
* `message` — краткое описание без секретов.
* `component` — подсистема: `listener`, `http`, `router`, `filters`, `proxy`, `static`, `tls`, `admin`, `runtime`.
* `phase` — фаза пайплайна: `parse`, `match`, `pre`, `action`, `post`, `write`, `upstream_connect`, `upstream_io`, `compress`, `sendfile`.
* `trace_id`, `span_id` — шестнадцатеричные строки (W3C Trace-Context).
* `route_id` — стабильный хэш маршрута.
* `route_pattern` — опционально шаблон пути (например, `/users/:id`), может отключаться политикой.
* `action` — `fixed`, `echo`, `static`, `proxy`.
* `method`, `scheme`, `host`, `port`, `path`, `query` — как в запросе; `scheme` — `http|https`.
* `status` — код ответа.
* `duration_ms` — общая длительность запроса (целое, миллисекунды).
* `bytes_rx` / `bytes_tx` — принятые/отданные байты тел.
* `client_ip`, `client_port` — сетевые реквизиты клиента (маскирование — политикой).
* `user_agent` — строка без парсинга (может отсутствовать).
* `upstream` — логическое имя/адрес апстрима; `upstream_ip` — фактический адрес соединения (если есть).
* `attempt` — номер попытки при ретраях (целое, начиная с `1`).
* `retry` — булево, флаг повтора.
* `policy` — краткое имя политики, если применялась (например, `rate_limit`, `ip_deny`, `cors`).
* `secure` — булево, включена ли `secure_error_log`.
* `labels` — объект низкой кардинальности для копирования в Loki-лейблы (см. ниже).
* `extras` — объект с дополнительными полями плагинов (ключи неймспейсить: `plg_<name>_*`).

> Поля со значением `null` **не сериализуются**. Пустые строки — также не выводить; использовать отсутствие поля.

### Пример JSON-записи

```json
{
  "ts":"2025-10-17T13:37:42.123456789Z",
  "level":"INFO",
  "severity":30,
  "code":"AG-OK",
  "message":"request served",
  "component":"proxy",
  "phase":"action",
  "trace_id":"9f09e6e1c8a44b58",
  "span_id":"b1f3aa22e9f1d011",
  "route_id":"r:7d2c9e4b",
  "route_pattern":"/v1/*",
  "action":"proxy",
  "method":"GET",
  "scheme":"https",
  "host":"api.example.com",
  "port":443,
  "path":"/v1/items",
  "query":"q=1",
  "status":200,
  "duration_ms":12,
  "bytes_rx":0,
  "bytes_tx":342,
  "client_ip":"203.0.113.10",
  "client_port":51820,
  "user_agent":"curl/8.5.0",
  "upstream":"auth:8000",
  "upstream_ip":"10.0.0.12:8000",
  "attempt":1,
  "retry":false,
  "policy":"",
  "secure":true,
  "labels":{"env":"prod","service":"anygate","version":"1.0.0"},
  "extras":{"plg_logger_sampled":true}
}
```

### CBOR

* Формат — **канонический CBOR** (RFC 8949, детерминированная сортировка ключей по байтовому порядку, текстовые ключи UTF-8).
* Набор ключей и порядок соответствуют JSON-порядку (для канонизации — сортировка обеспечивается энкодером; при выводе JSON — использовать фиксированный список).

### Loki: лейблы и парсинг

* Базовые лейблы для потока: `app="anygate"`, `env`, `cluster`, `instance`, `version`.
  Допустимые дополнительные: `component`, `action`, `status_class`.
  **Не** использовать как лейблы: `path`, `trace_id`, `user_agent`, `upstream_ip` (высокая кардинальность).
* Рекомендуемый фрагмент promtail:

```yaml
pipeline_stages:
  - json:
      expressions:
        level: level
        component: component
        action: action
        status: status
        route_id: route_id
        trace_id: trace_id
        code: code
  - labels:
      level:
      component:
      action:
  - replace:
      expression: '(\\d)\\d\\d$'
      source: status
      replace: '${1}xx'
      # итог перенести в label status_class при необходимости
```

---

## Трейсинг (OpenTelemetry)

### Ресурсные атрибуты

* `service.name = "anygate"`
* `service.version = "<semver>"`
* `deployment.environment = "<env>"`
* `service.instance.id = "<host>_<pid>"`

### Имена спанов (внутри одного запроса)

* `server.request` — корневой спан запроса.
* `http.parse`
* `router.match`
* `filters.pre`
* `action.fixed` / `action.echo` / `action.static` / `action.proxy`
* `upstream.connect` — дочерний к `action.proxy`.
* `upstream.request` / `upstream.response` — дочерние к `action.proxy`.
* `filters.post`
* `http.write`

### Ключевые атрибуты спанов

* Общие:
  `http.method`, `http.target`, `http.route` (шаблон), `http.status_code`, `anygate.route_id`, `anygate.action`, `enduser.id` (если присутствует), `net.peer.ip`, `net.peer.port`, `client.address` (опционально).
* Proxy-спаны:
  `server.address` (upstream host), `server.port`, `network.protocol.version` (`http/1.1|h2`), `anygate.upstream`.
* Ошибки:
  `error.type` (домен, напр. `UPS`, `RTE`), `error.code` (`AG-XXX-####`), `error.message`.

### События (span events)

* `retry` — поля: `attempt`, `reason`.
* `circuit_open` — поля: `upstream`, `window_s`.
* `spa_fallback` — поля: `index`, `path`.

### Корреляция с метриками

* Экземплары для `anygate_http_request_duration_seconds` прикрепляют `trace_id` (и, по возможности, ссылку на трейс).

---

## Политики сокращения/безопасности в логах

* Маскирование секретов по списку заголовков и JSON-ключей: `authorization`, `cookie`, `set-cookie`, `x-api-key`, `token`, `password` → `***`.
* `client_ip` может быть обрезан до /24 (IPv4) или /56 (IPv6) по флагу `privacy_mask`.
* Отключаемые поля: `route_pattern`, `user_agent`, `query` — через конфиг `obs.redact`.

---

## Соответствие с каталогом ошибок

* Поле `code` всегда присутствует: `AG-OK` при успехе или код из `ERROR_CATALOG.md`.
* Для `ERROR` — `message` краткий, детали в `extras` и трейсинге.

---

## Карта соответствий «фаза → метрика/спан/лог»

| Фаза                            | Метрика                                                  | Спан                        | Лог `phase`        |
| ------------------------------- | -------------------------------------------------------- | --------------------------- | ------------------ |
| Разбор запроса                  | `anygate_pipeline_phase_duration_seconds{phase="parse"}` | `http.parse`                | `parse`            |
| Маршрутизация                   | `…{phase="match"}`                                       | `router.match`              | `match`            |
| PRE-фильтры                     | `…{phase="pre"}`                                         | `filters.pre`               | `pre`              |
| Action: fixed/echo/static/proxy | `…{phase="action"}`                                      | `action.*`                  | `action`           |
| Upstream connect                | `anygate_proxy_upstream_connect_duration_seconds`        | `upstream.connect`          | `upstream_connect` |
| Upstream IO                     | —                                                        | `upstream.request/response` | `upstream_io`      |
| Компрессия                      | —                                                        | —                           | `compress`         |
| sendfile                        | `anygate_static_sendfile_bytes_total`                    | —                           | `sendfile`         |
| POST-фильтры                    | `…{phase="post"}`                                        | `filters.post`              | `post`             |
| Запись ответа                   | `…{phase="write"}`                                       | `http.write`                | `write`            |

---

## Соглашения об именах и отсутствующие дубли

* Каждая метрика уникальна по имени и семантике; пересекающиеся значения отражаются через лейблы, а не новые метрики.
* Нет пары метрик с одинаковым смыслом и разными именами (например, нет дублирования для байтов ответа).
* Списки `result`, `state`, `reason`, `kind` фиксированы и документированы, расширяются минорно.

---

## Приёмочные проверки

* Гистограммы имеют стабильные buckets (конфигурируемые глобально, но одинаковые для однотипных метрик).
* Любая запись лога сериализуется с точным порядком полей и без `null`.
* Loki успешно парсит JSON с указанным `pipeline_stages`; кардинальность лейблов остаётся в безопасных пределах.
* Трейсы содержат именованные спаны и атрибуты из раздела, события `retry/circuit_open/spa_fallback` фиксируются.

---

## Примеры сценариев

### Успешный `static` с `sendfile`

* Метрики: инкременты `anygate_http_requests_total{status_class="2xx"}`, `anygate_static_responses_total{kind="hit"}`, рост `anygate_static_sendfile_bytes_total`.
* Лог: `component=static`, `phase=sendfile`, `code=AG-OK`, `status=200`.
* Трейс: `action.static` без дочерних спанов.

### Прокси с ретраем и брейкером

* Метрики: `anygate_proxy_retries_total{reason="read_timeout"}` + `anygate_proxy_circuit_open_total`.
* Лог: события с `retry=true`, `attempt>1`, затем `code=AG-UPS-2006`.
* Трейс: события `retry` и `circuit_open` в `action.proxy`.

---

## Эволюция таксономии

* Новые метрики добавляются с документированными лейблами и без изменения существующих имен.
* Расширение наборов значений лейблов (`result`, `reason` и т. п.) — минорно, с обратной совместимостью.
* Изменение порядка/состава полей логов — только через RFC и мажорный релиз; ключ `log_version` добавляется при необходимости миграций.

---

## Рекомендации по дашбордам и алертам

* **Дашборды:**
  — Карта RPS/latency по `route_id` и `action`;
  — Ошибки `5xx` с разбивкой `result` для апстрима;
  — Состояние брейкера и пулы соединений;
  — Статика: hit/miss, сжатые/несжатые байты.
* **Алерты:**
  — Спайк `5xx`/`4xx` на route;
  — Увеличение `anygate_http_request_duration_seconds` p95;
  — Частые `timeout_*` в `anygate_proxy_upstream_requests_total`;
  — `anygate_admin_config_reloads_total{result="failure"}` > 0.

---

## Приложение: поля `labels` для Loki

Рекомендуемый набор низкой кардинальности:

```
labels = {
  "env":        "<dev|staging|prod>",
  "service":    "anygate",
  "component":  "proxy|static|router|filters|tls|runtime|admin|http",
  "action":     "proxy|static|fixed|echo",
  "version":    "<semver>"
}
```

Опционально (с осторожностью): `status_class`, `route_id` (если число маршрутов ограничено).
