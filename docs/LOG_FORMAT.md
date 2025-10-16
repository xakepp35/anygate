# LOG_FORMAT.md — структура логов и маскирование секретов

## Цель

Единый, детерминированный формат логов для потоковой доставки в Loki/Elastic и офлайн-разбор. Гарантируется фиксированный порядок полей, строгие типы, маскирование секретов и отсутствие `null`/многострочных значений. Все записи — UTF-8, одна строка на событие (JSON Lines) или канонический CBOR.

---

## Таймштампы

* Поле `ts` — **RFC3339Nano** с **ровно девятью** долями секунды, зона — **UTC**, пример:
  `2025-10-17T13:37:42.123456789Z`
* Регулярное выражение в линтере:
  `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z$`

---

## Порядок и состав полей (жёстко фиксированы)

Поля сериализуются **строго в следующем порядке**; отсутствующие значения **опускаются** (ключ не выводится):

`ts`, `level`, `severity`, `code`, `message`, `component`, `phase`, `trace_id`, `span_id`, `route_id`, `route_pattern`, `action`, `method`, `scheme`, `host`, `port`, `path`, `query`, `status`, `duration_ms`, `bytes_rx`, `bytes_tx`, `client_ip`, `client_port`, `user_agent`, `upstream`, `upstream_ip`, `attempt`, `retry`, `policy`, `secure`, `labels`, `extras`

### Типы и ограничения

| Поле            | Тип     | Ограничения/заметки                                                                                  |
| --------------- | ------- | ---------------------------------------------------------------------------------------------------- |
| `ts`            | string  | RFC3339Nano, UTC                                                                                     |
| `level`         | string  | `TRACE` `DEBUG` `INFO` `WARN` `ERROR`                                                                |
| `severity`      | integer | `TRACE=10` `DEBUG=20` `INFO=30` `WARN=40` `ERROR=50`                                                 |
| `code`          | string  | `AG-[A-Z]{3}-\d{4}` или `AG-OK`                                                                      |
| `message`       | string  | Одна строка, без переносов, без секретов                                                             |
| `component`     | string  | `listener` `http` `router` `filters` `proxy` `static` `tls` `admin` `runtime`                        |
| `phase`         | string  | `parse` `match` `pre` `action` `post` `write` `upstream_connect` `upstream_io` `compress` `sendfile` |
| `trace_id`      | string  | 16/32 hex (W3C Trace-Context)                                                                        |
| `span_id`       | string  | 16 hex                                                                                               |
| `route_id`      | string  | Стабильный идентификатор правила (хэш)                                                               |
| `route_pattern` | string  | Шаблон пути (`/users/:id`), может быть вырезан политикой                                             |
| `action`        | string  | `fixed` `echo` `static` `proxy`                                                                      |
| `method`        | string  | Любой HTTP-метод (токен)                                                                             |
| `scheme`        | string  | `http` `https`                                                                                       |
| `host`          | string  | Имя хоста запроса                                                                                    |
| `port`          | integer | Порт клиента/сервера по месту поля                                                                   |
| `path`          | string  | Путь запроса, без декодирования                                                                      |
| `query`         | string  | Строка запроса, **с редакцией секретов**                                                             |
| `status`        | integer | HTTP-код ответа                                                                                      |
| `duration_ms`   | integer | Общая длительность запроса (мс)                                                                      |
| `bytes_rx`      | integer | Принято байт тела                                                                                    |
| `bytes_tx`      | integer | Отдано байт тела                                                                                     |
| `client_ip`     | string  | IP клиента (маскирование опционально)                                                                |
| `client_port`   | integer | Порт клиента                                                                                         |
| `user_agent`    | string  | Без переносов                                                                                        |
| `upstream`      | string  | Логическое имя/`host:port` апстрима (без userinfo)                                                   |
| `upstream_ip`   | string  | Фактический адрес соединения                                                                         |
| `attempt`       | integer | Номер попытки (ретраи), начиная с `1`                                                                |
| `retry`         | boolean | Признак повтора                                                                                      |
| `policy`        | string  | Короткое имя политики (`rate_limit`, `ip_deny`, …)                                                   |
| `secure`        | boolean | Флаг включённого режима secure-логирования                                                           |
| `labels`        | object  | Низкая кардинальность: `env`, `service`, `component`, `action`, `version`                            |
| `extras`        | object  | Доп. поля плагинов; ключи неймспейсить: `plg_<name>_*`                                               |

