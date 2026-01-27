#!/bin/bash
#
# Gobot Installer & Setup Script
# ================================
#
# This single script handles both remote installation and local setup.
#
# REMOTE INSTALLATION (via curl):
#   curl -fsSL https://raw.githubusercontent.com/almatuck/gobot/main/install.sh | bash -s -- myapp
#
# LOCAL SETUP (after git clone):
#   ./install.sh myapp
#
# The script auto-detects which mode to run based on context.
#

set -e

# ============================================================
# Configuration
# ============================================================
OLD_NAME="gobot"
OLD_NAME_PROPER="Gobot"
VERSION="1.0.0"
REPO_URL="https://github.com/almatuck/gobot.git"

# ============================================================
# Colors & Output Helpers
# ============================================================
if [[ -t 1 ]]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  BLUE='\033[0;34m'
  BOLD='\033[1m'
  NC='\033[0m'
else
  RED='' GREEN='' YELLOW='' BLUE='' BOLD='' NC=''
fi

print_step()    { echo -e "${BLUE}▶${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_error()   { echo -e "${RED}✗${NC} $1"; }
print_info()    { echo -e "  $1"; }

die() { print_error "$1"; exit 1; }

# ============================================================
# Help
# ============================================================
show_help() {
  cat << EOF
${BOLD}Gobot Installer v${VERSION}${NC}

${BOLD}REMOTE INSTALLATION:${NC}
  curl -fsSL https://raw.githubusercontent.com/almatuck/gobot/main/install.sh | bash -s -- <app-name>

${BOLD}LOCAL SETUP (after git clone):${NC}
  ./install.sh <app-name> [options]

${BOLD}ARGUMENTS:${NC}
  app-name      Your app name (lowercase, letters and numbers only)
                Example: myapp, coolsaas, app123

${BOLD}OPTIONS:${NC}
  --admin-email EMAIL    Admin email address (required, will prompt if not provided)
  --admin-password PASS  Admin password (min 8 chars, will prompt if not provided)
  --levee                Enable Levee mode (will prompt for API key and URL)
  --levee-key KEY        Levee API key (requires --levee)
  --levee-url URL        Levee base URL (default: https://api.levee.sh)
  --smtp                 Enable SMTP email (will prompt for settings)
  --smtp-host HOST       SMTP host (e.g., smtp.gmail.com)
  --smtp-user USER       SMTP username
  --smtp-pass PASS       SMTP password
  --oauth                Enable OAuth login (will prompt for Google/GitHub credentials)
  --google-id ID         Google OAuth client ID
  --google-secret SECRET Google OAuth client secret
  --github-id ID         GitHub OAuth client ID
  --github-secret SECRET GitHub OAuth client secret
  --no-orgs              Disable team/organization features
  --no-notifications     Disable in-app notifications
  --dir PATH             Install directory (remote mode only, default: current dir)
  --no-deps              Skip installing dependencies
  --no-env               Skip creating .env file
  --no-start             Don't offer to start the app after setup
  --force                Overwrite existing .env file
  --help                 Show this help message

${BOLD}EXAMPLES:${NC}
  # Remote: Install in ./myapp
  curl ... | bash -s -- myapp

  # Remote: Install in specific directory
  curl ... | bash -s -- myapp --dir ~/projects

  # Local: Set up after manual clone
  git clone https://github.com/almatuck/gobot.git myapp
  cd myapp
  ./install.sh myapp

  # Local: Just rename, skip deps
  ./install.sh myapp --no-deps

${BOLD}WHAT IT DOES:${NC}
  1. Checks prerequisites (Go, Node.js, pnpm)
  2. Clones repository (remote mode only)
  3. Renames project from 'gobot' to your app name
  4. Generates secure JWT secret
  5. Creates .env with working defaults
  6. Installs all dependencies
  7. Initializes git repository

EOF
  exit 0
}

# ============================================================
# Parse Arguments
# ============================================================
NEW_NAME=""
INSTALL_DIR=""
INSTALL_DEPS=true
CREATE_ENV=true
FORCE_ENV=false
OFFER_START=true
ADMIN_EMAIL=""
ADMIN_PASSWORD=""
USE_LEVEE=false
LEVEE_API_KEY=""
LEVEE_BASE_URL="https://api.levee.sh"
USE_SMTP=false
SMTP_HOST=""
SMTP_PORT="587"
SMTP_USER=""
SMTP_PASS=""

# OAuth settings
USE_OAUTH=false
GOOGLE_CLIENT_ID=""
GOOGLE_CLIENT_SECRET=""
GITHUB_CLIENT_ID=""
GITHUB_CLIENT_SECRET=""

# Feature toggles
ENABLE_ORGANIZATIONS=true
ENABLE_NOTIFICATIONS=true

while [[ $# -gt 0 ]]; do
  case $1 in
    --help|-h)
      show_help
      ;;
    --admin-email)
      ADMIN_EMAIL="$2"
      shift 2
      ;;
    --admin-password)
      ADMIN_PASSWORD="$2"
      shift 2
      ;;
    --levee)
      USE_LEVEE=true
      shift
      ;;
    --levee-key)
      LEVEE_API_KEY="$2"
      shift 2
      ;;
    --levee-url)
      LEVEE_BASE_URL="$2"
      shift 2
      ;;
    --smtp)
      USE_SMTP=true
      shift
      ;;
    --smtp-host)
      SMTP_HOST="$2"
      USE_SMTP=true
      shift 2
      ;;
    --smtp-user)
      SMTP_USER="$2"
      USE_SMTP=true
      shift 2
      ;;
    --smtp-pass)
      SMTP_PASS="$2"
      USE_SMTP=true
      shift 2
      ;;
    --oauth)
      USE_OAUTH=true
      shift
      ;;
    --google-id)
      GOOGLE_CLIENT_ID="$2"
      USE_OAUTH=true
      shift 2
      ;;
    --google-secret)
      GOOGLE_CLIENT_SECRET="$2"
      USE_OAUTH=true
      shift 2
      ;;
    --github-id)
      GITHUB_CLIENT_ID="$2"
      USE_OAUTH=true
      shift 2
      ;;
    --github-secret)
      GITHUB_CLIENT_SECRET="$2"
      USE_OAUTH=true
      shift 2
      ;;
    --no-orgs)
      ENABLE_ORGANIZATIONS=false
      shift
      ;;
    --no-notifications)
      ENABLE_NOTIFICATIONS=false
      shift
      ;;
    --dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    --no-deps)
      INSTALL_DEPS=false
      shift
      ;;
    --no-env)
      CREATE_ENV=false
      shift
      ;;
    --no-start)
      OFFER_START=false
      shift
      ;;
    --force)
      FORCE_ENV=true
      shift
      ;;
    -*)
      die "Unknown option: $1 (use --help for usage)"
      ;;
    *)
      if [[ -z "$NEW_NAME" ]]; then
        NEW_NAME="$1"
      else
        die "Too many arguments. Use --help for usage."
      fi
      shift
      ;;
  esac
