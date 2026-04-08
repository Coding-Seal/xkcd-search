# Отчёт по модульному тестированию

**Проект:** XKCD Comics Aggregator (xs)
**Дата:** 08.04.2026
**Команда:** 2 человека

---

## 1. Описание выполненной работы

### 1.1 Объект тестирования

REST API + веб-фронтенд для агрегации и поиска комиксов XKCD. Проект состоит из двух исполняемых файлов:
- `cmd/xkcd/` — REST API сервер (порт 8080), реализованный на Go
- `cmd/web/` — Веб-сервер фронтенда (порт 8090), проксирует запросы к API

Архитектура — гексагональная (ports-and-adapters):
- `internal/core/services/` — бизнес-логика (Search, Fetcher, User, Comic)
- `internal/adapters/repos/` — репозитории (SQLite + in-memory индекс)
- `internal/adapters/rest/` — HTTP-обработчики, JWT-аутентификация, rate limiting
- `web/` — шаблоны и REST-клиент веб-фронтенда
- `pkg/` — вспомогательные утилиты

### 1.2 Используемые инструменты

| Инструмент | Назначение |
|------------|------------|
| `go test` | Встроенный фреймворк тестирования Go |
| `testify/assert` | Удобные assert-функции |
| `testify/require` | Критичные проверки (останавливают тест при ошибке) |
| `net/http/httptest` | HTTP-сервер/рекордер для тестирования обработчиков |
| `database/sql` + SQLite in-memory | Изолированная тестовая БД |
| `-race` флаг | Детектор гонок данных |
| `-coverprofile` | Измерение покрытия кода |

Тестирование запускается автоматически при сборке:

```
make test → go test -race -coverprofile build/cover.out ./...
```

### 1.3 Классификация тестов

#### По уровню изоляции

Модульные тесты обеспечивают **полную изоляцию** тестируемого модуля. Все внешние зависимости (репозитории, сервисы, HTTP-клиенты, БД) заменены локальными моками, реализующими соответствующие интерфейсы из `internal/core/ports/`. SQLite-репозитории тестируются с базой данных in-memory.

#### По подходу к тестированию

| Подход | Описание | Применение |
|--------|----------|------------|
| **Белый ящик (white-box)** | Тестировщик знает внутреннюю реализацию, проверяет конкретные ветвления | Сервисы (проверяем вызовы моков), репозитории (проверяем SQL-поведение), rate limiter (проверяем состояние счётчиков) |
| **Чёрный ящик (black-box)** | Проверяется только интерфейс: входные данные → выходные данные | HTTP-обработчики через `httptest`, шаблоны (подаём данные → проверяем HTML), парсеры конфига |

#### По назначению теста

| Назначение | Маркер | Кол-во | Описание |
|------------|--------|--------|----------|
| Штатный сценарий | EC (Expected Case) | 112 | Проверка основного поведения с корректными входными данными |
| Путь ошибки | EP (Error Path) | 47 | Проверка обработки ошибок: несуществующие объекты, сетевые ошибки, невалидный ввод |
| Граничное значение | BV (Boundary Value) | 36 | Тестирование на границах: нулевые ID, пустые строки, пустые репозитории |

#### По объекту тестирования

| Категория | Файлы | Тесты | Объект |
|-----------|-------|-------|--------|
| Сервисы | `*_test.go` в `internal/core/services/` | 19 | Search, Comic, User, Fetcher сервисы |
| Репозитории | `*_test.go` в `internal/adapters/repos/` | 37 | SQLite store для комиксов/пользователей, in-memory индекс, fetcher-репозиторий |
| REST API | `*_test.go` в `internal/adapters/rest/` | 34 | HTTP-обработчики, JWT-аутентификация, middleware, роутер, сервер |
| Веб-слой | `*_test.go` в `web/` | 41 | Веб-обработчики, HTML-шаблоны, REST-клиент, роутер |
| Утилиты | `*_test.go` в `pkg/` | 36 | Rate limiter, JSONL, stemmer/stop-words, XKCD-клиент, HTTP-хелперы |
| Конфигурация | `config/yaml_test.go` | 4 | Загрузка YAML-конфига, дефолтные значения |
| База данных | `db/setup_test.go` | 8 | Подключение, миграции вверх/вниз, idempotency |
| Контекст | `internal/contextutil/*_test.go` | 16 | Хранение/извлечение user ID, admin flag, request ID из контекста |
| **ИТОГО** | **30 файлов** | **195** | |

