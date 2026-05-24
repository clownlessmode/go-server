# Beeline Admin API

Документация REST API для управления SIM-картами Билайн и историей платежей.  
Используется админкой / конфиг-вебом; через этот API данные попадают в MITM-прокси и подменяют ответы мобильного приложения Билайн.

### Изменения для фронта

**Breaking:** из ответов `GET /sims`, `GET /sims/{number}` и `POST /sims` удалено поле `balance`.

**Breaking:** удалён `PATCH /sims/{number}/config/balance` — баланс **нельзя** задавать вручную.

**Breaking:** из `GET /sims/{number}/config` удалено поле `baseBalance`. Добавлено `incomingTotal`. Баланс считается из snapshot детализации + платежи за период.

**Breaking:** новый `GET /sims/{number}/detalization` — полная история как в приложении Билайн (`data.transactions`, `data.balances`, `data.categories`). У каждой транзакции есть `id` и `source` (`beeline` | `payment`).

**Breaking:** новый `DELETE /sims/{number}/detalization/transactions/{id}` — скрыть **реальную** операцию Билайн (`source=beeline`). Повторный вызов безопасен (`204`), операция **не вернётся** даже после повторного запроса детализации или обновления snapshot из приложения. Наши платежи (`source=payment`) скрывать этим роутом нельзя — только `DELETE /payments/{id}`.

---

## Базовые сведения

| Параметр | Значение |
|----------|----------|
| Base URL | `http://<host>:8080` |
| Prefix | `/banks/beeline` |
| Content-Type | `application/json` |
| Авторизация | **нет** (эндпоинты открыты) |
| Swagger | `GET /swagger/index.html` (теги `beeline sims`, `beeline config`, `beeline payments`) |

Все даты в **ответах** — ISO 8601 UTC (`2026-05-23T09:07:47Z`).  
В **запросах** поле `paidAt` — строка в формате **RFC3339** (можно с offset, напр. `2026-05-23T12:07:47+03:00`).

---

## Номер SIM

- Формат: **10 цифр**, без `+7`, без пробелов.
- Примеры валидных значений: `9680659702`
- Сервер нормализует номер из path/body:
  - убирает `+7`, ведущую `7` или `8` (если 11 цифр)
  - `79680659702` → `9680659702`

Ошибка при неверном формате: `400` + `{ "error": "number must be 10 digits without +7" }`

---

## Модели данных

### Sim

```json
{
  "number": "9680659702",
  "createdAt": "2026-05-22T10:00:00Z",
  "updatedAt": "2026-05-22T12:00:00Z"
}
```

| Поле | Тип | Описание |
|------|-----|----------|
| `number` | string | 10-значный номер |
| `createdAt` | string | Дата создания |
| `updatedAt` | string | Дата обновления |

> Баланс **не возвращается** в эндпоинтах SIM. Просмотр — только `GET /sims/{number}/config`.

---

### Config

```json
{
  "number": "9680659702",
  "balance": 48000.54,
  "paymentsTotal": 14999.46,
  "incomingTotal": 13000,
  "createdAt": "2026-05-22T10:00:00Z",
  "updatedAt": "2026-05-22T12:00:00Z"
}
```

| Поле | Тип | Описание |
|------|-----|----------|
| `number` | string | Номер SIM |
| `balance` | number \| null | Текущий баланс (как в приложении). Считается из snapshot детализации Билайн + платежи за период snapshot |
| `paymentsTotal` | number | Сумма исходящих платежей (`total`) за период snapshot |
| `incomingTotal` | number | Сумма входящих платежей (`amount`) за период snapshot |
| `createdAt` | string | |
| `updatedAt` | string | |

**Как считается `balance`:** берётся baseline из snapshot детализации (сохраняется, когда пользователь открывает детализацию в приложении), затем к нему применяются платежи из БД за тот же период — та же логика, что в MITM-прокси.

Если snapshot ещё нет (детализацию в приложении не открывали), `balance` будет `null`. После первого открытия детализации в приложении через прокси snapshot появится автоматически.

**Изменить баланс можно только через платежи** — `POST/PATCH/DELETE /sims/{number}/payments`.

---

### Detalization (полная история)

