# AI Development Instructions

This document provides instructions for AI coding assistants working on this codebase. Follow these rules exactly.

---

## /init - Interactive Project Setup

When the user runs `/init`, execute this complete setup flow.

### Phase 1: Environment Setup

1. **Run the install script** (if not already done):
   ```bash
   ./install.sh <project-name>
   ```
   This handles: project renaming, admin credentials, port configuration, dependency installation.

2. **Verify the setup**:
   - Check `.env` exists with required variables
   - Verify `make air` can start the backend
   - Verify `cd app && pnpm dev` can start the frontend

### Phase 2: Discover the Business

Ask the user ONE of these questions (use AskUserQuestion tool):

**"What would you like to build?"**

Options:
- **A) I have a domain** → "I own the domain [XYZ.com]"
- **B) I have expertise** → "My expertise is in [industry/field]"
- **C) I have a specific idea** → "I want to build [specific product]"
- **D) Skip research** → Proceed directly to manual configuration

### Phase 3: Research (Optional but Recommended)

If the user chose A, B, or C, offer to run the research process:

**"Would you like me to run the startup validation research? This takes 5-10 minutes and generates a comprehensive market analysis, competitor insights, and landing page copy."**

If YES:
1. Read `RESEARCH-PLAN.md` for the complete 14-step research protocol
2. Create `./plan/` directory
3. Execute all 14 prompts sequentially, writing output to `./plan/01-*.md` through `./plan/14-*.md`
4. Check for STOP conditions at steps 2, 3, 8, and 9 (see RESEARCH-PLAN.md for details)
5. On completion, write `./plan/00-EXECUTION-COMPLETE.md`

Key research outputs for customization:
- `./plan/09-big-idea.md` - Core positioning and belief shift
- `./plan/14-landing-page-copy.md` - Complete landing page sections
- `./plan/04-audience-discovery.md` - Target customer profile
- `./plan/07-positioning.md` - Unique differentiator

### Phase 4: Auto-Customize the Product

After research completes (or if skipped, ask these questions manually):

#### 4.1 Configure site.ts

Update `app/src/lib/config/site.ts` with:
- `name` - Product name (from research or ask user)
- `tagline` - Short value proposition (from `./plan/09-big-idea.md`)
- `description` - SEO description (from research summary)
- `url` - Production URL (ask user)
- `supportEmail` - Support email (ask user)
- `twitter` - Twitter handle (ask user, optional)
- `social.*` - Social links (ask user, optional)
- `legal.companyName` - Company name (ask user)

#### 4.2 Update the Landing Page

Edit `app/src/routes/(www)/+page.svelte` using content from `./plan/14-landing-page-copy.md`:

1. **Hero Section**: Update headline, subheadline, and CTA from research
2. **Features**: Map the top 3 benefits to feature cards with appropriate icons
3. **How It Works**: Adapt the steps to match the product flow
4. **What's Included**: List actual product features
5. **Testimonial**: Replace with placeholder or real testimonial
6. **FAQ**: Update with product-specific questions from `./plan/14-landing-page-copy.md`
7. **CTA**: Match the final call-to-action from research

#### 4.3 Theme the Design

Edit `app/src/app.css` @theme section:

Ask the user: **"What's your brand's primary color?"** (provide color picker or hex input)

Update these CSS custom properties:
```css
--color-primary: #[user-choice];      /* Main brand color */
--color-primary-light: [lighter];      /* Auto-calculate */
--color-primary-dark: [darker];        /* Auto-calculate */
--color-secondary: [complementary];    /* Suggest based on primary */
```

Optionally ask about:
- Font preference (keep defaults or suggest alternatives)
- Dark/light mode preference

#### 4.4 Update README.md

Replace the boilerplate README with product-specific content:
- Product name and description
- Quick start instructions
- Features list
- Tech stack (keep existing)
- License

#### 4.5 Configure Pricing (etc/gobot.yaml)

Ask the user:
1. **"What pricing tiers do you want?"** (Free, Pro, Enterprise patterns)
2. **"What's the Pro plan price?"** (monthly amount)
3. **"Offer annual billing?"** (yes/no, calculate 17% discount)
4. **"Offer a free trial?"** (yes/no, how many days)

Update `etc/<appname>.yaml` Products section accordingly.

### Phase 5: Verify and Launch

1. **Build check**:
   ```bash
   make build
   cd app && pnpm check && pnpm build
   ```

2. **Start development**:
   ```bash
   make dev
   ```

3. **Show the user**:
   - Frontend URL (from .env port)
   - Backend URL (from .env port)
   - Admin backoffice URL
   - Next steps (Stripe setup, deployment)

### Quick Reference: Files to Customize

