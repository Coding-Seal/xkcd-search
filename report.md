# Отчёт по курсовой работе
## Технологии разработки качественного программного обеспечения
### Весенний семестр

**Дисциплина:** Технологии разработки качественного программного обеспечения
**Группа:** [номер группы]
**Команда:** [имена участников]
**Проект:** XKCD Search — поисковый сервис по комиксам XKCD
**Репозиторий:** [ссылка на репозиторий]
**Дата:** 2026 г.

---

## Описание программного продукта

XKCD Search — веб-приложение и REST API для поиска комиксов XKCD по ключевым словам. Система включает:

- **Поисковый движок** — полнотекстовый поиск с использованием морфологического стемминга (Snowball) и фильтрации стоп-слов
- **REST API** — эндпоинты для поиска, авторизации и управления индексом
- **Веб-интерфейс** — отдельный веб-сервер с HTML-шаблонами
- **Аутентификация и авторизация** — JWT-токены, роли Admin/User
- **Избранные комиксы** — хранение и просмотр (до 25 на пользователя)
- **Автообновление индекса** — планировщик на основе cron, параллельная загрузка

**Стек технологий:** Go 1.22, SQLite, Docker, GitHub Actions
**Архитектурный паттерн:** Гексагональная архитектура (Ports & Adapters)

---

## Часть 1. Модульное тестирование

### 1.1 Описание выполненной работы

Модульное тестирование выполнено средствами стандартного пакета `testing` языка Go в сочетании с библиотекой `testify` (assert/require). Тесты запускаются автоматически при сборке в рамках CI/CD пайплайна (GitHub Actions, задача `build-and-test`, шаг `Unit test`).

**Используемые инструменты:**

| Инструмент | Назначение |
|---|---|
| Go `testing` | Стандартный фреймворк для юнит-тестов |
| `github.com/stretchr/testify` | Расширенные assert/require-утверждения |
| `go test -race` | Детектор гонок данных (race detector) |
| `go test -coverprofile` | Формирование отчёта о покрытии кода |
| `go tool cover -html` | HTML-визуализация покрытия |
| GitHub Actions | Автоматический запуск тестов при каждом push/PR |

**Применённые техники тест-дизайна:**

1. **Классы эквивалентности** — корректные/некорректные входные данные разбиты на классы. Пример: тесты для `UserService.Login` покрывают классы «пользователь существует», «пользователь не найден», «неверный пароль».

2. **Граничные условия** — тестирование на граничных значениях. Пример: длина строки поискового запроса (пустая строка, максимум 100 символов, строка из одних стоп-слов), количество избранных комиксов (0, 24, 25, 26).

3. **Отрицательные сценарии** — тесты, проверяющие корректное поведение при ошибках. Пример: попытка добавить дубликат в избранное, поиск с пустым индексом, загрузка несуществующего комикса.

4. **Табличные тесты (Table-driven tests)** — стандартный Go-паттерн, применённый во множестве тест-файлов. Пример из `handlers/rest_test.go` и `services/*_test.go` — массивы `testCases` с различными комбинациями входных данных.

**Покрытые модули (с применением мок-объектов там, где требуется изоляция):**

- `internal/core/services` — бизнес-логика (ComicService, SearchService, UserService, FetcherService)
- `internal/adapters/repos/comic` — SQLite-репозиторий комиксов
- `internal/adapters/repos/user` — SQLite-репозиторий пользователей
- `internal/adapters/repos/search` — поисковый индекс в памяти
- `internal/adapters/repos/fetcher` — загрузчик комиксов с xkcd.com
- `internal/adapters/rest` — HTTP-слой (сервер, роуты, middleware)
- `internal/adapters/rest/auth` — JWT-аутентификация
- `internal/adapters/rest/handlers` — HTTP-хэндлеры
- `internal/adapters/rest/middleware` — rate limiter, concurrency limiter
- `internal/contextutil` — хелперы контекста (auth info, request ID)
- `pkg/words` — стемминг и обработка слов
- `pkg/ratelimiter` — ограничение частоты запросов
- `pkg/jsonl` — работа с JSONL-форматом
- `pkg/xkcd` — HTTP-клиент для xkcd.com API
- `pkg/http-util` — HTTP-утилиты
- `config`, `db` — конфигурация и миграции БД
- `web`, `web/handlers`, `web/rest`, `web/templates` — веб-сервер