```json
{
  "number": "9609177131",
  "periodStart": "2026-04-24T00:00:00Z",
  "periodEnd": "2026-05-23T23:59:59Z",
  "balance": 48000.54,
  "data": {
    "transactions": [
      {
        "id": "a1b2c3...",
        "source": "beeline",
        "dateTime": "2026-05-22T16:20:40",
        "category": "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
        "name": "списание за мобильную коммерцию",
        "balances": [{ "changeValue": -70, "startValue": 105.56, "endValue": 35.56 }]
      },
      {
        "id": "payment-uuid",
        "source": "payment",
        "dateTime": "2026-05-23T12:07:47",
        "name": "списание за мобильную коммерцию",
        "balances": [{ "changeValue": -13845 }]
      }
    ],
    "balances": [{ "startValue": 50000, "endValue": 48000.54 }],
    "categories": []
  },
  "createdAt": "...",
  "updatedAt": "..."
}
```

| Поле | Описание |
|------|----------|
| `data` | Полный объект детализации как в API Билайн (`/mobile/api/v2/detalization`) |
| `data.transactions[].id` | Стабильный id транзакции |
| `data.transactions[].source` | `beeline` — реальная операция из Билайн; `payment` — наша из БД |
| `balance` | Итоговый баланс после скрытых + наших платежей |

Скрыть реальную операцию Билайн (`source=beeline`):

```http
DELETE /banks/beeline/sims/{number}/detalization/transactions/{id}
```

- Idempotent: повторный `DELETE` → `204`, транзакция не возвращается
- Сохраняется в БД навсегда (по `id` + fingerprint), вычищается из snapshot
- Не работает для `source=payment` → `400`, используй `DELETE /payments/{id}`

---

### Payment

```json
{
  "id": "a1b2c3d4e5f6789012345678901234ab",
  "simNumber": "9680659702",
  "direction": "outgoing",
  "receiverCard": "220094**0028",
  "amount": 13000,
  "commission": 845,
  "total": 13845,
  "source": "manual",
  "paidAt": "2026-05-23T09:07:47Z",
  "createdAt": "2026-05-23T09:08:00Z",
  "updatedAt": "2026-05-23T09:08:00Z"
}
```

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | string | UUID-like hex (32 символа), генерируется сервером |
| `simNumber` | string | Номер SIM |
| `direction` | `"outgoing"` \| `"incoming"` | Тип операции |
| `receiverCard` | string | Маска карты получателя. Только для `outgoing`. Формат: `220094**0028` |
| `amount` | number | Сумма платежа (без комиссии для outgoing) |
| `commission` | number | Комиссия 6.5% для outgoing, `0` для incoming |
| `total` | number | `amount + commission` (outgoing) или `amount` (incoming) |
| `source` | `"manual"` \| `"payment_flow"` | `manual` — создан через API; `payment_flow` — перехвачен из реального платежа в приложении |
| `paidAt` | string | Дата/время операции |
| `createdAt` | string | |
| `updatedAt` | string | |

**Комиссия (outgoing):** `commission = round(amount * 0.065, 2)`  
**Минимальная сумма outgoing:** `924` руб.

**Incoming (пополнение):**
- `receiverCard` не нужна (пустая)
- `commission = 0`, `total = amount`
- `amount > 0`

---

### Error

```json
{
  "error": "sim not found"
}
```

---

## Эндпоинты

### SIM-карты

#### `GET /banks/beeline/sims`

Список всех зарегистрированных SIM.

**Response `200`:** массив `Sim[]`, сортировка по номеру ASC.

```http
GET /banks/beeline/sims
```

---

#### `POST /banks/beeline/sims`

Регистрация новой SIM.

**Body:**

```json
{
  "number": "9680659702"
}
```

| Поле | Обязательное | Описание |
|------|--------------|----------|
| `number` | да | 10 цифр |

**Response `201`:** `Sim`

**Ошибки:**

| Code | error |
|------|-------|
| 400 | `invalid request body` |
| 400 | `number must be 10 digits without +7` |
| 409 | `sim already exists` |

---

#### `GET /banks/beeline/sims/{number}`

Получить SIM по номеру.

**Response `200`:** `Sim`

**Ошибки:**

| Code | error |
|------|-------|
| 404 | `sim not found` |

---

#### `DELETE /banks/beeline/sims/{number}`

Удалить SIM и **всю** историю платежей (каскадно).

**Response `204`:** без тела

**Ошибки:**

