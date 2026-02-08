# SQL Playground - Setup & Usage Guide

## Overview

The SQL Playground is a unique feature that lets users query CTF data using SQL. It runs **entirely in the browser** using DuckDB WASM, making it safe and performant.

## Why Doesn't It Work on Localhost?

DuckDB WASM is loaded from a CDN (jsDelivr). Browser security policies block loading web workers from external domains on localhost, causing a CORS error.

## ✅ Solutions

### Solution 1: Use Docker (Easiest, Recommended)

Docker containers properly handle CORS, so the SQL Playground works out-of-the-box:

```bash
docker compose -f docker-compose.dev.yml up
```

Then visit: http://localhost:8090/sql

**Why this works**: Docker's network setup bypasses the localhost CORS restrictions.

---

### Solution 2: Setup Local DuckDB Files

Download DuckDB WASM files locally so they're served from your own server:

#### **Linux/macOS:**
```bash
task setup-sql
```

#### **Windows:**
```batch
scripts/setup-duckdb-local.bat
```

#### **Manual:**
```bash
# Create directory
mkdir -p internal/views/static/duckdb

# Download files
curl -L https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/dist/duckdb-mvp.wasm \
  -o internal/views/static/duckdb/duckdb-mvp.wasm

curl -L https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/dist/duckdb-browser-mvp.worker.js \
  -o internal/views/static/duckdb/duckdb-browser-mvp.worker.js
```

Then build and run:
```bash
task build
task run
```

Visit: http://localhost:8090/sql

**How it works**:
- App tries to load DuckDB from CDN first (best for production)
- If that fails, falls back to local files
- Shows which source is being used in the UI

---

### Solution 3: Deploy to Production

The SQL Playground works perfectly in production because real domains don't have CORS restrictions:

```bash
task build-prod
# Deploy binary to your server
```

---

## How It Works

### Architecture

```
┌─────────────────────────────────────────────────────┐
│ Browser (Client-Side)                               │
├─────────────────────────────────────────────────────┤
│                                                       │
│  1. Load DuckDB WASM (from CDN or local)            │
│     ↓                                                │
│  2. Fetch CTF data (/api/sql/snapshot)              │
│     ↓                                                │
│  3. Create in-memory database                        │
│     ↓                                                │
│  4. Execute user queries (NO server access)         │
│     ↓                                                │
│  5. Display results                                  │
│                                                       │
└─────────────────────────────────────────────────────┘
```

### Data Flow

1. **User types SQL query** → `SELECT * FROM challenges WHERE difficulty = 'hard'`

2. **Query executes in browser** using DuckDB WASM
   - 100% client-side
   - No SQL injection risk (runs in sandbox)
   - No server load

3. **Results displayed** as HTML table

### Data Sources

The playground can query these tables:

| Table | Fields | Source |
|-------|--------|--------|
| **challenges** | id, name, category, difficulty, visible, created_at | Public challenges only |
| **questions** | id, challenge_id, name, points, flag_mask | Without actual flags |
| **submissions** | id, question_id, user_id, is_correct, created_at | Correct submissions only |
| **users** | id, name, team_id, created_at | Public user info |

---

## Usage Examples

### Example Queries

**Find hardest challenges:**
```sql
SELECT c.name, c.difficulty, COUNT(s.id) as solves
FROM challenges c
LEFT JOIN questions q ON c.id = q.challenge_id
LEFT JOIN submissions s ON q.id = s.question_id AND s.is_correct = 1
GROUP BY c.id, c.name, c.difficulty
ORDER BY solves ASC
LIMIT 10
```

**Top users by points:**
```sql
SELECT u.name, SUM(q.points) as total_points, COUNT(*) as solve_count
FROM users u
JOIN submissions s ON u.id = s.user_id
JOIN questions q ON s.question_id = q.id
WHERE s.is_correct = 1
GROUP BY u.id, u.name
ORDER BY total_points DESC
LIMIT 10
```

**Challenge difficulty distribution:**
```sql
SELECT difficulty, COUNT(*) as count
FROM challenges
WHERE visible = 1
GROUP BY difficulty
```

