# Getting Started with hCTF2

Welcome! Here's how to get hCTF2 running in less than 5 minutes.

## Choice 1: Docker (Recommended - Fastest)

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/hctf2.git
cd hctf2

# 2. Start with one command
docker compose -f docker-compose.dev.yml up -d

# 3. Open browser
open http://localhost:8090

# 4. Login
Email: admin@hctf.local
Password: changeme
```

**Done!** Your CTF platform is running.

### Stop it
```bash
docker compose -f docker-compose.dev.yml down
```

## Choice 2: Native Go (For Development)

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/hctf2.git
cd hctf2

# 2. Install Task (if needed)
go install github.com/go-task/task/v3/cmd/task@latest

# 3. Build and run
task deps
task run

# 4. Open browser
open http://localhost:8090

# 5. Login
Email: admin@hctf.local
Password: changeme
```

**Done!** Your CTF platform is running.

## First Time Setup

### 1. Change Admin Password
After logging in, change the default password immediately.

### 2. Create Your First Challenge

Use the API to create a challenge (admin UI coming in Phase 2):

```bash
# Login as admin
curl -X POST http://localhost:8090/api/auth/login \
  -d "email=admin@hctf.local&password=changeme" \
  -c cookies.txt

# Create a challenge
curl -X POST http://localhost:8090/api/admin/challenges \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "Welcome Challenge",
    "description": "Your first CTF challenge",
    "category": "misc",
    "difficulty": "easy",
    "visible": true
  }' | jq .

# Note the returned challenge ID (e.g., "abc123...")
# Use it in the next step
```

### 3. Create a Question

```bash
curl -X POST http://localhost:8090/api/admin/questions \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "challenge_id": "YOUR_CHALLENGE_ID_HERE",
    "name": "Find the Flag",
    "description": "The flag is hidden in plain sight!",
    "flag": "FLAG{welcome_to_hctf2}",
    "case_sensitive": false,
    "points": 100
  }'
```

### 4. Test It

1. Register a new user at http://localhost:8090/register
2. Navigate to Challenges
3. Find your challenge
4. Submit the flag: `FLAG{welcome_to_hctf2}`
5. Check the Scoreboard

**Success!** You've created your first CTF challenge.

## Features Available Now

✅ **User Management**
- Register, login, logout
- JWT authentication
- Secure password hashing

✅ **Challenges**
- Create/edit/delete challenges
- Multiple questions per challenge
- Flag validation with masking
- Difficulty levels and categories

✅ **Scoring**
- Point-based system
- Live scoreboard
- Solve tracking

✅ **Unique Feature: SQL Playground**
- Query challenge data with SQL
- Uses DuckDB WASM (runs in browser)
- Safe by design (no server-side SQL execution)
- Example queries included

✅ **Modern UI**
- Dark theme by default
- HTMX for smooth interactions
- Responsive design
- No JavaScript framework bloat

## Common Tasks

### Check Challenges
```bash
curl http://localhost:8090/api/challenges
```

### View Scoreboard
```bash
curl http://localhost:8090/api/scoreboard
```

### Get SQL Snapshot
```bash
curl http://localhost:8090/api/sql/snapshot
```

### View Logs (Docker)
```bash
docker compose logs -f hctf2
```

### Stop Server (Docker)
```bash
docker compose down
```

### Reset Database
```bash
# With Docker
docker compose down -v
docker compose up -d

# With native
rm hctf2.db
task run
```

## Useful Links

- **Homepage**: http://localhost:8090
- **Challenges**: http://localhost:8090/challenges
- **Scoreboard**: http://localhost:8090/scoreboard
- **SQL Playground**: http://localhost:8090/sql
- **Documentation**: See README.md
- **API Reference**: See API.md
- **Docker Guide**: See DOCKER.md
- **Full Docs**: See INSTALL.md, QUICKSTART.md

## What's Next?

### Phase 1 (Now) ✅
- User authentication
- Challenge management (API)
- Flag submission
- Scoreboard
- SQL playground

### Phase 2 (Coming Soon)
- Admin web UI (no more API)
- Team management
- Hints system
- File uploads
- Markdown support

### Phase 3 (Planned)
- Dynamic scoring
- Challenge dependencies
- Real-time updates (WebSockets)
- Export/import challenges

## Troubleshooting

### Docker Won't Start
```bash
# Check logs
docker compose logs hctf2

# Verify port is free
lsof -i :8090

# Try different port in docker-compose.yml
```

### Can't Access from Browser
```bash
# Check if container is running
docker ps | grep hctf2

# Test with curl
curl http://localhost:8090

# Check firewall
sudo ufw status
```

### Default Password Isn't Working
```bash
# Docker dev setup creates these:
Email: admin@hctf.local
Password: changeme

# If custom setup, check your compose file or logs
docker compose logs hctf2
```

### Database Issues
```bash
# Reset everything
docker compose down -v
rm -rf data/
docker compose up -d
```

## Security Notes

1. **Change Default Password**: Do this immediately in production
2. **Use HTTPS**: See DOCKER.md for nginx reverse proxy setup
3. **Regular Backups**: Keep copies of your database
4. **Update Regularly**: Pull latest changes and rebuild

## Getting Help

- **Documentation**: Check `*.md` files in project root
- **API Reference**: See `API.md`
- **Architecture**: See `ARCHITECTURE.md`
- **Docker Guide**: See `DOCKER.md`
- **Issues**: https://github.com/yourusername/hctf2/issues

## Architecture

```
┌─────────────────────────────────┐
│      Browser (Dark UI)          │
│   HTMX + Tailwind + Alpine      │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│   Go HTTP Server (Chi Router)   │
│   - Authentication (JWT)        │
│   - Challenge Management        │
│   - Flag Validation             │
│   - Scoreboard                  │
│   - SQL Snapshot API            │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│  SQLite Database                │
│  - Users                        │
│  - Challenges & Questions       │
│  - Submissions & Scoring        │
│  - Hints System                 │
└─────────────────────────────────┘
```

## Single Command Examples

```bash
# Docker: One command setup
docker compose -f docker-compose.dev.yml up -d && \
  sleep 2 && \
  open http://localhost:8090

# Native: One command setup
cd ~/Projects/hCTF2 && \
  task run &

# Check it's working
curl http://localhost:8090 | head -5
```

## Stats

- **Binary Size**: 13 MB
- **Docker Image**: ~20 MB
- **Startup Time**: < 2 seconds
- **Memory Usage**: 50-100 MB
- **CPU Usage**: < 1% idle
- **Database**: SQLite (single file)
- **Language**: Go 1.25+
- **License**: MIT

## Ready to Go! 🚀

You now have a fully functional CTF platform. Start creating challenges and hosting competitions!

Questions? Check the documentation or create an issue on GitHub.

Happy CTFing! 🎯