| File | What to Change |
|------|----------------|
| `app/src/lib/config/site.ts` | Branding, SEO, social links |
| `app/src/routes/(www)/+page.svelte` | Landing page content |
| `app/src/app.css` | Colors, fonts, theme |
| `app/static/favicon.svg` | Site favicon (replace with your logo) |
| `etc/<appname>.yaml` | Products, pricing, features |
| `README.md` | Project documentation |
| `.env` | API keys, secrets (already done by install.sh) |

---

## Project Overview

**Gobot** is a full-stack SaaS boilerplate with:
- **Backend**: Go-Zero framework (Go 1.25+)
- **Frontend**: SvelteKit 2 + Svelte 5 + Tailwind v4
- **Database**: SQLite (standalone) or Levee (managed platform)
- **Payments**: Stripe

The frontend is statically built and embedded into a single Go binary.

## Architecture

```
├── gobot.api                 # API definition (routes, request/response types)
├── gobot.go                  # Entry point
├── etc/gobot.yaml            # Backend config + Products/Pricing
├── internal/
│   ├── handler/               # AUTO-GENERATED - DO NOT EDIT
│   ├── types/                 # AUTO-GENERATED - DO NOT EDIT
│   ├── logic/                 # Business logic - EDIT HERE
│   ├── svc/                   # Service context
│   ├── db/                    # SQLite (standalone mode)
│   ├── local/                 # Local auth/billing (standalone mode)
│   ├── config/                # Config structs
│   └── middleware/            # CORS, JWT, rate limiting
├── app/                       # SvelteKit frontend
│   ├── src/
│   │   ├── routes/            # File-based routing
│   │   │   ├── (www)/         # Marketing pages (public)
│   │   │   ├── (auth)/        # Auth pages (login, register)
│   │   │   ├── (app)/         # App pages (authenticated)
│   │   │   └── (minimal)/     # Minimal layout pages
│   │   └── lib/
│   │       ├── config/site.ts # Branding/SEO - single source of truth
│   │       ├── api/           # AUTO-GENERATED TypeScript client
│   │       ├── stores/        # Svelte stores (auth, subscription)
│   │       ├── components/    # UI components
│   │       └── utils/         # Utilities (seo.ts, etc.)
│   └── static/                # Static assets (favicon.svg, images)
└── docs/                      # Documentation
```

## Critical Rules

### 1. Code Generation

**NEVER run `goctl` commands directly.** Always use:
```bash
make gen
```

This generates:
- Go handlers in `internal/handler/`
- Go types in `internal/types/`
- TypeScript API client in `app/src/lib/api/`

### 2. Hot Reloading

**Do NOT restart the server.** We use `air` for hot reloading:
```bash
make air
```

### 3. Package Manager

**Use pnpm only.** Never use npm or yarn:
```bash
cd app && pnpm install
cd app && pnpm dev
cd app && pnpm build
```

### 4. Styling

**All styles go in `app/src/app.css` or component-specific CSS files.**
- NEVER use inline styles
- NEVER use `<style>` blocks in Svelte files
- Use Tailwind v4 utility classes

### 5. Svelte 5 Syntax

Use Svelte 5 runes, NOT Svelte 4 syntax:

```svelte
<!-- CORRECT - Svelte 5 -->
<script lang="ts">
  let count = $state(0);
  let doubled = $derived(count * 2);
  let { name, age }: { name: string; age: number } = $props();

  $effect(() => {
    console.log('count changed:', count);
  });
</script>

<!-- WRONG - Svelte 4 -->
<script lang="ts">
  export let name;  // DON'T USE
  let count = 0;    // DON'T USE for reactive state
  $: doubled = count * 2;  // DON'T USE
</script>
```

For slots, use snippets:
```svelte
<!-- CORRECT - Svelte 5 -->
<script lang="ts">
  import type { Snippet } from 'svelte';
  let { children }: { children: Snippet } = $props();
</script>
{@render children()}

<!-- WRONG - Svelte 4 -->
<slot />
```

### 6. Go Code Style

**Idiomatic Go only.** One function with parameters, not multiple variations:

```go
// CORRECT
func Register(token string, opts ...Option) error

// WRONG - Don't create multiple functions
func Register() error
func RegisterWithToken(token string) error
```

### 7. Dual-Mode Support

All logic handlers must support both standalone and Levee modes:

```go
func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
    // Check mode first
    if l.svcCtx.UseLocal() {
        return l.loginLocal(req)
    }

    // Levee mode
    if l.svcCtx.Levee == nil {
        return nil, fmt.Errorf("auth service not configured")
    }
    // ... Levee implementation
}
```

Key methods:
- `l.svcCtx.UseLocal()` - Returns true for standalone mode (SQLite + Stripe)
- `l.svcCtx.UseLevee()` - Returns true for Levee mode
- `l.svcCtx.DB` - SQLite database (nil in Levee mode)
- `l.svcCtx.Levee` - Levee SDK client (nil in standalone mode)

### 8. Minimal Changes

