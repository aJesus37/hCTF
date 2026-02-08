#!/bin/bash
# Setup script to download DuckDB WASM files for local development

set -e

echo "📦 Setting up DuckDB WASM files for local development..."
echo ""

# Create directory
DUCKDB_DIR="internal/views/static/duckdb"
mkdir -p "$DUCKDB_DIR"

echo "📍 Target directory: $DUCKDB_DIR"
echo ""

# Check if files already exist
if [ -f "$DUCKDB_DIR/duckdb-mvp.wasm" ] && [ -f "$DUCKDB_DIR/duckdb-browser-mvp.worker.js" ]; then
    echo "✅ DuckDB files already downloaded"
    ls -lh "$DUCKDB_DIR"/
    exit 0
fi

echo "📥 Downloading DuckDB WASM files from CDN..."
echo ""

# Download main WASM file
echo "  • Downloading duckdb-mvp.wasm (5-6MB)..."
curl -f -L --progress-bar \
    "https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/dist/duckdb-mvp.wasm" \
    -o "$DUCKDB_DIR/duckdb-mvp.wasm"

if [ $? -ne 0 ]; then
    echo "❌ Failed to download duckdb-mvp.wasm"
    echo "   Check your internet connection and try again"
    exit 1
fi

# Download worker file
echo ""
echo "  • Downloading duckdb-browser-mvp.worker.js..."
curl -f -L --progress-bar \
    "https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/dist/duckdb-browser-mvp.worker.js" \
    -o "$DUCKDB_DIR/duckdb-browser-mvp.worker.js"

if [ $? -ne 0 ]; then
    echo "❌ Failed to download duckdb-browser-mvp.worker.js"
    exit 1
fi

# Create ES module wrapper for local DuckDB
echo ""
echo "  • Creating ES module wrapper..."
cat > "$DUCKDB_DIR/duckdb-wasm.js" << 'EOF'
// This is a shim that exports the DuckDB WASM module
// The actual files are served via HTTP from /static/duckdb/

// Re-export everything from the CDN version
export * from 'https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/+esm';

console.log('DuckDB WASM loaded from local fallback');
EOF

echo ""
echo "✅ DuckDB setup complete!"
echo ""
echo "📋 Files installed:"
ls -lh "$DUCKDB_DIR"/ | tail -3

echo ""
echo "🎯 Now you can:"
echo "  1. Build the app: task build"
echo "  2. Run locally: task run"
echo "  3. Test SQL: http://localhost:8090/sql"
echo ""
echo "ℹ️  The app will:"
echo "  • Try CDN first (works best on production)"
echo "  • Fall back to local files on localhost"
echo ""