**Запреты:** `null`, пустые ключи, многострочные строки, сырые бинарные данные.

---

## Каналы вывода

* **JSON Lines:** по одной записи на строку, без пробельного форматирования.
* **CBOR:** канонический (RFC 8949) — детерминированная сортировка ключей; те же ключи/типы, что и в JSON.

---

## Маскирование секретов (redaction)

### Источники секрета

* Заголовки запроса/ответа.
* Строка запроса (`query`), параметры URI.
* Параметры/значения из `extras` (плагины).
* Политически чувствительные сетевые поля (опционально).

### Правила

####

**Ключи для безусловного сокрытия** (регистронезависимо, для заголовков, параметров URL, JSON-ключей в `extras`):

```
authorization, proxy-authorization, cookie, set-cookie,
x-api-key, x-api-token, x-auth-token, x-amz-security-token,
api_key, api-key, access_token, id_token, refresh_token, token,
password, passwd, secret, client_secret, private_key, signing_key,
session, session_id
```

####

**Стратегии редактирования**

* Значение заменяется на строку `***`.
* Для `cookie`/`set-cookie`:

  * в логах остаются **только имена** куков; значения маскируются: `name=***; Path=/; HttpOnly`
  * атрибуты `Expires/Max-Age/Domain/SameSite/Secure/HttpOnly` сохраняются.
* Для `authorization`/`proxy-authorization` — всегда `***`.
* Для `query`: парсинг пары `k=v`; для секретных ключей — `k=***`; остальное — как есть.
* Для `upstream`/URI — **userinfo удаляется** (если встречается).
* Для `client_ip` при включённом `privacy_mask`:

  * IPv4 → обрезка до `/24` (последний октет `0`), IPv6 → до `/56`.

####

**Алгоритм применения**

1. Источник данных → нормализация ключей к нижнему регистру.
2. Сопоставление с «чёрным списком».
3. Замена значения на `***`.
4. Сборка обратно в компактную строку (знаки `&`, `;`, `,` и т. п. сохраняются).

---

## Линт-правила (лог-линтер)

* `ts` соответствует `RFC3339Nano` (ровно 9 знаков долей секунды, `Z`).
* `level` и `severity` согласованы (таблица сопоставления).
* `code` соответствует `AG-[A-Z]{3}-\d{4}` или `AG-OK`.
* Поля идут **строго** в заданном порядке; левые/правые пробелы отсутствуют.
* Запрещены `\n`, `\r`, `\t` внутри значений (кроме `\t` в `query`, если пришёл из клиента — линтер предупреждает и нормализует).
* Отсутствуют `null` и пустые строки; отсутствующие значения не сериализуются.
* `labels` — ключи/значения строковые; не более десяти ключей; только низкая кардинальность.
* `extras` — все ключи начинаются с `plg_`; отсутствуют секретные ключи (после редакции допускаются, но значения — `***`).
* Для `status>=400` — `level` не ниже `WARN`; для внутренних сбоев — `ERROR`.

---

## Тесты редактирования (приёмочные фикстуры)

### Заголовки и query

**Вход (фрагменты источника):**

```
Authorization: Bearer abc.def.ghi
Cookie: sid=xyz; other=ok
X-API-Key: secretkey
GET /v1?a=1&access_token=123&user=john
```

**Ожидаемый лог (фрагменты полей):**

```json
{
  "query":"a=1&access_token=***&user=john",
  "extras":{"plg_example_note":"***"}  // если плагин передал секрет
}
```

*(В теле записи `Cookie` в явном виде не логируется; если логируется — значения маскируются: `sid=***; other=***`.)*

### Set-Cookie

**Источник ответа:**

```
Set-Cookie: sid=abcd; HttpOnly; Path=/; Secure
```

**Ожидание в логе:**

```
"extras":{"plg_headers_set_cookie":"sid=***; HttpOnly; Path=/; Secure"}
```