### 1.4 Применённые техники тест-дизайна

**Классы эквивалентности (КЭ)** — разбиение входных данных на классы с одинаковым ожидаемым поведением.

Пример: `TestComicService_Comic_Found` — ID существующего комикса; `TestComicService_Comic_NotFound` — несуществующий ID; `TestSqliteStore_Comics_Empty` — пустое хранилище; `TestClient_Login_Success` — корректные учётные данные.

**Граничные значения (ГЗ)** — тестирование на границах допустимых диапазонов.

Пример: `TestSqliteStore_Comic_ZeroID` — ID = 0 (ниже минимума); `TestSearch_ZeroLimit` — лимит = 0; `TestUserID_Zero` / `TestUserID_Negative` — граничные значения user ID в контексте; `TestParsePhrase_EmptyString` — пустая строка на вход парсеру.

**Попарное тестирование (ПП)** — проверка взаимодействия нескольких параметров одновременно.

Пример: `TestIntegration_Login_MissingCredentials` покрывает комбинации `{пустой login, пустой password}`: {пусто, заполнено}, {заполнено, пусто}, {оба пусты}; `TestContext_AllThreeValues_Coexist` — хранение трёх независимых значений в контексте одновременно.

**Путь ошибки (ПО)** — проверка обработки исключительных ситуаций.

Пример: `TestClient_Login_NetworkError` — сеть недоступна; `TestFetcher_LastComicID_UnreachableServer` — XKCD-сервер недоступен; `TestSqliteStore_Store_DuplicateIsError` — вставка дубликата; `TestMigrateDown_Idempotent` — повторный откат миграции.

### 1.5 Классы эквивалентности по модулям

#### Search-сервис (`internal/core/services/search.go`)

| Входной параметр | Класс эквивалентности | Тест-представитель | Значение |
|------------------|-----------------------|--------------------|----------|
| `query` | Непустой, есть совпадения | `TestSearch_SearchComics` | `"test"` → список комиксов |
| `query` | Непустой, нет совпадений | `TestSearch_Empty` | `"nonexistent"` → пустой список |
| `limit` | Ноль | `TestSearch_ZeroLimit` | `limit=0` → пустой результат |
| `comic` | Добавление нового | `TestSearch_AddComic` | ID=1, title="test" → индексировано |
| `repo` | Ошибка при построении | `TestSearch_SearchComics_RepoErrors` | репо возвращает error |

#### In-memory индекс (`internal/adapters/repos/search/index.go`)

| Входной параметр | Класс эквивалентности | Тест-представитель | Значение |
|------------------|-----------------------|--------------------|----------|
| `comic.Title` | Содержит ключевые слова | `TestIndex_SearchBySafeTitle` | поиск по заголовку |
| `comic.Transcription` | Содержит ключевые слова | `TestIndex_SearchByTranscription` | поиск по тексту |
| `query` | Пустой (после стемминга) | `TestIndex_SearchEmpty_AfterAddNoKeywords` | `""` → пусто |
| `query` | Промах кэша | `TestIndex_SearchMiss` | слово отсутствует в индексе |
| `repo` | Пустой репозиторий | `TestIndex_Build_EmptyRepo` | `Build()` → нет паники |
| `конкурентность` | Параллельные операции | `TestIndex_ConcurrentAddAndSearch` | гонки отсутствуют |
| `ранжирование` | Несколько совпадений | `TestIndex_Ranking_MoreMatchesHigherScore` | больше совпадений → выше в списке |