| Code | error |
|------|-------|
| 404 | `sim not found` |

---

### Конфиг и баланс

#### `GET /banks/beeline/sims/{number}/config`

Конфиг SIM: баланс из snapshot детализации + платежи за период.

**Response `200`:** `Config`

**Ошибки:**

| Code | error |
|------|-------|
| 404 | `sim not found` |

---

### Детализация (полная история)

#### `GET /banks/beeline/sims/{number}/detalization`

Полная история как в приложении Билайн: snapshot + наши платежи − скрытые операции.

**Response `200`:** `DetalizationResponse`

**Ошибки:**

| Code | error |
|------|-------|
| 404 | `sim not found` |
| 404 | `detalization snapshot not found, open detalization in Beeline app first` |

---

#### `DELETE /banks/beeline/sims/{number}/detalization/transactions/{id}`

Скрыть реальную операцию Билайн (`source=beeline`). Idempotent — повторный вызов тоже `204`, операция не появится снова.

**Response `204`:** без тела

**Ошибки:**

| Code | error |
|------|-------|
| 404 | `sim not found` |
| 400 | `use DELETE /payments/{id} to remove configured payments` |

---

### Платежи

Список платежей: **новые сверху** (`paidAt DESC`).

#### `GET /banks/beeline/sims/{number}/payments`

**Response `200`:** `Payment[]`

**Ошибки:**

| Code | error |
|------|-------|
| 404 | `sim not found` |

---

#### `POST /banks/beeline/sims/{number}/payments`

Создать платёж вручную.

**Body (outgoing — списание через мобильную коммерцию):**

```json
{
  "direction": "outgoing",
  "receiverCard": "220094**0028",
  "amount": 13000,
  "paidAt": "2026-05-23T12:07:47+03:00"
}
```

**Body (incoming — пополнение баланса):**

```json
{
  "direction": "incoming",
  "amount": 5000,
  "paidAt": "2026-05-23T12:07:47+03:00"
}
```

| Поле | Обязательное | Описание |
|------|--------------|----------|
| `direction` | нет (default: `outgoing`) | `outgoing` или `incoming` |
| `receiverCard` | да для outgoing | `^\d{6}\*\*\d{4}$` |
| `amount` | да | для outgoing ≥ 924 |
| `paidAt` | да | RFC3339 |

**Response `201`:** `Payment`

**Ошибки:**

| Code | error |
|------|-------|
| 400 | `invalid request body` |
| 400 | `invalid paidAt, expected RFC3339` |
| 400 | `invalid direction, expected outgoing or incoming` |
| 400 | `receiverCard must match format 220094**0028` |
| 400 | `amount must be at least 924` |
| 400 | `invalid payment amount` |
| 404 | `sim not found` |

---

#### `GET /banks/beeline/sims/{number}/payments/{id}`

**Response `200`:** `Payment`

**Ошибки:**

| Code | error |
|------|-------|
| 404 | `payment not found` |

---

#### `PATCH /banks/beeline/sims/{number}/payments/{id}`

Частичное обновление. Передаются только изменяемые поля.

**Body (все поля опциональны):**

```json
{
  "direction": "outgoing",
  "receiverCard": "220094**0028",
  "amount": 15000,
  "paidAt": "2026-05-23T14:00:00+03:00"
}
```

При смене `amount` у outgoing комиссия пересчитывается автоматически.  
При смене direction на `incoming` — `receiverCard` очищается, комиссия = 0.

**Response `200`:** `Payment`

**Ошибки:** те же validation-ошибки, что при создании + `404 payment not found`

---

#### `DELETE /banks/beeline/sims/{number}/payments/{id}`

**Response `204`:** без тела

**Ошибки:**

| Code | error |
|------|-------|
| 404 | `payment not found` |

---

## Рекомендуемые экраны для фронтенда

### 1. Список SIM

- `GET /banks/beeline/sims`
- Кнопка «Добавить SIM» → модалка с полем номера → `POST /banks/beeline/sims`
- Клик по SIM → экран деталей
- Swipe / кнопка удаления → `DELETE /banks/beeline/sims/{number}` (confirm)

### 2. Дашборд SIM

- `GET /banks/beeline/sims/{number}/config`
- Показать:
  - **Текущий баланс** (`balance`) — `null`, если snapshot ещё не захвачен из приложения
  - **Списано (outgoing)** (`paymentsTotal`)
  - **Пополнено (incoming)** (`incomingTotal`)
