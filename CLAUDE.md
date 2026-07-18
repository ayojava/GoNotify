# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

GoNotify reads a session schedule from a Google Sheet and sends WhatsApp reminders (via Twilio) to
whoever is leading each upcoming session. It's a single-shot CLI (`go run .`), not a long-running
service — it's meant to be invoked once per day by a scheduled job.

## Commands

```bash
go build ./...          # build everything
go test ./...           # run all tests
go test -run TestName ./...       # run a single test by name (works from repo root; package-qualify if ambiguous)
go test -v ./internal/config/...  # run just the config package tests, verbose
go run .                # run the reminder script locally (needs secrets/application.yaml, see below)
```

There is no lint step configured in CI beyond `go build`/`go test`.

## Local setup

`go run .` loads config from `secrets/application.yaml` (path is `DefaultConfigPath` in `main.go`).
That whole `secrets/` directory is gitignored, so it won't exist on a fresh clone — create it with:

```yaml
google:
  application_credentials: secrets/service_account.json
  sheet_id: <sheet id>
  sheet_range: notification!B2:E
twilio:
  account_sid: <sid>
  auth_token: <token>
  whatsapp_from: "whatsapp:+14155238886"
```

plus a `secrets/service_account.json` Google service account key with read access to the sheet.
In CI these same values come from GitHub Actions secrets as env vars instead (see below) — the
config file is optional and only an error if present-but-malformed.

## Architecture

The whole `main` package is small and split by responsibility, wired together in `main.go`. Each
piece is an interface with one real implementation and one fake used in tests (`mocks_test.go`) —
there's no DI framework, just constructor injection:

- **`ScheduleRepository`** (`schedule_repository.go`) — loads `[]Session` from a Google Sheet.
  `GoogleSheetsRepository` authenticates via a JWT service-account client and expects each row as
  `[name, phone, date, time]` with `date` in `time.DateOnly` (`YYYY-MM-DD`) format and `time` a
  free-form string (e.g. `18:00`). Row 0 is treated as a header and skipped; malformed rows (fewer
  than 4 columns, bad date) are logged and skipped rather than failing the whole load. The sheet
  range must cover all 4 columns (see `sheet_range` above) or every data row gets skipped as
  malformed.
- **`Session`** (`session.go`) — the domain record (name, phone, date, time). `DaysUntil(now)`
  truncates both sides to midnight before diffing, so "same calendar day" reliably yields `0`
  regardless of time-of-day.
- **`MessageBuilder`** (`message_builder.go`) — turns a `Session` + `daysUntil` into the WhatsApp
  message text. `ReminderMessageBuilder` has distinct copy for "today" (`daysUntil == 0`, which also
  includes `Session.Time`) vs. future days.
- **`Notifier`** (`notifier.go`) — sends a message to a phone number. `TwilioNotifier` wraps the
  official `twilio-go` SDK client.
- **`ReminderService`** (`reminder_service.go`) — the orchestrator. For each session it computes
  `DaysUntil`, checks it against a configured `ReminderDays` list (currently `[]int{1, 0}` — day-
  before and day-of), builds the message, and sends it. A failed send for one recipient is logged
  and does not abort the run for the rest; `Run` returns the count of successful sends.
- **`internal/config`** — Viper-based config loading (`Load(configPath)`). Reads YAML if present,
  always overlays env vars (`.` in keys maps to `_`, e.g. `google.sheet_id` → `GOOGLE_SHEET_ID`),
  then validates that every required field ended up populated regardless of source. This is what
  lets the same binary run locally from a YAML file and in CI from secrets-as-env-vars with no code
  branching. Two Viper gotchas this code works around, both only visible when the file is missing
  (the CI case): (1) `viper.SetConfigFile` with an explicit path makes a missing file return a raw
  `*fs.PathError` instead of `viper.ConfigFileNotFoundError`, so `Load` checks `os.IsNotExist` too;
  (2) `viper.AutomaticEnv` only overrides keys Viper already knows about, normally learned from the
  file — with no file, `Unmarshal` silently comes back empty unless every key is bound explicitly
  via `viper.BindEnv`, which `Load` does for all six required fields.

Tests are plain-Go (`_test.go` files colocated with the code, no testify/mock framework) and use the
hand-written fakes in `mocks_test.go` against the real interfaces — new tests for `ReminderService`
should extend those fakes rather than introducing a mocking library.

## CI

GitHub Actions workflows does the following (build/test, materialize the service account
JSON from a base64 secret, run the script, clean up the credentials file) but on different triggers:

- `.github/workflows/go.yml` — on push/PR to `master`, includes the `go build`/`go test` steps.
  this is the actual production trigger for sending reminders, and skips the build/test steps.

Both expect these repo secrets: `GOOGLE_SERVICE_ACCOUNT_JSON_BASE64`, `GOOGLE_SHEET_ID`,
`GOOGLE_SHEET_RANGE`, `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_WHATSAPP_FROM`.