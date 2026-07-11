# Abystral VPN Platform — Development Roadmap

> Single source of truth for project development.
>
> **Project**: Transform 3x-ui v3.4.2 into a multi-provider VPN management
> platform with unified API consumed by the Abystral Telegram bot.
>
> **Repositories**:
> - `3x-ui-fork` — Go/Gin panel (this repo), forked from MHSanaei/3x-ui v3.4.2
> - `Abystral` — Python 3.13+ Telegram bot (aiogram 3.x, SQLAlchemy 2.x async)
>
> **Rules**:
> - Implement tasks sequentially within each milestone unless stated otherwise.
> - Never skip dependencies.
> - Mark tasks after successful implementation and verification.
> - Update this document after every completed milestone.
> - Keep synchronized with the codebase at all times.

---

## Task Status Legend

- ☐ Not Started
- ◐ In Progress
- ☑ Completed

---

## Overall Progress

| Milestone | Title | Status | Progress |
|-----------|-------|--------|----------|
| 1 | Foundation (Abystral) | ☑ Complete | 100% |
| 2 | VPN Platform | ◐ In Progress | 85% |
| 3 | BPB Integration | ☐ Not Started | 0% |
| 4 | Unified API | ☐ Not Started | 0% |
| 5 | Admin Experience | ☐ Not Started | 0% |
| 6 | User Experience | ☐ Not Started | 0% |
| 7 | Future Providers | ☐ Not Started | 0% |
| 8 | Production | ☐ Not Started | 0% |
| 9 | Nice-to-Have | ☐ Not Started | 0% |

**Current Milestone**: 2 — VPN Platform
**Next Milestone**: 3 — BPB Integration
**Overall Progress**: ~11% (Milestone 1 complete)

---

# Milestone 1 — Foundation

**Objective**: Production-grade Telegram VPN SaaS with wallet-based payments,
referral system, trial support, promo codes, multi-subscription management, and
full admin panel. Serves as the stable base for the multi-provider extension.

**Status**: ☑ Complete
**Complexity**: High
**Dependencies**: None

## Features (all delivered)

### 1.1 Core Architecture
- ☑ Project structure, Docker, configuration (Pydantic Settings)
- ☑ SQLAlchemy 2.x async engine + session factory
- ☑ Alembic async migrations
- ☑ Base model, User, Subscription, Server, Transaction models
- ☑ Repository pattern (base + per-entity repositories)
- ☑ Service layer (UserService, SubscriptionService, VPSService, PlanService)
- ☑ Dependency injection via aiogram dispatcher

### 1.2 3X-UI API Client
- ☑ `XUIBaseClient` — async HTTPX client with Bearer token / session auth
- ☑ Retry with exponential backoff, circuit breakers
- ☑ Inbound CRUD, client CRUD, subscription links, traffic stats, server status
- ☑ Pydantic models: `ClientTraffic`, `Inbound`, `Client`, `ClientCreate`, `ClientAttachPayload`, `ServerStatus`

### 1.3 Telegram Bot
- ☑ aiogram 3.x bot with middleware stack (Throttle → DbSession → Language)
- ☑ Admin guard middleware, callback registry, FSM
- ☑ User features: /start, subscription panel, config delivery (.conf/QR/link), traffic display, profile management
- ☑ Multiple concurrent subscriptions per user

### 1.4 Wallet & Payments
- ☑ `wallets` table — per-user balance in nanoton (1 TON = 1e9)
- ☑ `wallet_transactions` append-only ledger (DEPOSIT/PURCHASE/REFUND/REFERRAL/ADMIN_CREDIT/ADMIN_DEBIT)
- ☑ `WalletService` — atomic credit/debit with ledger row
- ☑ Wallet UI (balance, deposit with QR, paginated history)

### 1.5 TON Deposits
- ☑ `TonProvider` — TONAPI v2 client (TON/USD rate + incoming transfers)
- ☑ `TonDepositService.reconcile` — on-chain memo matching, idempotent credit
- ☑ `TonDepositWorker` — polls, credits, expires stale deposits, DMs users

### 1.6 Referral System
- ☑ Referral fields on `users` (referral_code, referred_by)
- ☑ Capture on `/start ref_<code>` deep link
- ☑ 10% commission on referred deposits, atomic credit
- ☑ Invite & Earn screen (code, link, count, earnings)

### 1.7 Promo Codes
- ☑ `promo_codes` table + `PromoService` (PERCENT/FIXED/DAYS)
- ☑ Wired into wallet-funded purchase flow (validation, discount, redemption)

