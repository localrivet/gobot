# Configuration Guide

Complete reference for all configuration in Gobot.

## Configuration Architecture

Gobot uses a simple 3-file configuration system:

| File | Purpose | Git Tracked | When to Edit |
|------|---------|-------------|--------------|
| `app/src/lib/config/site.ts` | Frontend branding, SEO | Yes | Rebrand your product |
| `etc/gobot.yaml` | Backend config + products | Yes | Change pricing, features |
| `.env` | Secrets only | No | Per-environment setup |

## 1. Frontend Branding (site.ts)

Single source of truth for all frontend branding and SEO.

**Location:** `app/src/lib/config/site.ts`

```typescript
export const site = {
  // Basic info
  name: 'Gobot',
  tagline: 'Ship your SaaS in days, not months',
  description: 'The fastest way to launch your SaaS...',
  url: 'https://example.com',
  supportEmail: 'support@example.com',

  // SEO
  ogImage: '/images/og-default.png',
  twitter: '@yourhandle',
  keywords: ['saas', 'startup', 'boilerplate'],
  locale: 'en_US',

  // Social links (leave empty to hide)
  social: {
    twitter: 'https://twitter.com/yourapp',
    github: 'https://github.com/yourapp',
    linkedin: '',
    discord: ''
  },

  // Legal
  legal: {
    companyName: 'Your Company, Inc.',
    companyAddress: '123 Main St, City, State 12345',
    privacyUrl: '/privacy',
    termsUrl: '/terms'
  }
} as const;
```

### What Uses site.ts

- SEO meta tags (title, description, OG images)
- Footer (social links, legal info)
- JSON-LD structured data
- Email templates (company info)

## 2. Backend Configuration (gobot.yaml)

Backend settings and product/pricing definitions.

**Location:** `etc/gobot.yaml`

### Server Configuration

```yaml
Name: gobot
Host: 0.0.0.0
Port: 8888

App:
  BaseURL: ${APP_BASE_URL}
  Domain: ${APP_DOMAIN}
  ProductionMode: ${PRODUCTION_MODE}
```

### Authentication

```yaml
Auth:
  AccessSecret: ${ACCESS_SECRET}
  AccessExpire: 7200          # 2 hours
  RefreshTokenExpire: 604800  # 7 days
```

### Database (Standalone Mode)

```yaml
Database:
  SQLitePath: ${SQLITE_PATH}
```

### Stripe (Standalone Mode)

```yaml
Stripe:
  SecretKey: ${STRIPE_SECRET_KEY}
  PublishableKey: ${STRIPE_PUBLISHABLE_KEY}
  WebhookSecret: ${STRIPE_WEBHOOK_SECRET}
  SuccessURL: "/app/account"
  CancelURL: "/pricing"
```

### Products & Pricing

```yaml
Products:
  - slug: "free"
    name: "Free"
    description: "Get started for free"
    default: true  # Default plan for new users
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
    description: "For professionals"
    features:
      - "Unlimited projects"
      - "Advanced analytics"
      - "Priority support"
      - "API access"
    prices:
      - slug: "monthly"
        amount: 2900       # $29.00 in cents
        currency: "usd"
        interval: "month"
        trialDays: 14
      - slug: "yearly"
        amount: 29000      # $290.00
        currency: "usd"
        interval: "year"
        trialDays: 14

  - slug: "team"
    name: "Team"
    description: "For teams"
    features:
      - "Everything in Pro"
      - "Team collaboration"
      - "Admin dashboard"
    prices:
      - slug: "monthly"
        amount: 7900
        currency: "usd"
        interval: "month"
      - slug: "yearly"
        amount: 79000
        currency: "usd"
        interval: "year"
```

#### Product Fields

| Field | Required | Description |
|-------|----------|-------------|
| `slug` | Yes | Unique identifier (e.g., "free", "pro") |
| `name` | Yes | Display name |
| `description` | No | Description for pricing page |
| `default` | No | Set `true` for default plan |
| `features` | No | List of feature strings |
| `prices` | Yes | List of pricing options |

#### Price Fields

| Field | Required | Description |
|-------|----------|-------------|
| `slug` | Yes | Price identifier (e.g., "monthly", "yearly") |
| `amount` | Yes | Price in cents (2900 = $29.00) |
| `currency` | No | Currency code (default: "usd") |
| `interval` | No | "month", "year", or omit for one-time |
| `trialDays` | No | Free trial days (default: 0) |

### Security