### Userinfo в URI

**Вход:**

```
upstream: "https://user:pass@up.example.com:443"
```

**Ожидание:**

```
"upstream":"up.example.com:443"
```

---

## Примеры записей

### Успешный запрос (proxy)

```json
{"ts":"2025-10-17T13:37:42.123456789Z","level":"INFO","severity":30,"code":"AG-OK","message":"request served","component":"proxy","phase":"action","trace_id":"9f09e6e1c8a44b58","span_id":"b1f3aa22e9f1d011","route_id":"r:7d2c9e4b","route_pattern":"/v1/*","action":"proxy","method":"GET","scheme":"https","host":"api.example.com","port":443,"path":"/v1/items","query":"q=1&access_token=***","status":200,"duration_ms":12,"bytes_rx":0,"bytes_tx":342,"client_ip":"203.0.113.10","client_port":51820,"user_agent":"curl/8.5.0","upstream":"auth:8000","upstream_ip":"10.0.0.12:8000","attempt":1,"retry":false,"policy":"","secure":true,"labels":{"env":"prod","service":"anygate","component":"proxy","action":"proxy","version":"1.0.0"},"extras":{"plg_logger_sampled":true}}
```

### Ошибка апстрима (timeout + retry)

```json
{"ts":"2025-10-17T13:37:43.000000111Z","level":"ERROR","severity":50,"code":"AG-UPS-2003","message":"Gateway timeout: upstream did not respond","component":"proxy","phase":"upstream_io","trace_id":"1a2b3c4d5e6f7a8b","span_id":"0000aa11bb22cc33","route_id":"r:7d2c9e4b","route_pattern":"/v1/*","action":"proxy","method":"GET","scheme":"https","host":"api.example.com","port":443,"path":"/v1/items","status":504,"duration_ms":1005,"bytes_rx":0,"bytes_tx":0,"client_ip":"203.0.113.10","client_port":51820,"upstream":"auth:8000","upstream_ip":"10.0.0.12:8000","attempt":2,"retry":true,"policy":"retry","secure":true,"labels":{"env":"prod","service":"anygate","component":"proxy","action":"proxy","version":"1.0.0"}}
```

### Статика (sendfile, частичный контент)

```json
{"ts":"2025-10-17T13:38:00.999999999Z","level":"INFO","severity":30,"code":"AG-OK","message":"static range served","component":"static","phase":"sendfile","trace_id":"abcdabcdabcdabcd","span_id":"efefefefefefefef","route_id":"r:12aa34bb","route_pattern":"/","action":"static","method":"GET","scheme":"http","host":"site.example.com","port":80,"path":"/assets/app.js","status":206,"duration_ms":3,"bytes_rx":0,"bytes_tx":65536,"client_ip":"198.51.100.5","client_port":44321,"secure":true,"labels":{"env":"prod","service":"anygate","component":"static","action":"static","version":"1.0.0"}}
```

---

## Рекомендации для Loki

* Лейблы берём из поля `labels`. Добавочные лейблы на стадии promtail: `level`, `component`, `action`, производный `status_class` (из `status`).
* Не поднимайте в лейблы высококардинальные поля (`trace_id`, `path`, `upstream_ip`, `user_agent`).

---

## Производительность и устойчивость парсинга

* Ключи фиксированы; порядок полей стабилен независимо от содержимого.
* Значения — атомарные; отсутствуют вложенные структуры, кроме `labels` и `extras`.
* Секреты редактируются **до** сериализации, чтобы регэкспы разборщиков не «горели» на больших строках.

---

## Приёмка (линт + redaction)

* Линтер подтверждает: формат `ts`, порядок/типы полей, допустимые значения `level/severity/code`, отсутствие `null`/переносов, кардинальность `labels`.
* Набор redaction-фикстур проходит: заголовки/куки/параметры/URI полностью отмаскированы (`***`), userinfo удалён, query собран обратно корректно.
* Любое добавление поля/изменение порядка допускается только через RFC и мажорный релиз; для миграций разрешается временное дублирование ключей в `extras` с префиксом `compat_*` (не рекомендуется).