- Баланс меняется только через платежи (экраны 3–5)

### 3. История платежей

- `GET /banks/beeline/sims/{number}/payments`
- Таблица/список: дата, direction, amount, commission, total, source, receiverCard
- Фильтр/бейдж по `direction`: incoming / outgoing
- Бейдж `source`: manual / auto (payment_flow)

### 4. Создание платежа

- Переключатель direction: outgoing / incoming
- Outgoing: поля amount, receiverCard (маска `XXXXXX**XXXX`), paidAt (datetime picker)
- Incoming: amount, paidAt
- Live-preview: commission и total для outgoing
- Submit → `POST .../payments` → обновить config и список

### 5. Редактирование / удаление платежа

- `GET .../payments/{id}` для формы редактирования
- `PATCH .../payments/{id}`
- `DELETE .../payments/{id}`

---

## Примеры curl

```bash
# Список SIM
curl -s http://localhost:8080/banks/beeline/sims

# Создать SIM
curl -s -X POST http://localhost:8080/banks/beeline/sims \
  -H 'Content-Type: application/json' \
  -d '{"number":"9680659702"}'

# Конфиг с балансом
curl -s http://localhost:8080/banks/beeline/sims/9680659702/config

# Создать outgoing платёж
curl -s -X POST http://localhost:8080/banks/beeline/sims/9680659702/payments \
  -H 'Content-Type: application/json' \
  -d '{
    "direction": "outgoing",
    "receiverCard": "220094**0028",
    "amount": 13000,
    "paidAt": "2026-05-23T12:07:47+03:00"
  }'

# Создать incoming (пополнение)
curl -s -X POST http://localhost:8080/banks/beeline/sims/9680659702/payments \
  -H 'Content-Type: application/json' \
  -d '{
    "direction": "incoming",
    "amount": 5000,
    "paidAt": "2026-05-23T12:07:47+03:00"
  }'

# Список платежей
curl -s http://localhost:8080/banks/beeline/sims/9680659702/payments

# Удалить платёж
curl -s -X DELETE http://localhost:8080/banks/beeline/sims/9680659702/payments/{id}
```

---

## TypeScript типы (для фронтенда)

```typescript
export interface Sim {
  number: string;
  createdAt: string;
  updatedAt: string;
}

export interface Config {
  number: string;
  balance: number | null;
  paymentsTotal: number;
  incomingTotal: number;
  createdAt: string;
  updatedAt: string;
}

export type PaymentDirection = 'outgoing' | 'incoming';
export type PaymentSource = 'manual' | 'payment_flow';

export interface Payment {
  id: string;
  simNumber: string;
  direction: PaymentDirection;
  receiverCard?: string;
  amount: number;
  commission: number;
  total: number;
  source: PaymentSource;
  paidAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface ApiError {
  error: string;
}

export interface CreateSimRequest {
  number: string;
}

export interface CreatePaymentRequest {
  direction?: PaymentDirection;
  receiverCard?: string;
  amount: number;
  paidAt: string;
}

export interface UpdatePaymentRequest {
  direction?: PaymentDirection;
  receiverCard?: string;
  amount?: number;
  paidAt?: string;
}
```

---

## Связь с приложением Билайн (контекст для UI)

Эти данные не просто хранятся в БД — прокси подменяет ответы API Билайн:

- **Баланс** в приложении берётся из snapshot детализации + расчёт по платежам
- **Детализация** (`/mobile/api/v2/detalization`) дополняется платежами из БД
- Платежи с `source: "payment_flow"` создаются автоматически при реальной оплате через приложение (API их только читает/редактирует/удаляет)

Фронтенду достаточно работать только с REST API выше; прокси настраивается отдельно на сервере.

---

## Чеклист валидации на клиенте

- [ ] Номер SIM: ровно 10 цифр
- [ ] `receiverCard`: regex `^\d{6}\*\*\d{4}$` для outgoing
- [ ] `amount` ≥ 924 для outgoing
- [ ] `amount` > 0 для incoming
- [ ] `paidAt`: валидный ISO datetime (RFC3339)
- [ ] Preview commission: `Math.round(amount * 0.065 * 100) / 100`
- [ ] Preview total (outgoing): `amount + commission`