done

# ============================================================
# Detect Mode: Remote (curl) vs Local (already cloned)
# ============================================================
# We're in "local mode" if we can find gobot project files
if [[ -f "go.mod" ]] && [[ -f "${OLD_NAME}.go" || -f "${OLD_NAME}.api" || -d "internal" ]]; then
  MODE="local"
else
  MODE="remote"
fi

# ============================================================
# Validation
# ============================================================
if [[ -z "$NEW_NAME" ]]; then
  echo ""
  echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
  echo -e "${GREEN}║${NC}              🚀 Gobot - Ship Your SaaS Fast              ${GREEN}║${NC}"
  echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
  echo ""
  echo -e "${BOLD}Usage:${NC}"
  echo ""
  if [[ "$MODE" == "remote" ]]; then
    echo "  curl -fsSL https://raw.githubusercontent.com/almatuck/gobot/main/install.sh | bash -s -- <app-name>"
  else
    echo "  ./install.sh <app-name>"
  fi
  echo ""
  echo -e "${BOLD}Example:${NC}"
  echo "  ${MODE == 'remote' && 'curl ... | bash -s -- myapp' || './install.sh myapp'}"
  echo ""
  echo "Run with --help for more options."
  exit 1
fi

# Validate app name
if [[ ! "$NEW_NAME" =~ ^[a-z][a-z0-9]*$ ]]; then
  die "Invalid app name: '$NEW_NAME'

App name must:
  • Start with a lowercase letter
  • Contain only lowercase letters and numbers
  • Be a valid Go package name

Examples: myapp, coolsaas, app123"
fi

# ============================================================
# OS Detection for sed compatibility
# ============================================================
if [[ "$OSTYPE" == "darwin"* ]]; then
  sed_inplace() { sed -i '' "$@"; }
else
  sed_inplace() { sed -i "$@"; }
fi

# ============================================================
# Start
# ============================================================
echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║${NC}              🚀 Gobot - Ship Your SaaS Fast              ${GREEN}║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${BOLD}App name:${NC} $NEW_NAME"
echo -e "  ${BOLD}Mode:${NC}     $MODE"
echo ""