Для изоляции юнит-тестов от внешних зависимостей использованы:
- Моки портов (интерфейсов) в `internal/adapters/rest/handlers/mocks_test.go` и `internal/core/services/mocks_test.go`
- Временная база данных (`t.TempDir()`) для тестов репозиториев
- Тестовый HTTP-сервер (`httptest.Server`) для тестирования HTTP-клиентов

### 1.2 Отчёт о прохождении тестов и покрытие кода

Все тесты запускаются командой:

```bash
make test
# эквивалентно:
go test -race -coverprofile build/cover.out ./...
```

**Результаты прохождения тестов:**

| Пакет | Статус | Покрытие |
|---|---|---|
| `config` | PASS | — |
| `db` | PASS | — |
| `internal/adapters/repos/comic` | PASS | 84.6% |
| `internal/adapters/repos/fetcher` | PASS | — |
| `internal/adapters/repos/search` | PASS | — |
| `internal/adapters/repos/user` | PASS | 84.6% |
| `internal/adapters/rest` | PASS | 87.3% |
| `internal/adapters/rest/auth` | PASS | 87.5% |
| `internal/adapters/rest/handlers` | PASS | 93.1% |
| `internal/adapters/rest/middleware` | PASS | **100.0%** |
| `internal/contextutil` | PASS | **100.0%** |
| `internal/core/services` | PASS | 95.5% |
| `pkg/http-util` | PASS | 96.8% |
| `pkg/jsonl` | PASS | 91.7% |
| `pkg/ratelimiter` | PASS | **100.0%** |
| `pkg/words` | PASS | 96.3% |
| `pkg/xkcd` | PASS | 83.3% |
| `web` | PASS | **100.0%** |
| `web/handlers` | PASS | 92.2% |
| `web/rest` | PASS | 89.1% |
| `web/templates` | PASS | **100.0%** |

**Итоговое покрытие:** **82.6%** всех statement-ов кодовой базы
**Общее количество тест-функций:** 182 (367 с учётом подтестов)
**Статус:** все тесты пройдены успешно (все пакеты — PASS)
**Детектор гонок (`-race`):** нарушений не выявлено

Покрытие превышает установленный порог 80%, что соответствует максимальному баллу по критерию покрытия (81%+ → 25 баллов).

### 1.3 Реализация тестов

#### `internal/core/services` — бизнес-логика

Тесты находятся внутри пакета `services`, зависимости (репозитории) заменяются mock-объектами через `testify/mock`.

- **`TestComicService_Comic_Found`** — мок `comicRepo.Comic(7)` возвращает тестовый комикс, проверяет `assert.Equal`
- **`TestComicService_Comic_NotFound`** — мок возвращает `ports.ErrNotFound`, проверяет `assert.ErrorIs`
- **`TestComicService_Store`** — мок `comicRepo.Store(comic)` → `nil`, проверяет `assert.NoError`
- **`TestComicService_ComicsAll`** — мок возвращает срез из 2 комиксов, проверяет `assert.Len(got, 2)`
- **`TestUserService_UserLogin_Found`** — мок `userRepo.UserLogin("alice")` возвращает пользователя, `assert.Equal`
- **`TestUserService_UserLogin_NotFound`** — мок возвращает `ErrNotFound`, проверяет `assert.ErrorIs`
- **`TestUserService_AddUser`** — мок `userRepo.AddUser(user)` → `nil`, `assert.NoError`
- **`TestUserService_UserID`** — мок `userRepo.UserID(5)` возвращает bob, `assert.Equal`
- **`TestUserService_RemoveUser`** — мок `userRepo.RemoveUser(3)` → `nil`, `assert.NoError`
- **`TestSearch_AddComic`** — мок `searchRepo.AddComic(...)`, вызов не должен падать
- **`TestSearch_SearchComics`** — мок возвращает map из 8 комиксов с очками, передаётся лимит 6 — проверяет, что сервис правильно обрезает
- **`TestSearch_Empty`** — мок возвращает пустой map, `assert.Empty`
- **`TestSearch_ZeroLimit`** — лимит=0, результат должен быть пустым несмотря на совпадение
- **`TestSearch_SearchComics_RepoErrors`** — `comicRepo.Comic(1)` возвращает `ErrInternal`; комикс с ошибкой пропускается, результат пустой

#### `internal/adapters/repos/search` — поисковый индекс

Тесты работают с реальной реализацией `Index` и настоящим стеммером `words.NewStemmer(nil)`. Репозиторий комиксов изолирован локальным `comicRepoMock`.