### 1.8 Free Trial
- ☑ `TrialService.claim` — one-time trial (1 day, 100 MB)
- ☑ Trial button gated by runtime `trial_enabled` flag
- ☑ Admin global toggle + per-user reset

### 1.9 Subscription Management
- ☑ Lifecycle statuses: ACTIVE/EXPIRED/REVOKED/PENDING_RETRY/FAILED_FINAL
- ☑ `VPNProvisioningService` — server failover, circuit breaker integration
- ☑ `RetryProvisioningWorker` — backoff-gated retry, gives up at max_retries=5
- ☑ `SubscriptionExpiryWorker`, `TrafficSyncWorker`, `ExpiryReminderWorker`

### 1.10 Localization
- ☑ Catalogs for English, Arabic, Persian
- ☑ `t(lang, key)` with English fallback
- ☑ `LanguageMiddleware` + profile language switcher

### 1.11 Admin Panel
- ☑ Admin menu, user management, subscription management
- ☑ Server management with live connectivity test
- ☑ Transaction management, statistics dashboard
- ☑ Live server monitor (CPU, RAM, network, Xray state)
- ☑ Subscription-client management (lookup, enable/disable, reset, extend, remove, re-provision)
- ☑ Admin wallet management (lookup, credit/debit/refund, deposits, tx search)
- ☑ Runtime app settings (`app_settings` key/value table)

### 1.12 Production Hardening (Partial)
- ☑ docker-compose, Redis rate limiting, health endpoint
- ☑ Per-user throttle middleware, per-server circuit breakers
- ☐ Redis caching (10.3)
- ☐ Error tracking (10.5)

## Acceptance Criteria
- ☑ Users can register, deposit TON, purchase VPN subscriptions, receive configs
- ☑ Multiple subscriptions per user with independent lifecycle
- ☑ Referral commissions credited atomically with deposits
- ☑ Promo codes apply discounts at checkout
- ☑ Admin can manage all entities through Telegram bot
- ☑ Workers handle expiry, traffic sync, deposit polling, retry provisioning
- ☑ 11/11 automated tests pass (resilience layer)

---

# Milestone 2 — VPN Platform

**Objective**: Transform the 3x-ui fork into a provider-based VPN management
platform. Add a provider abstraction layer so the panel can manage subscriptions
across multiple backends (native Xray, BPB Worker, future providers) through a
single unified interface. Existing Xray functionality must continue without
modification.

**Status**: ☐ Not Started
**Complexity**: High
**Dependencies**: Milestone 1 (complete)

## 2.1 Fork Setup
- ☐ Verify `go build` produces working binary from v3.4.2 source
- ☐ Configure fork to run on port 2054 (testing alongside production on 2053)
- ☐ Commit clean baseline

**Complexity**: Low

## 2.2 Provider Interface

Define the core provider abstraction in `internal/provider/provider.go`:

- ☐ `ProviderType` enum: `NATIVE`, `EXTERNAL`, `STATIC`
- ☐ `Provider` interface with methods:
  - `Name() string`
  - `Type() ProviderType`
  - `CreateSubscription(ctx, req) (*Subscription, error)`
  - `RenewSubscription(ctx, id, req) error`
  - `SuspendSubscription(ctx, id) error`
  - `ResumeSubscription(ctx, id) error`
  - `DeleteSubscription(ctx, id) error`
  - `GetSubscription(ctx, id) (*Subscription, error)`
  - `GetSubscriptionURL(ctx, id) (string, error)`
  - `GetStatus(ctx, id) (*SubStatus, error)`
  - `SyncUsage(ctx, id) (*UsageReport, error)`
  - `IsHealthy(ctx) bool`
- ☐ Request/response structs: `CreateSubscriptionRequest`, `RenewRequest`, `Subscription`, `SubStatus`, `UsageReport`
- ☐ Error types: `ErrProviderUnavailable`, `ErrSubscriptionNotFound`, `ErrQuotaExceeded`

**Complexity**: Medium

## 2.3 Provider Registry

- ☐ `ProviderRegistry` — thread-safe map of provider name to `Provider` instance
- ☐ `Register(name, provider)`, `Get(name)`, `List()`, `Remove(name)`
- ☐ Startup registration from database `providers` table
- ☐ Auto-register the native Xray provider on initialization

**Complexity**: Medium

## 2.4 Native Xray Provider

Wrap existing client CRUD as a `Provider` implementation in `internal/provider/xray.go`:

- ☐ `XrayProvider` struct implementing `Provider` interface
- ☐ `CreateSubscription` delegates to existing `ClientService.Create`
- ☐ `RenewSubscription` updates client expiry/traffic via `ClientService.Update`
- ☐ `SuspendSubscription` disables client via `ClientService.Update` (enable=false)
- ☐ `ResumeSubscription` re-enables client
- ☐ `DeleteSubscription` delegates to `ClientService.Delete`
- ☐ `GetSubscription` reads from `ClientRecord` + `ClientTraffic`
- ☐ `GetSubscriptionURL` builds URL from sub base + subId
- ☐ `SyncUsage` reads `ClientTraffic` (already tracked by Xray stats)
- ☐ `IsHealthy` always true (local panel)
- ☐ Zero behavior change for existing clients (wrapper only)

**Complexity**: Medium

## 2.5 Database Extensions

New GORM models (existing tables untouched):

- ☐ `providers` table:
  - `id` (int, PK, auto-increment)
  - `name` (string, unique)
  - `type` (string: native/external/static)
  - `config` (JSON text, provider-specific settings)
  - `is_enabled` (bool, default true)
  - `created_at`, `updated_at` (auto-managed)
- ☐ `external_subscriptions` table:
  - `id` (string UUID, PK)
  - `provider_id` (int, FK to providers)
  - `client_email` (string, index)
  - `subscription_url` (text)
  - `plan_name` (string)
  - `status` (string: active/suspended/expired/deleted)
  - `traffic_limit_bytes` (int64)
  - `traffic_used_bytes` (int64)
  - `device_limit` (int)
  - `expires_at` (int64, unix timestamp)
  - `metadata` (JSON text, provider-specific data)
  - `notes` (text)
  - `created_at`, `updated_at` (auto-managed)
- ☐ `subscription_events` table:
  - `id` (int, PK, auto-increment)
  - `subscription_id` (string, index)
  - `provider_id` (int, FK to providers)
  - `event_type` (string: created/renewed/suspended/resumed/deleted/usage_synced/error)
  - `details` (JSON text)
  - `created_at` (auto-managed)
- ☐ Register new models in `database.initModels()`
- ☐ Seed default native Xray provider row on first startup

**Complexity**: Medium

## 2.6 API Extensions

New endpoints under `/panel/api/providers/`:

- ☐ `GET /panel/api/providers/list` — list all providers
- ☐ `GET /panel/api/providers/get/:id` — get provider details
- ☐ `POST /panel/api/providers/add` — register a new provider
- ☐ `POST /panel/api/providers/update/:id` — update provider config
- ☐ `POST /panel/api/providers/del/:id` — remove provider (soft)
- ☐ `POST /panel/api/providers/test/:id` — health check

All use existing 3x-ui auth (Bearer token / session cookie) and response
envelope `{success, msg, obj}`.

**Complexity**: Medium

## 2.7 Subscription API

New endpoints under `/panel/api/subscriptions/`:

- ☐ `POST /panel/api/subscriptions/create` — create subscription (provider-routed)
- ☐ `GET /panel/api/subscriptions/get/:id` — get subscription (any provider)
- ☐ `POST /panel/api/subscriptions/update/:id` — update subscription
- ☐ `POST /panel/api/subscriptions/renew/:id` — renew subscription
- ☐ `POST /panel/api/subscriptions/suspend/:id` — suspend subscription
- ☐ `POST /panel/api/subscriptions/resume/:id` — resume subscription
- ☐ `POST /panel/api/subscriptions/delete/:id` — delete subscription
- ☐ `GET /panel/api/subscriptions/url/:id` — get subscription URL
- ☐ `GET /panel/api/subscriptions/status/:id` — get status + usage
- ☐ `POST /panel/api/subscriptions/sync/:id` — sync usage from provider
- ☐ `GET /panel/api/subscriptions/list` — list subscriptions (filterable by provider)

**Complexity**: Medium

## Acceptance Criteria
- ☐ `go build` produces working binary
- ☐ Native Xray clients continue working exactly as before
- ☐ Provider interface defined and Xray provider wraps existing CRUD
- ☐ New tables created via GORM auto-migrate without breaking existing schema
- ☐ Provider CRUD API returns correct responses with existing auth
- ☐ Subscription API routes requests to the correct provider
- ☐ Default native Xray provider seeded on fresh install

---

# Milestone 3 — BPB Integration

**Objective**: Implement BPB Worker Panel as the first external provider. The
panel communicates with BPB's Cloudflare Worker API to create, manage, and
track external subscriptions. Subscriptions appear alongside native Xray
subscriptions through the unified API.

**Status**: ☐ Not Started
**Complexity**: High
**Dependencies**: Milestone 2

