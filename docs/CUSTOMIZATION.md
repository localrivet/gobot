# Customization Guide

How to customize Gobot for your brand and product.

## Configuration System

Gobot uses a 3-file configuration system:

| File | Purpose | When to Edit |
|------|---------|--------------|
| `app/src/lib/config/site.ts` | Branding, SEO, social links | Rebrand your product |
| `etc/gobot.yaml` | Products, pricing, backend settings | Change pricing, features |
| `.env` | Secrets only | Per-environment setup |

## Branding (site.ts)

All branding is centralized in `app/src/lib/config/site.ts`:

```typescript
export const site = {
  // Basic info
  name: 'YourApp',
  tagline: 'Your tagline here',
  description: 'SEO meta description',
  url: 'https://yourapp.com',
  supportEmail: 'support@yourapp.com',

  // SEO
  ogImage: '/images/og-default.png',
  twitter: '@yourhandle',

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

This is automatically used in:
- Page titles and meta tags
- Footer (social links, legal info)
- JSON-LD structured data
- Navigation

## Theme & Colors

All styling is centralized in `app/src/app.css` using CSS custom properties.

### Color Palette

```css
/* app/src/app.css */
:root {
  /* Base colors (dark theme) */
  --color-base-50: #f8fafc;
  --color-base-100: #f1f5f9;
  --color-base-200: #e2e8f0;
  --color-base-300: #cbd5e1;
  --color-base-400: #94a3b8;
  --color-base-500: #64748b;
  --color-base-600: #475569;
  --color-base-700: #334155;
  --color-base-800: #1e293b;
  --color-base-900: #0f172a;
  --color-base-950: #020617;

  /* Primary brand color */
  --color-primary: #6366f1;
  --color-primary-light: #818cf8;
  --color-primary-dark: #4f46e5;

  /* Accent colors */
  --color-accent-primary: #8b5cf6;
  --color-accent-secondary: #06b6d4;

  /* Semantic colors */
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-error: #ef4444;

  /* Text colors */
  --color-text: #f8fafc;
  --color-text-muted: #94a3b8;
}
```

### Typography

```css
:root {
  /* Font families */
  --font-sans: 'Inter', system-ui, sans-serif;
  --font-display: 'Inter', system-ui, sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
}
```

## Logo & Assets

Replace these in `app/static/`:

```
app/static/
├── favicon.png      # Browser tab icon (32x32 or 64x64)
├── favicon.svg      # SVG favicon (preferred)
├── logo.svg         # Site logo (navigation)
├── logo-dark.svg    # Logo for light backgrounds
└── og-image.png     # Social sharing image (1200x630)
```

## Products & Pricing

Products are defined in `etc/gobot.yaml`:

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
        amount: 2900        # $29.00 in cents
        currency: "usd"
        interval: "month"
        trialDays: 14
      - slug: "yearly"
        amount: 29000       # $290.00
        currency: "usd"
        interval: "year"
        trialDays: 14
```

Pricing is:
- Read at build time and embedded in static HTML (no API call needed)
- Synced to Stripe on startup (standalone mode)
- Available at `/pricing` automatically

## Landing Page

### Hero Section

Edit `app/src/routes/(www)/+page.svelte`:

```svelte
<section class="hero-section">
  <h1 class="hero-title">
    Your Compelling<br />
    <span class="text-gradient">Headline Here</span>
  </h1>
  <p class="hero-subtitle">
    Your value proposition goes here. Explain what makes your product unique.
  </p>
</section>
```

### Features

```svelte
<script lang="ts">
  import { Zap } from 'lucide-svelte';

  const features = [
    {
      icon: Zap,
      title: 'Feature One',
      description: 'Explain the benefit to users.',
      color: 'var(--color-primary)'
    },
    // Add more features...
  ];
</script>
```

### Testimonials

```svelte
<div class="quote-card">
  <Quote class="w-8 h-8 text-primary mb-4" />
  <p class="quote-text">"Your customer testimonial here."</p>
  <div class="quote-author">
    <div class="quote-avatar">JD</div>
    <div>
      <p class="quote-name">Jane Doe</p>
      <p class="quote-role">CEO, Company</p>
    </div>
  </div>
</div>
```

## Navigation

### Header Menu

Edit `app/src/routes/(www)/+layout.svelte`:

```svelte
<script lang="ts">
  const navItems = [
    { label: 'Pricing', href: '/pricing' },
    { label: 'Features', href: '/features' },
    { label: 'Blog', href: '/blog' },
  ];
</script>
```

### Footer Links

Edit `app/src/lib/components/Footer.svelte`:

```svelte
<script lang="ts">
  const footerLinks = {
    product: [
      { label: 'Pricing', href: '/pricing' },
      { label: 'Features', href: '/features' },
    ],
    legal: [
      { label: 'Privacy Policy', href: '/privacy' },
      { label: 'Terms of Service', href: '/terms' },
    ]
  };
</script>
```