#### SQLite-репозиторий комиксов (`internal/adapters/repos/comic/sqlite.go`)

| Входной параметр | Класс эквивалентности | Тест-представитель | Значение |
|------------------|-----------------------|--------------------|----------|
| `id` | Существующий | `TestSqliteStore_Comic` | ID=1 → комикс найден |
| `id` | Несуществующий | `TestSqliteStore_Comic_NotFound` | ID=999 → ошибка |
| `id` | Граничный (ноль) | `TestSqliteStore_Comic_ZeroID` | ID=0 → ошибка |
| `comics` | Порядок | `TestSqliteStore_Comics_Order` | возврат в порядке ID |
| `store` | Дубликат | `TestSqliteStore_Store_DuplicateIsError` | повторная вставка → ошибка |
| `store` | Все поля | `TestSqliteStore_Store_PreservesFields` | все поля сохранены корректно |

#### JWT-аутентификация (`internal/adapters/rest/auth/`)

| Входной параметр | Класс эквивалентности | Тест-представитель | Значение |
|------------------|-----------------------|--------------------|----------|
| `token` | Отсутствует | `TestAuthenticate_NoToken` | нет заголовка → 401 |
| `token` | Валидный | `TestAuthenticate_Happy` | корректный JWT → 200 |
| `token` | Без expiration | `TestAuthenticate_NoExpiration` | нет `exp` поля → отклонён |
| `role` | Admin | `TestAuthorize_Happy` | `is_admin=true` → пропуск |
| `role` | Non-admin | `TestAuthorize` | `is_admin=false` → 401 |

#### Rate Limiter (`pkg/ratelimiter/limiter.go`)

| Входной параметр | Класс эквивалентности | Тест-представитель | Значение |
|------------------|-----------------------|--------------------|----------|
| `userID` | Первый запрос | `TestRateLimiter_AllowFirst` | всегда разрешён |
| `userID` | Разные пользователи | `TestRateLimiter_DifferentUsers_Independent` | счётчики независимы |
| `cleanup` | После истечения | `TestPerUser_CleanUp` | старые записи удаляются |

#### Stemmer / Stop words (`pkg/words/stemmer.go`)

| Входной параметр | Класс эквивалентности | Тест-представитель | Значение |
|------------------|-----------------------|--------------------|----------|
| `phrase` | Нормальная фраза | `TestStemmer` | `"running"` → `"run"` |
| `phrase` | Короткие слова | `TestStemmer_ShortWordsFiltered` | слова < 3 букв отфильтрованы |
| `phrase` | Stop-слова | `TestStemmer_StopWordFiltered` | `"the"`, `"a"` отфильтрованы |
| `phrase` | Пустая строка | `TestStemmer_EmptyInput` | `""` → пустой список |
| `phrase` | Только спецсимволы | `TestParsePhrase_OnlySpecialChars` | `"!@#"` → пустой список |
| `stopwords` | Пустой файл | `TestParseStopWords_Empty` | корректная обработка |

---

## 2. Результаты тестирования

### 2.1 Общие результаты

| Метрика | Значение |
|---------|----------|
| Всего тестов | 195 |
| Пройдено | 195 (100%) |
| Провалено | 0 |
| Обнаруженных гонок данных (`-race`) | 0 |
| Покрытие кода | **~88%** |
| Порог покрытия (требование) | 80% — пройден |

### 2.2 Покрытие по пакетам

