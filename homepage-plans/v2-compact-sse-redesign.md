# Akatsuki Landing Page Redesign v2

## Goal
Create a modern, dense, single-page homepage that:
1. Feels alive with real-time activity
2. Shows aggregate stats as floating overlays on the hero
3. Uses a mode selector to filter activity feeds (no 3x duplication)
4. Fits above the fold (~900px) or provides compelling content to scroll

---

## Design Concept

```
┌─ NAVBAR ──────────────────────────────────────────────────────┐
│ Logo                              [Vanilla|Relax|AP] Register │
├───────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌─ HERO with Background Image ─────────────────────────────┐ │
│  │                                                           │ │
│  │  ┌─────────┐                           ┌─────────┐       │ │
│  │  │ 103K    │      AKATSUKI             │ 288K    │       │ │
│  │  │ players │      [wordmark]           │ maps    │       │ │
│  │  └─────────┘                           └─────────┘       │ │
│  │                   🟢 X online                             │ │
│  │                                                           │ │
│  │  ┌─────────┐   [Register] [Connect]   ┌─────────┐       │ │
│  │  │ 78M+    │                           │ 1,208   │       │ │
│  │  │ scores  │                           │ years   │       │ │
│  │  └─────────┘                           └─────────┘       │ │
│  │                                                           │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                │
│  ┌─ LIVE SCORE TICKER (SSE) ─────────────────────────────────┐│
│  │ 🔴 PlayerX just set 456pp on MapName • PlayerY got #1... │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                │
│  ┌─ ACTIVITY ─────────────────┐ ┌─ TRENDING ────────────────┐ │
│  │ Recent First Places        │ │ Hot Maps This Week        │ │
│  │ (filtered by mode selector)│ │ [card] [card] [card]      │ │
│  │ • Player → #1 | 456pp     │ │ [card] [card] [card]      │ │
│  │ • Player → #1 | 523pp     │ │                           │ │
│  │ • ...                      │ │                           │ │
│  │ [High PP Today] tab        │ │                           │ │
│  └────────────────────────────┘ └───────────────────────────┘ │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## Key Features

### 1. Floating Overlay Stats on Hero
Semi-transparent glass-morphism cards positioned around the wordmark:
- **Players**: 103K registered
- **Maps**: 288K ranked beatmaps
- **Scores**: 78M+ total (aggregate across all modes)
- **Playtime**: 1,208 years

CSS: `backdrop-filter: blur(8px); background: rgba(0,0,0,0.4)`

### 2. Mode Selector (Shared Component)
Reuse the existing mode selector pattern from leaderboard/profile pages:
- Located in navbar OR above activity section
- Filters: First Places, High PP feeds by selected mode
- Aggregate stats in hero don't change (total across all modes)

### 3. Live Score Ticker (SSE)
Horizontal scrolling ticker showing real-time activity:
- "PlayerX just set 456pp on SongName"
- "PlayerY claimed #1 on MapName"
- Auto-updates via Server-Sent Events

**Backend**: New SSE endpoint in hanayo that subscribes to Redis pub/sub

### 4. Compact Activity Section
Two-column layout below hero:
- **Left**: First Places / High PP (tabbed, mode-filtered)
- **Right**: Trending Maps (card grid, 2x3)

### 5. Single Page Target
- Navbar: ~55px
- Hero with overlays: ~320px
- Ticker: ~40px
- Activity section: ~400px
- **Total: ~815px** (fits 900px viewport with breathing room)

---

## Architecture Changes

### New: SSE Endpoint for Live Scores

**File: `hanayo/app/handlers/api_sse.go`** (new)

```go
// GET /api/v1/sse/scores - Server-Sent Events stream
func SSEScores(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // Subscribe to Redis pub/sub channel
    pubsub := services.RD.Subscribe("akatsuki:live_scores")
    defer pubsub.Close()

    for msg := range pubsub.Channel() {
        fmt.Fprintf(c.Writer, "data: %s\n\n", msg.Payload)
        c.Writer.Flush()
    }
}
```

**Publisher**: score-service already publishes to Redis on score submission.
Add new channel `akatsuki:live_scores` with formatted messages.

### New Redis Keys

| Key | Type | Source | Description |
|-----|------|--------|-------------|
| `akatsuki:first_places:{rx}` | JSON | new-cron | First places filtered by mode (0/1/2) |
| `akatsuki:high_pp:{rx}` | JSON | new-cron | High PP plays filtered by mode |
| `akatsuki:live_scores` | Pub/Sub | score-service | Real-time score events |

---

## Implementation Phases

### Phase 1: Layout Restructure
1. Remove stacked sections, create two-column layout
2. Hero with overlay stat cards (glass-morphism)
3. Mode selector in navbar or activity header
4. Compact activity feeds

### Phase 2: SSE Infrastructure
1. Add SSE endpoint in hanayo
2. Modify score-service to publish to `akatsuki:live_scores`
3. JavaScript EventSource client for ticker

### Phase 3: Mode-Filtered Data
1. Update new-cron to create per-mode Redis keys
2. JavaScript to reload activity on mode change
3. API endpoint for fetching filtered data

### Phase 4: Polish
1. Glass-morphism CSS for overlay cards
2. Ticker animation (smooth horizontal scroll)
3. Responsive breakpoints
4. Loading states

---

## Files to Modify

| File | Changes |
|------|---------|
| `hanayo/web/templates/homepage.html` | Complete restructure with overlay layout |
| `hanayo/web/src/css/pages/home.css` | Glass-morphism, ticker, two-column |
| `hanayo/web/src/js/pages/home.js` | SSE client, mode selector, ticker |
| `hanayo/app/handlers/api_sse.go` | New SSE endpoint |
| `hanayo/app/router.go` | Register SSE route |
| `new-cron/main.py` | Add per-mode Redis keys |
| `score-service` | Publish to live_scores channel (if not already) |

---

## Verification

1. Load homepage - should fit above fold on 900px viewport
2. Overlay stats visible on hero background
3. Mode selector filters activity feeds
4. SSE ticker shows live scores (test with score submission)
5. Responsive: stacks properly on mobile
6. Performance: no layout shift, smooth animations

---

## Deferred / Future

- Interactive stat cards (click to see breakdown by mode)
- Personalized "Welcome back" section for logged-in users
- Achievement unlocks in ticker
- WebSocket upgrade if SSE proves limiting
