# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Reference

```bash
# Development (hot reload - no restart needed)
make air              # Backend with hot reload
cd app && pnpm dev    # Frontend dev server

# Code generation (CRITICAL: NEVER run goctl directly)
make gen              # Regenerate handlers/types from .api file

# Testing
go test -v ./internal/logic/...                        # All Go tests
go test -v -run TestName ./internal/logic/auth/        # Single test
cd app && pnpm check                                   # TypeScript check
cd app && pnpm test:unit                               # Frontend tests

# Database (standalone mode)
make migrate-up       # Run pending migrations
make migrate-down     # Rollback last migration
make migrate-status   # Check migration status

# Before committing
make build && cd app && pnpm build
```

## Architecture

**Dual-mode system** - switches between standalone (SQLite + Stripe) and Levee (managed platform):

```
gobot.api                   → API definition (routes, types) - EDIT HERE to add endpoints
├── internal/handler/        → AUTO-GENERATED from .api (DO NOT EDIT)
├── internal/types/          → AUTO-GENERATED from .api (DO NOT EDIT)
├── internal/logic/          → Business logic - EDIT HERE for implementation
├── internal/svc/            → ServiceContext with UseLocal()/UseLevee() mode check
├── internal/db/             → SQLite (standalone mode)
└── internal/local/          → Local auth/billing services (standalone mode)

app/src/
├── routes/(www)/            → Marketing pages (public)
├── routes/(auth)/           → Auth pages (login, register)
├── routes/(app)/            → App pages (authenticated)
├── lib/api/                 → AUTO-GENERATED TypeScript client
├── lib/config/site.ts       → Branding/SEO (single source of truth)
└── lib/stores/              → Svelte stores (auth, subscription)
```

## Adding API Endpoints

1. Define in `gobot.api`:
```
@server(prefix: /api/v1, jwt: Auth)
service gobot {
    @handler GetWidget
    get /widgets/:id (GetWidgetRequest) returns (GetWidgetResponse)
}

type GetWidgetRequest { Id string `path:"id"` }
type GetWidgetResponse { Name string `json:"name"` }
```

2. Run `make gen`

3. Implement in `internal/logic/getwidgetlogic.go`:
```go
func (l *GetWidgetLogic) GetWidget(req *types.GetWidgetRequest) (*types.GetWidgetResponse, error) {
    if l.svcCtx.UseLocal() {
        // SQLite implementation
    }
    // Levee implementation
}
```

4. Frontend types auto-available: `import { getWidget } from '$lib/api'`

## Mode-Aware Logic Pattern

All handlers must support both modes:

```go
func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
    if l.svcCtx.UseLocal() {
        return l.loginLocal(req)  // SQLite + local JWT
    }
    if l.svcCtx.Levee == nil {
        return nil, fmt.Errorf("auth service not configured")
    }
    // Levee SDK implementation
}
```

Key methods: `l.svcCtx.UseLocal()`, `l.svcCtx.UseLevee()`, `l.svcCtx.DB`, `l.svcCtx.Auth`, `l.svcCtx.Billing`, `l.svcCtx.Levee`

## Critical Rules

- **NEVER run goctl directly** - Use `make gen`
- **pnpm only** - Never npm or yarn
- **Styles in app.css only** - No inline styles or `<style>` blocks
- **Svelte 5 runes** - `$state`, `$derived`, `$props`, `$effect` (not Svelte 4 `export let`, `$:`, `<slot>`)
- **DaisyUI components** - Use DaisyUI classes for UI components (btn, card, modal, etc.)
- **Idiomatic Go** - One function with parameters, not multiple variations
- **Minimal changes** - Never remove code that appears unused without asking
- **Support both modes** - Logic handlers must work with UseLocal() and UseLevee()

## Configuration

| File | Purpose |
|------|---------|
| `app/src/lib/config/site.ts` | Branding, SEO, social links |
| `etc/gobot.yaml` | Products, pricing, backend settings |
| `.env` | Secrets only (API keys, JWT secret) |

## /init Flow

When user runs `/init`, follow the interactive setup in `AI.md`:
1. Environment setup (install.sh)
2. Business discovery - ask what they want to build
3. Research (optional) - run 14-step validation from `RESEARCH-PLAN.md`, outputs to `./plan/`
4. Auto-customize - update site.ts, landing page, theme (app.css), pricing (gobot.yaml)
5. Verify and launch - `make build && cd app && pnpm build`

## Admin Backoffice

Access admin dashboard at `/admin` (requires `ADMIN_USERNAME` and `ADMIN_PASSWORD` from .env).
Admin API routes use JWT + basic auth middleware (`internal/middleware/adminauth.go`).