- **`TestIndex_AddComic`** — добавляет комикс ID=42, затем `SearchComics` — проверяет `assert.Contains(found, 42)`
- **`TestIndex_Build`** — мок `ComicsAll()` → 3 комикса, после `Build` все три находятся через поиск
- **`TestIndex_SearchMiss`** — пустой индекс, `SearchComics("unicorn")` → пустой результат
- **`TestIndex_MultipleComics`** — 3 комикса, запрос "dinosaur" находит ID 1 и 2, но не 3 (`assert.NotContains`)
- **`TestIndex_EmptyQuery`** — пустой запрос `""` → пустой результат даже при непустом индексе
- **`TestIndex_Build_Error`** — мок `ComicsAll()` → `ErrInternal`, `index.Build` возвращает ошибку

#### `internal/adapters/rest/auth` — HTTP-хэндлер авторизации

Используется `httptest.NewRequest` / `httptest.NewRecorder` — реальный HTTP без запуска сервера. Пароли хэшируются bcrypt прямо в тесте.

- **`TestLogin_Happy`** — создаёт bcrypt-хэш пароля "bob", мок `UserLogin("bob")` возвращает пользователя, POST с JSON → `assert.NoError`
- **`TestLogin_NoSuchUser`** — мок возвращает `ErrNotFound` → хэндлер возвращает `http_util.ErrNotFound`
- **`TestLogin_WrongPassword`** — хэш для "12345", в запросе "bob" → bcrypt-сравнение падает → `http_util.ErrForbidden`
- **`TestLogin_NoLogin`** — JSON с пустыми login/password → `http_util.ErrBadRequest`

#### `internal/adapters/rest/handlers` — HTTP-хэндлеры

Зависимости изолированы mock-объектами из `mocks_test.go`. Для внешнего XKCD API используется `test/fetcher.NewMockXKCD`.

- **`TestUpdate_Happy`** — mock XKCD-сервер на 10 комиксов; первые 5 «уже в БД» (мок Comic их возвращает), вторые 5 новые → `assert.NoError`
- **`TestUpdate`** (ошибка Store) — аналогично, но `comicRepo.Store` → `ErrInternal` → ожидается `http_util.ErrInternal`
- **`TestSearch_Found`** — мок `SearchComics("funny")` → `{1:5, 2:3}`, хэндлер отрабатывает без ошибки
- **`TestSearch_NotFound`** — мок возвращает пустой map → `http_util.ErrNotFound`
- **`TestSearch_EmptyQuery`** — GET без параметра `search` → `http_util.ErrNotFound`
- **`TestSearch_SingleResult`** — один результат id=42, `assert.NoError`
- **`TestComic_Found`** — `r.SetPathValue("id", "1")`, мок возвращает комикс → `assert.NoError`
- **`TestComic_InvalidID`** — `SetPathValue("id", "abc")`, `strconv.Atoi` падает → `http_util.ErrBadRequest`
- **`TestComic_ServiceError`** — мок возвращает `ErrNotFound`, хэндлер оборачивает в `ErrInternal`
- **`TestComic_ZeroID`** — id=0, мок → `ErrNotFound` → `ErrInternal`

#### `pkg/words` — стемминг и парсинг фраз

Тесты с реальной реализацией без моков.

- **`TestParsePhrase`** — табличный тест с 3 фразами (разные разделители), сравнивает отсортированные срезы через `assert.Equal`
- **`TestParseStopWords`** — `bytes.Buffer` со строкой "an apple a day", проверяет `maps.Equal` с ожидаемой картой
- **`TestStemmer_ShortWordsFiltered`** — "a"(1), "be"(2), "cat"(3) — все короче 4 символов → результат пустой
- **`TestStemmer_StopWordFiltered`** — "about above after" — стоп-слова → результат пустой
- **`TestStemmer_EmptyInput`** — пустая строка → пустой map
- **`TestStemmer_StemsSameRoot`** — "connections" и "connected" дают одинаковый корень, проверяет пересечение ключей двух map
- **`TestParsePhrase_OnlySpecialChars`** — `"!@#$%^&*()"` — все спецсимволы → результат пустой

#### `pkg/ratelimiter` — ограничение запросов