- **NEVER remove code that appears unused** - It may be called from frontend or other services
- **NEVER add features beyond what's requested** - No over-engineering
- **NEVER add comments to code you didn't change**
- **Ask before deleting** - If something seems unused, ask first

### 9. Build Before Push

Always verify before committing:
```bash
make build
cd app && pnpm check && pnpm build
```

## Configuration System

Three config files, each with a specific purpose:

| File | Purpose | Contains |
|------|---------|----------|
| `app/src/lib/config/site.ts` | Frontend branding/SEO | Site name, tagline, social links, legal info |
| `etc/gobot.yaml` | Backend config + pricing | Products, prices, features, backend settings |
| `.env` | Secrets only | API keys, JWT secret, database credentials |

### Branding Changes

Edit `app/src/lib/config/site.ts`:
```typescript
export const site = {
  name: 'YourApp',
  tagline: 'Your tagline',
  url: 'https://yourapp.com',
  // ...
}
```

### Pricing Changes

Edit `etc/gobot.yaml`:
```yaml
Products:
  - slug: "pro"
    name: "Pro"
    prices:
      - slug: "monthly"
        amount: 2900  # cents
        interval: "month"
```

Pricing is read from YAML at build time and embedded in static HTML.

## Adding New Endpoints

1. Define in `gobot.api`:
```
@server(
    prefix: /api/v1
    middleware: JwtAuth
)
service gobot {
    @handler GetWidget
    get /widgets/:id (GetWidgetRequest) returns (GetWidgetResponse)
}

type GetWidgetRequest {
    Id string `path:"id"`
}

type GetWidgetResponse {
    Id   string `json:"id"`
    Name string `json:"name"`
}
```

2. Run `make gen`

3. Implement in `internal/logic/getwidgetlogic.go`

4. Frontend types auto-available in `$lib/api`

## File Locations

| Task | Location |
|------|----------|
| Add API endpoint | `gobot.api` → `make gen` → `internal/logic/` |
| Add page | `app/src/routes/` |
| Add component | `app/src/lib/components/` |
| Add store | `app/src/lib/stores/` |
| Change branding | `app/src/lib/config/site.ts` |
| Change favicon | `app/static/favicon.svg` |
| Change pricing | `etc/gobot.yaml` |
| Add middleware | `internal/middleware/` |
| Database migrations | `internal/db/` |

## Route Groups

| Group | Purpose | Layout |
|-------|---------|--------|
| `(www)` | Marketing pages | Full header/footer |
| `(auth)` | Login, register, reset password | Minimal, centered |
| `(app)` | Authenticated app pages | App shell with sidebar |
| `(minimal)` | Email confirmations, errors | Minimal layout |

## Stores

| Store | Purpose |
|-------|---------|
| `auth` | Authentication state, login/logout functions |
| `subscription` | Subscription state, checkout functions |
| `currentUser` | Derived from auth store |
| `isAuthenticated` | Derived boolean |

## Environment Variables

### Required (Both Modes)
```bash
ACCESS_SECRET=xxx        # JWT secret (openssl rand -hex 32)
APP_BASE_URL=xxx         # Frontend URL
```

### Standalone Mode
```bash
LEVEE_ENABLED=false      # or unset
SQLITE_PATH=./data/gobot.db
STRIPE_SECRET_KEY=sk_xxx
STRIPE_PUBLISHABLE_KEY=pk_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
```

### Levee Mode
```bash
LEVEE_ENABLED=true
LEVEE_API_KEY=lvk_xxx
LEVEE_BASE_URL=https://api.levee.sh
```

## Testing

```bash
# Go tests
go test -v ./internal/logic/...
go test -v -run TestFunctionName ./internal/logic/auth/

# Frontend tests
cd app && pnpm test:unit
cd app && pnpm test:unit -- MyTest
```

## Common Patterns

### API Client Usage (Frontend)
```typescript
import { login, getMe, createCheckout } from '$lib/api';

// Auth
const { token } = await login({ email, password });

// User
const user = await getMe();

// Subscription
const { checkoutUrl } = await createCheckout({ planName: 'pro' });
```

### Protected Routes
```svelte
<script lang="ts">
  import { isAuthenticated } from '$lib/stores/auth';
  import { goto } from '$app/navigation';

  $effect(() => {
    if (!$isAuthenticated) {
      goto('/auth/login');
    }
  });
</script>
```

### Form Handling
```svelte
<script lang="ts">
  let loading = $state(false);
  let error = $state('');

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    loading = true;
    error = '';
    try {
      await someApiCall();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }
</script>
```

## Do NOT

- Run `goctl` directly (use `make gen`)
- Use npm or yarn (use `pnpm`)
- Use Svelte 4 syntax (`export let`, `$:`, `<slot>`)
- Add inline styles or `<style>` blocks
- Create multiple function variations in Go
- Remove code without asking
- Over-engineer solutions
- Skip the build check before commits
- Create `.cursorrules`, `.windsurfrules`, or similar files
