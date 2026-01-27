# Stripe Integration Guide

This guide covers setting up Stripe for Gobot in **standalone mode** (without Levee).

## Overview

Gobot supports two modes for billing:

1. **Standalone Mode**: Direct Stripe integration with products defined in YAML config
2. **Levee Mode**: Billing handled by Levee platform (see [LEVEE_INTEGRATION.md](./LEVEE_INTEGRATION.md))

This guide covers standalone mode.

## Quick Setup

### 1. Create a Stripe Account

1. Go to [stripe.com](https://stripe.com) and create an account
2. Complete account verification for live payments

### 2. Get Your API Keys

1. Go to [Stripe Dashboard → Developers → API keys](https://dashboard.stripe.com/apikeys)
2. Copy your **Secret key** (`sk_test_xxx` or `sk_live_xxx`)
3. Copy your **Publishable key** (`pk_test_xxx` or `pk_live_xxx`)

### 3. Configure Environment Variables

Add to your `.env` file:

```bash
# Stripe API Keys
STRIPE_SECRET_KEY=sk_test_your_key_here
STRIPE_PUBLISHABLE_KEY=pk_test_your_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret  # Set up in step 5
```

### 4. Configure Products

Products are defined in `etc/gobot.yaml`. The app syncs them to Stripe on startup.

```yaml
Products:
  # Free tier (no Stripe subscription created)
  - slug: "free"
    name: "Free"
    description: "Get started for free"
    default: true # Default plan for new users
    features:
      - "5 projects"
      - "Basic analytics"
      - "Community support"
    prices:
      - slug: "default"
        amount: 0
        currency: "usd"
        interval: "month"

  # Pro tier with monthly and yearly pricing
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
        amount: 2900 # $29.00 in cents
        currency: "usd"
        interval: "month"
        trialDays: 14 # 14-day free trial
      - slug: "yearly"
        amount: 29000 # $290.00 (save ~17%)
        currency: "usd"
        interval: "year"
        trialDays: 14

  # Team tier
  - slug: "team"
    name: "Team"
    description: "For teams that need collaboration"
    features:
      - "Everything in Pro"
      - "Team collaboration"
      - "Admin dashboard"
      - "SAML SSO"
      - "Dedicated support"
    prices:
      - slug: "monthly"
        amount: 7900 # $79.00
        currency: "usd"
        interval: "month"
      - slug: "yearly"
        amount: 79000 # $790.00
        currency: "usd"
        interval: "year"
```

### 5. Set Up Webhooks

Webhooks notify your app when subscription events occur (payments, cancellations, etc.).

#### Development (Local Testing)

Use Stripe CLI to forward webhooks to your local server:

```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Login to your Stripe account
stripe login

# Forward webhooks to your local server
stripe listen --forward-to localhost:8888/api/webhook/stripe
```

Copy the webhook signing secret (`whsec_xxx`) and add it to your `.env`:

```bash
STRIPE_WEBHOOK_SECRET=whsec_xxx
```

#### Production

1. Go to [Stripe Dashboard → Developers → Webhooks](https://dashboard.stripe.com/webhooks)
2. Click "Add endpoint"
3. Enter your endpoint URL: `https://yourdomain.com/api/webhook/stripe`
4. Select events to listen for:
   - `checkout.session.completed`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.payment_succeeded`
   - `invoice.payment_failed`
5. Click "Add endpoint"
6. Click "Reveal" to see the signing secret
7. Copy the signing secret to your production `.env`

## Product Configuration Reference

### Product Fields

| Field         | Required | Description                                     |
| ------------- | -------- | ----------------------------------------------- |
| `slug`        | Yes      | Unique identifier (e.g., "free", "pro", "team") |
| `name`        | Yes      | Display name shown to customers                 |
| `description` | No       | Product description for pricing page            |
| `features`    | No       | List of features for this plan                  |
| `default`     | No       | Set `true` for the default plan for new users   |
| `prices`      | Yes      | List of pricing options                         |

### Price Fields

| Field           | Required | Description                                                          |
| --------------- | -------- | -------------------------------------------------------------------- |
| `slug`          | Yes      | Price identifier (e.g., "monthly", "yearly", "default")              |
| `amount`        | Yes      | Price in cents (e.g., 2900 = $29.00)                                 |
| `currency`      | No       | Currency code (default: "usd")                                       |
| `interval`      | No       | Billing interval: "month", "year", or "one_time" (default: "month")  |
| `intervalCount` | No       | Frequency (default: 1). Set to 3 with interval "month" for quarterly |
| `trialDays`     | No       | Free trial days (default: 0)                                         |

## How Product Sync Works

On startup, Gobot:

1. Reads products from `etc/gobot.yaml`
2. Checks if each product exists in Stripe (by metadata)
3. Creates new products/prices if they don't exist
4. Updates product names/descriptions if they changed
5. Stores Stripe IDs in memory for checkout lookups

Products are identified in Stripe by metadata:

- Products: `gobot_slug: "pro"`
- Prices: `gobot_key: "pro-monthly"`

**Important**: Stripe prices are immutable. If you need to change a price amount:

1. Update the `slug` (e.g., "monthly" → "monthly-v2")
2. A new price will be created in Stripe
3. Existing subscribers keep their old price

## Frontend Integration

### Checkout Flow

```javascript
// Create checkout session (in your frontend)
const response = await fetch("/api/v1/subscription/checkout", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  },
  body: JSON.stringify({
    planName: "pro", // Product slug
    billingCycle: "monthly", // Price slug (optional, defaults to first price)
  }),
});

const { checkoutUrl } = await response.json();
window.location.href = checkoutUrl; // Redirect to Stripe Checkout
```

### Billing Portal

```javascript
// Open Stripe billing portal for subscription management
const response = await fetch("/api/v1/subscription/billing-portal", {
  method: "POST",
  headers: {
    Authorization: `Bearer ${token}`,
  },
});

const { portalUrl } = await response.json();
window.location.href = portalUrl;
```

## API Endpoints

| Method | Path                                   | Description                   |
| ------ | -------------------------------------- | ----------------------------- |
| GET    | `/api/v1/subscription/plans`           | List available plans          |
| GET    | `/api/v1/subscription`                 | Get current subscription      |
| POST   | `/api/v1/subscription/checkout`        | Create checkout session       |
| POST   | `/api/v1/subscription/billing-portal`  | Create billing portal session |
| POST   | `/api/v1/subscription/cancel`          | Cancel subscription           |
| POST   | `/api/v1/subscription/check-feature`   | Check feature access          |
| GET    | `/api/v1/subscription/usage`           | Get usage statistics          |
| GET    | `/api/v1/subscription/billing-history` | Get billing history           |

## Webhook Events

The webhook handler processes these Stripe events:

| Event                           | Action                                       |
| ------------------------------- | -------------------------------------------- |
| `checkout.session.completed`    | Activate subscription, update user record    |
| `customer.subscription.updated` | Sync subscription status and period          |
| `customer.subscription.deleted` | Downgrade user to free plan                  |
| `invoice.payment_succeeded`     | Log successful payment                       |
| `invoice.payment_failed`        | Log failed payment (send email notification) |

## Testing

### Test Cards

Use these test card numbers in test mode:

| Card Number         | Result                   |
| ------------------- | ------------------------ |
| 4242 4242 4242 4242 | Successful payment       |
| 4000 0000 0000 0002 | Card declined            |
| 4000 0000 0000 3220 | 3D Secure authentication |

Use any future expiry date and any 3-digit CVC.

### Testing Webhooks Locally

```bash
# Terminal 1: Run your app
make air

# Terminal 2: Forward webhooks
stripe listen --forward-to localhost:8888/api/webhook/stripe

# Terminal 3: Trigger test events
stripe trigger checkout.session.completed
stripe trigger customer.subscription.updated
stripe trigger invoice.payment_failed
```

## Troubleshooting

### Products not syncing

Check that:

1. `STRIPE_SECRET_KEY` is set correctly
2. Products are defined in `etc/gobot.yaml`
3. You're running in standalone mode (`LEVEE_ENABLED=false`)

Look for startup logs:

```
Synced 3 products to Stripe
```

### Webhooks not received

Check that:

1. `STRIPE_WEBHOOK_SECRET` matches your endpoint
2. Webhook URL is correct: `/api/webhook/stripe`
3. In development: Stripe CLI is running and forwarding

Test webhook signature:

```bash
# View recent webhook attempts in Stripe Dashboard
# Dashboard → Developers → Webhooks → Select endpoint → Recent events
```

### Checkout fails

Check that:

1. Products have synced (Stripe IDs are set)
2. User is authenticated
3. `APP_BASE_URL` is set correctly for redirects

## Going Live

Before launching:

1. [ ] Switch to live API keys (`sk_live_xxx`, `pk_live_xxx`)
2. [ ] Create production webhook endpoint
3. [ ] Configure production `STRIPE_WEBHOOK_SECRET`
4. [ ] Review pricing in `etc/gobot.yaml`
5. [ ] Test complete checkout flow
6. [ ] Test subscription management (upgrades, cancellations)
7. [ ] Verify webhook events are received

## Further Reading

- [Stripe Checkout Documentation](https://stripe.com/docs/payments/checkout)
- [Stripe Billing Documentation](https://stripe.com/docs/billing)
- [Stripe Webhooks Documentation](https://stripe.com/docs/webhooks)
- [Stripe CLI Documentation](https://stripe.com/docs/stripe-cli)