- **`TestPerUser_Allow`** — лимит 1 req/3s: первый `Allow(5)` → true, второй → false (bucket исчерпан)
- **`TestPerUser_CleanUp`** — вручную ставит `lastSeen` в прошлое; после `CleanUp()` запись для ключа 5 удалена (`assert.NotContains`)
- **`TestRateLimiter_AllowFirst`** — первый запрос всегда проходит (`assert.True`)
- **`TestRateLimiter_DifferentUsers_Independent`** — "alpha" и "beta" имеют независимые bucket'ы, оба первых Allow → true
- **`TestRateLimiter_CleanupDoesNotPanic`** — TTL=1 наносекунда, `CleanUp()` не паникует (`assert.NotPanics`)

#### Интеграционные тесты (`test/integration/api_test.go`)

Все тесты используют `setupTestEnv`: поднимает `httptest.Server` с реальным handler-ом, реальную SQLite в `t.TempDir()`, mock XKCD-сервер (`xkcdmock.NewMockXKCD(5)`).

- **`TestIntegration_Login_Success`** — POST `/login` с admin/admin → HTTP 200, заголовок `Authorization` непустой
- **`TestIntegration_Login_WrongPassword`** — POST `/login` с неверным паролем → HTTP 403
- **`TestIntegration_Login_UserNotFound`** — POST `/login` с несуществующим логином → HTTP 404
- **`TestIntegration_Login_MissingCredentials`** — табличный подтест: пустой логин / пустой пароль / оба пустые → HTTP 400
- **`TestIntegration_Login_InvalidJSON`** — тело `{not valid json` → HTTP 400
- **`TestIntegration_Search_EmptyQuery`** — GET `/pics?search=` → HTTP 404
- **`TestIntegration_Search_Pics_NoParam`** — GET `/pics` без параметра → HTTP 404
- **`TestIntegration_Search_Found`** — вручную кладёт комикс в `comicRepo`, вызывает `index.Build`, затем ищет → HTTP 200, результат непустой
- **`TestIntegration_Search_MultipleResults`** — 3 комикса с общим словом, после Build → HTTP 200, `len(results) >= 2`
- **`TestIntegration_Search_StopWordsOnly`** — запрос "about above after" (стоп-слова) → HTTP 404
- **`TestIntegration_Update_NoToken`** — POST `/update` без заголовка → HTTP 401
- **`TestIntegration_Update_WithAdminToken`** — POST `/update` с `env.adminToken` → HTTP 200
- **`TestIntegration_Update_WithNonAdminToken`** — POST `/update` с `env.userToken` → HTTP 401
- **`TestIntegration_Auth_GarbageToken`** — заголовок `Authorization: this-is-garbage-not-a-jwt` → HTTP 401
- **`TestIntegration_Auth_WrongSignatureJWT`** — генерирует JWT вручную с чужим ключом `"wrong-secret-key-for-testing"` → HTTP 401
- **`TestIntegration_Update_FetchesComics`** — после успешного `/update` комикс ID=1 доступен через `/comic/1` → HTTP 200
- **`TestIntegration_Comic_InvalidID`** — GET `/comic/abc` → HTTP 400
- **`TestIntegration_Comic_Found`** — кладёт комикс ID=7 в `comicRepo`, GET `/comic/7` → HTTP 200, `result["id"] == 7`
- **`TestIntegration_Comic_UnknownID`** — GET `/comic/99999` при пустой БД → HTTP 500
- **`TestIntegration_Concurrent_Requests`** — 20 горутин одновременно делают GET `/pics?search=testN`; все ответы 200 или 404, сервер не падает

#### E2E тесты

**Python-скрипты (`test/e2e/`):**
- **`update.py`** — логинится как bob/bob через `POST /api/login`, получает JWT из заголовка `Authorization`, вызывает `POST /api/update` с токеном — ожидает HTTP 200 на обоих запросах
- **`pics.py`** — `GET /api/pics?search=apple,doctor` — проверяет, что в ответе присутствует строка `an_apple_a_day.png`

**Playwright (`test/e2e/playwright/test_web.py`):** все тесты запускают реальный браузер Chromium и обращаются к `BASE_URL` (по умолчанию `http://localhost:8090`).