## 3.1 BPB Provider Implementation

`internal/provider/bpb.go`:

- ☐ `BPBProvider` struct implementing `Provider` interface
- ☐ `BPBConfig` — worker URL, auth token, default settings
- ☐ `CreateSubscription` — call BPB API to create subscription, store URL + metadata in `external_subscriptions`
- ☐ `RenewSubscription` — call BPB API to extend, update local record
- ☐ `SuspendSubscription` — call BPB API to disable + update local status
- ☐ `ResumeSubscription` — call BPB API to re-enable + update local status
- ☐ `DeleteSubscription` — call BPB API to revoke + soft-delete local record
- ☐ `GetSubscription` — read from `external_subscriptions` table
- ☐ `GetSubscriptionURL` — return stored URL from `external_subscriptions`
- ☐ `GetStatus` — read local status + optional BPB API poll
- ☐ `SyncUsage` — poll BPB API for traffic data where available; no-op fallback
- ☐ `IsHealthy` — HTTP ping to BPB worker endpoint

**Complexity**: High

## 3.2 BPB HTTP Client

- ☐ `internal/provider/bpb_client.go` — HTTP client for BPB Worker API
- ☐ Request/response types matching BPB API contract
- ☐ Retry with backoff for transient failures
- ☐ Timeout configuration
- ☐ Error mapping (BPB errors to provider error types)

**Complexity**: Medium

## 3.3 External Subscription Lifecycle

- ☐ Status transitions: created to active to suspended/expired/deleted
- ☐ Event logging to `subscription_events` on every state change
- ☐ Expiration handling (local timer, not dependent on BPB polling)
- ☐ Graceful degradation when BPB is unreachable

**Complexity**: Medium

## 3.4 Plan Assignment

- ☐ Map Abystral plan codes to BPB subscription parameters
- ☐ Traffic limit translation (bytes to BPB format)
- ☐ Duration mapping (days to BPB expiry format)
- ☐ Device/IP limit passthrough

**Complexity**: Low

## 3.5 Provider Synchronization

- ☐ Background sync worker — periodically poll BPB for usage/status updates
- ☐ Conflict resolution (local vs remote status divergence)
- ☐ Sync interval configuration per provider
- ☐ Sync event logging

**Complexity**: Medium

## 3.6 Admin Interface for BPB

- ☐ Add/configure BPB provider through API
- ☐ View BPB subscriptions alongside native ones
- ☐ Manual sync trigger
- ☐ BPB health status in provider list

**Complexity**: Low

## 3.7 Testing

- ☐ Unit tests for BPB provider (mocked HTTP)
- ☐ Integration test: create BPB subscription, verify in `external_subscriptions`
- ☐ Integration test: suspend/resume/delete lifecycle
- ☐ Integration test: sync worker updates usage
- ☐ Verify native Xray subscriptions unaffected

**Complexity**: Medium

## Acceptance Criteria
- ☐ BPB provider registered and healthy via `/providers/test/:id`
- ☐ Create subscription via unified API creates actual BPB subscription
- ☐ Subscription URL returned correctly for BPB subscriptions
- ☐ Suspend/resume/delete propagated to BPB
- ☐ Usage sync worker populates traffic_used_bytes
- ☐ Native Xray subscriptions continue working unchanged
- ☐ All lifecycle events logged to `subscription_events`

---

# Milestone 4 — Unified API

**Objective**: Ensure one consistent API surface for all operations regardless
of the backing provider. The Abystral bot communicates exclusively through
these endpoints and never contains provider-specific logic.

**Status**: ☐ Not Started
**Complexity**: Medium
**Dependencies**: Milestone 3

## 4.1 Abystral Client Updates

Changes in `Abystral/src/vpn/client.py`:

- ☐ `create_subscription(provider, plan, email, ...)` via `POST /subscriptions/create`
- ☐ `get_subscription(id)` via `GET /subscriptions/get/:id`
- ☐ `renew_subscription(id, ...)` via `POST /subscriptions/renew/:id`
- ☐ `suspend_subscription(id)` via `POST /subscriptions/suspend/:id`
- ☐ `resume_subscription(id)` via `POST /subscriptions/resume/:id`
- ☐ `delete_subscription(id)` via `POST /subscriptions/delete/:id`
- ☐ `get_subscription_url(id)` via `GET /subscriptions/url/:id`
- ☐ `get_subscription_status(id)` via `GET /subscriptions/status/:id`
- ☐ `sync_subscription_usage(id)` via `POST /subscriptions/sync/:id`
- ☐ `list_subscriptions(...)` via `GET /subscriptions/list`
- ☐ `list_providers()` via `GET /providers/list`

