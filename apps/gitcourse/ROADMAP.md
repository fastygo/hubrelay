# GitCourse: публичная дорожная карта

Открытая обучающая платформа на базе Git. Один пользователь (admin) добавляет публичные репозитории как совместимые курсы, проходит их в своём IDE, отслеживает прогресс в кабинете и общается с AI-ассистентом через HubRelay.

* * *

## Модель

### Компоненты

*   **Git-репозиторий курса** — публичный репозиторий с `course.json`, `verify.sh`, стартовым кодом и CI workflow
*   **Git-репозиторий ученика** — создаётся из шаблона (Use this template), CI проверяет шаги и обновляет `progress.json`
*   **GitCourse app** — Go-приложение (SSR, копия архитектуры `apps/dashboard`): витрина курсов, кабинет ученика, ASK AI
*   **HubRelay** — связующее звено между GitCourse и AI-провайдером; чистый проводник команд
*   **Qdrant** — пользовательская векторная БД; GitCourse индексирует в неё содержимое курсов для RAG при общении с ASK AI

### Принципы

*   Курс полностью живёт в Git-репозитории; платформа — витрина и верификатор, не runtime
*   Прогресс фиксируется CI/CD в репозитории ученика через `.course/progress.json`
*   Cursor-агент работает с файлами `.course/` напрямую; MCP-сервер не нужен
*   ASK AI: GitCourse делает RAG-поиск в Qdrant, формирует контекст и отправляет через HubRelay SDK к AI-провайдеру
*   HubRelay не знает про Qdrant и курсы — он чистый relay/proxy для команд
*   Один пользователь `admin`; масштабирование заложено в структурах, но не реализовано
*   Git-провайдер абстрагирован через интерфейс; MVP работает с любым публичным репозиторием по raw HTTP

### Архитектура ASK AI

```
Ученик задаёт вопрос в кабинете
  │
  ▼
GitCourse app
  ├─ rag/ → Qdrant (пользовательский)
  │   └─ similarity search по вопросу + контексту урока
  ├─ формирует обогащённый промпт (вопрос + RAG-контекст)
  └─ relay/ → HubRelay SDK → Execute("ask", {prompt: ...})
       └─ HubRelay → AI provider → SSE stream → ответ
```

* * *

## Этап 1: Контракты и первый курс

Цель: определить формат курса и валидировать цикл `course.json -> verify.sh -> CI -> progress.json`.

Стартовый шаблон: `apps/gitcourse/example-vite` (React 19 + Vite 6 + Tailwind 4 + TypeScript).

### Структура репозитория-шаблона

```
example-vite/
  .course/
    course.json
    progress.json
    ci/
      verify.sh
  .cursor/
    rules/
      course.mdc
  .github/
    workflows/
      course-check.yml
  src/
  package.json
  vite.config.ts
  tsconfig.json
```

### Контракт course.json

```json
{
  "id": "vite-react-starter",
  "version": "1.0.0",
  "title": "Vite + React: первое приложение",
  "language": "typescript",
  "sections": [
    {
      "id": "basics",
      "title": "Основы",
      "lessons": [
        {
          "id": "001",
          "title": "Запуск проекта",
          "objective": "Установить зависимости и запустить dev-сервер",
          "checklist": [
            { "id": "node_modules", "label": "npm install завершён", "verify": "dir_exists:node_modules" },
            { "id": "build_ok", "label": "vite build проходит", "verify": "build_succeeds:npm run build" }
          ],
          "ask_context": "Ученик начинает Vite+React проект, устанавливает зависимости"
        }
      ]
    }
  ]
}
```

### Контракт verify.sh

Вход: ничего. Выход: JSON с результатами проверок. CI и платформа используют один формат.

### Контракт .cursor/rules/course.mdc

Инструкция для Cursor-агента: читай `.course/course.json` для контекста урока, `.course/progress.json` для статуса.

### Уроки первого курса (5-7 уроков)

1.  Клонирование шаблона, `npm install`, `npm run dev`
2.  Создание компонента Header с навигацией
3.  Добавление react-router и базовых страниц
4.  Страница каталога с mock-данными
5.  Подключение API (fetch из JSON-заглушки)
6.  Стилизация через Tailwind
7.  Сборка: `npm run build` + `npm run preview`

### Результат этапа

*   Готовый template-репозиторий первого курса
*   CI workflow, который обновляет `progress.json`
*   Cursor rules для AI-агента
*   Валидированный формат `course.json` / `verify.sh`

* * *

## Этап 2: GitCourse app (MVP)

Цель: Go-приложение в стиле `apps/dashboard` — витрина курсов и кабинет ученика.

### Каркас приложения

Копия архитектуры dashboard: `cmd/server`, `internal/config`, `internal/handlers`, `internal/modules`, `internal/presenter`, `internal/source`, `fixtures/`, `views/`, `static/`.

Модули:
*   `course` — витрина курсов + кабинет ученика + прогресс
*   `ask` — ASK AI через HubRelay SDK

### Реализация MVP

*   `internal/store` — `CourseStore` c JSON-backed реестром (`data/courses.json`) для публичных курсов, snapshot hash и привязок student repo
*   `internal/git` — raw HTTP reader для `course.json`, `progress.json`, `verify.sh` и CI workflow
*   `internal/progress` — `sync.Map` + TTL cache для публичного `progress.json`
*   `internal/source` — fixture/live слой: fixture для разработки без HubRelay, live для real Git + SDK
*   `views/catalog.templ`, `views/course.templ`, `views/lesson.templ`, `views/ask.templ` — SSR страницы GitCourse
*   `example-git/` — publishable шаблон первого курса, который потом выносится в отдельный публичный репозиторий

