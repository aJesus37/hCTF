# OpenAPI UI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a built-in ReDoc UI at `/docs` to visualize the OpenAPI specification.

**Architecture:** Create a new HTML template that loads ReDoc from CDN, embeds the OpenAPI YAML spec, and serves it via a new handler at `/docs`. The UI will be accessible from the navigation menu.

**Tech Stack:** ReDoc (CDN), Go handlers, HTML template with embedded YAML

---

### Task 1: Create OpenAPI UI Template

**Files:**
- Create: `internal/views/templates/docs.html`

**Step 1: Create the ReDoc HTML template**

```html
{{define "docs-content"}}
<div class="min-h-screen bg-white dark:bg-dark-surface">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-6">API Documentation</h1>
        <p class="text-gray-600 dark:text-gray-400 mb-4">
            This documentation describes the hCTF2 REST API. You can use this spec to integrate with the platform.
        </p>
        <div id="redoc-container" class="border border-gray-200 dark:border-dark-border rounded-lg overflow-hidden"></div>
    </div>
</div>

<script src="https://cdn.jsdelivr.net/npm/redoc@2.1.3/bundles/redoc.standalone.js"></script>
<script>
    // Fetch and display the OpenAPI spec
    fetch('/api/openapi.yaml')
        .then(response => response.text())
        .then(spec => {
            Redoc.init(
                '/api/openapi.yaml',
                {
                    theme: {
                        colors: {
                            primary: {
                                main: '#3b82f6'
                            }
                        },
                        typography: {
                            fontFamily: 'system-ui, -apple-system, sans-serif'
                        }
                    },
                    hideDownloadButton: false,
                    expandResponses: '200,201',
                    jsonSampleExpandLevel: 2
                },
                document.getElementById('redoc-container')
            );
        })
        .catch(error => {
            document.getElementById('redoc-container').innerHTML = 
                '<p class="p-4 text-red-600 dark:text-red-400">Failed to load API documentation. Please try again later.</p>';
            console.error('Error loading OpenAPI spec:', error);
        });
</script>
{{end}}

{{define "docs-page"}}
<!DOCTYPE html>
<html lang="en" class="{{if eq .Theme "dark"}}dark{{end}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            darkMode: 'class',
            theme: {
                extend: {
                    colors: {
                        dark: {
                            bg: '#0f172a',
                            surface: '#1e293b',
                            border: '#334155',
                        }
                    }
                }
            }
        }
    </script>
    <link rel="stylesheet" href="/static/css/custom.css">
</head>
<body class="bg-gray-50 dark:bg-dark-bg text-gray-900 dark:text-white min-h-screen flex flex-col">
    {{template "navbar" .}}
    <main class="flex-grow">
        {{template "docs-content" .}}
    </main>
    {{template "footer" .}}
    <script src="/static/js/theme.js"></script>
</body>
</html>
{{end}}
```

**Step 2: Verify template syntax**

Run: `cat internal/views/templates/docs.html | head -20`
Expected: Template content displayed without errors

**Step 3: Commit**

```bash
git add internal/views/templates/docs.html
git commit -m "feat: add ReDoc template for OpenAPI UI"
```

---

### Task 2: Add Handler for Docs Page

**Files:**
- Modify: `main.go` (add handler method)

**Step 1: Add handler method to Server struct**

In `main.go`, add after the `handleOpenAPISpec` method (around line 610):

```go
// handleDocsPage serves the OpenAPI documentation UI
func (s *Server) handleDocsPage(w http.ResponseWriter, r *http.Request) {
    claims := auth.GetUserFromContext(r.Context())
    
    data := map[string]interface{}{
        "Title": "API Documentation",
        "User":  claims,
        "Page":  "docs",
    }
    
    s.render(w, "docs-page", data)
}
```

**Step 2: Register the route**

Find the route registration section (around line 318) and add:

```go
// OpenAPI Spec and Docs
r.Get("/api/openapi.yaml", s.handleOpenAPISpec)
r.Get("/docs", s.handleDocsPage)
```

**Step 3: Verify Go syntax**

Run: `go build -o hctf2 .`
Expected: Build succeeds with no errors

**Step 4: Commit**

```bash
git add main.go
git commit -m "feat: add handler and route for OpenAPI docs UI"
```

---

### Task 3: Add Navigation Link

**Files:**
- Modify: `internal/views/templates/base.html` (navbar section)

**Step 1: Find the navbar template**

Locate the navigation links in `base.html` (around the Challenges, Scoreboard links).

**Step 2: Add Docs link**

Add a new nav link after "SQL Playground":

```html
<a href="/docs" class="text-gray-600 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition {{if eq .Page "docs"}}text-blue-600 dark:text-blue-400{{end}}">
    API Docs
</a>
```

**Step 3: Verify template renders**

Run server and check any page for the new nav link.

**Step 4: Commit**

```bash
git add internal/views/templates/base.html
git commit -m "feat: add API Docs link to navigation"
```

---

### Task 4: Test the UI

**Step 1: Build and run**

```bash
go build -o hctf2 . && ./hctf2 &
```

**Step 2: Verify page loads**

Navigate to `http://localhost:8090/docs`
Expected: ReDoc UI loads showing the API documentation

**Step 3: Verify navigation works**

Click on "API Docs" in the navbar from any page.
Expected: Successfully navigates to the docs page

**Step 4: Test dark mode**

Toggle dark mode on the docs page.
Expected: UI adapts to dark mode properly

**Step 5: Final commit**

```bash
git add docs/plans/2026-02-18-openapi-ui.md
git commit -m "docs: add OpenAPI UI implementation plan"
```

---

## Verification Checklist

- [ ] `/docs` page loads with ReDoc UI
- [ ] OpenAPI spec is rendered correctly
- [ ] Navigation link works from all pages
- [ ] Dark mode works properly
- [ ] Mobile responsive
- [ ] No console errors in browser