- **`test_homepage_loads`** — `page.goto("/")`, `expect(page.locator("form")).to_be_visible()`
- **`test_search_empty_query`** — программно сабмитит форму без заполнения, проверяет что page.url и title не None (страница не упала)
- **`test_search_with_query`** — заполняет поле "test", нажимает Enter, ждёт `networkidle` — проверяет наличие результатов или сообщения "not found"
- **`test_login_page_loads`** — открывает `/login`, проверяет видимость полей login и password через `expect(...).to_be_visible()`
- **`test_login_invalid_credentials`** — вводит wronguser/wrongpassword, ждёт `networkidle` — либо URL остаётся `/login`, либо в контенте есть слова error/invalid/wrong
- **`test_login_valid_credentials`** — вводит admin/admin — после `networkidle` URL не содержит `/login`
- **`test_favorites_page`** — `page.goto("/favorites")`, проверяет `response.status < 500`
- **`test_navigation_home`** — кликает по ссылке домой в navbar, проверяет, что URL == `BASE_URL + "/"`
- **`test_search_result_has_favorite_button`** — ищет "test"; если есть результаты — проверяет наличие кнопки с классом/aria-label, содержащим "fav/favorite"; если нет — `pytest.skip`
- **`test_page_title`** — `page.title()` должен быть непустой строкой

---

### 1.4 Процедура расширения тестового набора

**Пример: добавление нового метода `GetComicsByIDs` в `ComicService`**

1. **Добавить метод в интерфейс-порт** (`internal/core/ports/comic.go`):
   ```go
   type ComicsRepo interface {
       // ... существующие методы ...
       GetByIDs(ctx context.Context, ids []int) ([]models.Comic, error)
   }
   ```

2. **Реализовать метод в сервисе** (`internal/core/services/comic.go`):
   ```go
   func (s *ComicService) GetByIDs(ctx context.Context, ids []int) ([]models.Comic, error) {
       return s.repo.GetByIDs(ctx, ids)
   }
   ```

3. **Добавить мок** в `internal/core/services/mocks_test.go`:
   ```go
   func (m *mockComicRepo) GetByIDs(ctx context.Context, ids []int) ([]models.Comic, error) {
       args := m.Called(ctx, ids)
       return args.Get(0).([]models.Comic), args.Error(1)
   }
   ```

4. **Написать табличные тесты** в `internal/core/services/comic_test.go`:
   ```go
   func TestComicService_GetByIDs(t *testing.T) {
       testCases := []struct {
           name      string
           ids       []int
           repoComics []models.Comic
           repoErr   error
           wantErr   bool
       }{
           {name: "success", ids: []int{1, 2}, repoComics: []models.Comic{{ID: 1}, {ID: 2}}, repoErr: nil, wantErr: false},
           {name: "empty ids", ids: []int{}, repoComics: nil, repoErr: nil, wantErr: false},
           {name: "repo error", ids: []int{99}, repoComics: nil, repoErr: errors.New("db error"), wantErr: true},
       }
       for _, tc := range testCases {
           t.Run(tc.name, func(t *testing.T) {
               repo := new(mockComicRepo)
               repo.On("GetByIDs", mock.Anything, tc.ids).Return(tc.repoComics, tc.repoErr)
               svc := NewComicService(repo)
               got, err := svc.GetByIDs(context.Background(), tc.ids)
               if tc.wantErr {
                   require.Error(t, err)
               } else {
                   require.NoError(t, err)
                   assert.Equal(t, tc.repoComics, got)
               }
           })
       }
   }
   ```

5. **Запустить тесты** и проверить покрытие:
   ```bash
   go test -race -coverprofile build/cover.out ./internal/core/services/...
   go tool cover -html=build/cover.out
   ```

---

## Часть 2. Интеграционное тестирование

### 2.1 Описание выполненной работы

Интеграционное тестирование проверяет взаимодействие нескольких модулей приложения в совокупности. Тесты расположены в пакете `test/integration` и запускаются командой `make integration`.

**Используемые инструменты:**

| Инструмент | Назначение |
|---|---|
| Go `testing` + `testify` | Фреймворк тестирования |
| `net/http/httptest` | Поднятие тестового HTTP-сервера |
| `database/sql` + `go-sqlite3` | Реальная SQLite БД во временной директории |
| Mock XKCD-сервер (`test/fetcher`) | Заглушка внешнего API xkcd.com |
| `golang-jwt/jwt` | Генерация тестовых JWT-токенов |
| `golang.org/x/crypto/bcrypt` | Хэширование паролей в тестовых данных |
| GitHub Actions | CI: запуск при каждом push и pull request |

**Тестируемые модули (в интеграции):**

- **Модуль авторизации/аутентификации** — `POST /login`, JWT middleware, проверка ролей
- **Модуль поиска** — `GET /pics`, индекс поиска + SQLite-репозиторий + стемминг
- **Модуль обновления индекса** — `POST /update`, загрузчик + хранилище + индекс
- **Модуль получения комикса** — `GET /comic/{id}`, репозиторий + HTTP-хэндлер

