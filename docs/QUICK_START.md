# Quick Start Guide

Launch your SaaS in under 10 minutes.

## Recommended: Claude Code + /init

We recommend using [Claude Code](https://claude.ai/code) for the best development experience:

```bash
# 1. Clone the repository
git clone https://github.com/almatuck/gobot.git myapp
cd myapp

# 2. Open in Claude Code, then run:
/init
```

The `/init` command will interactively:
1. Rename the project to your app name
2. Prompt for admin email and password
3. Generate secure secrets
4. Configure ports automatically
5. Install all dependencies
6. Start the development environment

## Alternative: One-Line Install

```bash
curl -fsSL https://raw.githubusercontent.com/almatuck/gobot/main/install.sh | bash -s -- myapp
```

This single command:
1. Checks you have Node.js and pnpm installed
2. Clones the repository
3. Renames everything from `gobot` to `myapp`
4. Prompts for admin email and password
5. Generates a secure JWT secret
6. Creates `.env` with working defaults
7. Installs all dependencies

**Then visit http://localhost:YOUR_PORT** (shown after install)

## Prerequisites

The installer checks for these automatically:

- **Docker** - [Download](https://docker.com/) (for backend development)
- **Node.js 20+** - [Download](https://nodejs.org/) (for frontend)
- **pnpm** - Auto-installed if npm is available

**Note:** Go runs inside Docker, so you don't need it installed locally.

## Alternative: Manual Setup

If you prefer to clone manually:

```bash
git clone https://github.com/almatuck/gobot.git myapp
cd myapp
./install.sh myapp
make dev
```

The install script auto-detects whether it's running remotely (via curl) or locally (after clone) and does the right thing.

## Two Modes of Operation

Gobot works out of the box in **Standalone Mode**. No configuration needed to start developing.

### Standalone Mode (Default)

- **Database**: SQLite (auto-created)
- **Auth**: Built-in JWT authentication
- **Billing**: Direct Stripe integration
- **Zero external dependencies**

This is set up automatically by the installer.

### Levee Mode (Optional)

For advanced features (email sequences, CMS, AI gateway), edit `.env`:

```bash
LEVEE_ENABLED=true
LEVEE_API_KEY=lvk_your_api_key_here  # Get from https://dashboard.levee.sh
LEVEE_BASE_URL=https://api.levee.sh
```

See [Levee Integration Guide](./LEVEE_INTEGRATION.md) for details.

---

## Customizing Your App

### Step 1: Configure Your Brand (5 minutes)

Edit `app/src/lib/config/site.ts`:

```typescript
export const site = {
  name: 'MyApp',
  tagline: 'Your awesome tagline',
  description: 'A brief description for SEO',
  url: 'https://myapp.com',  // Your production URL
  supportEmail: 'support@myapp.com',
  ogImage: '/images/og-default.png',
  twitter: '@myapp',
  social: {
    twitter: 'https://twitter.com/myapp',
    github: 'https://github.com/myapp',
    linkedin: '',
    discord: ''
  },
  legal: {
    companyName: 'MyApp, Inc.',
    companyAddress: '123 Main St, City, State 12345',
    privacyUrl: '/privacy',
    termsUrl: '/terms'
  }
}
```

### Step 2: Configure Products & Pricing

Edit `etc/gobot.yaml` (or `etc/myapp.yaml` after renaming):

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
    description: "For professionals"
    features:
      - "Unlimited projects"
      - "Advanced analytics"
      - "Priority support"
    prices:
      - slug: "monthly"
        amount: 2900  # $29.00
        currency: "usd"
        interval: "month"
        trialDays: 14
      - slug: "yearly"
        amount: 29000  # $290.00
        currency: "usd"
        interval: "year"
        trialDays: 14
```

### Step 3: Set Up Stripe Webhooks (For Payments)

### Development

```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Forward webhooks (run in separate terminal)
stripe listen --forward-to localhost:8888/api/webhook/stripe
```

Copy the webhook secret (`whsec_xxx`) to your `.env`:

```bash
STRIPE_WEBHOOK_SECRET=whsec_xxx
```

### Production

1. Go to [Stripe Dashboard → Webhooks](https://dashboard.stripe.com/webhooks)
2. Add endpoint: `https://yourdomain.com/api/webhook/stripe`
3. Select events: `checkout.session.completed`, `customer.subscription.*`, `invoice.*`
4. Copy signing secret to production `.env`

---

## Running the App

### Easiest Way

```bash
make dev
```

This starts both backend and frontend using Docker (or natively if Docker isn't available).

### Manual (Two Terminals)

**Terminal 1 - Backend:**
```bash
make air
```

**Terminal 2 - Frontend:**
```bash
cd app && pnpm dev
```

---

## Verify Everything Works

1. Open http://localhost:5173
2. Click "Get Started Free" to register
3. Check the dashboard at http://localhost:5173/app
4. Try the pricing page at http://localhost:5173/pricing
5. Test checkout flow (use Stripe test card: `4242 4242 4242 4242`)

---

## Customize the UI

### Colors & Theme

Edit `app/src/app.css`:

```css
:root {
  --color-primary: #your-brand-color;
  --color-accent: #your-accent-color;
}
```

### Logo & Favicon

Replace files in `app/static/`:
- `favicon.svg` - Browser tab icon
- `logo.svg` - Site logo

### Landing Page

Edit `app/src/routes/(www)/+page.svelte`:
- Update hero section
- Modify feature cards
- Customize FAQ

## What's Next?

- [Configuration Guide](./CONFIGURATION.md) - All settings explained
- [Deployment Guide](./DEPLOYMENT.md) - Production deployment with auto-SSL
- [Stripe Integration](./STRIPE.md) - Payment setup
- [Levee Integration](./LEVEE_INTEGRATION.md) - Platform features (optional)
- [Customization Guide](./CUSTOMIZATION.md) - Theming and components

## Troubleshooting

### Build Errors

```bash
# Clear caches and rebuild
cd app
rm -rf node_modules .svelte-kit
pnpm install
pnpm build
```

### Port Already in Use

```bash
# Check ports
lsof -i :8888
lsof -i :5173

# Kill process or change ports in .env
```

### Stripe Webhooks Not Working

1. Ensure Stripe CLI is running: `stripe listen --forward-to localhost:8888/api/webhook/stripe`
2. Check `STRIPE_WEBHOOK_SECRET` matches the CLI output
3. Verify endpoint URL is correct

### Database Issues

```bash
# Reset SQLite database
rm -f ./data/myapp.db
make air  # Will recreate on startup
```
