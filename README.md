# LinkSnap

A URL shortener written in Go. Paste a long URL, get a short one back.
Every click on a short link is recorded with IP and user agent so the
data is there for analytics later.

The backend is a small HTTP API over Postgres. The frontend is a single
HTML page embedded into the binary at build time — no JS build step,
no framework.

## Running it

Requires Go 1.22+ and Docker.

```sh
cp .env.example .env
docker compose up -d
go run ./cmd/server
```

Then open http://localhost:8080.

Migrations run automatically on startup. Config is read from `.env`.

## API

```sh
# create a short link
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev"}'
# -> 201 {"code":"AbCdEfG","short_url":"http://localhost:8080/AbCdEfG"}

# follow it (302 to the original; records a click)
curl -i http://localhost:8080/AbCdEfG

# delete it
curl -X DELETE http://localhost:8080/api/AbCdEfG

# health
curl http://localhost:8080/health
```

## Layout

```
cmd/server/             entry point + migration runner
internal/config/        .env / env-var loader
internal/handler/       HTTP handlers
internal/repository/    pgxpool-backed Postgres repository
migrations/             golang-migrate SQL files
web/                    embedded frontend (go:embed index.html)
```

## Notes

- **7-character codes from a 62-symbol alphabet** — ~3.5B possibilities,
  far more than this will ever need. On the rare unique-index violation
  the insert retries up to 5 times before giving up.

- **Clicks are recorded fire-and-forget** — a goroutine with a 2s
  context. The redirect doesn't wait for the insert, so a slow DB can't
  slow user navigation. Acceptable for analytics; would not be for
  anything billable.

- **No service layer.** Handlers call the repository directly. There's
  no business logic between them worth abstracting yet. If validation
  rules grow beyond URL parsing, a service layer will earn its place.

- **Frontend embedded into the binary.** No separate static-asset
  serving, no CDN config — deploys are a single binary.

## What's next

Real-time click analytics over WebSocket. Clients will subscribe to a
short code and receive a push each time it's clicked.

## Stack

Go, Postgres, pgx/v5, golang-migrate, godotenv. Vanilla HTML / CSS / JS
on the frontend.