## SEO Configuration

SEO uses values from `site.ts`. For per-page customization:

```svelte
<script lang="ts">
  import { site } from '$lib/config/site';
</script>

<svelte:head>
  <title>Page Title - {site.name}</title>
  <meta name="description" content="Page-specific description" />
  <meta property="og:title" content="Page Title - {site.name}" />
  <meta property="og:description" content="Page-specific description" />
  <meta property="og:image" content="{site.url}{site.ogImage}" />
</svelte:head>
```

## Adding New Pages

### Marketing Page (Public)

Create `app/src/routes/(www)/your-page/+page.svelte`:

```svelte
<script lang="ts">
  import { site } from '$lib/config/site';
  import { Button } from '$lib/components/ui';
</script>

<svelte:head>
  <title>Your Page - {site.name}</title>
  <meta name="description" content="Page description" />
</svelte:head>

<section class="hero-section">
  <h1 class="hero-title">Your Page Title</h1>
  <p class="hero-subtitle">Your page content</p>
</section>
```

### Authenticated Page (App)

Create in `app/src/routes/(app)/app/`:

```svelte
<!-- app/src/routes/(app)/app/your-feature/+page.svelte -->
<script lang="ts">
  import { currentUser } from '$lib/stores/auth';
</script>

{#if $currentUser}
  <h1>Welcome, {$currentUser.name}</h1>
{/if}
```

## Components

### UI Components

Located in `app/src/lib/components/ui/`:

- `Button` - Primary, secondary, danger, ghost variants
- `Card` - Container with optional click handling
- `Input` - Form input with sizes
- `Select` - Dropdown select
- `Modal` - Dialog component
- `Badge` - Status badges
- `Tooltip` - Hover tooltips
- And more...

### Using Components

```svelte
<script lang="ts">
  import { Button, Card, Input, Modal } from '$lib/components/ui';

  let name = $state('');

  function submit() {
    // Handle submission
  }
</script>

<Card title="Card Title">
  <Input bind:value={name} placeholder="Enter name" />
  <Button variant="primary" onclick={submit}>Submit</Button>
</Card>
```

### Component Styling

Component styles are in `app/src/app.css`:

```css
/* Buttons */
.btn-primary { ... }
.btn-secondary { ... }

/* Cards */
.card { ... }
.card-header { ... }

/* Forms */
.input { ... }
.select { ... }
```

**Important**: Never use inline styles or `<style>` blocks in Svelte files. All styles go in `app.css`.

## API Customization

### Adding Endpoints

1. Define in `gobot.api`:

```go
type YourRequest {
    Field string `json:"field"`
}

type YourResponse {
    Result string `json:"result"`
}

@server(
    prefix: /api/v1
    middleware: JwtAuth
)
service gobot {
    @handler YourHandler
    post /your-endpoint (YourRequest) returns (YourResponse)
}
```

2. Generate code:

```bash
make gen
```

3. Implement logic in `internal/logic/yourhandlerlogic.go`

4. Frontend types are auto-generated in `app/src/lib/api/`

## Custom Domains

### Development

No changes needed - runs on `localhost:5173`.

### Production

1. Point DNS to your server
2. Set environment variables:

```bash
PRODUCTION_MODE=true
APP_DOMAIN=myapp.com
APP_ADMIN_EMAIL=admin@myapp.com
```

3. SSL certificates are obtained automatically via Let's Encrypt

See [Deployment Guide](./DEPLOYMENT.md) for details.

## Analytics Integration

### Google Analytics

Add to `app/src/routes/+layout.svelte`:

```svelte
<svelte:head>
  {#if browser}
    <script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXX"></script>
    <script>
      window.dataLayer = window.dataLayer || [];
      function gtag(){dataLayer.push(arguments);}
      gtag('js', new Date());
      gtag('config', 'G-XXXXXXX');
    </script>
  {/if}
</svelte:head>
```

### Plausible (Privacy-Friendly)

```svelte
<svelte:head>
  <script defer data-domain="myapp.com" src="https://plausible.io/js/script.js"></script>
</svelte:head>
```

## Email Templates (Levee Mode)

When using Levee, email templates are managed in the Levee Dashboard:

1. Go to Levee Dashboard > Emails > Templates
2. Create templates for:
   - `welcome` - New user welcome
   - `password-reset` - Password reset link
   - `email-verification` - Verify email address

### Template Variables

```
{{ .Name }}        - User's name
{{ .Email }}       - User's email
{{ .Link }}        - Action link (verify, reset, etc.)
{{ .CompanyName }} - Your company name
```

## Feature Gating (Svelte 5)

Control access based on subscription plan:

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