**Complexity**: Medium

## 4.2 Provisioning Service Update

Changes in `Abystral/src/services/vpn_provisioning.py`:

- ☐ `_provision_on_server` calls `create_subscription(provider=...)` instead of `add_client()`
- ☐ Provider selection logic: read provider from server config or use default
- ☐ Subscription ID returned from unified API stored in `subscriptions.inbound_uuid`
- ☐ Backward-compatible: existing subscriptions keep working

**Complexity**: Medium

## 4.3 Config Service Update

Changes in `Abystral/src/services/vpn_config.py`:

- ☐ `get_subscription_url` calls unified API instead of manual URL construction
- ☐ `get_config` works for both native and external subscriptions
- ☐ Provider-agnostic URL resolution

**Complexity**: Low

## 4.4 Legacy Compatibility Layer

- ☐ Existing `/panel/api/clients/*` endpoints remain functional
- ☐ Existing `/panel/api/inbounds/*` endpoints remain functional
- ☐ Native Xray operations continue through both old and new API paths
- ☐ No breaking changes to existing Abystral handlers

**Complexity**: Low

## 4.5 API Documentation

Complete endpoint reference:

### Provider Endpoints
| Method | Path | Description |
|--------|------|-------------|
| GET | `/panel/api/providers/list` | List all registered providers |
| GET | `/panel/api/providers/get/:id` | Get provider details |
| POST | `/panel/api/providers/add` | Register new provider |
| POST | `/panel/api/providers/update/:id` | Update provider config |
| POST | `/panel/api/providers/del/:id` | Remove provider |
| POST | `/panel/api/providers/test/:id` | Health check provider |

### Subscription Endpoints
| Method | Path | Description |
|--------|------|-------------|
| POST | `/panel/api/subscriptions/create` | Create subscription |
| GET | `/panel/api/subscriptions/get/:id` | Get subscription |
| POST | `/panel/api/subscriptions/update/:id` | Update subscription |
| POST | `/panel/api/subscriptions/renew/:id` | Renew subscription |
| POST | `/panel/api/subscriptions/suspend/:id` | Suspend subscription |
| POST | `/panel/api/subscriptions/resume/:id` | Resume subscription |
| POST | `/panel/api/subscriptions/delete/:id` | Delete subscription |
| GET | `/panel/api/subscriptions/url/:id` | Get subscription URL |
| GET | `/panel/api/subscriptions/status/:id` | Get status + usage |
| POST | `/panel/api/subscriptions/sync/:id` | Sync usage from provider |
| GET | `/panel/api/subscriptions/list` | List (filterable) |

### Existing Endpoints (unchanged)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/panel/api/clients/list` | List all Xray clients |
| POST | `/panel/api/clients/add` | Create Xray client |
| POST | `/panel/api/clients/update/:email` | Update Xray client |
| POST | `/panel/api/clients/del/:email` | Delete Xray client |
| GET | `/panel/api/clients/subLinks/:subId` | Get subscription links |
| GET | `/panel/api/inbounds/list` | List inbounds |
| GET | `/panel/api/server/status` | Server status |

**Complexity**: Low (documentation)

## Acceptance Criteria
- ☐ Abystral bot provisions subscriptions through unified API only
- ☐ Bot code contains zero provider-specific logic
- ☐ Creating a native Xray subscription via unified API produces identical result to direct client creation
- ☐ Creating a BPB subscription via unified API works end-to-end
- ☐ Existing Abystral handlers (purchase, wallet, admin, referral, trial, promo) unchanged
- ☐ All existing functionality preserved (backward compatibility verified)

---

# Milestone 5 — Admin Experience

**Objective**: Enhance the admin panel for multi-provider management. Admins can
manage providers, view unified subscriptions, configure plans per provider, and
monitor the health of the entire platform.

**Status**: ☐ Not Started
**Complexity**: Medium
**Dependencies**: Milestone 4

## 5.1 Provider Management (Telegram Bot)

- ☐ List providers with status indicators
- ☐ Add provider (FSM flow: type, name, config, test, save)
- ☐ Edit provider configuration
- ☐ Enable/disable provider
- ☐ Remove provider (with subscription migration prompt)
- ☐ Health check on demand

**Complexity**: Medium

## 5.2 Unified Subscription Management

- ☐ List subscriptions across all providers with provider badge
- ☐ Filter by provider, status, user
- ☐ View subscription detail (provider-specific metadata included)
- ☐ Suspend/resume/delete from unified view
- ☐ Manual usage sync trigger
- ☐ Re-provision failed subscriptions

