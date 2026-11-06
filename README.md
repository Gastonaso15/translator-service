# Translator Service

High-performance Go microservice for local text and localization (`locale`) file translation powered by **Ollama** (configured by default with `qwen3.5:4b`).

Features structured JSON output, graceful shutdown, and support for multiple languages while strictly preserving identifiers (`id`), technical tags, and interpolation placeholders (such as `{0}`, `{{name}}`, `%s`).

---

## Project Structure

The project adheres to the standard Go layout:

```text
├── cmd/
│   └── translator-service/
│       └── main.go           # Application entrypoint & graceful shutdown lifecycle
├── pkg/
│   ├── config/               # Environment configuration loader & validation
│   ├── handler/              # HTTP routing & REST API handlers
│   ├── llm/                  # Decoupled LLM client & Ollama API integration
│   └── models/               # Request/Response DTOs & domain models
├── bin/                      # Compiled binary outputs (ignored in git)
├── .golangci.yml             # Linter configuration
├── Makefile                  # Automated build and testing targets
└── README.md                 # Project documentation
```

---

## Configuration

Configuration is loaded from environment variables with sensible defaults:

| Environment Variable | Default Value | Description |
|---|---|---|
| `PORT` | `8080` | Port on which the HTTP server listens. |
| `OLLAMA_URL` | `http://localhost:11434` | Base URL of the Ollama service. |
| `LLM_MODEL` | `qwen3.5:4b` | Model identifier to use for translation. |
| `HTTP_TIMEOUT` | `180s` | HTTP client timeout for LLM inference requests. |

---

## Developer Commands (`Makefile`)

The project includes standard build automation via `make`:

```bash
# Run unit tests with the race condition detector enabled
make test

# Generate code coverage report
make coverage

# Compile the executable binary into bin/translator-service
make build

# Run service directly from source
make run

# Clean build artifacts and coverage files
make clean
```

---

## API Endpoints

### 1. Health Check
Checks the service operational status and reports the active LLM model.

- **Method**: `GET`
- **Route**: `/health`
- **Response**:
```json
{
  "status": "ok",
  "model": "qwen3.5:4b"
}
```

**Example request:**
```bash
curl -X GET http://localhost:8080/health
```

---

### 2. Single Message Translation
Translates an individual UI text message or string between two languages.

- **Method**: `POST`
- **Route**: `/api/v1/translate/message`
- **Supported Languages**: Standard ISO language codes (e.g., `es`, `en`, `ja`, `pt`, `fr`, `de`, etc.).

#### Request Payload
```json
{
  "source": "en",
  "target": "es",
  "text": "Welcome back! You have 3 unread notifications."
}
```

#### Response
```json
{
  "translated_text": "¡Bienvenido de nuevo! Tienes 3 notificaciones no leídas."
}
```

#### Usage Examples with `curl`:

**English to Spanish (`es`):**
```bash
curl -X POST http://localhost:8080/api/v1/translate/message \
  -H "Content-Type: application/json" \
  -d '{
    "source": "en",
    "target": "es",
    "text": "Welcome back! You have 3 unread notifications."
  }'
```

**Spanish to Japanese (`ja`):**
```bash
curl -X POST http://localhost:8080/api/v1/translate/message \
  -H "Content-Type: application/json" \
  -d '{
    "source": "es",
    "target": "ja",
    "text": "¿Qué tal tu día?"
  }'
```

**Spanish to Portuguese (`pt`):**
```bash
curl -X POST http://localhost:8080/api/v1/translate/message \
  -H "Content-Type: application/json" \
  -d '{
    "source": "es",
    "target": "pt",
    "text": "¿Qué tal tu día?"
  }'
```

---

### 3. Locale File Translation
Translates localization files in a single batch while strictly preserving entry identifiers (`id`) and technical placeholders like `{0}`, `{{name}}`, `%s`, or `<tag>`.

- **Method**: `POST`
- **Route**: `/api/v1/translate/localefile`

#### Request Payload
```json
{
  "source": "en",
  "target": "es",
  "texts": [
    {"id": "btn.save", "text": "Save changes"},
    {"id": "btn.cancel", "text": "Cancel"},
    {"id": "msg.greeting", "text": "Hello, {0}!"}
  ]
}
```

#### Response
```json
{
  "texts": [
    {"id": "btn.save", "text": "Guardar cambios"},
    {"id": "btn.cancel", "text": "Cancelar"},
    {"id": "msg.greeting", "text": "¡Hola, {0}!"}
  ]
}
```

#### Usage Examples with `curl`:

**English to Spanish (`es`):**
```bash
curl -X POST http://localhost:8080/api/v1/translate/localefile \
  -H "Content-Type: application/json" \
  -d '{
    "source": "en",
    "target": "es",
    "texts": [
      {"id": "btn.save", "text": "Save changes"},
      {"id": "btn.cancel", "text": "Cancel"},
      {"id": "msg.greeting", "text": "Hello, {0}!"}
    ]
  }'
```

**Spanish to Japanese (`ja`):**
```bash
curl -X POST http://localhost:8080/api/v1/translate/localefile \
  -H "Content-Type: application/json" \
  -d '{
    "source": "es",
    "target": "ja",
    "texts": [
      {"id": "msg.greeting", "text": "¡Hola, {0}!"},
      {"id": "btn.save", "text": "Guardar cambios"},
      {"id": "btn.cancel", "text": "Cancelar"}
    ]
  }'
```

**Spanish to Portuguese (`pt`):**
```bash
curl -X POST http://localhost:8080/api/v1/translate/localefile \
  -H "Content-Type: application/json" \
  -d '{
    "source": "es",
    "target": "pt",
    "texts": [
      {"id": "msg.greeting", "text": "¡Hola, {0}!"},
      {"id": "btn.save", "text": "Guardar cambios"},
      {"id": "btn.cancel", "text": "Cancelar"}
    ]
  }'
```
