@echo off
REM Setup script to download DuckDB WASM files for local development (Windows)

setlocal enabledelayedexpansion

echo.
echo 📦 Setting up DuckDB WASM files for local development...
echo.

set DUCKDB_DIR=internal\views\static\duckdb

echo 📍 Target directory: %DUCKDB_DIR%
echo.

REM Create directory
if not exist "%DUCKDB_DIR%" (
    mkdir "%DUCKDB_DIR%"
)

REM Check if files already exist
if exist "%DUCKDB_DIR%\duckdb-mvp.wasm" (
    if exist "%DUCKDB_DIR%\duckdb-browser-mvp.worker.js" (
        echo ✅ DuckDB files already downloaded
        dir /s "%DUCKDB_DIR%"
        exit /b 0
    )
)

echo 📥 Downloading DuckDB WASM files from CDN...
echo.

REM Download main WASM file
echo   * Downloading duckdb-mvp.wasm (5-6MB)...
powershell -Command "& {(New-Object System.Net.WebClient).DownloadFile('https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/dist/duckdb-mvp.wasm', '%DUCKDB_DIR%\duckdb-mvp.wasm')}"

if errorlevel 1 (
    echo ❌ Failed to download duckdb-mvp.wasm
    echo    Check your internet connection and try again
    exit /b 1
)

REM Download worker file
echo.
echo   * Downloading duckdb-browser-mvp.worker.js...
powershell -Command "& {(New-Object System.Net.WebClient).DownloadFile('https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/dist/duckdb-browser-mvp.worker.js', '%DUCKDB_DIR%\duckdb-browser-mvp.worker.js')}"

if errorlevel 1 (
    echo ❌ Failed to download duckdb-browser-mvp.worker.js
    exit /b 1
)

REM Create ES module wrapper
echo.
echo   * Creating ES module wrapper...
(
    echo // This is a shim that exports the DuckDB WASM module
    echo // The actual files are served via HTTP from /static/duckdb/
    echo.
    echo // Re-export everything from the CDN version
    echo export * from 'https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/+esm';
    echo.
    echo console.log('DuckDB WASM loaded from local fallback');
) > "%DUCKDB_DIR%\duckdb-wasm.js"

echo.
echo ✅ DuckDB setup complete!
echo.
echo 📋 Files installed:
dir "%DUCKDB_DIR%"

echo.
echo 🎯 Now you can:
echo   1. Build the app: task build
echo   2. Run locally: task run
echo   3. Test SQL: http://localhost:8090/sql
echo.
echo ℹ️  The app will:
echo   * Try CDN first (works best on production)
echo   * Fall back to local files on localhost
echo.

endlocal