**Complexity**: Medium

## 5.3 Plan Management

- ☐ Assign plans to specific providers or make provider-agnostic
- ☐ Per-provider plan parameter overrides (traffic format, limits)
- ☐ Plan availability per provider

**Complexity**: Low

## 5.4 Traffic & IP Limits

- ☐ Display traffic usage from all providers in unified format
- ☐ IP limit display and enforcement status per provider
- ☐ Traffic reset across providers

**Complexity**: Low

## 5.5 Expiration Management

- ☐ View upcoming expirations across all providers
- ☐ Bulk extend subscriptions
- ☐ Provider-aware expiration handling

**Complexity**: Low

## 5.6 Analytics

- ☐ Subscription count by provider
- ☐ Active/suspended/expired breakdown per provider
- ☐ Total traffic usage by provider
- ☐ Revenue per provider (from transaction data)
- ☐ Provider health history

**Complexity**: Medium

## 5.7 Logs & Events

- ☐ Subscription event log viewer (from `subscription_events`)
- ☐ Filter by provider, event type, subscription
- ☐ Provider sync logs
- ☐ Error aggregation per provider

**Complexity**: Low

## 5.8 Search

- ☐ Search subscriptions by email, UUID, user ID across all providers
- ☐ Search by subscription URL (external providers)
- ☐ Cross-provider user lookup

**Complexity**: Low

## 5.9 Bulk Operations

- ☐ Bulk suspend/resume subscriptions by provider
- ☐ Bulk traffic reset by provider
- ☐ Bulk extend by provider
- ☐ Provider migration (move subscriptions between providers)

**Complexity**: Medium

## Acceptance Criteria
- ☐ Admin can add, configure, and test a BPB provider through Telegram
- ☐ Subscription list shows both Xray and BPB subscriptions with badges
- ☐ All CRUD operations work from the unified admin interface
- ☐ Analytics show per-provider breakdown
- ☐ Event logs viewable and filterable
- ☐ Bulk operations execute across provider boundaries

---

# Milestone 6 — User Experience

**Objective**: Improve the Telegram bot user experience for the multi-provider
platform. Users interact with subscriptions seamlessly regardless of the
backing provider.

**Status**: ☐ Not Started
**Complexity**: Medium
**Dependencies**: Milestone 4

## 6.1 Menu Improvements

- ☐ Provider-aware subscription display (icon/badge per provider type)
- ☐ Grouped subscription list by status (active, expiring, expired)
- ☐ Quick actions from subscription list (renew, config, share)
- ☐ Pagination for large subscription lists

**Complexity**: Low

## 6.2 Purchase Flow

- ☐ Plan selection shows available providers (if multiple serve a plan)
- ☐ Provider auto-selection (admin-configured default or round-robin)
- ☐ Clear confirmation with provider info before purchase
- ☐ Promo code + provider compatibility check

**Complexity**: Medium

## 6.3 Wallet UX

- ☐ Balance display with pending deposits indicator
- ☐ Deposit instructions with copy-to-clipboard memo
- ☐ Transaction history with provider context
- ☐ Refund display for failed provisioning

**Complexity**: Low

## 6.4 Subscription List

- ☐ Active subscriptions with days remaining, traffic used
- ☐ Provider type indicator (Xray, BPB, etc.)
- ☐ One-tap renewal for expiring subscriptions
- ☐ Status badges (active, expiring soon, suspended, expired)

**Complexity**: Low

## 6.5 QR & Config Delivery

- ☐ QR code generation for subscription URLs (all providers)
- ☐ Deep link for Xray/V2Ray/Clash apps
- ☐ Config format selection (link, QR, file) per provider capability
- ☐ Subscription URL copy button

**Complexity**: Low

## 6.6 Subscription Sharing

- ☐ Generate shareable invite link
- ☐ Referral deep link with subscription context
- ☐ Share button with formatted message

**Complexity**: Low

## 6.7 Localization Improvements

- ☐ Provider-specific strings (BPB, Xray labels) in all 3 languages
- ☐ Dynamic plan names localization
- ☐ Error messages for provider-specific failures
- ☐ Date/time formatting per locale

**Complexity**: Low

## Acceptance Criteria
- ☐ Users purchase and receive subscriptions without knowing the provider
- ☐ Subscription list clearly shows status and provider for each subscription
- ☐ QR codes and config links work for both Xray and BPB subscriptions
- ☐ All new strings localized in English, Arabic, Persian
- ☐ Purchase flow handles provider selection transparently

---

# Milestone 7 — Future Providers

