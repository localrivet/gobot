# Gobot

Ship your SaaS in days, not months. A production-ready full-stack boilerplate with Go-Zero backend and SvelteKit frontend.

## Features

- **Single Binary Deployment** - Go backend with embedded static frontend
- **Two Modes** - Standalone (SQLite + Stripe) or Levee-powered (managed platform)
- **Authentication** - JWT auth with refresh tokens, email verification, password reset
- **Billing** - Stripe integration with subscriptions, trials, and billing portal
- **Email** - SMTP support for transactional emails (password reset, welcome, etc.)
- **Modern Frontend** - SvelteKit 2 + Svelte 5 + Tailwind v4
- **Type-Safe API** - Auto-generated TypeScript client from Go API definitions
- **Production Ready** - Rate limiting, CORS, security headers, health checks

## Quick Start

### Recommended: Use Claude Code

We recommend [Claude Code](https://claude.ai/code) for the best development experience. After cloning, run `/init` to set up your project interactively.

```bash
# Clone the repository
git clone https://github.com/almatuck/gobot.git myapp
cd myapp

# Open in Claude Code, then run:
/init
```

### Alternative: One-Line Install

```bash
curl -fsSL https://raw.githubusercontent.com/almatuck/gobot/main/install.sh | bash -s -- myapp
```

This will:
- Clone the repository
- Rename everything to `myapp`
- Prompt for admin email and password
- Auto-generate secure JWT secret
- Install all dependencies
- Create `.env` with working defaults

Then visit **http://localhost:YOUR_PORT** (shown after install)

### Already Cloned?

```bash
./install.sh myapp    # Rename and configure
make dev              # Start everything
```

### Manual Start

```bash
make air              # Backend with hot reload (port 8888)
cd app && pnpm dev    # Frontend dev server (port 5173)
```

## Project Structure

```
myapp/
├── myapp.api                  # API definition (routes, types)
├── myapp.go                   # Entry point
├── etc/
│   └── gobot.yaml            # Backend config + Products/Pricing
├── internal/
│   ├── handler/               # Auto-generated handlers (DO NOT EDIT)
│   ├── types/                 # Auto-generated types (DO NOT EDIT)
│   ├── logic/                 # Business logic (EDIT HERE)
│   │   ├── auth/              # Login, register, password reset
│   │   ├── user/              # Profile, preferences, account
│   │   └── subscription/      # Billing, checkout, usage
│   ├── svc/                   # Service context
│   ├── db/                    # SQLite setup (standalone mode)
│   ├── local/                 # Local auth & billing (standalone mode)
│   └── middleware/            # CORS, JWT, rate limiting
├── app/                       # SvelteKit frontend
│   ├── src/
│   │   ├── routes/            # Pages and API routes
│   │   │   ├── (www)/         # Marketing pages
│   │   │   ├── (auth)/        # Auth pages (login, register)
│   │   │   └── (app)/         # Authenticated app pages
│   │   └── lib/
│   │       ├── config/
│   │       │   └── site.ts    # Branding/SEO config
│   │       ├── api/           # Auto-generated API client
│   │       ├── stores/        # Svelte stores
│   │       └── components/    # UI components
│   └── static/                # Static assets
└── docs/                      # Documentation
```

## Configuration

Gobot uses a simple 3-file configuration system:

| File | Purpose | When to Edit |
|------|---------|--------------|
| `app/src/lib/config/site.ts` | Frontend branding, SEO, social links | Rebrand your product |
| `etc/gobot.yaml` | Backend config + products/pricing | Add features, change pricing |
| `.env` | Secrets (API keys, JWT secret) | Per-environment setup |

### 1. Branding & SEO (site.ts)

Edit `app/src/lib/config/site.ts` to customize your brand:

```typescript
export const site = {
  name: 'YourApp',
  tagline: 'Your tagline here',
  description: 'Your meta description for SEO',
  url: 'https://yourapp.com',
  supportEmail: 'support@yourapp.com',
  ogImage: '/images/og-default.png',
  twitter: '@yourhandle',
  social: {
    twitter: 'https://twitter.com/yourapp',
    github: 'https://github.com/yourapp',
    linkedin: '',
    discord: ''
  },
  legal: {
    companyName: 'Your Company, Inc.',
    companyAddress: '123 Main St, City, State 12345',
    privacyUrl: '/privacy',
    termsUrl: '/terms'
  }
}
```

### 2. Products & Pricing (gobot.yaml)

Edit `etc/gobot.yaml` to configure your pricing plans:

```yaml
Products:
  - slug: "free"
    name: "Free"
    description: "Get started for free"
    default: true
    features:
      - "5 projects"
      - "Basic analytics"
      - "Community support"
    prices:
      - slug: "default"
        amount: 0
        currency: "usd"
        interval: "month"

  - slug: "pro"
    name: "Pro"
    description: "For professionals and growing teams"
    features:
      - "Unlimited projects"
      - "Advanced analytics"
      - "Priority support"
      - "API access"
    prices:
      - slug: "monthly"
        amount: 2900  # $29.00 in cents
        currency: "usd"
        interval: "month"
        trialDays: 14
      - slug: "yearly"
        amount: 29000  # $290.00 (save ~17%)
        currency: "usd"
        interval: "year"
        trialDays: 14
```

**Note:** Pricing is automatically embedded into the static frontend at build time - no API call required.

### 3. Environment Variables (.env)

Choose ONE mode. The installer creates a working `.env` for Standalone Mode by default.

**Standalone Mode** (default - no external services):

```bash
ACCESS_SECRET=your-jwt-secret-here    # Auto-generated by installer
APP_BASE_URL=http://localhost:5173
LEVEE_ENABLED=false
SQLITE_PATH=./data/gobot.db

# Stripe
STRIPE_SECRET_KEY=sk_test_xxx
STRIPE_PUBLISHABLE_KEY=pk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx

# Email (optional - for password reset, welcome emails)
SMTP_HOST=smtp.gmail.com              # Or SendGrid, Mailgun, SES, etc.
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
EMAIL_FROM_ADDRESS=noreply@yourapp.com
EMAIL_FROM_NAME=YourApp
```

**Levee Mode** (managed platform features):

```bash
ACCESS_SECRET=your-jwt-secret-here    # Auto-generated by installer
APP_BASE_URL=http://localhost:5173
LEVEE_ENABLED=true
LEVEE_API_KEY=lvk_xxx                 # Get from https://dashboard.levee.sh
LEVEE_BASE_URL=https://api.levee.sh
```

## Two Modes of Operation

### Standalone Mode (Default)

Zero external dependencies. Everything runs locally:

- **Database**: SQLite with WAL mode
- **Auth**: Local JWT auth with bcrypt passwords
- **Billing**: Direct Stripe integration
- **Products**: Synced to Stripe on startup from `gobot.yaml`

```bash
LEVEE_ENABLED=false  # or just don't set it
```

### Levee Mode

Full platform features via [Levee](https://levee.sh):

- **Auth**: Managed authentication with email verification
- **Billing**: Stripe subscriptions, trials, usage tracking
- **Email**: Transactional emails, lists, sequences
- **LLM**: Cost-tracked AI gateway

```bash
LEVEE_ENABLED=true
LEVEE_API_KEY=lvk_xxx
```

## Development Commands

```bash
# Backend
make air              # Hot reload development
make build            # Build binary
make test             # Run tests
make gen              # Regenerate handlers from .api file (NEVER run goctl directly)

# Frontend
cd app
pnpm dev              # Dev server (port 5173)
pnpm build            # Production build
pnpm check            # Type checking
```

## Adding API Endpoints

1. **Define the endpoint** in `gobot.api`:

```api
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

2. **Generate handlers**: `make gen`

3. **Implement logic** in `internal/logic/getwidgetlogic.go`:

```go
func (l *GetWidgetLogic) GetWidget(req *types.GetWidgetRequest) (*types.GetWidgetResponse, error) {
    return &types.GetWidgetResponse{
        Id:   req.Id,
        Name: "My Widget",
    }, nil
}
```

4. **Use in frontend** (TypeScript client auto-generated):

```typescript
import { getWidget } from '$lib/api';
const widget = await getWidget({ id: '123' });
```

## Deployment

### Single Binary with Auto-SSL

Gobot deploys as a single binary - **no nginx, Apache, or reverse proxy needed**:

```bash
make build
cd app && pnpm build
# Deploy ./bin/gobot + .env anywhere
./bin/gobot
```

In production mode, the binary automatically:
- Obtains Let's Encrypt SSL certificates
- Serves HTTPS on port 443
- Redirects HTTP (port 80) to HTTPS
- Redirects www to non-www
- Enables HTTP/2 and gzip compression

```bash
# Production .env
PRODUCTION_MODE=true
APP_DOMAIN=myapp.com
APP_ADMIN_EMAIL=admin@myapp.com
```

### Docker

```bash
docker build -t myapp .
docker run -p 80:80 -p 443:443 --env-file .env myapp
```

### Production (Docker Compose)

```bash
docker compose --profile production up -d --build
```

See [Deployment Guide](./docs/DEPLOYMENT.md) for platform-specific instructions (Fly.io, Railway, DigitalOcean, AWS).

## API Endpoints

### Auth (Public)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/refresh` | Refresh token |
| POST | `/api/v1/auth/forgot-password` | Request password reset |
| POST | `/api/v1/auth/reset-password` | Reset password |
| POST | `/api/v1/auth/verify-email` | Verify email |

### User (Authenticated)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/user/me` | Get current user |
| PUT | `/api/v1/user/me` | Update profile |
| DELETE | `/api/v1/user/me` | Delete account |
| POST | `/api/v1/user/me/change-password` | Change password |
| GET | `/api/v1/user/me/preferences` | Get preferences |
| PUT | `/api/v1/user/me/preferences` | Update preferences |

### Subscription (Authenticated)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/subscription` | Get current subscription |
| GET | `/api/v1/subscription/plans` | List available plans |
| POST | `/api/v1/subscription/checkout` | Create checkout session |
| POST | `/api/v1/subscription/billing-portal` | Open billing portal |
| POST | `/api/v1/subscription/cancel` | Cancel subscription |

## Documentation

- [Quick Start Guide](./docs/QUICK_START.md) - Get running in under an hour
- [Configuration Guide](./docs/CONFIGURATION.md) - All settings explained
- [Deployment Guide](./docs/DEPLOYMENT.md) - Production deployment with auto-SSL
- [Stripe Integration](./docs/STRIPE.md) - Standalone mode billing setup
- [Levee Integration](./docs/LEVEE_INTEGRATION.md) - Managed platform features
- [Customization](./docs/CUSTOMIZATION.md) - Theming, branding, components
- [Research Runner](./docs/RESEARCH.md) - Validate your SaaS idea (bonus tool)

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.25+, go-zero framework |
| Frontend | SvelteKit 2, Svelte 5, Tailwind v4 |
| Database | SQLite (standalone) or Levee (managed) |
| Payments | Stripe |
| Build | Single binary with embedded static frontend |

## License

MIT