### Source interface

```go
type Source interface {
    Courses(ctx context.Context) ([]Course, error)
    Course(ctx context.Context, id string) (Course, error)
    Ask(ctx context.Context, prompt, model string) (CommandResult, error)
    AskStream(ctx context.Context, prompt, model string) (AskStream, error)
}
```

Fixture-реализация для офлайн-разработки. Live-реализация через `relay.Client` (SDK HubRelay).

### Конфиг

*   `APP_BIND`, `APP_ADMIN_USER`, `APP_ADMIN_PASS`, `APP_AUTH_DISABLED`
*   `APP_DATA_SOURCE=fixture|live`
*   `APP_DATA_DIR` — путь для `courses.json`
*   `APP_WEBHOOK_TOKEN` — shared secret для webhook прогресса
*   `HUBRELAY_TRANSPORT`, `HUBRELAY_BASE_URL` — подключение к HubRelay
*   `QDRANT_URL` — опциональная векторная БД для RAG

### Маршруты

*   `GET /` — каталог курсов (список из fixtures или добавленных репозиториев)
*   `GET /course/:id` — страница курса с программой и прогрессом
*   `GET /course/:id/lesson/:lessonId` — урок с чеклистом
*   `POST /courses/add` — admin добавляет URL публичного репозитория
*   `GET /ask` — ASK AI (SSE stream через HubRelay)
*   `POST /api/webhook/progress` — webhook от CI ученика

### Валидация курса при добавлении

При добавлении URL репозитория платформа:
1.  Читает `course.json` по raw URL
2.  Проверяет соответствие контракту (обязательные поля, структура)
3.  Проверяет наличие `verify.sh` и CI workflow
4.  Сохраняет hash оригинальных файлов
5.  Добавляет курс в каталог

### Прогресс

*   Admin указывает URL своего fork/clone репозитория
*   Платформа читает `progress.json` по raw URL (публичный репо)
*   Опционально: webhook от CI при каждом push
*   Кэш в памяти (`sync.Map`, TTL 5 мин)

### Результат этапа

*   Работающее Go-приложение с SSR-витриной
*   Добавление курсов через URL
*   Отображение прогресса из публичного репозитория
*   ASK AI через HubRelay в fixture и live режиме
*   Темы light/dark, локализация en/ru
*   Deployment path через `.paas/extensions/deploy-gitcourse.yml` и `.paas/deploy-gitcourse-clean.sh`

* * *

## Этап 3: Qdrant и RAG для ASK AI

Цель: ASK AI отвечает с учётом содержимого курсов ученика.

### Интеграция Qdrant

*   Конфиг: `QDRANT_URL` (Docker или облако — решает пользователь)
*   При добавлении курса: индексация `course.json` (уроки, описания, чеклисты, подсказки) в Qdrant
*   При вопросе в ASK AI: similarity search → формирование контекста → отправка через HubRelay

### RAG-пайплайн

```
Вопрос ученика
  ├─ rag.Search(question, courseID) → релевантные фрагменты из Qdrant
  ├─ buildPrompt(question, ragContext, lessonContext)
  └─ relay.AskStream(prompt, model) → HubRelay → AI → SSE
```

### Индексация

*   `internal/rag/indexer.go` — при добавлении курса разбивает `course.json` на чанки и индексирует
*   `internal/rag/searcher.go` — similarity search по вопросу + `course_id` как фильтр
*   Переиндексация при обновлении курса (hash changed)

### Результат этапа

*   ASK AI отвечает с контекстом из курсов
*   Qdrant подключается опционально; без него ASK AI работает без RAG
*   Индексация автоматическая при добавлении курса

* * *

## Этап 4: Production-readiness публичной версии

### Приёмочные проверки

1.  Витрина: курсы отображаются, фильтры работают
2.  Прогресс: `progress.json` читается и отображается корректно
3.  ASK AI: стриминг работает в fixture и live режиме
4.  RAG: ответы релевантны контексту курса
5.  Валидация: невалидный `course.json` отклоняется при добавлении

### Production-чеклист

*   SEO: canonical URL, meta-теги для страниц курсов
*   Безопасность: rate-limit на ASK AI, CSRF для форм, заголовки
*   Наблюдаемость: логирование, базовые метрики
*   CI/CD: Dockerfile, smoke-тесты
*   Документация: README, инструкция по добавлению курса

### Результат этапа

*   Публичная версия GitCourse готова к деплою
*   Один пользователь admin, публичные репозитории, ASK AI с RAG
*   Формат курсов стабилен и задокументирован

* * *

## Граница публичного и приватного

После завершения этапа 4 публичная часть стабильна. Всё, что дальше, разрабатывается в приватном репозитории:

*   Мульти-пользовательская регистрация и авторизация
*   Кабинет наставника и продажа курсов
*   Платёжная система и комиссия
*   Верификация и сертификаты
*   Приватные репозитории и платные курсы
*   Абстракция Git-провайдера (GitLab, Gitea)
*   Масштабирование инфраструктуры

## После публикации example-git

После выноса `apps/gitcourse/example-git/` в отдельный публичный репозиторий:

*   удалить `apps/gitcourse/example-git/` из монорепозитория
*   удалить `apps/gitcourse/example-vite/`, так как он нужен только как ранний локальный стартовый шаблон
*   оставить в репозитории только GitCourse app, дорожную карту и ссылки на внешний template repo