**Objective**: Design and document the integration framework for future VPN
providers. No implementations, only interface contracts, documentation, and
placeholder registrations.

**Status**: ☐ Not Started
**Complexity**: Low
**Dependencies**: Milestone 3 (BPB validates the pattern)

## 7.1 Provider Integration Guide

- ☐ Document the `Provider` interface contract
- ☐ Step-by-step guide for implementing a new provider
- ☐ Config schema documentation
- ☐ Error handling requirements
- ☐ Testing requirements for new providers

**Complexity**: Low

## 7.2 Provider Placeholders

Document integration approach for each (no code):

- ☐ **MoaV** — Cloudflare Worker-based, similar to BPB; document API differences
- ☐ **WireGuard** — Native kernel module; document wg-quick integration path
- ☐ **Hiddify** — Full panel with REST API; document DriverABS compatibility
- ☐ **Outline** — Jigsaw's Shadowsocks server; document Access Key API
- ☐ **OpenVPN** — Certificate-based; document OpenVPN Access Server API
- ☐ **Shadowsocks** — Standalone ss-server; document multi-user management
- ☐ **Generic VLESS** — Any VLESS-compatible server; document share-link format

**Complexity**: Low

## 7.3 Provider Config Schemas

- ☐ Define JSON config schema for each provider type
- ☐ Validation rules per provider
- ☐ Required vs optional fields
- ☐ Sensitive field handling (tokens, keys)

**Complexity**: Low

## Acceptance Criteria
- ☐ Integration guide complete and accurate (validated against BPB implementation)
- ☐ Each future provider has documented integration approach
- ☐ Config schemas defined for all planned providers
- ☐ A developer can read the guide and implement a new provider without ambiguity

---

# Milestone 8 — Production

**Objective**: Harden the platform for production deployment. Security audit,
performance optimization, monitoring, backup strategy, and deployment automation.

**Status**: ☐ Not Started
**Complexity**: High
**Dependencies**: Milestone 4

## 8.1 Security Audit

- ☐ API authentication review (Bearer tokens, session handling)
- ☐ Provider config encryption at rest (API tokens, secrets)
- ☐ Input validation on all new endpoints
- ☐ SQL injection prevention (GORM parameterized queries)
- ☐ Rate limiting on subscription creation
- ☐ CSRF protection on mutation endpoints (existing middleware)
- ☐ Audit provider HTTP client for SSRF
- ☐ Secrets rotation procedure for provider API tokens

**Complexity**: High

## 8.2 Performance Optimization

- ☐ Database indexing on `external_subscriptions` (provider_id, status, client_email, expires_at)
- ☐ Database indexing on `subscription_events` (subscription_id, created_at)
- ☐ Provider health check caching (avoid per-request pings)
- ☐ Subscription list pagination (offset/limit)
- ☐ Connection pooling for external provider HTTP clients
- ☐ Redis caching for provider list and subscription status (Abystral side)

**Complexity**: Medium

## 8.3 Backup Strategy

- ☐ SQLite database backup automation (existing x-ui backup + new tables)
- ☐ Provider config backup (encrypted)
- ☐ External subscription state export/import
- ☐ Backup to Telegram bot (existing `BackuptoTgbot` extended)

**Complexity**: Low

## 8.4 Monitoring

- ☐ Provider health dashboard endpoint
- ☐ Subscription count metrics by provider
- ☐ Sync worker success/failure metrics
- ☐ Error rate per provider
- ☐ Alerting on provider health degradation (Telegram bot notification)

**Complexity**: Medium

## 8.5 Logging

- ☐ Structured logging for provider operations
- ☐ Request/response logging for external provider API calls (redacted)
- ☐ Subscription lifecycle audit trail
- ☐ Log rotation configuration

**Complexity**: Low

## 8.6 CI/CD

- ☐ Go build + test pipeline for 3x-ui fork
- ☐ Python test pipeline for Abystral
- ☐ Binary artifact build (multi-arch: amd64, arm64)
- ☐ Docker image build and push
- ☐ Version tagging convention

**Complexity**: Medium

## 8.7 Docker Improvements

- ☐ Multi-stage Dockerfile for the fork (build to runtime)
- ☐ docker-compose with fork + Abystral + PostgreSQL + Redis
- ☐ Environment variable documentation
- ☐ Volume mappings for database and config persistence
- ☐ Health check probes in compose

**Complexity**: Medium

## 8.8 Deployment Guide

- ☐ Fresh installation procedure
- ☐ Migration from stock 3x-ui to fork
- ☐ Provider setup walkthrough (BPB)
- ☐ Abystral bot configuration for unified API
- ☐ Rollback procedure