| Пакет | Покрытие |
|-------|----------|
| `internal/adapters/repos/search` | 100% |
| `internal/adapters/rest/middleware` | 100% |
| `internal/contextutil` | 100% |
| `pkg/ratelimiter` | 100% |
| `web` | 100% |
| `web/templates` | 100% |
| `internal/core/services` | 95.5% |
| `internal/adapters/repos/fetcher` | 94.1% |
| `test/fetcher` | 93.8% |
| `internal/adapters/rest/handlers` | 93.1% |
| `web/handlers` | 92.2% |
| `pkg/jsonl` | 91.7% |
| `config` | 90.0% |
| `web/rest` | 89.1% |
| `internal/adapters/rest` | 87.3% |
| `internal/adapters/rest/auth` | 87.5% |
| `internal/adapters/repos/comic` | 85.2% |
| `internal/adapters/repos/user` | 84.6% |
| `pkg/xkcd` | 83.3% |
| `pkg/words` | 96.3% |
| `pkg/http-util` | 96.8% |
| `db` | 75.0% |
| `cmd/xkcd`, `cmd/web` | 0.0% (точки входа, только `main()`) |
| `internal/core/models`, `internal/core/ports` | — (только типы/интерфейсы) |

### 2.3 Распределение тестов по файлам

| Тестовый файл | Кол-во | Объект тестирования |
|---------------|--------|---------------------|
| `web/handlers/handlers_test.go` | 17 | Веб-обработчики (поиск, логин, избранное) |
| `web/rest/client_test.go` | 15 | REST-клиент (логин, поиск, обновление, комиксы) |
| `internal/adapters/repos/search/index_test.go` | 13 | In-memory инвертированный индекс |
| `pkg/words/stemmer_test.go` | 11 | Стеммер и stop-слова |
| `internal/adapters/rest/handlers/rest_test.go` | 10 | HTTP-обработчики поиска/обновления/комиксов |
| `internal/contextutil/auth_test.go` | 9 | Admin-флаг в контексте |
| `db/setup_test.go` | 8 | Подключение и миграции SQLite |
| `internal/adapters/repos/comic/sqlite_test.go` | 8 | CRUD комиксов в SQLite |
| `internal/adapters/repos/fetcher/fetcher_test.go` | 8 | HTTP-клиент XKCD API |
| `internal/adapters/repos/user/sqlite_test.go` | 8 | CRUD пользователей в SQLite |
| `internal/adapters/rest/server_test.go` | 8 | Инициализация сервера и маршрутов |
| `pkg/http-util/handlers_test.go` | 7 | WrapHandler: маппинг кодов ошибок |
| `internal/contextutil/id_test.go` | 7 | Request ID в контексте |
| `web/templates/templates_test.go` | 7 | HTML-шаблоны (Pics, Login, Update, Favorites) |
| `internal/core/services/user_test.go` | 6 | UserService (логин, поиск, добавление) |
| `internal/adapters/rest/middleware/limiter_test.go` | 5 | Rate-limit middleware по user/IP |
| `pkg/ratelimiter/limiter_test.go` | 5 | Rate limiter per user |
| `internal/adapters/rest/auth/auth_test.go` | 5 | JWT-аутентификация и авторизация |
| `internal/core/services/search_test.go` | 5 | Search сервис |
| `internal/core/services/comic_test.go` | 5 | Comic сервис |
| `pkg/jsonl/scanner_test.go` | 4 | JSONL-сканер |
| `internal/adapters/rest/auth/handlers_test.go` | 4 | Auth-обработчики (моки) |
| `pkg/http-util/middleware_test.go` | 4 | Chain, Logging, RequestID middleware |
| `config/yaml_test.go` | 4 | Загрузка YAML-конфига |
| `internal/core/services/fetcher_test.go` | 3 | Fetcher сервис |
| `pkg/xkcd/fetcher_test.go` | 3 | XKCD HTTP-клиент (бенчмарки) |
| `pkg/jsonl/writer_test.go` | 2 | JSONL-записатель |
| `web/routes_test.go` | 2 | Роутер веб-фронтенда |
| `internal/adapters/rest/routes_test.go` | 1 | API-роутер |
| `internal/adapters/rest/logger_test.go` | 1 | Настройка логгера |
| **ИТОГО** | **195** | |

### 2.4 Результаты на сервере непрерывной интеграции