**User progress:**
```sql
SELECT
    u.name,
    COUNT(DISTINCT s.question_id) as questions_solved,
    COUNT(*) as total_attempts,
    MAX(s.created_at) as last_attempt
FROM users u
LEFT JOIN submissions s ON u.id = s.user_id
GROUP BY u.id, u.name
```

---

## Troubleshooting

### "SQL Playground could not load"

**Problem**: DuckDB failed to load from both CDN and local fallback

**Solutions**:
1. Check internet connection
2. Run `task setup-sql` to download local files
3. Try Docker: `docker compose -f docker-compose.dev.yml up`
4. Check browser console (F12) for detailed error message

### "Could not load CDN" but "Local files work"

**This is normal!** The app will:
- Show "✅ Database loaded (from Local)"
- SQL Playground will work fine
- It means you're using the local fallback

### Performance issues

DuckDB WASM is very fast, but:
- First load takes ~2-3 seconds (loading WASM module)
- Subsequent queries are instant
- Works best with <50MB of data (typical CTF size)

---

## Technical Details

### DuckDB WASM vs Alternatives

| Feature | DuckDB | sql.js | Server-side |
|---------|--------|--------|-------------|
| Features | Full SQL, CTEs, window functions | Basic SQL | Full SQL |
| Performance | Excellent | Good | Excellent |
| Security | Sandboxed | Sandboxed | Depends |
| Server Load | None | None | High |
| Complexity | Medium | Low | High |

We use DuckDB because it offers the best combination of features and performance.

### Data Freshness

- Snapshot loaded once when page opens
- Reflects data at that moment
- Refresh page to get latest data
- Good for learning; use `/api/challenges` for real-time data

### Browser Compatibility

| Browser | Status |
|---------|--------|
| Chrome | ✅ Works |
| Firefox | ✅ Works |
| Safari | ✅ Works |
| Edge | ✅ Works |
| IE 11 | ❌ Not supported |

---

## Configuration

### Customize Loaded Tables

Edit `internal/database/queries.go` → `GetSQLSnapshot()` to control what data is exposed.

### Add Custom Example Queries

Edit `internal/views/templates/sql.html` → `exampleQueries` array to add more examples.

### Change Load Timeout

The DuckDB module has a default timeout. To customize, edit the script in sql.html.

---

## Security Notes

✅ **Why it's safe:**
- SQL runs in browser (no server access)
- User can only query public data
- Cannot modify data (read-only)
- No SQL injection risk (DuckDB handles it)
- No sensitive data exposed (passwords, flags hidden)

⚠️ **Limitations:**
- Cannot delete/update/insert data
- Cannot access tables from other connections
- Cannot execute arbitrary JavaScript (sandbox)

---

## Performance Characteristics

### Load Times

| Step | Time | Notes |
|------|------|-------|
| Load DuckDB WASM | 2-3s | First time only |
| Load snapshot data | 500ms | Typical CTF data |
| Parse into DB | 100ms | Creates tables |
| Simple query | 10ms | Most queries |
| Complex query | 50-200ms | Joins, aggregations |

### Memory Usage

- DuckDB module: 5-10MB
- Data snapshot: 1-5MB (varies)
- Total: ~10-15MB for typical CTF

---

## Future Improvements

- [ ] Real-time data updates (WebSocket)
- [ ] Query history/save favorite queries
- [ ] Query optimization hints
- [ ] Visual query builder
- [ ] Export results to CSV
- [ ] Query execution analytics
- [ ] Support for team-specific queries

---

## Getting Help

**SQL Playground not working?**
1. Try Docker first (most reliable)
2. Run `task setup-sql` for local files
3. Check browser console (F12) for errors
4. Ensure internet connection is stable

**Want to contribute?**
- Add new example queries
- Improve documentation
- Report bugs
- Suggest features

---

## References

- [DuckDB Documentation](https://duckdb.org/docs/)
- [DuckDB WASM](https://duckdb.org/docs/api/wasm/overview.html)
- [SQLite WASM](https://sqlite.org/wasm/)
- [CORS Restrictions](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
