# CLAUDE.md - Hanayo

## Development Server

**IMPORTANT**: Always use `./run-server.sh` or restart the full server after any asset changes. Never run `gulp` alone without restarting the server.

Hanayo uses content-hashed asset filenames (e.g., `akatsuki-cd83ec9f8f.min.css`), and the Go binary reads the asset manifest at startup. Running only `gulp` will generate new hashed filenames, but the running server will still reference the old hashes, causing 404s for all JS/CSS.

```bash
# Preferred: builds assets, Go binary, signs, and starts server
./run-server.sh

# If you need to rebuild manually, always restart the server after:
cd web && gulp && cd ..
go build -o hanayo .
./hanayo
```

## Project Structure

- `app/handlers/` - HTTP request handlers (profiles, beatmaps, clans, sessions, etc.)
- `app/usecases/` - Business logic (auth, templates, user management)
- `app/middleware/` - Rate limiting, logging, blacklist
- `app/models/` - Data models (session, pages, messages)
- `app/states/` - Application state (settings, services)
- `internal/` - Internal utilities (CSRF, BBCode, locale)
- `web/templates/` - Go HTML templates
- `web/src/css/` - Source CSS files
- `web/src/js/` - Source JavaScript files
- `web/static/` - Compiled static assets (output of gulp)
- `main.go` - Server setup with Gin router

## Design Mockups

When exploring visual design options (CSS styling, component layouts, color schemes), create standalone HTML mockup files in `web/static/` to compare alternatives side-by-side before implementing:

```bash
# Example: create mockups for a new component
web/static/component-mockups.html
```

Include multiple design variations in a single file with the same background/context as the target page. This allows quick visual comparison without modifying production code. Once a design is chosen, implement it in the actual CSS/templates.