**Complexity**: Low

## Acceptance Criteria
- ☐ No critical or high-severity security findings
- ☐ Subscription list loads in <200ms with 10k records
- ☐ Automated backup runs daily
- ☐ Provider outage triggers admin notification within 60s
- ☐ CI pipeline passes on every commit
- ☐ Deployment guide tested on fresh server

---

# Milestone 9 — Nice-to-Have

**Objective**: Future enhancements and ideas. No timeline, no commitments.
Tracked here for reference and prioritization.

**Status**: ☐ Not Started
**Complexity**: Varies
**Dependencies**: Milestone 8

## 9.1 Smart Routing
- ☐ Route users to optimal provider based on location, latency, load
- ☐ Provider weight/priority configuration
- ☐ Geo-aware provider selection

## 9.2 Auto Failover
- ☐ Detect provider failure, auto-migrate active subscriptions to healthy provider
- ☐ Configurable failover policy (immediate, after N failures, manual)
- ☐ User notification on failover

## 9.3 Multi-Region Balancing
- ☐ Provider tagging by region
- ☐ User region detection (from Telegram or IP)
- ☐ Region-aware plan catalog

## 9.4 Provider Health Monitoring
- ☐ Continuous health probing with configurable intervals
- ☐ Health history and uptime calculation
- ☐ Degraded state detection (slow but alive)

## 9.5 Automatic Provider Selection
- ☐ ML-based provider scoring (latency, packet loss, throughput)
- ☐ A/B testing between providers
- ☐ User preference learning

## 9.6 Usage Analytics
- ☐ Per-user usage trends over time
- ☐ Peak usage detection
- ☐ Bandwidth forecasting
- ☐ Cost-per-GB analysis by provider

## 9.7 User Dashboard (Web)
- ☐ Web-based dashboard for users (subscription management, usage)
- ☐ OAuth login via Telegram
- ☐ Real-time usage graphs
- ☐ Multi-device config delivery

## 9.8 Mobile App API
- ☐ REST API for native mobile apps
- ☐ Push notification support
- ☐ In-app config import
- ☐ Subscription status polling

## 9.9 Reseller System
- ☐ Reseller accounts with margin management
- ☐ White-label subscription delivery
- ☐ Per-reseller plan pricing
- ☐ Commission tracking and payout

## 9.10 Public REST API
- ☐ API key management for third-party integrations
- ☐ Rate-limited public endpoints
- ☐ Webhook support for subscription events
- ☐ OpenAPI 3.0 documentation

## 9.11 Web Dashboard (Admin)
- ☐ Full-featured web admin replacing/supplementing Telegram admin
- ☐ Real-time metrics and charts
- ☐ User/subscription search and management
- ☐ Provider configuration UI

## Acceptance Criteria
- ☐ Each feature is self-contained and does not break existing functionality
- ☐ Features are prioritized based on user demand and business value

---

# Status Summary

**Current Overall Progress**: ~11%
**Current Milestone**: 2 — VPN Platform
**Next Milestone**: 3 — BPB Integration

## Known Issues
- Abystral Milestone 10 incomplete: Redis caching (10.3) and error tracking (10.5) outstanding
- No automated tests for wallet, referral, TON, purchase, trial, promo, provisioning flows
- `users.referral_rewarded` column exists but unused (leftover from first-purchase design)
- `TrialService` docstring says "7 days" but effective constant is 1 day

## Technical Debt
- Static `domain/plans.py` removed but some references may remain
- Abystral `src/vpn/client.py` will need unified API methods added (Milestone 4)
- 3x-ui fork build not yet verified from source
- Provider config contains sensitive data (API tokens) stored as plaintext JSON in SQLite

## Long-term Vision

Abystral evolves from a single-panel VPN bot into a **multi-provider VPN
platform** where:

1. **Providers are pluggable** — adding a new VPN backend requires implementing
   one Go interface, no changes to the Telegram bot or existing providers.

2. **The bot is provider-agnostic** — users purchase, renew, and manage
   subscriptions through a single unified experience regardless of the backing
   technology (Xray, BPB, WireGuard, etc.).

3. **Admins manage everything from one place** — providers, subscriptions, users,
   plans, and analytics through the Telegram bot and eventually a web dashboard.

4. **Scale is horizontal** — new providers bring new capacity without changing
   the platform. Smart routing and auto-failover distribute load across
   providers for reliability and performance.

5. **The platform is self-sustaining** — wallet-based payments, referral
   commissions, and reseller margins create a self-growing ecosystem.
