# Deployment Guide

Gobot deploys as a **single binary with automatic SSL** - no nginx, Apache, or reverse proxy required.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Gobot Binary                            │
│  ┌──────────────────┐  ┌────────────────────────────────┐   │
│  │   HTTP Server    │  │         HTTPS Server           │   │
│  │     :80          │  │           :443                 │   │
│  │  ┌────────────┐  │  │  ┌──────────────────────────┐  │   │
│  │  │ ACME       │  │  │  │  Go-Zero API Server      │  │   │
│  │  │ challenges │  │  │  │  (/api/* endpoints)      │  │   │
│  │  ├────────────┤  │  │  ├──────────────────────────┤  │   │
│  │  │ HTTP→HTTPS │  │  │  │  Embedded SvelteKit SPA  │  │   │
│  │  │ redirect   │  │  │  │  (static files)          │  │   │
│  │  └────────────┘  │  │  └──────────────────────────┘  │   │
│  └──────────────────┘  │  + Let's Encrypt auto-SSL      │   │
│                        │  + HTTP/2 support              │   │
│                        │  + Gzip compression            │   │
│                        └────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## No External Reverse Proxy Needed

Gobot handles everything internally:

- **SSL/TLS Termination** - Automatic Let's Encrypt certificates
- **Certificate Renewal** - Auto-renewed before expiry
- **HTTP to HTTPS Redirect** - All port 80 traffic → port 443
- **www to non-www Redirect** - `www.example.com` → `example.com`
- **Static File Serving** - Frontend embedded in binary
- **Gzip Compression** - Automatic for text/html/css/js
- **HTTP/2 Support** - Modern protocol for better performance
- **WebSocket Support** - Full duplex communication

You do **NOT** need:
- nginx
- Apache
- Caddy
- HAProxy
- Any reverse proxy

## Quick Deploy

### 1. Build the Binary

```bash
# Build Go binary
make build

# Build and embed frontend
cd app && pnpm install && pnpm build && cd ..

# Result: single binary at ./bin/gobot
```

### 2. Configure Production Environment

Create `.env` on your server:

```bash
# Required for production
PRODUCTION_MODE=true
APP_DOMAIN=myapp.com           # Your domain (without www)
APP_BASE_URL=https://myapp.com
APP_ADMIN_EMAIL=admin@myapp.com  # For Let's Encrypt notifications

# JWT Secret
ACCESS_SECRET=your-256-bit-secret-here  # openssl rand -hex 32

# Choose ONE mode:

# Option A: Standalone (SQLite + Stripe)
LEVEE_ENABLED=false
SQLITE_PATH=/data/gobot.db
STRIPE_SECRET_KEY=sk_live_xxx
STRIPE_PUBLISHABLE_KEY=pk_live_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx

# Option B: Levee
LEVEE_ENABLED=true
LEVEE_API_KEY=lvk_xxx
```

### 3. Point DNS to Your Server

Add A/AAAA records pointing to your server IP:

```
myapp.com     A     123.45.67.89
www.myapp.com A     123.45.67.89
```

### 4. Run the Binary

```bash
./gobot
```

On first request to your domain, Let's Encrypt will automatically issue a certificate.

## Auto-SSL Details

### How It Works

1. When `PRODUCTION_MODE=true` and `APP_DOMAIN` is set:
   - HTTP server starts on port 80 (ACME challenges + redirect)
   - HTTPS server starts on port 443 (main application)

2. On first HTTPS request:
   - `autocert.Manager` contacts Let's Encrypt
   - ACME challenge served on port 80
   - Certificate issued and cached in `./certs/` directory

3. Automatic renewal:
   - Certificates checked on each TLS handshake
   - Auto-renewed ~30 days before expiry

### Certificate Storage

Certificates are cached in `./certs/` directory (created automatically):

```
certs/
├── myapp.com          # Certificate + private key
└── www.myapp.com      # Certificate for www variant
```

**Important**: Back up the `certs/` directory to avoid rate limit issues with Let's Encrypt.

### Domain Whitelisting

Only the configured domain and its www variant receive certificates:

```go
HostPolicy: autocert.HostWhitelist(c.App.Domain, "www."+c.App.Domain)
```

Requests to other domains will be rejected.

## Deployment Options

### Option 1: Direct Binary (Recommended for VPS)

Upload and run the binary directly:

```bash
# On your server
scp ./bin/gobot .env user@server:/app/
ssh user@server
cd /app
./gobot
```

### Option 2: Systemd Service

Create `/etc/systemd/system/gobot.service`:

```ini
[Unit]
Description=Gobot Application
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/app
ExecStart=/app/gobot
Restart=always
RestartSec=5
EnvironmentFile=/app/.env

# Give time for graceful shutdown
TimeoutStopSec=30

# Allow binding to ports 80 and 443
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable gobot
sudo systemctl start gobot
sudo journalctl -u gobot -f  # View logs
```

### Option 3: Docker

```dockerfile
# Already included Dockerfile
docker build -t myapp .
docker run -d \
  --name myapp \
  -p 80:80 \
  -p 443:443 \
  -v /data/certs:/app/certs \
  -v /data/db:/app/data \
  --env-file .env \
  myapp
```

### Option 4: Docker Compose (Production)

```bash
docker compose --profile production up -d --build
```

## Platform-Specific Guides

### Fly.io

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Launch app
fly launch

# Set secrets
fly secrets set ACCESS_SECRET=xxx STRIPE_SECRET_KEY=sk_live_xxx ...

# Deploy
fly deploy
```

Note: Fly.io handles SSL at the edge, so set `PRODUCTION_MODE=false` and let Fly terminate SSL.

### Railway

1. Connect your GitHub repo
2. Add environment variables in Railway dashboard
3. Deploy automatically on push

Railway handles SSL, so `PRODUCTION_MODE=false`.

### DigitalOcean Droplet

1. Create Ubuntu droplet
2. Install Go (if building on server) or upload binary
3. Configure firewall:
   ```bash
   ufw allow 80
   ufw allow 443
   ufw allow 22
   ufw enable
   ```
4. Set up systemd service (see above)
5. Run with `PRODUCTION_MODE=true`

### AWS EC2

1. Launch EC2 instance (Ubuntu recommended)
2. Configure security group:
   - Inbound: 80, 443, 22 (SSH)
3. Point domain to Elastic IP
4. Upload binary and run with systemd

## Security Checklist

Before going live:

- [ ] `ACCESS_SECRET` is a strong random value (not the default)
- [ ] `PRODUCTION_MODE=true` is set
- [ ] `APP_DOMAIN` matches your actual domain
- [ ] Firewall only allows ports 80, 443, 22
- [ ] Database is backed up (if using SQLite)
- [ ] `certs/` directory is backed up
- [ ] Stripe is using live keys (not test keys)
- [ ] Webhook secrets are configured

## Troubleshooting

### SSL Certificate Not Issued

1. Check DNS is pointing to correct IP: `dig myapp.com`
2. Ensure ports 80 and 443 are open
3. Check `APP_DOMAIN` matches exactly (no `https://`)
4. Review Let's Encrypt rate limits: 50 certs/domain/week

### Binary Won't Bind to Port 80/443

On Linux, binding to ports below 1024 requires root or capabilities:

```bash
# Option 1: Run as root (not recommended)
sudo ./gobot

# Option 2: Add capability (recommended)
sudo setcap 'cap_net_bind_service=+ep' ./gobot

# Option 3: Use systemd with AmbientCapabilities (see above)
```

### Connection Refused

1. Check firewall: `ufw status`
2. Check server is running: `ps aux | grep gobot`
3. Check logs for errors
4. Verify `.env` has correct values

### Database Errors (Standalone Mode)

Ensure `SQLITE_PATH` directory exists and is writable:

```bash
mkdir -p /data
chown www-data:www-data /data
```

## Monitoring

### Health Check Endpoint

The app exposes a health check at `/api/v1/health`:

```bash
curl https://myapp.com/api/v1/health
# {"status": "ok"}
```

### Log Output

In production, the app logs to stdout:

```
Starting HTTPS server on :443 with Let's Encrypt auto-certificate...
Auto-redirect: www → non-www
Auto-redirect: HTTP → HTTPS
API routes: /api/* (proxied to backend)
Static SPA: /* (served directly from embedded FS)
Optimizations: HTTP/2, connection pooling, compression
```

### Uptime Monitoring

Use external services like:
- UptimeRobot
- Pingdom
- StatusCake

Point them at your health check endpoint.

## Graceful Shutdown

The server handles `SIGINT` and `SIGTERM` signals gracefully:

1. Stops accepting new connections
2. Waits up to 30 seconds for active requests
3. Closes database connections
4. Exits cleanly

This ensures zero downtime during deployments with proper orchestration.