Тесты запускаются автоматически через `make test` при каждом коммите. Запуск включает флаг `-race` для обнаружения гонок данных.

```
make test
Running tests ...
ok  yadro-go-course/config              coverage: 90.0%
ok  yadro-go-course/db                  coverage: 75.0%
ok  yadro-go-course/internal/adapters/repos/comic    coverage: 85.2%
ok  yadro-go-course/internal/adapters/repos/fetcher  coverage: 94.1%
ok  yadro-go-course/internal/adapters/repos/search   coverage: 100.0%
ok  yadro-go-course/internal/adapters/repos/user     coverage: 84.6%
ok  yadro-go-course/internal/adapters/rest           coverage: 87.3%
ok  yadro-go-course/internal/adapters/rest/auth      coverage: 87.5%
ok  yadro-go-course/internal/adapters/rest/handlers  coverage: 93.1%
ok  yadro-go-course/internal/adapters/rest/middleware coverage: 100.0%
ok  yadro-go-course/internal/contextutil             coverage: 100.0%
ok  yadro-go-course/internal/core/services           coverage: 95.5%
ok  yadro-go-course/pkg/http-util                    coverage: 96.8%
ok  yadro-go-course/pkg/jsonl                        coverage: 91.7%
ok  yadro-go-course/pkg/ratelimiter                  coverage: 100.0%
ok  yadro-go-course/pkg/words                        coverage: 96.3%
ok  yadro-go-course/pkg/xkcd                         coverage: 83.3%
ok  yadro-go-course/web                              coverage: 100.0%
ok  yadro-go-course/web/handlers                     coverage: 92.2%
ok  yadro-go-course/web/rest                         coverage: 89.1%
ok  yadro-go-course/web/templates                    coverage: 100.0%
```

Статус последнего запуска: **195/195 пройдено**, гонок данных: **0**, покрытие: **~88%**.

---

## 3. Процедура расширения тестового набора

При добавлении нового сервиса (например, `internal/core/services/stats.go`):

1. Создать файл `internal/core/services/stats_test.go` рядом с тестируемым файлом
2. Определить классы эквивалентности входных данных (позитивный, негативный, граничный)
3. Добавить мок нового порта в `mocks_test.go` — локальную структуру, реализующую интерфейс из `internal/core/ports/`
4. Написать тесты в формате `TestStats_<Сценарий>` с паттерном Arrange-Act-Assert
5. Добавить граничные случаи: пустой репозиторий, нулевые/отрицательные значения, ошибка зависимости
6. Проверить покрытие: `go test ./internal/core/services/... -run TestStats -v -cover`

**Пример:** для функции `GetTopComics(ctx, limit int) ([]Comic, error)`:

```go
func TestStats_GetTopComics_Found(t *testing.T) {
    // EC: непустой репозиторий, limit > 0
    repo := &mockComicRepo{comics: []models.Comic{{ID: 1}, {ID: 2}}}
    svc := services.NewStats(repo)
    result, err := svc.GetTopComics(context.Background(), 10)
    require.NoError(t, err)
    assert.Len(t, result, 2)
}

func TestStats_GetTopComics_ZeroLimit(t *testing.T) {
    // BV: limit = 0
    svc := services.NewStats(&mockComicRepo{})
    result, err := svc.GetTopComics(context.Background(), 0)
    require.NoError(t, err)
    assert.Empty(t, result)
}

func TestStats_GetTopComics_RepoError(t *testing.T) {
    // EP: репозиторий возвращает ошибку
    repo := &mockComicRepo{err: errors.New("db error")}
    svc := services.NewStats(repo)
    _, err := svc.GetTopComics(context.Background(), 10)
    assert.Error(t, err)
}
```

При добавлении нового HTTP-обработчика: использовать `net/http/httptest.NewRecorder()` + реальный роутер с мок-сервисом, проверять статус код и тело ответа через `testify/assert`.