# ============================================================
# Step 1: Collect Admin Credentials (before any file operations)
# ============================================================
# Collect these FIRST so user can Ctrl+C and retry without cleanup
if [[ "$CREATE_ENV" == true ]]; then
  print_step "Admin account setup (for /backoffice access)"
  echo -e "  ${YELLOW}Note: Password is stored in plain text in .env${NC}"
  echo -e "  ${YELLOW}This is safe as long as .env remains in your docker compose env_file${NC}"
  echo -e "  ${YELLOW}and is not committed to git (already in .gitignore)${NC}"
  echo ""

  # Validate email provided via flag
  if [[ -n "$ADMIN_EMAIL" ]]; then
    if [[ ! "$ADMIN_EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
      print_warning "Invalid email format: $ADMIN_EMAIL"
      ADMIN_EMAIL=""  # Clear to trigger interactive prompt
    fi
  fi

  # Validate password provided via flag
  if [[ -n "$ADMIN_PASSWORD" ]]; then
    if [[ ${#ADMIN_PASSWORD} -lt 8 ]]; then
      print_warning "Password too short (min 8 chars)"
      ADMIN_PASSWORD=""  # Clear to trigger interactive prompt
    fi
  fi

  # Prompt for admin email if not provided or invalid
  if [[ -z "$ADMIN_EMAIL" ]]; then
    echo -e "  Admin username must be a valid email address"
    echo ""
    while true; do
      read -p "  Enter admin email: " ADMIN_EMAIL < /dev/tty
      if [[ "$ADMIN_EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        break
      else
        print_warning "Please enter a valid email address"
      fi
    done
  fi

  # Prompt for admin password if not provided or invalid
  if [[ -z "$ADMIN_PASSWORD" ]]; then
    while true; do
      read -s -p "  Enter admin password (min 8 chars): " ADMIN_PASSWORD < /dev/tty
      echo ""
      if [[ ${#ADMIN_PASSWORD} -ge 8 ]]; then
        read -s -p "  Confirm password: " ADMIN_PASSWORD_CONFIRM < /dev/tty
        echo ""
        if [[ "$ADMIN_PASSWORD" == "$ADMIN_PASSWORD_CONFIRM" ]]; then
          break
        else
          print_warning "Passwords do not match"
        fi
      else
        print_warning "Password must be at least 8 characters"
      fi
    done
  fi

  print_success "Admin credentials collected"
  echo ""

  # Ask about Levee mode
  if [[ "$USE_LEVEE" != true ]]; then
    echo -e "  ${BOLD}Mode selection:${NC}"
    echo -e "  • ${GREEN}Standalone${NC} (default): SQLite + direct Stripe - zero dependencies"
    echo -e "  • ${BLUE}Levee${NC}: Managed auth, billing, email, AI gateway"
    echo ""
    read -p "  Enable Levee mode? [y/N] " LEVEE_CHOICE < /dev/tty
    if [[ "$LEVEE_CHOICE" =~ ^[Yy]$ ]]; then
      USE_LEVEE=true
    fi
  fi

  # If using Levee, collect API key
  if [[ "$USE_LEVEE" == true ]]; then
    echo ""
    print_step "Levee configuration"
    echo -e "  Get your API key from ${BLUE}https://dashboard.levee.sh${NC}"
    echo ""

    # Prompt for API key if not provided
    if [[ -z "$LEVEE_API_KEY" ]]; then
      while true; do
        read -p "  Enter Levee API key (lvk_...): " LEVEE_API_KEY < /dev/tty
        if [[ "$LEVEE_API_KEY" =~ ^lvk_ ]]; then
          break
        else
          print_warning "API key should start with 'lvk_'"
        fi
      done
    elif [[ ! "$LEVEE_API_KEY" =~ ^lvk_ ]]; then
      print_warning "Invalid API key format: $LEVEE_API_KEY"
      while true; do
        read -p "  Enter Levee API key (lvk_...): " LEVEE_API_KEY < /dev/tty
        if [[ "$LEVEE_API_KEY" =~ ^lvk_ ]]; then
          break
        else
          print_warning "API key should start with 'lvk_'"
        fi
      done
    fi

    # Ask about custom URL (most users will use default)
    read -p "  Levee URL [${LEVEE_BASE_URL}]: " LEVEE_URL_INPUT < /dev/tty
    if [[ -n "$LEVEE_URL_INPUT" ]]; then
      LEVEE_BASE_URL="$LEVEE_URL_INPUT"
    fi

    print_success "Levee configured"
    echo ""
  fi

  # Ask about SMTP email (skip if using Levee - Levee handles email)
  if [[ "$USE_LEVEE" != true ]] && [[ "$USE_SMTP" != true ]]; then
    echo -e "  ${BOLD}Email configuration (optional):${NC}"
    echo -e "  SMTP is needed for password reset and welcome emails"
    echo ""
    read -p "  Configure SMTP now? [y/N] " SMTP_CHOICE < /dev/tty
    if [[ "$SMTP_CHOICE" =~ ^[Yy]$ ]]; then
      USE_SMTP=true
    fi
  fi

  if [[ "$USE_SMTP" == true ]]; then
    echo ""
    print_step "SMTP configuration"
    echo -e "  ${YELLOW}Press Enter to skip any field${NC}"
    echo ""

    # SMTP Host
    if [[ -z "$SMTP_HOST" ]]; then
      read -p "  SMTP host (e.g., smtp.gmail.com): " SMTP_HOST < /dev/tty
    fi

    # SMTP Port
    read -p "  SMTP port [587]: " SMTP_PORT_INPUT < /dev/tty
    if [[ -n "$SMTP_PORT_INPUT" ]]; then
      SMTP_PORT="$SMTP_PORT_INPUT"
    fi

    # SMTP User
    if [[ -z "$SMTP_USER" ]]; then
      read -p "  SMTP username: " SMTP_USER < /dev/tty
    fi

    # SMTP Password
    if [[ -z "$SMTP_PASS" ]]; then
      read -s -p "  SMTP password: " SMTP_PASS < /dev/tty
      echo ""
    fi

    if [[ -n "$SMTP_HOST" ]]; then
      print_success "SMTP configured"
    else
      print_warning "SMTP skipped (can configure later in .env)"
      USE_SMTP=false
    fi
    echo ""
  fi

  # Ask about OAuth social login (skip if using Levee - Levee handles OAuth)
  if [[ "$USE_LEVEE" != true ]] && [[ "$USE_OAUTH" != true ]]; then
    echo -e "  ${BOLD}OAuth / Social Login (optional):${NC}"
    echo -e "  Allow users to sign in with Google and/or GitHub"
    echo ""
    read -p "  Configure OAuth now? [y/N] " OAUTH_CHOICE < /dev/tty
    if [[ "$OAUTH_CHOICE" =~ ^[Yy]$ ]]; then
      USE_OAUTH=true
    fi
  fi

  if [[ "$USE_OAUTH" == true ]]; then
    echo ""
    print_step "OAuth configuration"
    echo -e "  ${YELLOW}Press Enter to skip any provider${NC}"
    echo ""

    # Google OAuth
    echo -e "  ${BOLD}Google OAuth${NC} (from https://console.cloud.google.com/apis/credentials)"
    if [[ -z "$GOOGLE_CLIENT_ID" ]]; then
      read -p "  Google Client ID: " GOOGLE_CLIENT_ID < /dev/tty
    fi
    if [[ -n "$GOOGLE_CLIENT_ID" ]] && [[ -z "$GOOGLE_CLIENT_SECRET" ]]; then
      read -s -p "  Google Client Secret: " GOOGLE_CLIENT_SECRET < /dev/tty
      echo ""
    fi

    echo ""

    # GitHub OAuth
    echo -e "  ${BOLD}GitHub OAuth${NC} (from https://github.com/settings/developers)"
    if [[ -z "$GITHUB_CLIENT_ID" ]]; then
      read -p "  GitHub Client ID: " GITHUB_CLIENT_ID < /dev/tty
    fi
    if [[ -n "$GITHUB_CLIENT_ID" ]] && [[ -z "$GITHUB_CLIENT_SECRET" ]]; then
      read -s -p "  GitHub Client Secret: " GITHUB_CLIENT_SECRET < /dev/tty
      echo ""
    fi

    if [[ -n "$GOOGLE_CLIENT_ID" ]] || [[ -n "$GITHUB_CLIENT_ID" ]]; then
      print_success "OAuth configured"
    else
      print_warning "OAuth skipped (can configure later in .env)"
      USE_OAUTH=false
    fi
    echo ""
  fi
fi

# ============================================================
# Step 2: Check Prerequisites
# ============================================================
print_step "Checking prerequisites..."

MISSING=""
WARNINGS=""

# Check Go
if command -v go &> /dev/null; then
  GO_VERSION=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')
  print_success "Go $GO_VERSION"
else
  MISSING="$MISSING Go"
fi

# Check Node.js
if command -v node &> /dev/null; then
  NODE_VERSION=$(node --version 2>/dev/null)
  print_success "Node.js $NODE_VERSION"
else
  MISSING="$MISSING Node.js"
fi

# Check pnpm (try to install if missing)
if command -v pnpm &> /dev/null; then
  PNPM_VERSION=$(pnpm --version 2>/dev/null)
  print_success "pnpm $PNPM_VERSION"
elif command -v npm &> /dev/null; then
  print_warning "pnpm not found - will install via npm"
  WARNINGS="$WARNINGS pnpm"
else
  MISSING="$MISSING pnpm"
fi

# Check git
if command -v git &> /dev/null; then
  print_success "git $(git --version 2>/dev/null | awk '{print $3}')"
else
  MISSING="$MISSING git"
fi

if [[ -n "$MISSING" ]]; then
  echo ""
  print_error "Missing required tools:$MISSING"
  echo ""
  echo "Please install:"
  [[ "$MISSING" == *"Go"* ]] && echo "  • Go: https://go.dev/dl/"
  [[ "$MISSING" == *"Node"* ]] && echo "  • Node.js: https://nodejs.org/"
  [[ "$MISSING" == *"pnpm"* ]] && echo "  • pnpm: npm install -g pnpm"
  [[ "$MISSING" == *"git"* ]] && echo "  • git: https://git-scm.com/"
  exit 1
fi

# Install pnpm if needed
if [[ "$WARNINGS" == *"pnpm"* ]]; then
  print_info "Installing pnpm..."
  if npm install -g pnpm 2>/dev/null; then
    print_success "pnpm installed"
  else
    print_warning "Could not install pnpm globally (may need sudo)"
  fi
fi

echo ""

# ============================================================
# Step 2: Clone Repository (Remote Mode Only)
# ============================================================
if [[ "$MODE" == "remote" ]]; then
  print_step "Cloning Gobot repository..."

  # Determine target directory
  if [[ -n "$INSTALL_DIR" ]]; then
    TARGET="$INSTALL_DIR/$NEW_NAME"
  else
    TARGET="./$NEW_NAME"
  fi

  # Check if target exists
  if [[ -d "$TARGET" ]]; then
    die "Directory already exists: $TARGET

Remove it first or choose a different name."
  fi

  # Create parent directory if needed
  if [[ -n "$INSTALL_DIR" ]]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || die "Cannot create directory: $INSTALL_DIR"
  fi

  # Clone
  if ! git clone --depth 1 "$REPO_URL" "$TARGET" 2>/dev/null; then
    die "Failed to clone repository.

Check your internet connection and try again."
  fi

  # Enter directory
  cd "$TARGET" || die "Cannot enter directory: $TARGET"

  # Remove git history (fresh start)
  rm -rf .git

  print_success "Repository cloned to $(pwd)"
  echo ""
fi

# ============================================================
# Step 3: Verify We're in a Gobot Project
# ============================================================
if [[ ! -f "go.mod" ]]; then
  die "Not in a Gobot project directory.

Expected to find: go.mod

Make sure you're running this from the project root."
fi

# Check if already set up with a different name
CURRENT_MODULE=$(grep "^module " go.mod 2>/dev/null | awk '{print $2}')
if [[ "$CURRENT_MODULE" != "$OLD_NAME" ]] && [[ "$CURRENT_MODULE" != "$NEW_NAME" ]]; then
  die "Project already renamed to '$CURRENT_MODULE'.

To use a different name, start fresh:
  rm -rf $(pwd)
  curl -fsSL $REPO_URL/install.sh | bash -s -- $NEW_NAME"
fi

# ============================================================
# Step 4: Rename Project Files
# ============================================================

# Generate proper case name (capitalize first letter)
NEW_NAME_PROPER="$(echo "${NEW_NAME:0:1}" | tr '[:lower:]' '[:upper:]')${NEW_NAME:1}"

if [[ -f "${OLD_NAME}.go" ]] || [[ -f "${OLD_NAME}.api" ]] || [[ -f "etc/${OLD_NAME}.yaml" ]]; then
  print_step "Renaming project to '$NEW_NAME'..."

  # Main files
  [[ -f "${OLD_NAME}.go" ]] && mv "${OLD_NAME}.go" "${NEW_NAME}.go" && print_info "${OLD_NAME}.go → ${NEW_NAME}.go"
  [[ -f "${OLD_NAME}.api" ]] && mv "${OLD_NAME}.api" "${NEW_NAME}.api" && print_info "${OLD_NAME}.api → ${NEW_NAME}.api"
  [[ -f "etc/${OLD_NAME}.yaml" ]] && mv "etc/${OLD_NAME}.yaml" "etc/${NEW_NAME}.yaml" && print_info "etc/${OLD_NAME}.yaml → etc/${NEW_NAME}.yaml"

  print_success "Files renamed"
else
  print_success "Files already renamed"
fi

# ============================================================
# Step 5: Replace ALL occurrences in ALL files
# ============================================================
print_step "Replacing all occurrences of 'gobot' and 'Gobot'..."

# Rename frontend API files first
if [[ -d "app/src/lib/api" ]]; then
  [[ -f "app/src/lib/api/${OLD_NAME}.ts" ]] && mv "app/src/lib/api/${OLD_NAME}.ts" "app/src/lib/api/${NEW_NAME}.ts"
  [[ -f "app/src/lib/api/${OLD_NAME}Components.ts" ]] && mv "app/src/lib/api/${OLD_NAME}Components.ts" "app/src/lib/api/${NEW_NAME}Components.ts"
fi

print_info "Updating all project files..."

# Helper function to process files of a given pattern
process_files() {
  local pattern="$1"
  find . -type f -name "$pattern" ! -path "./.git/*" ! -path "./vendor/*" ! -path "./node_modules/*" 2>/dev/null | while IFS= read -r file; do
    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' "s/${OLD_NAME_PROPER}/${NEW_NAME_PROPER}/g" "$file" 2>/dev/null || true
      sed -i '' "s/${OLD_NAME}/${NEW_NAME}/g" "$file" 2>/dev/null || true
    else
      sed -i "s/${OLD_NAME_PROPER}/${NEW_NAME_PROPER}/g" "$file" 2>/dev/null || true
      sed -i "s/${OLD_NAME}/${NEW_NAME}/g" "$file" 2>/dev/null || true
    fi
  done
}

# Process each file type separately (most reliable)
process_files "*.go"
process_files "*.api"
process_files "*.yaml"
process_files "*.yml"
process_files "*.md"
process_files "*.sql"
process_files "*.ts"
process_files "*.svelte"
process_files "*.css"
process_files "*.json"
process_files "*.toml"
process_files "*.sh"
process_files "*.webmanifest"
process_files "*.example"
process_files "Makefile"
process_files "Dockerfile"
process_files "go.mod"
process_files "go.sum"

print_success "All replacements complete"

# ============================================================
# Step 6: Configure Available Ports
# ============================================================
print_step "Configuring ports..."

# Function to check if port is in use
is_port_in_use() {
  local port=$1
  if command -v lsof &> /dev/null; then
    lsof -i :"$port" &>/dev/null && return 0
  elif command -v netstat &> /dev/null; then
    netstat -tuln 2>/dev/null | grep -q ":$port " && return 0
  elif command -v ss &> /dev/null; then
    ss -tuln 2>/dev/null | grep -q ":$port " && return 0
  fi
  return 1
}

# Function to generate random available port over 20000
random_available_port() {
  local port
  local attempts=0
  while [[ $attempts -lt 50 ]]; do
    port=$((20000 + RANDOM % 10000))
    if ! is_port_in_use $port; then
      echo "$port"
      return
    fi
    attempts=$((attempts + 1))
  done
  # Fallback if all random attempts fail
  echo "$port"
}

# Generate random non-sequential ports
BACKEND_PORT=$(random_available_port)
FRONTEND_PORT=$(random_available_port)
# Ensure they're different
while [[ $FRONTEND_PORT -eq $BACKEND_PORT ]]; do
  FRONTEND_PORT=$(random_available_port)
done

print_success "Backend port: $BACKEND_PORT"
print_success "Frontend port: $FRONTEND_PORT"

# Update compose.yaml with selected backend port (frontend runs locally, not in Docker)
if [[ -f "compose.yaml" ]]; then
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/8888:8888/${BACKEND_PORT}:8888/g" compose.yaml 2>/dev/null || true
    sed -i '' "s/localhost:5173/localhost:${FRONTEND_PORT}/g" compose.yaml 2>/dev/null || true
  else
    sed -i "s/8888:8888/${BACKEND_PORT}:8888/g" compose.yaml 2>/dev/null || true
    sed -i "s/localhost:5173/localhost:${FRONTEND_PORT}/g" compose.yaml 2>/dev/null || true
  fi
fi

# Update vite.config.ts with selected ports
if [[ -f "app/vite.config.ts" ]]; then
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/port: 5173/port: ${FRONTEND_PORT}/g" app/vite.config.ts 2>/dev/null || true
    sed -i '' "s/localhost:8888/localhost:${BACKEND_PORT}/g" app/vite.config.ts 2>/dev/null || true
  else
    sed -i "s/port: 5173/port: ${FRONTEND_PORT}/g" app/vite.config.ts 2>/dev/null || true
    sed -i "s/localhost:8888/localhost:${BACKEND_PORT}/g" app/vite.config.ts 2>/dev/null || true
  fi
fi

# NOTE: etc/*.yaml keeps Port: 8888 - that's the container's internal port
# The external port mapping is handled by compose.yaml (BACKEND_PORT:8888)

# Update Makefile with selected ports
if [[ -f "Makefile" ]]; then
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/localhost:8888/localhost:${BACKEND_PORT}/g" Makefile 2>/dev/null || true
    sed -i '' "s/localhost:5173/localhost:${FRONTEND_PORT}/g" Makefile 2>/dev/null || true
    sed -i '' "s/-p 8888:8888/-p ${BACKEND_PORT}:${BACKEND_PORT}/g" Makefile 2>/dev/null || true
  else
    sed -i "s/localhost:8888/localhost:${BACKEND_PORT}/g" Makefile 2>/dev/null || true
    sed -i "s/localhost:5173/localhost:${FRONTEND_PORT}/g" Makefile 2>/dev/null || true
    sed -i "s/-p 8888:8888/-p ${BACKEND_PORT}:${BACKEND_PORT}/g" Makefile 2>/dev/null || true
  fi
fi

# Update config.go with selected frontend port (for email links default)
if [[ -f "internal/config/config.go" ]]; then
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/localhost:5173/localhost:${FRONTEND_PORT}/g" internal/config/config.go 2>/dev/null || true
  else
    sed -i "s/localhost:5173/localhost:${FRONTEND_PORT}/g" internal/config/config.go 2>/dev/null || true
  fi
fi

# ============================================================
# Step 8: Generate .env
# ============================================================
if [[ "$CREATE_ENV" == true ]]; then
  print_step "Creating configuration..."

  if [[ -f ".env" ]] && [[ "$FORCE_ENV" != true ]]; then
    print_success ".env exists (use --force to overwrite)"
  else
    # Generate secure secrets
    generate_secret() {
      local length=${1:-32}
      local secret=""
      if command -v openssl &> /dev/null; then
        secret=$(openssl rand -hex "$length" 2>/dev/null)
      fi
      if [[ -z "$secret" ]] && [[ -r /dev/urandom ]]; then
        secret=$(head -c "$length" /dev/urandom | xxd -p 2>/dev/null | tr -d '\n')
      fi
      if [[ -z "$secret" ]] && [[ -r /dev/urandom ]]; then
        secret=$(head -c "$length" /dev/urandom | od -A n -t x1 2>/dev/null | tr -d ' \n')
      fi
      echo "$secret"
    }

    ACCESS_SECRET=$(generate_secret 32)

    if [[ -z "$ACCESS_SECRET" ]] || [[ ${#ACCESS_SECRET} -lt 32 ]]; then
      ACCESS_SECRET="CHANGE_THIS_$(date +%s)_INSECURE_PLEASE_REGENERATE"
      print_warning "Could not generate secure JWT secret - please regenerate with: openssl rand -hex 32"
    fi

    # Generate .env based on mode selection
    if [[ "$USE_LEVEE" == true ]]; then
      cat > .env << ENVEOF
# ============================================================
# ${NEW_NAME_PROPER} Configuration
# Generated: $(date)
# ============================================================

# Core Settings
APP_BASE_URL=http://localhost:${FRONTEND_PORT}
APP_DOMAIN=localhost
PRODUCTION_MODE=false

# JWT Secret (auto-generated - keep this safe!)
ACCESS_SECRET=${ACCESS_SECRET}

# Backoffice Admin (for SaaS metrics at /backoffice)
ADMIN_USERNAME="${ADMIN_EMAIL}"
ADMIN_PASSWORD="${ADMIN_PASSWORD}"

# CORS
ALLOWED_ORIGINS=http://localhost:${FRONTEND_PORT},http://localhost:${BACKEND_PORT}

# ============================================================
# Levee Mode - Full platform features
# ============================================================
LEVEE_ENABLED=true
LEVEE_API_KEY=${LEVEE_API_KEY}
LEVEE_BASE_URL=${LEVEE_BASE_URL}

# ============================================================
# Standalone Mode (disabled) - Uncomment to switch
# ============================================================
# LEVEE_ENABLED=false
# SQLITE_PATH=./data/${NEW_NAME}.db
# STRIPE_SECRET_KEY=sk_test_your_key_here
# STRIPE_PUBLISHABLE_KEY=pk_test_your_key_here
# STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
ENVEOF
    else
      cat > .env << ENVEOF
# ============================================================
# ${NEW_NAME_PROPER} Configuration
# Generated: $(date)
# ============================================================

# Core Settings
APP_BASE_URL=http://localhost:${FRONTEND_PORT}
APP_DOMAIN=localhost
PRODUCTION_MODE=false

# JWT Secret (auto-generated - keep this safe!)
ACCESS_SECRET=${ACCESS_SECRET}

# Backoffice Admin (for SaaS metrics at /backoffice)
ADMIN_USERNAME="${ADMIN_EMAIL}"
ADMIN_PASSWORD="${ADMIN_PASSWORD}"

# CORS
ALLOWED_ORIGINS=http://localhost:${FRONTEND_PORT},http://localhost:${BACKEND_PORT}

# ============================================================
# Standalone Mode (Default) - SQLite + Stripe
# ============================================================
LEVEE_ENABLED=false
SQLITE_PATH=./data/${NEW_NAME}.db

# Stripe - Get keys from https://dashboard.stripe.com/apikeys
# Leave as-is for local development without payments
STRIPE_SECRET_KEY=sk_test_your_key_here
STRIPE_PUBLISHABLE_KEY=pk_test_your_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

ENVEOF

      # Append SMTP configuration
      if [[ "$USE_SMTP" == true ]] && [[ -n "$SMTP_HOST" ]]; then
        cat >> .env << SMTPEOF

# ============================================================
# Email (SMTP Configured)
# ============================================================
SMTP_HOST=${SMTP_HOST}
SMTP_PORT=${SMTP_PORT}
SMTP_USER=${SMTP_USER}
SMTP_PASS=${SMTP_PASS}
EMAIL_FROM_ADDRESS=noreply@${NEW_NAME}.com
EMAIL_FROM_NAME=${NEW_NAME_PROPER}
SMTPEOF
      else
        cat >> .env << SMTPEOF

# ============================================================
# Email (Optional - not configured)
# ============================================================
# SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
# SMTP_USER=your-email@gmail.com
# SMTP_PASS=your-app-password
# EMAIL_FROM_ADDRESS=noreply@${NEW_NAME}.com
# EMAIL_FROM_NAME=${NEW_NAME_PROPER}
SMTPEOF
      fi

      # Append Levee section
      cat >> .env << LEVEEEOF

# ============================================================
# Levee Mode (Optional) - Full platform features
# ============================================================
# LEVEE_ENABLED=true
# LEVEE_API_KEY=lvk_your_api_key
# LEVEE_BASE_URL=https://api.levee.sh
LEVEEEOF

      # Append OAuth configuration
      if [[ "$USE_OAUTH" == true ]] && ([[ -n "$GOOGLE_CLIENT_ID" ]] || [[ -n "$GITHUB_CLIENT_ID" ]]); then
        cat >> .env << OAUTHEOF

# ============================================================
# OAuth / Social Login
# ============================================================
OAUTH_ENABLED=true
OAUTHEOF
        if [[ -n "$GOOGLE_CLIENT_ID" ]]; then
          cat >> .env << GOOGLEEOF
GOOGLE_OAUTH_ENABLED=true
GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
GOOGLEEOF
        else
          cat >> .env << GOOGLEEOF
# GOOGLE_OAUTH_ENABLED=true
# GOOGLE_CLIENT_ID=your-google-client-id
# GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLEEOF
        fi
        if [[ -n "$GITHUB_CLIENT_ID" ]]; then
          cat >> .env << GITHUBEOF
GITHUB_OAUTH_ENABLED=true
GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
GITHUBEOF
        else
          cat >> .env << GITHUBEOF
# GITHUB_OAUTH_ENABLED=true
# GITHUB_CLIENT_ID=your-github-client-id
# GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUBEOF
        fi
      else
        cat >> .env << OAUTHEOF

# ============================================================
# OAuth / Social Login (Optional)
# ============================================================
# To enable OAuth, set OAUTH_ENABLED=true and configure providers below
OAUTH_ENABLED=false
GOOGLE_OAUTH_ENABLED=false
# GOOGLE_CLIENT_ID=your-google-client-id
# GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_OAUTH_ENABLED=false
# GITHUB_CLIENT_ID=your-github-client-id
# GITHUB_CLIENT_SECRET=your-github-client-secret
OAUTHEOF
      fi

      # Append Features configuration
      cat >> .env << FEATURESEOF

# ============================================================
# Features
# ============================================================
ORGANIZATIONS_ENABLED=${ENABLE_ORGANIZATIONS}
NOTIFICATIONS_ENABLED=${ENABLE_NOTIFICATIONS}
FEATURESEOF
    fi

    print_success ".env created with secure secret"
  fi
fi

# ============================================================
# Step 10: Install Dependencies
# ============================================================
if [[ "$INSTALL_DEPS" == true ]]; then
  print_step "Installing dependencies..."

  # Go
  if command -v go &> /dev/null; then
    if go mod download 2>/dev/null && go mod tidy 2>/dev/null; then
      print_success "Go dependencies"
    else
      print_warning "Go deps may need: go mod tidy"
    fi
  fi

  # Frontend
  if [[ -d "app" ]] && command -v pnpm &> /dev/null; then
    if (cd app && pnpm install --silent 2>/dev/null); then
      print_success "Frontend dependencies"
    else
      print_warning "Frontend deps may need: cd app && pnpm install"
    fi
  fi
fi

# ============================================================
# Step 11: Create Data and Build Directories
# ============================================================
mkdir -p data 2>/dev/null || true
mkdir -p app/build 2>/dev/null || true
touch app/build/.keep 2>/dev/null || true

# ============================================================
# Step 12: Initialize Git (Remote Mode)
# ============================================================
if [[ "$MODE" == "remote" ]]; then
  print_step "Initializing git repository..."
  git init -q -b main 2>/dev/null
  git add . 2>/dev/null
  git commit -q -m "Initial commit - ${NEW_NAME} from Gobot" 2>/dev/null
  print_success "Git repository initialized"
fi

# ============================================================
# Done!
# ============================================================
echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║${NC}                  ✅ Setup Complete!                        ${GREEN}║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${BOLD}Your app:${NC}  $NEW_NAME"
echo -e "  ${BOLD}Location:${NC}  $(pwd)"
echo ""
echo -e "  ${BOLD}Start development:${NC}"
echo ""
echo "    make dev              # Starts backend + frontend"
echo ""
echo -e "  ${BOLD}Or manually:${NC}"
echo ""
echo "    make air              # Backend with hot reload"
echo "    cd app && pnpm dev    # Frontend (another terminal)"
echo ""
echo -e "  ${BOLD}Then visit:${NC} http://localhost:${FRONTEND_PORT}"
echo ""
echo -e "  ${BOLD}Admin Backoffice:${NC}"
echo "    URL:      http://localhost:${BACKEND_PORT}/backoffice"
echo "    Username: ${ADMIN_EMAIL}"
echo "    Password: (see ADMIN_PASSWORD in .env)"
echo ""
echo -e "  ${BOLD}Configure payments:${NC}"
echo "    1. Get keys from https://dashboard.stripe.com/apikeys"
echo "    2. Edit .env with your STRIPE_* values"
echo ""

# ============================================================
# Clean Up - Remove installer to prevent re-running
# ============================================================
# By this point we're always in the project directory
# Remove install.sh to prevent accidental re-installation
rm -f ./install.sh 2>/dev/null || true

# ============================================================
# Offer to Start
# ============================================================
if [[ "$OFFER_START" == true ]] && [[ -t 0 ]]; then
  echo -ne "  Start the app now? [Y/n] "
  read -r START_NOW
  if [[ ! "$START_NOW" =~ ^[Nn]$ ]]; then
    echo ""
    print_step "Starting ${NEW_NAME}..."
    echo ""
    exec make dev
  fi
fi