```yaml
Security:
  EnableSecurityHeaders: true
  ForceHTTPS: false
  RateLimitEnabled: true
  RateLimitRequests: 100
  RateLimitInterval: 60
  RateLimitBurst: 20
  AuthRateLimitRequests: 5
  AuthRateLimitInterval: 60
  AllowedOrigins: ${ALLOWED_ORIGINS}
```

### Email (SMTP)

```yaml
Email:
  SMTPHost: ${SMTP_HOST}
  SMTPPort: ${SMTP_PORT}
  SMTPUser: ${SMTP_USER}
  SMTPPass: ${SMTP_PASS}
  FromAddress: ${EMAIL_FROM_ADDRESS}
  FromName: ${EMAIL_FROM_NAME}
  ReplyTo: ${EMAIL_REPLY_TO}
```

### Levee Mode

```yaml
Levee:
  APIKey: ${LEVEE_API_KEY}
  BaseURL: ${LEVEE_BASE_URL}
  Enabled: true
  # ... additional Levee config
```

## 3. Environment Variables (.env)

Secrets and per-environment configuration.

### Required (Both Modes)

```bash
# JWT Secret - generate with: openssl rand -hex 32
ACCESS_SECRET=your-256-bit-secret-here

# Frontend URL (for redirects)
APP_BASE_URL=http://localhost:5173
```

### Standalone Mode

```bash
# Mode selection (or just don't set LEVEE_ENABLED)
LEVEE_ENABLED=false

# SQLite database path
SQLITE_PATH=./data/gobot.db

# Stripe API keys
STRIPE_SECRET_KEY=sk_test_xxx
STRIPE_PUBLISHABLE_KEY=pk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
```

### Levee Mode

```bash
LEVEE_ENABLED=true
LEVEE_API_KEY=lvk_xxx
LEVEE_BASE_URL=https://api.levee.sh
```

### Production

```bash
# Required for production
PRODUCTION_MODE=true
APP_DOMAIN=myapp.com

# For Let's Encrypt SSL
APP_ADMIN_EMAIL=admin@myapp.com
```

### Optional

```bash
# CORS (comma-separated origins)
ALLOWED_ORIGINS=http://localhost:5173,https://myapp.com

# SMTP Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user
SMTP_PASS=password
EMAIL_FROM_ADDRESS=noreply@myapp.com
EMAIL_FROM_NAME=MyApp
EMAIL_REPLY_TO=support@myapp.com
```

## Frontend Environment (app/.env)

The frontend only needs minimal configuration since branding is in `site.ts`.

**Location:** `app/.env` or `app/.env.local`

```bash
# Monitoring (optional)
VITE_ENVIRONMENT=development
VITE_APP_VERSION=1.0.0
VITE_ENABLE_DEBUG=false
VITE_ALERT_WEBHOOK_URL=https://hooks.slack.com/xxx
```

**Note:** The API URL is auto-detected from `window.location.origin` since the frontend is embedded in the Go binary.

## How Pricing Works

1. Products/prices are defined in `etc/gobot.yaml`
2. At build time, SvelteKit reads the YAML and embeds pricing in static HTML
3. At runtime (standalone mode), products sync to Stripe on startup
4. No API call needed to display pricing page

```
etc/gobot.yaml → pnpm build → pricing.html (static)
                → make air → Stripe sync (runtime)
```

## Configuration Precedence

1. Environment variables (highest)
2. `.env` file
3. `etc/gobot.yaml` defaults
4. Code defaults (lowest)

## Common Patterns

### Local Development

```bash
# .env
ACCESS_SECRET=dev-secret-not-for-production
APP_BASE_URL=http://localhost:5173
SQLITE_PATH=./data/dev.db
STRIPE_SECRET_KEY=sk_test_xxx
```

### Staging

```bash
# .env
ACCESS_SECRET=staging-secret-xxx
APP_BASE_URL=https://staging.myapp.com
PRODUCTION_MODE=true
APP_DOMAIN=staging.myapp.com
STRIPE_SECRET_KEY=sk_test_xxx
```

### Production

```bash
# .env
ACCESS_SECRET=production-secret-xxx
APP_BASE_URL=https://myapp.com
PRODUCTION_MODE=true
APP_DOMAIN=myapp.com
APP_ADMIN_EMAIL=admin@myapp.com
STRIPE_SECRET_KEY=sk_live_xxx
```

## Security Best Practices

1. **Never commit `.env`** - It's gitignored by default
2. **Rotate secrets** - Change `ACCESS_SECRET` periodically
3. **Use strong secrets** - Generate with `openssl rand -hex 32`
4. **Restrict CORS** - Only allow your actual domains
5. **Enable rate limiting** - Prevent abuse
6. **Use HTTPS** - Always in production
