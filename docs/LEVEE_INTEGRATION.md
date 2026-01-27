# Levee Integration Guide

**Note**: Levee is an **optional** mode. By default, Gobot runs in standalone mode with SQLite and direct Stripe integration. Use Levee mode when you need managed platform features like email sequences, CMS, and AI gateway.

This guide covers using [Levee](https://levee.sh) as the backend-as-a-service for authentication, billing, email, and AI features.

## When to Use Levee Mode

| Feature | Standalone Mode | Levee Mode |
|---------|----------------|------------|
| **Database** | SQLite (local) | Levee-managed |
| **Auth** | Built-in JWT | Levee Auth |
| **Billing** | Direct Stripe | Levee Billing |
| **Email** | Not included | Full email platform |
| **CMS** | Not included | Blog, pages, menus |
| **AI Gateway** | Not included | LLM with cost tracking |
| **Dependencies** | Zero external | Levee API |

Use **Standalone Mode** when you want zero external dependencies and simple billing.

Use **Levee Mode** when you need email sequences, CMS, or AI features.

## What Levee Provides

| Feature | Description |
|---------|-------------|
| **Authentication** | User registration, login, JWT tokens, email verification, password reset |
| **Billing** | Stripe integration, subscriptions, checkout sessions, billing portal |
| **Email** | Transactional emails, email lists, drip sequences |
| **CMS** | Blog posts, pages, navigation menus, site settings |
| **AI Gateway** | LLM chat with usage tracking and cost management |

## Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   SvelteKit     │────▶│    Go-Zero      │────▶│     Levee       │
│   Frontend      │     │    Backend      │     │     API         │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                       │                       │
        │                       │                       │
        ▼                       ▼                       ▼
   User Interface         Business Logic          Auth, Billing,
   API Client             Levee SDK               Email, CMS
```

## SDK Setup

### Backend (Go)

The Levee SDK is initialized in `internal/svc/servicecontext.go`:

```go
package svc

import (
    leveeSDK "github.com/almatuck/levee-go"
    "myapp/internal/config"
)

type ServiceContext struct {
    Config config.Config
    Levee  *leveeSDK.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    var leveeClient *leveeSDK.Client
    if c.Levee.Enabled && c.Levee.APIKey != "" {
        baseURL := c.Levee.BaseURL
        if baseURL == "" {
            baseURL = "https://api.levee.sh"
        }
        client, err := leveeSDK.NewClient(c.Levee.APIKey, baseURL)
        if err == nil {
            leveeClient = client
        }
    }

    return &ServiceContext{
        Config: c,
        Levee:  leveeClient,
    }
}
```

### Frontend (TypeScript)

The frontend uses Levee for CMS content at build time:

```typescript
// src/lib/server/levee.ts
import { Levee } from '@levee/sdk';
import { env } from '$env/dynamic/private';

export const levee = env.LEVEE_API_KEY
    ? new Levee(env.LEVEE_API_KEY, env.LEVEE_BASE_URL || 'https://api.levee.sh')
    : null;
```

## Authentication

### How It Works

1. User submits registration/login form
2. Frontend calls Go backend API
3. Backend calls Levee Auth API
4. Levee validates and returns JWT tokens
5. Backend returns tokens to frontend
6. Frontend stores tokens and redirects

### Backend Implementation

```go
// internal/logic/auth/registerlogic.go
func (l *RegisterLogic) Register(req *types.RegisterRequest) (*types.RegisterResponse, error) {
    if l.svcCtx.Levee == nil {
        return nil, fmt.Errorf("levee service not configured")
    }

    resp, err := l.svcCtx.Levee.Auth.Register(l.ctx, &levee.RegisterRequest{
        Email:    req.Email,
        Password: req.Password,
        Name:     req.Name,
    })
    if err != nil {
        return nil, err
    }

    return &types.RegisterResponse{
        Token:        resp.Token,
        RefreshToken: resp.RefreshToken,
        Customer:     mapCustomer(resp.Customer),
    }, nil
}
```

### Frontend Implementation

```typescript
// src/lib/stores/auth.ts
export async function register(email: string, password: string, name: string) {
    const response = await api.auth.register({ email, password, name });

    if (response.token) {
        setToken(response.token);
        setRefreshToken(response.refreshToken);
        goto('/app');
    }
}
```

## Billing

### Setting Up Products in Levee

1. Go to Levee Dashboard > Products
2. Create your pricing tiers:

| Product | Slug | Price | Interval |
|---------|------|-------|----------|
| Starter | starter | $0 | - |
| Pro | pro | $29 | month |
| Team | team | $99 | month |

3. Add features to each product
4. Set the "highlighted" flag on your recommended plan

### Checkout Flow

```go
// internal/logic/subscription/createcheckoutlogic.go
func (l *CreateCheckoutLogic) CreateCheckout(req *types.CreateCheckoutRequest) (*types.CreateCheckoutResponse, error) {
    if l.svcCtx.Levee == nil {
        return nil, fmt.Errorf("levee service not configured")
    }

    email, _ := auth.GetEmailFromContext(l.ctx)

    // Create order via Levee SDK (returns checkout URL)
    orderResp, err := l.svcCtx.Levee.Orders.CreateOrder(l.ctx, &levee.OrderRequest{
        Email:       email,
        ProductSlug: req.ProductSlug,
        SuccessUrl:  req.SuccessURL,
        CancelUrl:   req.CancelURL,
    })
    if err != nil {
        return nil, err
    }

    return &types.CreateCheckoutResponse{
        CheckoutUrl: orderResp.CheckoutUrl,
    }, nil
}
```

### Billing Portal

```go
// Let users manage their subscription
func (l *CreateBillingPortalLogic) CreateBillingPortal(req *types.CreateBillingPortalRequest) (*types.CreateBillingPortalResponse, error) {
    if l.svcCtx.Levee == nil {
        return nil, fmt.Errorf("levee service not configured")
    }

    email, _ := auth.GetEmailFromContext(l.ctx)

    portal, err := l.svcCtx.Levee.Billing.GetCustomerPortal(l.ctx, email, req.ReturnUrl)
    if err != nil {
        return nil, err
    }

    return &types.CreateBillingPortalResponse{Url: portal.URL}, nil
}
```

### Feature Gating

Check subscription status and gate features by plan:

```svelte
<script lang="ts">
    import { subscription } from '$lib/stores/subscription';
    import { Button } from '$lib/components/ui';

    let hasPro = $derived($subscription?.plan === 'pro');
</script>

{#if hasPro}
    <AdvancedFeature />
{:else}
    <div class="upgrade-prompt">
        <p>Upgrade to Pro to access this feature</p>
        <Button href="/pricing">View Plans</Button>
    </div>
{/if}
```

## Email

### Transactional Emails

```go
// Send a welcome email
func (l *WelcomeLogic) SendWelcome(email string, name string) error {
    _, err := l.svcCtx.Levee.Emails.SendEmail(&levee.SendEmailRequest{
        To:           email,
        TemplateSlug: "welcome",
        Data: map[string]interface{}{
            "name": name,
        },
    })
    return err
}
```

### Email Sequences

Enroll users in drip campaigns:

```go
// Enroll in onboarding sequence
func (l *OnboardingLogic) StartOnboarding(email string) error {
    _, err := l.svcCtx.Levee.Sequences.EnrollInSequence(&levee.EnrollSequenceRequest{
        Email:        email,
        SequenceSlug: "onboarding",
    })
    return err
}
```

### Email Lists

Subscribe users to newsletters:

```go
// Subscribe to newsletter
func (l *NewsletterLogic) Subscribe(email string) error {
    _, err := l.svcCtx.Levee.Lists.SubscribeToList("newsletter", &levee.SubscribeRequest{
        Email: email,
    })
    return err
}
```

## CMS Content

### Navigation Menus

Menus are fetched at build time and embedded in static HTML:

```typescript
// src/routes/(www)/+layout.server.ts
export async function load() {
    if (!levee) return { headerMenu: null };

    try {
        const headerMenu = await levee.site.getNavigationMenu('header');
        return { headerMenu };
    } catch {
        return { headerMenu: null };
    }
}
```

### Blog Posts

```typescript
// src/routes/(www)/blog/+page.server.ts
export async function load() {
    if (!levee) return { posts: [], categories: [] };

    const [postsRes, categoriesRes] = await Promise.all([
        levee.content.listContentPosts(),
        levee.content.listContentCategories()
    ]);

    return {
        posts: postsRes.posts,
        categories: categoriesRes.categories
    };
}
```

### Static Pages

```typescript
// src/routes/(www)/[...slug]/+page.server.ts
export async function load({ params }) {
    if (!levee) error(404, 'Page not found');

    const page = await levee.content.getContentPage(params.slug);
    return { page };
}
```

## Pricing Page Integration

The pricing page automatically fetches products from Levee:

```typescript
// src/routes/(www)/pricing/+page.server.ts
export async function load() {
    if (!levee) return { products: fallbackProducts };

    const products = await Promise.all(
        ['starter', 'pro', 'team'].map(slug =>
            levee.products.getProduct(slug).catch(() => null)
        )
    );

    return { products: products.filter(Boolean) };
}
```

## AI Chat (LLM Gateway)

Levee provides an LLM gateway with usage tracking:

```go
// internal/logic/ai/chatlogic.go
func (l *ChatLogic) Chat(req *types.ChatRequest) (*types.ChatResponse, error) {
    resp, err := l.svcCtx.Levee.LLM.Chat(&levee.LLMChatRequest{
        Model:    "gpt-4",
        Messages: req.Messages,
        CustomerEmail: l.ctx.Value("email").(string), // For usage tracking
    })
    if err != nil {
        return nil, err
    }

    return &types.ChatResponse{
        Message: resp.Message,
        Usage:   resp.Usage,
    }, nil
}
```

## Webhooks

The Levee SDK provides built-in webhook handlers for Stripe and SES events. Register them in your main server file:

```go
// saas-starter.go
if ctx.Levee != nil {
    ctx.Levee.RegisterHandlers(http.DefaultServeMux, "",
        levee.WithUnsubscribeRedirect("/unsubscribed"),
        levee.WithConfirmRedirect("/welcome"),
        levee.WithConfirmExpiredRedirect("/confirm-expired"),
    )
}
```

This registers handlers at:
- `/webhooks/stripe` - Stripe payment/subscription events
- `/webhooks/ses` - Amazon SES email events
- `/e/` - Email tracking (opens, clicks)
- `/confirm-email` - Email confirmation

The SDK automatically verifies webhook signatures using your Stripe/SES webhook secrets.

## Enabling Levee Mode

Add to your `.env`:

```bash
LEVEE_ENABLED=true
LEVEE_API_KEY=lvk_your_api_key_here
LEVEE_BASE_URL=https://api.levee.sh  # Optional, this is the default
```

Get your API key from the [Levee Dashboard](https://dashboard.levee.sh).

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `LEVEE_ENABLED` | Set to `true` to enable Levee mode | Yes |
| `LEVEE_API_KEY` | Your Levee API key | Yes |
| `LEVEE_BASE_URL` | Levee API URL (default: https://api.levee.sh) | No |
| `LEVEE_WEBHOOK_SECRET` | For webhook signature verification | For webhooks |

## Testing Without Levee

The boilerplate includes fallback behavior when Levee is not configured:

- **Pricing page**: Shows hardcoded placeholder products
- **CMS pages**: Returns 404 (no CMS content)
- **Auth/Billing**: Will fail (Levee is required for these features)

## Resources

- [Levee Documentation](https://docs.levee.sh)
- [Levee Go SDK](https://github.com/almatuck/levee-go)
- [Levee TypeScript SDK](https://github.com/almatuck/levee-ts)
- [Levee Dashboard](https://dashboard.levee.sh)