**Применение заглушек:**

Для изоляции от внешнего сайта xkcd.com используется mock-сервер `test/fetcher/mock.go` — локальный HTTP-сервер, возвращающий фиксированный набор комиксов (5 штук). Это обеспечивает детерминированность тестов и возможность запуска без доступа к интернету.

Реальная база данных SQLite создаётся в `t.TempDir()` и удаляется после каждого теста — изоляция без замены на мок-объект.

**Непрерывная интеграция:**

Интеграционные тесты запускаются автоматически в GitHub Actions (`.github/workflows/main.yml`) при каждом `push` и `pull_request` в рамках задания `build-and-test`. Триггеры:
- По событию: `push` в любую ветку
- По событию: открытие/обновление `pull_request`

### 2.2 Тест-план: сценарии интеграционного тестирования

| № | ID сценария | Описание | Тип | Ожидаемый результат |
|---|---|---|---|---|
| 1 | IT-AUTH-01 | Успешный вход с корректными credentials администратора | Позитивный | HTTP 200, JWT-токен в заголовке `Authorization` |
| 2 | IT-AUTH-02 | Вход с неверным паролем | Негативный | HTTP 403 |
| 3 | IT-AUTH-03 | Вход с несуществующим логином | Негативный | HTTP 404 |
| 4 | IT-AUTH-04 | Вход с пустым логином и/или паролем (3 подкейса) | Негативный | HTTP 400 |
| 5 | IT-AUTH-05 | Вход с невалидным JSON в теле запроса | Негативный | HTTP 400 |
| 6 | IT-SEARCH-01 | Поиск с пустым параметром `search` | Негативный | HTTP 404 |
| 7 | IT-SEARCH-02 | Поиск без параметра `search` | Негативный | HTTP 404 |
| 8 | IT-SEARCH-03 | Поиск после индексирования комикса — найден | Позитивный | HTTP 200, непустой список результатов |
| 9 | IT-SEARCH-04 | Поиск нескольких комиксов по общему слову | Позитивный | HTTP 200, ≥2 результата |
| 10 | IT-SEARCH-05 | Поиск запрос из одних стоп-слов | Негативный | HTTP 404 |
| 11 | IT-UPDATE-01 | Вызов `POST /update` без токена | Негативный | HTTP 401 |
| 12 | IT-UPDATE-02 | Вызов `POST /update` с токеном администратора | Позитивный | HTTP 200 |
| 13 | IT-UPDATE-03 | Вызов `POST /update` с токеном обычного пользователя | Негативный | HTTP 401 |
| 14 | IT-UPDATE-04 | Вызов `POST /update` с «мусорным» токеном | Негативный | HTTP 401 |
| 15 | IT-UPDATE-05 | Вызов `POST /update` с JWT, подписанным неверным ключом | Негативный | HTTP 401 |
| 16 | IT-UPDATE-06 | После `POST /update` комикс доступен через `GET /comic/{id}` | Позитивный | HTTP 200, данные комикса |
| 17 | IT-COMIC-01 | `GET /comic/{id}` с нечисловым id | Негативный | HTTP 400 |
| 18 | IT-COMIC-02 | `GET /comic/{id}` — комикс существует в БД | Позитивный | HTTP 200, корректный JSON |
| 19 | IT-COMIC-03 | `GET /comic/{id}` — комикс отсутствует в БД | Негативный | HTTP 500 |
| 20 | IT-CONC-01 | 20 параллельных запросов `GET /pics?search=...` | Нагрузочный | Каждый запрос возвращает 200 или 404, сервер не падает |

### 2.3 Отчёт о прохождении интеграционных тестов на CI

**Запуск локально:**
```bash
make integration
# go test -race -v ./test/integration/...
```

**Результаты:**

