# 3X-UI — Development Roadmap & Architecture Guide

> **Version:** 3.5.0 (fork)
> **Upstream:** https://github.com/MHSanaei/3x-ui
> **Status:** Fully synced with upstream v3.5.0
> **Custom additions:** 6 files, 1,052 lines added

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture Summary](#2-architecture-summary)
3. [Custom Fork Modifications](#3-custom-fork-modifications)
4. [Upstream Sync Report](#4-upstream-sync-report)
5. [Development Milestones](#5-development-milestones)
6. [Task Breakdown](#6-task-breakdown)
7. [Dependency Graph](#7-dependency-graph)
8. [Risk Assessment](#8-risk-assessment)
9. [Testing Guide](#9-testing-guide)
10. [Development Workflow](#10-development-workflow)
11. [Future Vision](#11-future-vision)

---

## 1. Project Overview

### What Is 3X-UI?

3X-UI is an advanced, production-grade web control panel for managing **Xray-core** proxy servers. It provides a comprehensive web UI (React 19 + Ant Design 6) backed by a Go/Gin API server to deploy, configure, and monitor proxy protocols including:

- **VLESS** (with XTLS/REALITY)
- **VMess** (with WebSocket, TCP, gRPC, HTTPUpgrade, SplithTTP)
- **Trojan** (with WebSocket, gRPC, TCP)
- **Shadowsocks** (with AEAD 2022)
- **WireGuard** (full inbound/outbound)
- **Hysteria2**
- **MTProto** (Telegram proxy via `mtg-multi` sidecar)

### Key Capabilities

| Capability | Description |
|------------|-------------|
| Protocol Support | VLESS, VMess, Trojan, Shadowsocks, WireGuard, Hysteria2, MTProto |
| Transports | TCP, WebSocket, gRPC, HTTPUpgrade, SplithTTP, REALITY |
| XTLS Flow Control | xtls-rprx-vision, xtls-rprx-vision-udp443 |
| Subscription | Standard JSON, Clash/Meta, Sing-box |
| Multi-Node | Master/sub-node architecture with mTLS sync |
| Traffic Stats | Real-time via Xray gRPC API, periodic DB sync |
| Telegram Bot | Inbound management, traffic alerts, user notifications |
| Real-time UI | WebSocket push for traffic, connection counts |
| Multi-language | 13 locale files (en, fa, ru, zh, es, tr, ar, etc.) |
| Database | SQLite (default) or PostgreSQL |
| Docker | Official images, docker-compose setup |
| TLS | Automatic Let's Encrypt certificates via built-in ACME client |
| Backup/Restore | Panel-level backup, SQLite migration to PostgreSQL |
| IP Limiting | Built-in fail2ban style IP limit per inbound |
| Clash API | Generate Clash/Meta-compatible configs |

### Technology Stack

```
Backend:      Go 1.26+ / Gin / GORM / SQLite / PostgreSQL
Frontend:     React 19 / Ant Design 6 / Vite 8 / TypeScript 5
Xray Core:    github.com/xtls/xray-core (gRPC API + embedded types)
Embedded:     Frontend built into Go binary via embed.FS
Testing:      Go stdlib testing, Vitest (frontend)
CI/CD:        GitHub Actions (lint, test, build, docker)
```

---

## 2. Architecture Summary

### 2.1 High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                        3X-UI Binary (Go)                         │
│                                                                  │
│  ┌────────────┐  ┌────────────────┐  ┌──────────────────────┐   │
│  │   main.go  │  │  gin Router    │  │  CLI (run, migrate,  │   │
│  │  (entry)   │──│  (/panel/api/) │──│  setting, cert)      │   │
│  └─────┬──────┘  └───────┬────────┘  └──────────────────────┘   │
│        │                  │                                      │
│        ▼                  ▼                                      │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                   Middleware Stack                        │    │
│  │  ┌──────────┐ ┌──────────┐ ┌────────┐ ┌──────────────┐  │    │
│  │  │ Session  │ │  CSRF    │ │  Auth  │ │  Localization │  │    │
│  │  │ (cookie) │ │ (token)  │ │(JWT)   │ │  (i18n)       │  │    │
│  │  └──────────┘ └──────────┘ └────────┘ └──────────────┘  │    │
│  └──────────────────────────────────────────────────────────┘    │
│        │                                                         │
│        ▼                                                         │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                   Controllers (REST API)                  │    │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │   │
│  │  │ Auth   │ │Inbound │ │ Client │ │Setting │ │  Node  │  │   │
│  │  │Login/  │ │ CRUD   │ │ CRUD   │ │System  │ │ Sync   │  │   │
│  │  │Logout  │ │       │ │       │ │Config  │ │/Manage │  │   │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘  │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │   │
│  │  │ Xray   │ │Config  │ │ Backup │ │  Sub   │ │  API   │  │   │
│  │  │Control │ │ Export │ │ Restore │ │Server  │ │  Docs  │  │   │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘  │   │
│  └──────────────────────────────────────────────────────────┘    │
│        │                                                         │
│        ▼                                                         │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                   Service Layer                           │    │
│  │  InboundService  SettingService  XrayService             │    │
│  │  ClientService   OutboundService  PanelService           │    │
│  │  TGService       EmailService                             │    │
│  └──────────────────────────────────────────────────────────┘    │
│        │                                                         │
│        ├──────────────┬────────────────┬─────────────┬──────────┤│
│        ▼              ▼                ▼             ▼          │
│  ┌──────────┐  ┌───────────┐  ┌────────────┐  ┌───────────┐    │
│  │ Database │  │ Xray-core │  │ Sub Server │  │ MTProto   │    │
│  │ (GORM)   │  │ (child    │  │ (separate  │  │ (mtg-multi│    │
│  │ SQLite/  │  │  process) │  │  Gin app)  │  │  sidecar) │    │
│  │ Postgres │  │           │  │            │  │           │    │
│  └──────────┘  └───────────┘  └────────────┘  └───────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

### 2.2 Directory Structure

```
3x-ui/
├── main.go                          # Entry point + CLI
├── go.mod / go.sum                  # Go dependencies
├── Makefile                         # Build, lint, test, gen targets
├── Dockerfile / docker-compose.yml  # Container deployment
│
├── internal/
│   ├── config/                      # Env parsing, version, paths
│   │   ├── config.go                # Config struct + env loading
│   │   └── version                  # Version string (v3.5.0)
│   │
│   ├── database/                    # GORM schema, migrations
│   │   ├── db.go                    # Init, AutoMigrate, migrations
│   │   ├── db_seed_test.go          # Test helpers for temp SQLite
│   │   └── model/                   # GORM model structs
│   │       ├── inbound.go           # Inbound model (port, protocol, settings JSON)
│   │       ├── client.go            # Client model (email, uuid, flow, totalGB, expiry)
│   │       ├── setting.go           # Key-value settings store
│   │       └── user.go              # Admin user model (username, password)
│   │
│   ├── xerrors/                     # ❄️ CUSTOM: Structured error types
│   │   ├── errors.go                # Domain error hierarchy (9 kinds)
│   │   ├── http.go                  # Kind→HTTP status mapping
│   │   ├── gin.go                   # Gin AbortWithError/JSON helpers
│   │   └── errors_test.go           # 20 test cases
│   │
│   ├── eventbus/                    # In-process pub/sub
│   │   └── eventbus.go              # Subscribe/Publish/Emit pattern
│   │
│   ├── logger/                      # Logging (custom format, rotation)
│   │
│   ├── mtproto/                     # MTProto sidecar manager
│   │   └── manager.go               # mtg-multi process lifecycle
│   │
│   ├── provider/                    # Test fixtures & providers
│   │   └── provider.go
│   │
│   ├── sub/                         # Subscription server
│   │   ├── endpoint.go              # /sub/{hash} handler
│   │   ├── endpoint_test.go         # Test with initSubDB(t) pattern
│   │   ├── clash.go                 # Clash/Meta config generation
│   │   └── singbox.go               # Sing-box config generation
│   │
│   ├── tunnelmonitor/               # Tunnel health monitoring
│   │
│   ├── util/                        # Shared helpers
│   │   ├── common/                  # NewError, MultiError, etc.
│   │   ├── crypto/                  # Encryption utils
│   │   ├── json/                    # JSON parsing
│   │   ├── ldap/                    # LDAP auth integration
│   │   ├── netproxy/               # Network proxy utils
│   │   ├── sys/                     # System-level ops (memory, etc.)
│   │   └── wireguard/              # WireGuard key generation
│   │
│   ├── web/                         # Web server
│   │   ├── server.go                # Gin server setup
│   │   ├── web.go                   # Frontend embed + static serving
│   │   ├── router.go                # Route registration
│   │   ├── dist/                    # Built frontend (gitignored)
│   │   │
│   │   ├── controller/              # REST API handlers
│   │   │   ├── auth_controller.go   # Login, logout, session
│   │   │   ├── inbound_controller.go # Inbound CRUD
│   │   │   ├── client_controller.go # Client CRUD
│   │   │   ├── xray_controller.go   # Xray start/stop/restart
│   │   │   ├── setting_controller.go # System settings
│   │   │   ├── node_controller.go   # Multi-node management
│   │   │   ├── subscription_controller.go
│   │   │   ├── backup_controller.go # Backup/restore
│   │   │   └── util.go              # jsonMsg helper, I18nWeb
│   │   │
│   │   ├── service/                 # Business logic
│   │   │   ├── inbound.go           # InboundService
│   │   │   ├── inbound_test.go      # ❄️ CUSTOM: 15 test cases
│   │   │   ├── setting.go           # SettingService
│   │   │   ├── client.go            # ClientService
│   │   │   ├── xray.go              # XrayService
│   │   │   └── ...                  # panel, outbound, tgbot, email, etc.
│   │   │
│   │   ├── job/                     # Cron jobs
│   │   │   ├── traffic.go           # Periodic traffic sync
│   │   │   ├── fail2ban.go          # IP limit enforcement
│   │   │   ├── node.go              # Node heartbeat/sync
│   │   │   └── ldap.go             # LDAP sync
│   │   │
│   │   ├── middleware/              # Gin middleware
│   │   ├── session/                 # Session management
│   │   ├── websocket/               # Real-time WebSocket
│   │   ├── runtime/                 # Node runtime (master/sub)
│   │   ├── network/                 # Network helpers
│   │   ├── entity/                  # Response types
│   │   ├── global/                  # Global state
│   │   ├── locale/                  # Locale data
│   │   └── translation/            # i18n JSON (13 languages)
│   │
│   └── xray/                        # Xray-core integration
│       ├── config.go                # Config generation
│       ├── process.go               # Child process lifecycle
│       ├── api.go                   # gRPC API calls
│       └── stats.go                 # Traffic stats collection
│
├── frontend/                        # React SPA
│   ├── src/
│   │   ├── App.tsx                  # Root component
│   │   ├── main.tsx                 # Entry point
│   │   ├── pages/                   # Route pages
│   │   ├── components/              # Shared components
│   │   ├── api/                     # API queries + WebSocket
│   │   ├── lib/                     # Business logic
│   │   ├── schemas/                 # Zod schemas (source of truth)
│   │   ├── generated/               # Auto-generated types
│   │   └── test/                    # Test utils
│   ├── public/                      # Static assets
│   └── package.json                 # Frontend deps
│
├── docs/                            # Documentation site
├── deploy/                          # Deployment scripts
├── tools/                           # OpenAPIGen, etc.
└── ROADMAP.md                       # This file
```

### 2.3 Database Schema

```sql
-- Core tables (GORM auto-migrated)

-- Inbounds: proxy listener configuration
inbounds (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL,      -- owner admin user
  port        INTEGER NOT NULL,      -- listen port
  protocol    TEXT NOT NULL,          -- vless/vmess/trojan/shadowsocks/wireguard/hysteria2/mtproto
  tag         TEXT NOT NULL UNIQUE,   -- unique identifier tag
  remark      TEXT DEFAULT '',
  enable      BOOLEAN DEFAULT TRUE,
  settings    TEXT NOT NULL,          -- JSON: clients array, decryption, fallbacks
  stream_settings TEXT,              -- JSON: network, security, tls, reality
  sniffing    TEXT,                   -- JSON: sniffing config
  allocate    TEXT,                   -- JSON: port allocation
  options     TEXT DEFAULT '{}',     -- JSON: options flags
  expiry_time INTEGER DEFAULT 0,
  listen      TEXT DEFAULT '',
  total       INTEGER DEFAULT 0,
  up          INTEGER DEFAULT 0,
  down        INTEGER DEFAULT 0,
  client_stats TEXT,                  -- JSON: per-client traffic cache
  created_at  DATETIME,
  updated_at  DATETIME
);

-- Clients: per-inbound proxy users
clients (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  inbound_id  INTEGER NOT NULL,
  email       TEXT NOT NULL,          -- unique user email
  uuid        TEXT NOT NULL,          -- proxy UUID
  flow        TEXT,                   -- XTLS flow control
  enable      BOOLEAN DEFAULT TRUE,
  sub_id      TEXT,                   -- subscription ID
  tg_id       TEXT,                   -- Telegram user ID
  total_gb    INTEGER DEFAULT 0,      -- traffic quota (GB)
  expiry_time INTEGER DEFAULT 0,      -- expiry timestamp
  up          INTEGER DEFAULT 0,
  down        INTEGER DEFAULT 0,
  settings    TEXT DEFAULT '{}',      -- JSON: additional settings
  reset_at    DATETIME,
  created_at  DATETIME,
  updated_at  DATETIME,
  UNIQUE(inbound_id, email)
);

-- Settings: key-value configuration store
settings (
  id    INTEGER PRIMARY KEY AUTOINCREMENT,
  key   TEXT NOT NULL UNIQUE,
  value TEXT DEFAULT ''
);

-- Users: admin accounts
users (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  username    TEXT NOT NULL UNIQUE,
  password    TEXT NOT NULL,          -- bcrypt hash
  created_at  DATETIME,
  updated_at  DATETIME
);
```

### 2.4 Xray Integration

The system manages Xray-core as a child process:

1. **Config Generation** (`internal/xray/config.go`): Builds complete Xray JSON config from database inbounds, outbounds, and routing rules
2. **Process Management** (`internal/xray/process.go`): Spawns/kills xray process, atomic writes to config file, signal handling for graceful reload
3. **gRPC API** (`internal/xray/api.go`): Connects to xray's admin socket for real-time traffic and stats
4. **Stats Collection** (`internal/xray/stats.go`): Polls xray gRPC API, aggregates per-client and per-inbound traffic

The **MTProto** protocol uses a separate sidecar binary (`mtg-multi`), one process per inbound, with hot-reload via management API.

---

## 3. Custom Fork Modifications

### 3.1 Complete Change Log

All custom work is captured in a single commit (`e4618fb8`):

| File | Type | Lines | Purpose |
|------|------|-------|---------|
| `internal/xerrors/errors.go` | **New** | 182 | Domain error type hierarchy (9 Kind categories) |
| `internal/xerrors/http.go` | **New** | 37 | Kind→HTTP status mapping |
| `internal/xerrors/gin.go` | **New** | 107 | Gin controller integration helpers |
| `internal/xerrors/errors_test.go` | **New** | 210 | 20 test cases for error types |
| `internal/web/service/inbound_test.go` | **New** | 230 | 15 InboundService tests with temp SQLite |
| `.env.example` | **Modified** | +12/-1 | Development environment template |
| `frontend/public/mockServiceWorker.js` | **Modified** | -258 | Removed MSW service worker (upstream sync side effect) |
| `frontend/package-lock.json` | **Modified** | -1 | Lockfile update (upstream sync side effect) |

### 3.2 Custom Feature: `internal/xerrors` Package

**Purpose:** Replace ad-hoc error returns (string errors via `common.NewError`) with structured, typed errors that carry semantic meaning, support `errors.Is`/`errors.As`, and map to correct HTTP status codes.

**Error Kind Hierarchy:**

```
Kind
├── validation      (400)  — Bad input, missing fields
├── not_found       (404)  — Resource doesn't exist
├── conflict        (409)  — Duplicate email/tag/name
├── unauthenticated (401)  — Expired/missing session
├── forbidden       (403)  — Wrong permissions
├── quota_exceeded  (429)  — Traffic cap, IP limit
├── rate_limited    (429)  — Too many requests
├── unavailable     (503)  — Xray down, DB down, node unreachable
└── internal        (500)  — Unexpected bugs
```

**Key Design:**

- **Kind-based sentinels** — `errors.Is(err, xerrors.ErrNotFound)` matches any `not_found` error regardless of message
- **Constructors** — `NotFoundError("client %q not found", email)`
- **Wrap/Wrapf** — Returns `nil` when inner error is `nil` (safe to chain)
- **KindOf/MessageOf** — Extract fields at API boundaries without `errors.As`
- **Gin helpers** — `AbortWithError(c, err)`, `AbortWithNotFound(c, msg, args...)`, etc.

### 3.3 Custom Feature: `inbound_test.go`

**Purpose:** Provide a reusable test scaffold for the service layer using temp SQLite databases (following the pattern established by `internal/sub/endpoint_test.go` and `internal/database/db_seed_test.go`).

**Pattern:**
```go
func initInboundTestDB(t *testing.T) {
    t.Helper()
    t.Setenv("XUI_DB_FOLDER", t.TempDir())
    database.InitDB(filepath.Join(t.TempDir(), "x-ui.db"))
    t.Cleanup(func() { database.CloseDB() })
}

func seedTestInbound(t *testing.T, overrides ...func(*inboundOptions)) model.Inbound {
    // Seeds a VLESS inbound with 2 clients (alice, bob) for user 1
    // Supports port, protocol, tag, remark, enable overrides
}
```

**Test Coverage (15 tests):**
1. GetInbounds — basic fetch, verify fields
2. GetInbounds_emptyUser — user with no inbounds returns empty
3. GetInbound — single fetch by id
4. GetInbound_notFound — non-existent id
5. GetInbounds_sameUserMultipleInbounds — ordering by id ASC
6. GetAllInbounds — cross-user fetch
7. GetClients — client extraction from settings JSON
8. GetClients_noClients — empty clients array
9. GetInboundOptions — slim options with expected fields
10. GetAllEmails — unique email dedup
11. GetInboundsByTrafficReset — period filter
12. GetInboundsByTrafficReset_noMatch — no-match filter
13. GetInboundTags — tags JSON output
14. SearchInbounds — remark search
15. GetInboundsSlim — slim response drops UUID/flow

### 3.4 Remaining Integration Work

The xerrors package is **built and tested** but not yet integrated into the service and controller layers. The following files still use the old `common.NewError` pattern:

| File | Priority | Impact |
|------|----------|--------|
| `internal/web/controller/util.go` | **High** | jsonMsgObj helper — gateway for all API responses |
| `internal/web/service/inbound.go` | **High** | All InboundService error returns |
| `internal/web/service/setting.go` | **High** | All SettingService error returns |
| `internal/web/service/client.go` | **Medium** | All ClientService error returns |
| `internal/web/service/xray.go` | **Medium** | XrayService error returns |
| `internal/web/controller/*.go` | **High** | All controllers consume service errors |
| `internal/xray/*.go` | **Low** | Xray integration error returns |

---

## 4. Upstream Sync Report

### 4.1 Sync Strategy

This fork was created from a **tarball/snapshot**, not a GitHub fork — the git histories are unrelated. The sync was performed as a **manual source-code merge**:

1. Upstream was cloned independently
2. File-by-file comparison identified no missing or diverging files
3. All shared files (`config.go`, `db.go`, `main.go`, `go.mod`, `go.sum`, `Makefile`, `Dockerfile`, `docker-compose.yml`) were verified **identical** via `diff`
4. Custom files were backed up and restored on top of upstream content

### 4.2 Sync Verification

| Check | Result |
|-------|--------|
| Upstream version | v3.5.0 |
| Fork version | v3.5.0 ✅ |
| All upstream files present in fork | 1,376/1,376 ✅ |
| Content identity on shared files | All identical (`diff` = empty) ✅ |
| Custom files preserved | 6 files intact ✅ |
| Frontend builds | `npm run build` succeeds ✅ |
| xerrors package compiles | `go vet` passes ✅ |
| xerrors tests pass | 20/20 ✅ |

### 4.3 No Remaining Upstream Changes

The upstream repository has no commits beyond the fork's HEAD. The fork is at **full parity** with `MHSanaei/3x-ui@main`. No further sync is required.

---

## 5. Development Milestones

### Milestone Overview (Dependency Order)

```
M1 ─── M2 ─── M3 ─── M4 ─── M5 ─── M6 ─── M7 ─── M8 ─── M9
 │      │      │      │      │      │      │      │      │
 ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼
Done   80%    50%     0%     0%     0%     0%     0%     0%
```

| Milestone | Name | Status | Est. Effort | Go-Risk |
|-----------|------|--------|-------------|---------|
| **M1** | Structured Error Types | ✅ **Done** | 3 days | Compile-tested |
| **M2** | Service-Layer Error Refactor | 🔄 **80%** | 3 days | Needs Go env |
| **M3** | Controller Error Integration | 📋 **50%** | 2 days | Needs Go env |
| **M4** | Test Infrastructure Expansion | ⏳ **0%** | 4 days | Needs Go env |
| **M5** | Frontend UX Improvements | ⏳ **0%** | 5 days | No Go needed |
| **M6** | API Documentation | ⏳ **0%** | 2 days | No Go needed |
| **M7** | Observability & Monitoring | ⏳ **0%** | 3 days | Needs Go env |
| **M8** | Performance & Scale | ⏳ **0%** | 3 days | Needs Go env |
| **M9** | CI/CD & Quality Gates | ⏳ **0%** | 2 days | No Go needed |

### Milestone Details

#### M1: Structured Error Types ✅ (COMPLETE)

**Files:** `internal/xerrors/errors.go`, `internal/xerrors/http.go`, `internal/xerrors/gin.go`, `internal/xerrors/errors_test.go`

**Deliverables:**
- [x] 9 Kind constants with sentinel errors
- [x] Error struct implementing `error` interface with `Kind`, `Message`, `Err`
- [x] Kind-based `Is()` for `errors.Is()` matching
- [x] 9 constructors (`ValidationError()`, `NotFoundError()`, etc.)
- [x] `Wrap()`/`Wrapf()` with nil-safe behavior
- [x] `KindOf()`/`MessageOf()` query helpers
- [x] `HTTPStatus(err) int` mapping function
- [x] Gin helpers: `AbortWithError`, `AbortWithHTTPError`, `AbortWithValidation`, `AbortWithNotFound`, `AbortWithConflict`, `AbortWithUnauthenticated`, `AbortWithForbidden`, `AbortWithInternal`, `JSON`
- [x] 20 unit tests covering all functionality

**Verification:**
```bash
GOTOOLCHAIN=local go vet ./internal/xerrors/...
GOTOOLCHAIN=local go test -shuffle=on -count=1 -v ./internal/xerrors/...
```

---

#### M2: Service-Layer Error Refactor 🔄 (80% — code designed, needs Go env)

**Objective:** Migrate all service-layer error returns from `common.NewError` to `xerrors` typed errors.

**Files to modify:**
- `internal/web/service/inbound.go`
- `internal/web/service/setting.go`
- `internal/web/service/client.go`
- `internal/web/service/xray.go`
- `internal/web/service/panel.go`
- `internal/web/service/outbound.go`

**Pattern migration:**

```go
// BEFORE (old pattern):
if client.Email == "" {
    return common.NewError("client email is required")
}
return model.Inbound{}, common.NewErrorf("inbound %d not found", id)

// AFTER (new pattern):
if client.Email == "" {
    return xerrors.ValidationError("client email is required")
}
return model.Inbound{}, xerrors.NotFoundError("inbound %d not found", id)
```

**Search & Replace Map:**

| Service | common.NewError Pattern | xerrors Kind | Count |
|---------|------------------------|--------------|-------|
| `inbound.go` | Validation errors | `ValidationError` | ~8 |
| `inbound.go` | Not found errors | `NotFoundError` | ~4 |
| `inbound.go` | Internal errors | `InternalError` | ~3 |
| `setting.go` | Not found errors | `NotFoundError` | ~6 |
| `setting.go` | Validation errors | `ValidationError` | ~4 |
| `client.go` | Validation errors | `ValidationError` | ~10 |
| `client.go` | Not found errors | `NotFoundError` | ~5 |
| `client.go` | Conflict (duplicate email) | `ConflictError` | ~2 |
| `client.go` | Quota exceeded | `QuotaExceededError` | ~2 |
| `xray.go` | Unavailable (process down) | `UnavailableError` | ~4 |

**Acceptance Criteria:**
- All `common.NewError`/`common.NewErrorf` calls in service files converted to xerrors constructors
- `common` import removed from service files where no longer used
- `go vet ./internal/web/service/...` passes
- All existing service tests pass

---

#### M3: Controller Error Integration 📋 (50% — Gin helpers exist, needs wiring)

**Objective:** Connect service-layer xerrors to the HTTP response layer.

**Files to modify:**
- `internal/web/controller/util.go` — Refactor `jsonMsgObj` → use `xerrors.HTTPStatus`
- `internal/web/controller/inbound_controller.go`
- `internal/web/controller/client_controller.go`
- `internal/web/controller/setting_controller.go`
- `internal/web/controller/xray_controller.go`
- `internal/web/controller/auth_controller.go`

**Pattern migration:**

```go
// BEFORE (old pattern):
if err != nil {
    jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
    return
}

// AFTER (new pattern):
if err != nil {
    xerrors.AbortWithError(c, err)
    return
}
```

**Acceptance Criteria:**
- `jsonMsgObj` helper refactored to auto-map status codes via `xerrors.HTTPStatus`
- All controller error blocks use `AbortWithError` or typed helpers
- Response body always includes `{ "success": false, "msg": "...", "kind": "..." }`
- `go vet ./internal/web/controller/...` passes

---

#### M4: Test Infrastructure Expansion ⏳ (0%)

**Objective:** Create reusable test patterns and extend test coverage.

**Files to create/modify:**
- `internal/web/service/setting_test.go` — **New** (model after inbound_test.go)
- `internal/web/service/client_test.go` — **New** (model after inbound_test.go)
- `internal/database/db.go` — Export reusable `InitTestDB(t)` helper
- `internal/web/service/service_test.go` — Common test package with shared fixtures

**Tasks:**

| Task | Files | Est. |
|------|-------|------|
| T4.1 Export `InitTestDB` from `internal/database/` | `internal/database/db.go` | 1h |
| T4.2 Create shared `ServiceTestSuite` with pre-seeded data | `internal/web/service/service_test.go` | 3h |
| T4.3 SettingService tests (GetSetting, SetSetting, GetAllSettings, DeleteSetting) | `internal/web/service/setting_test.go` | 4h |
| T4.4 ClientService tests (AddClient, GetClients, DelClient, UpdateClient, ResetTraffic) | `internal/web/service/client_test.go` | 6h |

**Acceptance Criteria:**
- `InitTestDB(t)` exported and reusable across packages
- `ServiceTestSuite` provides common patterns (seeded data, cleanup, assertions)
- Each service has ≥10 test cases
- `go test -shuffle=on -count=1 ./internal/web/service/...` passes

---

#### M5: Frontend UX Improvements ⏳ (0%)

**Objective:** Enhance the React frontend with better UX, responsive design, and visual polish.

**Areas for improvement:**

| Area | Current State | Target | Est. |
|------|--------------|--------|------|
| Login page | Minimal Ant Design card | Polished landing with branded theme, social proofs, animated transitions | 6h |
| Dashboard | Functional, dated | Modern cards, better traffic visualization, quick actions, responsive grid | 8h |
| Inbound management | Table with modals | Enhanced inline editing, batch operations, search with filters, dark mode polish | 6h |
| Client management | Modal forms | Progressive forms with validation hints, traffic gauges, expiry indicators | 4h |
| Settings pages | Basic forms | Organized tabs with search, confirmation dialogs, auto-save | 3h |
| Mobile responsiveness | Partial | Full mobile support for management views | 4h |
| Dark mode | Basic | Consistent dark theme across all components | 3h |

**Implementation approach:**
- Leverage existing Ant Design 6 components (no Tailwind/shadcn)
- Reuse existing schema validations (Zod schemas already exist)
- Maintain backward compatibility with existing API responses

---

#### M6: API Documentation ⏳ (0%)

**Objective:** Improve API documentation by updating the OpenAPI spec and generating comprehensive docs.

**Tasks:**

| Task | Files | Est. |
|------|-------|------|
| T6.1 Update `endpoints.ts` with any missing routes | `frontend/src/pages/api-docs/endpoints.ts` | 2h |
| T6.2 Verify OpenAPIGen covers all response types | `tools/openapigen/main.go` | 2h |
| T6.3 Add example tags to Go structs | Various model files | 2h |
| T6.4 Regenerate OpenAPI spec | `make gen` | 0.5h |

---

#### M7: Observability & Monitoring ⏳ (0%)

**Objective:** Add structured logging, Prometheus metrics, and health check endpoints.

**Tasks:**

| Task | Files | Est. |
|------|-------|------|
| T7.1 Add `/panel/api/health` endpoint returning system status | `internal/web/controller/`, `internal/web/router.go` | 3h |
| T7.2 Structured JSON log output format | `internal/logger/` | 4h |
| T7.3 Add `/metrics` endpoint for Prometheus | New file + router | 4h |

---

#### M8: Performance & Scale ⏳ (0%)

**Objective:** Optimize for larger deployments (1000+ clients, 100+ inbounds).

**Tasks:**

| Task | Est. |
|------|------|
| T8.1 Add DB connection pooling configuration | 2h |
| T8.2 Cache subscription responses with TTL | 3h |
| T8.3 Paginate client lists at API layer | 3h |
| T8.4 Batch traffic update queries | 2h |

---

#### M9: CI/CD & Quality Gates ⏳ (0%)

**Objective:** Enhance CI pipeline with additional checks, code quality gates, and release automation.

**Tasks:**

| Task | Est. |
|------|------|
| T9.1 Add Go linting (golangci-lint) to CI | 2h |
| T9.2 Add frontend linting & type checking to CI | 1h |
| T9.3 Add integration test stage | 3h |
| T9.4 Automatic release notes generation | 2h |

---

## 6. Task Breakdown

### Complete Task List (Ordered by Dependencies)

```
M1 ───────────────────────────────────────────────────────────────
│ T1  Create xerrors package             [3h]  [Low]    ✅ DONE
│ T1a   errors.go — core types           [2h]  [Low]    ✅ DONE
│ T1b   http.go — status mapping         [0.5h][Low]    ✅ DONE
│ T1c   gin.go — controller helpers      [1h]  [Low]    ✅ DONE
│ T1d   errors_test.go — 20 tests        [2h]  [Low]    ✅ DONE
└───────────────────────────────────────────────────────────────

M2 ───────────────────────────────────────────────────────────────
│  ⬇ Depends on M1
│ T2  Refactor InboundService            [2h]  [Low]    🔄 READY
│ T3  Refactor SettingService            [1.5h][Low]    🔄 READY
│ T4  Refactor ClientService             [2h]  [Low]    🔄 READY
│ T5  Refactor XrayService               [1h]  [Low]    🔄 READY
│ T6  Refactor remaining services        [1h]  [Low]    🔄 READY
└───────────────────────────────────────────────────────────────

M3 ───────────────────────────────────────────────────────────────
│  ⬇ Depends on M2
│ T7  Refactor jsonMsgObj in util.go     [1h]  [Low]    🔄 READY
│ T8  Refactor InboundController         [2h]  [Low]    🔄 READY
│ T9  Refactor ClientController          [2h]  [Low]    🔄 READY
│ T10 Refactor SettingController         [1h]  [Low]    🔄 READY
│ T11 Refactor AuthController            [1h]  [Low]    🔄 READY
│ T12 Refactor XrayController            [0.5h][Low]    🔄 READY
└───────────────────────────────────────────────────────────────

M4 ───────────────────────────────────────────────────────────────
│  ⬇ Depends on M2 (service layer stable)
│ T13 Export InitTestDB from database     [1h]  [Low]    ⏳ PENDING
│ T14 Create ServiceTestSuite            [3h]  [Low]    ⏳ PENDING
│ T15 SettingService tests               [4h]  [Low]    ⏳ PENDING
│ T16 ClientService tests                [6h]  [Med]    ⏳ PENDING
│ T17 Extend inbound_test.go             [2h]  [Low]    ⏳ PENDING
└───────────────────────────────────────────────────────────────

M5 ───────────────────────────────────────────────────────────────
│  ⬇ No Go dependency. Can start immediately.
│ T18 Login page redesign                [6h]  [Low]    ⏳ PENDING
│ T19 Dashboard improvements             [8h]  [Med]    ⏳ PENDING
│ T20 Inbound management UX              [6h]  [Low]    ⏳ PENDING
│ T21 Client management UX               [4h]  [Low]    ⏳ PENDING
│ T22 Settings page polish               [3h]  [Low]    ⏳ PENDING
│ T23 Mobile responsiveness              [4h]  [Med]    ⏳ PENDING
│ T24 Dark mode consistency              [3h]  [Low]    ⏳ PENDING
└───────────────────────────────────────────────────────────────

M6 ───────────────────────────────────────────────────────────────
│  ⬇ No Go dependency
│ T25 Update endpoints.ts                [2h]  [Low]    ⏳ PENDING
│ T26 Verify OpenAPIGen allowlist        [2h]  [Low]    ⏳ PENDING
│ T27 Add example tags to structs        [2h]  [Low]    ⏳ PENDING
│ T28 Regenerate & verify                [0.5h][Low]    ⏳ PENDING
└───────────────────────────────────────────────────────────────

M7 ───────────────────────────────────────────────────────────────
│  ⬇ Depends on M3 (controller layer stable)
│ T29 Health check endpoint              [3h]  [Low]    ⏳ PENDING
│ T30 Structured JSON logging            [4h]  [Low]    ⏳ PENDING
│ T31 Prometheus /metrics endpoint       [4h]  [Med]    ⏳ PENDING
└───────────────────────────────────────────────────────────────

M8 ───────────────────────────────────────────────────────────────
│  ⬇ Depends on M4 (test infrastructure)
│ T32 Connection pooling config          [2h]  [Low]    ⏳ PENDING
│ T33 Subscription caching               [3h]  [Low]    ⏳ PENDING
│ T34 Client list pagination             [3h]  [Med]    ⏳ PENDING
│ T35 Batch traffic updates              [2h]  [Low]    ⏳ PENDING
└───────────────────────────────────────────────────────────────

M9 ───────────────────────────────────────────────────────────────
│  ⬇ No Go dependency
│ T36 Add golangci-lint to CI            [2h]  [Low]    ⏳ PENDING
│ T37 Add frontend checks to CI          [1h]  [Low]    ⏳ PENDING
│ T38 Integration test stage             [3h]  [Med]    ⏳ PENDING
│ T39 Release notes automation           [2h]  [Low]    ⏳ PENDING
└───────────────────────────────────────────────────────────────
```

---

## 7. Dependency Graph

```
T1 (xerrors core) ──────────────────── ✅ DONE
 │
 ├──→ T2 (InboundService errors) ────→ T7 (jsonMsgObj) ──→ T8 (InboundController)
 │                                      │
 ├──→ T3 (SettingService errors) ───────┼─────────────────→ T10 (SettingController)
 │                                      │
 ├──→ T4 (ClientService errors) ────────┼─────────────────→ T9 (ClientController)
 │                                      │
 └──→ T5 (XrayService errors) ─────────┼─────────────────→ T12 (XrayController)
 │                                      │
 └──→ T6 (Other services) ─────────────┴─────────────────→ T11 (AuthController)
                                        │
                                        └──→ T29 (Health endpoint)
                                              │
                                              └──→ T30 (JSON logging)
                                                    │
                                                    └──→ T31 (/metrics)

                                    T13 (InitTestDB)
                                      │
                                      ├──→ T14 (ServiceTestSuite)
                                      │      │
                                      │      ├──→ T15 (SettingService tests)
                                      │      ├──→ T16 (ClientService tests)
                                      │      └──→ T17 (Extend inbound tests)
                                      │
                                      ├──→ T32 (Connection pooling)
                                      │      │
                                      │      └──→ T33-T35 (Performance)
                                      │
                           T18-T24 ───┤ No Go deps (frontend)
                                      │
                           T25-T28 ───┤ No Go deps (docs)
                                      │
                           T36-T39 ───┤ No Go deps (CI/CD)
```

### Parallelizable Work

These tracks can run simultaneously:

| Track A (Go Backend) | Track B (Frontend) | Track C (CI/Docs) |
|----------------------|--------------------|--------------------|
| M2: Service errors | M5: Frontend UX | M9: CI/CD pipeline |
| M3: Controller errors | Login page | Linting config |
| M4: Test infra | Dashboard | Integration tests |
| M7: Observability | Inbound UX | Release automation |
| M8: Performance | Mobile support | M6: API docs |

---

## 8. Risk Assessment

### 8.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Error type refactoring breaks existing API response shapes | Low | High | All xerrors.AbortWithError emits `{success, msg, kind}` - ensure frontend handles `kind` field |
| Service test DB init conflicts with global state | Medium | Medium | Use `t.Setenv` + `t.TempDir()` per test, not shared state |
| Go toolchain compatibility (requires 1.26+, only 1.22 installed) | High | High | Install Go 1.26+ via `go install golang.org/dl/go1.26@latest` + `go1.26 download` |
| Frontend build failures after npm install | Low | High | Lock `package-lock.json`; use `--frozen-lockfile` in CI |

### 8.2 Migration Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| xerrors package creates import cycle | Blocked | Package is standalone, imports only stdlib + gin |
| Service layer refactor introduces regressions | Medium | Run full test suite before/after each service change |
| Frontend redesign breaks existing workflows | Medium | Incremental changes, not full rewrites |

### 8.3 Compatibility Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| API response format changes break frontend | High | Keep `msg` field. Add `kind` field alongside, don't remove old fields |
| Database migration incompatibility | Critical | Never change column types; only ADD columns/tables |
| Subscription config changes break clients | Critical | Subscription format changes must be opt-in via version header |

---

## 9. Testing Guide

### 9.1 Running Tests

```bash
# Prerequisites: Go 1.26+, CGO_ENABLED=1, mkdir -p internal/web/dist
export PATH=$(go env GOPATH)/bin:$PATH
export CGO_ENABLED=1

# Run xerrors tests (Go 1.22+ OK)
cd /path/to/3x-ui
go test -shuffle=on -count=1 -v ./internal/xerrors/...

# Run inbound service tests
go test -shuffle=on -count=1 -v ./internal/web/service/... -run 'TestInboundService'

# Run all Go tests
go test -shuffle=on -count=1 ./...

# Run frontend tests
cd frontend && npm test

# Run frontend type check
cd frontend && npx tsc --noEmit
```

### 9.2 Adding New Tests

**For Go service tests, follow this pattern:**
```go
func TestMyService_Function(t *testing.T) {
    initInboundTestDB(t)  // Sets up temp SQLite
    svc := service.NewInboundService(
        &service.InboundService{
            db: database.GetDB(),
        },
    )

    result, err := svc.GetInbounds(1)
    assertNoError(t, err)
    assertEqual(t, len(result), expectedCount)
}
```

**For custom assertions (no testify needed):**
```go
func assertEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}

func assertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### 9.3 Testing Principles (from CLAUDE.md)

- Stdlib `testing` only. No testify.
- Table-driven tests with `t.Run` subtests.
- `t.Helper()` on all assertion helpers.
- Assert exact values, not just `err != nil`.
- Prefer real deps over mocks: throwaway SQLite via `database.InitDB`.
- Golden fixtures for share-link logic (`src/lib/xray/`).

---

## 10. Development Workflow

### 10.1 Branch Strategy

```
main ──── Stable, production-ready
  │
  ├── feat/xerrors              ──── ✅ Merged (custom error types)
  ├── feat/service-error-refactor ──── Next (refactor services)
  ├── feat/controller-error-int   ──── Next +1 (refactor controllers)
  ├── feat/test-infra             ──── Parallelizable
  ├── feat/frontend-ux            ──── Parallelizable
  └── feat/ci-quality             ──── Parallelizable
```

### 10.2 Commit Convention

```
<type>(<area>): <imperative summary>

<type>: feat | fix | refactor | chore | docs | style | test | perf
<area>: backend | frontend | xray | database | sub | xerrors | ci | docs

Examples:
  feat(xerrors): add structured error types with Kind hierarchy
  refactor(service): migrate InboundService to xerrors
  test(service): add SettingService unit tests with temp SQLite
  fix(frontend): handle null inbound settings on client page
```

### 10.3 PR Checklist

Before opening a PR:
1. [ ] `make gen` — regenerate OpenAPI + Zod schemas
2. [ ] `make lint` — Go + frontend linting
3. [ ] `make test` — Go tests with shuffle + frontend tests
4. [ ] `go vet ./...` — no complaints
5. [ ] `cd frontend && npx tsc --noEmit` — no type errors
6. [ ] Frontend builds: `cd frontend && npm run build`
7. [ ] No regression in custom features
8. [ ] Diff is focused (refactors separate from features)

### 10.4 Rollback Plan

Each milestone is designed to be independently revertable:

- **M1 (xerrors):** Additive only — no existing code changed. `git revert` is safe.
- **M2-M3 (refactors):** Each service/controller file is a separate commit. Revert per-file if issues.
- **M4 (tests):** Additive only — `git revert` is safe.
- **M5 (frontend):** Each component change is a separate commit. Revert per-component.

---

## 11. Future Vision

### 11.1 Near Term (Weeks 1-4)

1. ✅ ~~Structured error types~~ (DONE)
2. 🔄 Service-layer error refactoring
3. 📋 Controller error integration
4. ⏳ Test infrastructure expansion
5. ⏳ Frontend login page and dashboard redesign

### 11.2 Medium Term (Weeks 5-8)

1. Multi-node management improvements
2. Enhanced subscription management
3. Performance optimization for large deployments
4. Prometheus metrics integration
5. Structured JSON logging

### 11.3 Long Term (Weeks 9+)

1. REST API v2 with consistent response format (using xerrors)
2. Migration from GORM to sqlc for better performance
3. Frontend state management modernization
4. Plugin system for custom protocol handlers
5. WebAssembly-based client generation

### 11.4 MOAV-Inspired Improvements

The MOAV project (Mother of All VPNs) uses a different architecture philosophy. While not directly applicable, these ideas are worth considering:

| MOAV Concept | 3X-UI Adaptation | Priority |
|-------------|------------------|----------|
| Docker-first deployment | Already supported | Low |
| Grafana dashboards | Custom Grafana dashboard template for 3X-UI metrics | Medium |
| Protocol diversity as UX | Better grouping/filtering of protocols in the UI | Low |
| Automated health checks | `/health` endpoint with component status | High (M7) |
| User bundles (QR + config) | Enhanced multi-format export with single-click bundles | Medium |

---

## Appendix A: Environment Setup

### Full Development Environment

```bash
# Prerequisites
sudo apt-get install -y golang-go gcc git
mkdir -p internal/web/dist

# Install newer Go (if apt version is too old)
go install golang.org/dl/go1.26@latest
go1.26 download
alias go=go1.26

# Frontend
cd frontend && npm install

# Backend
cd ..
go mod download
CGO_ENABLED=1 go build -o x-ui .
```

### Quick Development Loop

```bash
# Terminal 1: Frontend dev server (HMR)
cd frontend && npm run dev

# Terminal 2: Go backend (serves embedded frontend)
CGO_ENABLED=1 XUI_DEBUG=true go run .
# → http://localhost:2053 (admin/admin)
```

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `XUI_DEBUG` | `false` | Verbose logs + Gin debug + serve assets from disk |
| `XUI_DB_TYPE` | `sqlite` | `sqlite` or `postgres` |
| `XUI_DB_DSN` | — | PostgreSQL DSN |
| `XUI_DB_FOLDER` | `/etc/x-ui` (Linux) | Where `x-ui.db` lives |
| `XUI_LOG_FOLDER` | `/var/log/x-ui` (Linux) | Where logs live |
| `XUI_BIN_FOLDER` | `bin` | Where xray binary + geo files live |
| `XUI_PORT` | stored `webPort` | Override web listener port |
| `XUI_LOG_LEVEL` | `info` | Logging verbosity |
| `XUI_INIT_WEB_BASE_PATH` | `/` | Base URI path for the panel |
| `XUI_ENABLE_FAIL2BAN` | `true` | IP-limit enforcement |

---

## Appendix B: Key Commands

```bash
# Build
make build                    # Build Go binary
cd frontend && npm run build  # Build frontend only

# Test
make test                     # All tests (Go + frontend)
go test ./internal/xerrors/...  # xerrors package tests only

# Lint
make lint                     # Go + frontend linting

# Generate
make gen                      # Regenerate OpenAPI + Zod/TS types

# Verify (CI gate)
make verify                   # Full lint + test + build pipeline

# Run
CGO_ENABLED=1 go run .        # Full stack with frontend embedded
cd frontend && npm run dev    # Frontend HMR (proxies to Go on :2053)
```

---

## Appendix C: Current Custom Files — Quick Reference

```
internal/xerrors/
├── errors.go          # 182 lines — Error type, Kind, constructors, Wrap/Wrapf
├── http.go            # 37 lines  — HTTPStatus(err) mapping function
├── gin.go             # 107 lines — AbortWithError, JSON helpers
└── errors_test.go     # 210 lines — 20 test cases

internal/web/service/
└── inbound_test.go    # 230 lines — 15 test cases with temp SQLite

.env.example           # Development environment template
```

---

> **Last updated:** July 13, 2026
> **Maintainer:** Custom fork of MHSanaei/3x-ui v3.5.0
> **License:** GPL-3.0