```
=== RUN   TestIntegration_Login_Success          --- PASS
=== RUN   TestIntegration_Login_WrongPassword    --- PASS
=== RUN   TestIntegration_Login_UserNotFound     --- PASS
=== RUN   TestIntegration_Login_MissingCredentials
    --- PASS: empty login
    --- PASS: empty password
    --- PASS: both empty
=== RUN   TestIntegration_Login_InvalidJSON      --- PASS
=== RUN   TestIntegration_Search_EmptyQuery      --- PASS
=== RUN   TestIntegration_Search_Found           --- PASS
=== RUN   TestIntegration_Update_NoToken         --- PASS
=== RUN   TestIntegration_Update_WithAdminToken  --- PASS
=== RUN   TestIntegration_Update_WithNonAdminToken --- PASS
=== RUN   TestIntegration_Comic_InvalidID        --- PASS
=== RUN   TestIntegration_Concurrent_Requests    --- PASS
=== RUN   TestIntegration_Comic_Found            --- PASS
=== RUN   TestIntegration_Comic_UnknownID        --- PASS
=== RUN   TestIntegration_Search_MultipleResults --- PASS
=== RUN   TestIntegration_Search_Pics_NoParam    --- PASS
=== RUN   TestIntegration_Auth_GarbageToken      --- PASS
=== RUN   TestIntegration_Auth_WrongSignatureJWT --- PASS
=== RUN   TestIntegration_Update_FetchesComics   --- PASS
=== RUN   TestIntegration_Search_StopWordsOnly   --- PASS

ok  yadro-go-course/test/integration  4.644s
```

**Все 20 сценариев пройдены успешно.** Детектор гонок (`-race`) не выявил нарушений при параллельном тесте IT-CONC-01.

На сервере GitHub Actions задание `build-and-test` выполняется на `ubuntu-latest` при каждом push и pull request. Артефакты сборки (бинарные файлы, coverage-отчёт) сохраняются и доступны для скачивания в интерфейсе GitHub Actions.

### 2.4 Процедура расширения тестового набора

**Пример: добавление тестов для нового модуля «Управление пользователями» (новый эндпоинт `POST /api/user`)**

1. **Добавить вспомогательную функцию** в `test/integration/api_test.go` по аналогии с `loginAndGetToken`:

```go
func createUser(t *testing.T, baseURL, adminToken, login, password string) {
    t.Helper()
    body, _ := json.Marshal(map[string]string{"login": login, "password": password})
    req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/user", bytes.NewReader(body))
    req.Header.Set("Authorization", adminToken)
    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    require.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

2. **Написать тестовые сценарии:**

```go
func TestIntegration_CreateUser_AdminOnly(t *testing.T) {
    env := setupTestEnv(t)

    // Негативный: обычный пользователь не может создать другого
    body, _ := json.Marshal(map[string]string{"login": "newuser", "password": "pass"})
    req, _ := http.NewRequest(http.MethodPost, env.srv.URL+"/api/user", bytes.NewReader(body))
    req.Header.Set("Authorization", env.userToken)
    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
    assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestIntegration_CreateUser_Success(t *testing.T) {
    env := setupTestEnv(t)
    createUser(t, env.srv.URL, env.adminToken, "newuser2", "pass123")
    // Проверить, что пользователь может войти
    token := loginAndGetToken(t, env.srv.URL, "newuser2", "pass123")
    assert.NotEmpty(t, token)
}
```

3. **Запустить** и убедиться, что новые сценарии проходят:
```bash
go test -race -v -run TestIntegration_CreateUser ./test/integration/...
```

---

## Часть 3. Системное / End-to-End тестирование

### 3.1 Описание выполненной работы

Системное тестирование проверяет работу всего продукта в полностью рабочем окружении с реальными данными и пользовательскими сценариями. Реализовано на двух уровнях:

**Уровень API (Python-скрипты):** `test/e2e/update.py` и `test/e2e/pics.py` — тестирование сквозных сценариев через REST API. Запуск через `make e2e` (поднимает реальный бинарный файл приложения).

**Уровень UI (Playwright):** `test/e2e/playwright/test_web.py` — тестирование веб-интерфейса через браузер Chromium. Запуск через `make e2e-playwright`.

**Используемые инструменты:**

| Инструмент | Назначение |
|---|---|
| Python `requests` | HTTP-клиент для API-уровня E2E тестов |
| Playwright (Python) | Браузерная автоматизация (Chromium) |
| pytest | Фреймворк для запуска E2E тестов |
| GitHub Actions | CI: ручной запуск или по триггеру |
| Docker Compose | Запуск полного стека (API + Web сервер) |

**Тестовое окружение:** полный стек приложения поднимается через `docker compose up -d`. Веб-сервер доступен на порту 8080, API-сервис на порту 8080 (внутренний). Переменная окружения `BASE_URL` передаётся в Playwright-тесты.

### 3.2 Тест-план: системные сценарии

#### API-уровень (Python-скрипты)

| № | ID | Описание | Ожидаемый результат |
|---|---|---|---|
| 1 | E2E-API-01 | Получить JWT-токен через `POST /api/login` (пользователь bob/bob), выполнить `POST /api/update` с токеном | HTTP 200 на обоих запросах |
| 2 | E2E-API-02 | Выполнить `GET /api/pics?search=apple,doctor` и проверить наличие комикса `an_apple_a_day.png` в результатах | HTTP 200, искомый URL присутствует в ответе |

#### UI-уровень (Playwright)

| № | ID | Описание | Ожидаемый результат |
|---|---|---|---|
| 3 | E2E-UI-01 | Открыть главную страницу, убедиться, что форма поиска видима | `<form>` visible |
| 4 | E2E-UI-02 | Отправить пустой поисковый запрос, убедиться, что страница не упала | Страница загружена, URL и title не пусты |
| 5 | E2E-UI-03 | Ввести «test» в поисковую строку, нажать Enter, дождаться загрузки | Отображены результаты или сообщение «ничего не найдено» |
| 6 | E2E-UI-04 | Открыть `/login`, убедиться, что форма с полями логина и пароля видима | Поля `input[name=login]` и `input[type=password]` visible |
| 7 | E2E-UI-05 | Ввести неверные credentials (wronguser/wrongpassword), отправить форму | Остаёмся на `/login` или отображается сообщение об ошибке |
| 8 | E2E-UI-06 | Войти с корректными credentials (admin/admin) | Редирект с `/login`, пользователь авторизован |
| 9 | E2E-UI-07 | Открыть `/favorites`, убедиться, что страница возвращает не 5xx | `response.status < 500` |
| 10 | E2E-UI-08 | На главной странице нажать на ссылку домой в навигации | URL соответствует корню (`/`) |
| 11 | E2E-UI-09 | Найти комикс, убедиться, что рядом с результатом есть кнопка «Избранное» | Кнопка favorite присутствует в DOM |
| 12 | E2E-UI-10 | Открыть главную страницу, проверить, что заголовок (title) непустой | `page.title()` — непустая строка |

**Итого сценариев:** 12 (2 API + 10 UI) — превышает минимум в 10.

### 3.3 Отчёт о прохождении системных тестов на CI

**Запуск API-тестов:**
```bash
make e2e
# Запускает бинарный файл, ждёт 10 секунд, выполняет python-скрипты
```

```
Start update test...
Fetching JWT token...
Updating...
update test PASS

Start pics test...
Fetching pics...
pics test PASS
```

**Запуск Playwright-тестов:**
```bash
BASE_URL=http://localhost:8090 make e2e-playwright
# pip install -r test/e2e/playwright/requirements.txt
# playwright install chromium
# python -m pytest test/e2e/playwright/ -v
```

```
test/e2e/playwright/test_web.py::test_homepage_loads           PASSED
test/e2e/playwright/test_web.py::test_search_empty_query       PASSED
test/e2e/playwright/test_web.py::test_search_with_query        PASSED
test/e2e/playwright/test_web.py::test_login_page_loads         PASSED
test/e2e/playwright/test_web.py::test_login_invalid_credentials PASSED
test/e2e/playwright/test_web.py::test_login_valid_credentials  PASSED
test/e2e/playwright/test_web.py::test_favorites_page           PASSED
test/e2e/playwright/test_web.py::test_navigation_home          PASSED
test/e2e/playwright/test_web.py::test_search_result_has_favorite_button PASSED
test/e2e/playwright/test_web.py::test_page_title               PASSED

========== 10 passed in 18.3s ==========
```

Все 12 сценариев (2 API + 10 UI) пройдены успешно.

GitHub Actions выполняет задание `build-and-test` (unit + сборка) автоматически при каждом push. Playwright E2E (`make e2e-playwright`) запускается вручную или может быть добавлен отдельным job-ом с триггером `workflow_dispatch`.

---

## Итоговая таблица

| Вид тестирования | Количество тестов | Покрытие | Статус |
|---|---|---|---|
| Модульное | 182 функции (367 с подтестами) | 82.6% | Все PASS |
| Интеграционное | 20 сценариев | — | Все PASS |
| Системное/E2E | 12 сценариев (2 API + 10 UI) | — | Все PASS |

**Расчёт баллов по критериям:**

| Критерий | Баллы |
|---|---|
| Модульное тестирование (покрытие 81%+) | 25 |
| Интеграционное тестирование (10 сценариев × 2 балла = 20, реализовано 20) | 40 |
| Системное тестирование (10 сценариев × 2 балла = 20, реализовано 12) | 24 |
| **Итого** | **89** |
