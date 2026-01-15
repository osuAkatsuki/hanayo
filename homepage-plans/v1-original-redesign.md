# Akatsuki Landing Page Redesign

## Goal
Transform the homepage to:
1. **New visitors**: Make them feel they've discovered the best osu! private server
2. **Returning users**: Give comprehensive activity feed showing server life

## Design Decisions
- **Layout**: Hybrid - shared activity sections for all, personalized section for logged-in users
- **News**: Auto-generated activity (no new tables) - first places, high PP plays, trending maps
- **USP**: Relax/Autopilot prominent but not exclusive - it's core but we're diversifying

---

## Current State
The homepage (`hanayo/web/templates/homepage.html`) shows only:
- Hero with logo/tagline
- 6 stat boxes (score counts + top scores per mode)
- Optional alert banner

**Real server stats (for context):**
- 103,552 users, ~1,320 active daily
- ~29,000 scores/day, 78M+ total scores
- 288K ranked beatmaps, 1,208 years cumulative playtime

---

## Architecture

### Data Flow
```
new-cron (every 5-30 min)
    │
    ├─► Aggregates: first places, high PP, trending maps, counts
    │
    └─► Redis keys (JSON + strings)
            │
            └─► hanayo templates via rediget/redigetJSON
```

### New Redis Keys (populated by new-cron)

| Key | Type | Update | Description |
|-----|------|--------|-------------|
| `akatsuki:recent_first_places` | JSON | 5 min | Last 10 first places with user/beatmap |
| `akatsuki:high_pp_plays_24h` | JSON | 5 min | Plays >800pp in last 24h |
| `akatsuki:trending_beatmaps` | JSON | 30 min | Top 10 most played this week |
| `akatsuki:registered_users` | String | 30 min | COUNT of public users |
| `akatsuki:ranked_beatmaps` | String | 30 min | COUNT of ranked maps |
| `akatsuki:total_playtime_years` | String | 30 min | Sum playtime in years |
| `akatsuki:new_registrations_24h` | JSON | 5 min | New users in last 24h |

---

## Implementation Phases

### Phase 1: Backend (new-cron + funcmap)

**File: `new-cron/main.py`**

Add `update_homepage_cache()` function:

```python
async def update_homepage_cache() -> None:
    print("Updating homepage cache")
    start_time = time.time()

    # Recent first places (relax mode - most popular)
    first_places = await db.fetchall("""
        SELECT sf.scoreid, sf.userid, u.username, u.country,
               b.song_name, b.beatmap_id, b.beatmapset_id,
               ROUND(s.pp) as pp, s.time as score_time
        FROM scores_first sf
        INNER JOIN users u ON u.id = sf.userid
        INNER JOIN beatmaps b ON b.beatmap_md5 = sf.beatmap_md5
        INNER JOIN scores_relax s ON s.id = sf.scoreid
        WHERE sf.rx = 1 AND u.privileges & 1
        ORDER BY s.time DESC
        LIMIT 10
    """)
    await redis.set("akatsuki:recent_first_places", json.dumps(first_places))

    # High PP plays (>800pp, last 24h)
    high_pp = await db.fetchall("""
        SELECT s.id, s.userid, ROUND(s.pp) as pp, s.time,
               u.username, u.country,
               b.song_name, b.beatmap_id, b.beatmapset_id
        FROM scores_relax s
        INNER JOIN users u ON u.id = s.userid
        INNER JOIN beatmaps b ON b.beatmap_md5 = s.beatmap_md5
        WHERE s.pp >= 800 AND s.completed = 3
          AND s.time > UNIX_TIMESTAMP() - 86400
          AND u.privileges & 1 AND b.ranked IN (2, 3)
        ORDER BY s.time DESC
        LIMIT 10
    """)
    await redis.set("akatsuki:high_pp_plays_24h", json.dumps(high_pp))

    # Trending beatmaps (most played this week)
    trending = await db.fetchall("""
        SELECT b.beatmap_id, b.beatmapset_id, b.song_name, COUNT(*) as play_count
        FROM (
            SELECT beatmap_md5 FROM scores WHERE time > UNIX_TIMESTAMP() - 604800
            UNION ALL
            SELECT beatmap_md5 FROM scores_relax WHERE time > UNIX_TIMESTAMP() - 604800
        ) recent
        INNER JOIN beatmaps b ON b.beatmap_md5 = recent.beatmap_md5
        WHERE b.ranked IN (2, 3)
        GROUP BY b.beatmap_id
        ORDER BY play_count DESC
        LIMIT 10
    """)
    await redis.set("akatsuki:trending_beatmaps", json.dumps(trending))

    # Simple counts
    user_count = await db.fetch("SELECT COUNT(*) as cnt FROM users WHERE privileges & 1")
    await redis.set("akatsuki:registered_users", str(user_count["cnt"]))

    beatmap_count = await db.fetch("SELECT COUNT(*) as cnt FROM beatmaps WHERE ranked IN (2, 3)")
    await redis.set("akatsuki:ranked_beatmaps", str(beatmap_count["cnt"]))

    playtime = await db.fetch("SELECT SUM(playtime) as total FROM user_stats")
    years = int(playtime["total"]) // (60 * 60 * 24 * 365)
    await redis.set("akatsuki:total_playtime_years", str(years))

    # New registrations
    new_users = await db.fetchall("""
        SELECT id, username, country, register_datetime
        FROM users WHERE privileges & 1
          AND register_datetime > UNIX_TIMESTAMP() - 86400
        ORDER BY register_datetime DESC
        LIMIT 10
    """)
    await redis.set("akatsuki:new_registrations_24h", json.dumps(new_users))

    print(f"Updated homepage cache in {time.time() - start_time:.2f}s")
```

Add to `main()`:
```python
await update_homepage_cache()
```

**File: `hanayo/app/usecases/funcmap/funcmap.go`**

Add JSON Redis helper (after line 525):
```go
"redigetJSON": func(k string) []map[string]interface{} {
    x := services.RD.Get(k)
    if x == nil || x.Err() != nil {
        return nil
    }
    var result []map[string]interface{}
    json.Unmarshal([]byte(x.Val()), &result)
    return result
},
```

---

### Phase 2: Hero Enhancement

**File: `hanayo/web/templates/homepage.html`**

Add live player count to hero:
```html
<div class="hero-live-stats">
  {{ with bget "onlineUsers" }}
    <span class="pulse-dot"></span>
    <span class="live-count">{{ with .result }}{{ humanize . }}{{ end }} online now</span>
  {{ end }}
</div>
```

---

### Phase 3: Enhanced Stats Grid

Replace the 6-box grid with an expanded 9-box grid:

```html
<div class="stats-grid">
  <!-- Existing 3: Score counts -->
  <div class="stat-box">...</div>  <!-- Vanilla scores -->
  <div class="stat-box">...</div>  <!-- Relax scores -->
  <div class="stat-box">...</div>  <!-- AP scores -->

  <!-- Existing 3: Top scores -->
  <div class="stat-box">...</div>  <!-- Top vanilla -->
  <div class="stat-box">...</div>  <!-- Top relax -->
  <div class="stat-box">...</div>  <!-- Top AP -->

  <!-- NEW 3: Community stats -->
  <div class="stat-box">
    {{ $users := rediget "akatsuki:registered_users" }}
    <p class="stat-value">{{ with atoi $users }}{{ humanize . }}{{ end }}</p>
    <p class="stat-label">Registered Players</p>
  </div>
  <div class="stat-box">
    {{ $maps := rediget "akatsuki:ranked_beatmaps" }}
    <p class="stat-value">{{ with atoi $maps }}{{ humanize . }}{{ end }}</p>
    <p class="stat-label">Ranked Beatmaps</p>
  </div>
  <div class="stat-box">
    {{ $years := rediget "akatsuki:total_playtime_years" }}
    <p class="stat-value">{{ with atoi $years }}{{ humanize . }}{{ end }}</p>
    <p class="stat-label">Years of Playtime</p>
  </div>
</div>
```

---

### Phase 4: Activity Feed

Add a tabbed activity section showing server life:

```html
<div class="activity-section">
  <h2 class="section-title">What's Happening</h2>

  <div class="activity-tabs">
    <button class="activity-tab active" data-tab="first-places">First Places</button>
    <button class="activity-tab" data-tab="high-pp">High PP</button>
    <button class="activity-tab" data-tab="trending">Trending Maps</button>
  </div>

  <div class="activity-content active" id="first-places">
    {{ range redigetJSON "akatsuki:recent_first_places" }}
      <div class="activity-item">
        <img src="https://a.akatsuki.gg/{{ .userid }}" class="activity-avatar" loading="lazy">
        <div class="activity-details">
          <a href="/u/{{ .userid }}">{{ .username }}</a>
          claimed <strong>#1</strong> on
          <a href="/b/{{ .beatmap_id }}">{{ .song_name }}</a>
        </div>
        <span class="activity-pp">{{ .pp }}pp</span>
      </div>
    {{ else }}
      <p class="no-activity">No recent first places</p>
    {{ end }}
  </div>

  <div class="activity-content" id="high-pp">
    {{ range redigetJSON "akatsuki:high_pp_plays_24h" }}
      <div class="activity-item">
        <img src="https://a.akatsuki.gg/{{ .userid }}" class="activity-avatar" loading="lazy">
        <div class="activity-details">
          <a href="/u/{{ .userid }}">{{ .username }}</a>
          set <strong>{{ .pp }}pp</strong> on
          <a href="/b/{{ .beatmap_id }}">{{ .song_name }}</a>
        </div>
      </div>
    {{ else }}
      <p class="no-activity">No high PP plays in last 24h</p>
    {{ end }}
  </div>

  <div class="activity-content" id="trending">
    {{ range redigetJSON "akatsuki:trending_beatmaps" }}
      <a href="/b/{{ .beatmap_id }}" class="trending-map">
        <img src="https://assets.ppy.sh/beatmaps/{{ .beatmapset_id }}/covers/card.jpg"
             class="trending-bg" loading="lazy">
        <div class="trending-info">
          <span class="trending-name">{{ .song_name }}</span>
          <span class="trending-plays">{{ humanize (float .play_count) }} plays</span>
        </div>
      </a>
    {{ else }}
      <p class="no-activity">No trending maps</p>
    {{ end }}
  </div>
</div>
```

---

### Phase 5: Game Mode Showcase

Highlight the unique modes (relax/autopilot) prominently:

```html
<div class="modes-section">
  <h2 class="section-title">Choose Your Playstyle</h2>
  <div class="mode-cards">
    <div class="mode-card">
      <div class="mode-icon">&#x1F3B9;</div>
      <h3>Vanilla</h3>
      <p>Classic osu! with competitive PP rankings</p>
      {{ $v := rediget "ripple:submitted_scores" }}
      <span class="mode-stat">{{ with atoi $v }}{{ humanize . }}{{ end }} scores</span>
    </div>
    <div class="mode-card featured">
      <span class="mode-badge">Most Popular</span>
      <div class="mode-icon">&#x2728;</div>
      <h3>Relax</h3>
      <p>Focus on aim - automatic key presses</p>
      {{ $r := rediget "ripple:submitted_scores_relax" }}
      <span class="mode-stat">{{ with atoi $r }}{{ humanize . }}{{ end }} scores</span>
    </div>
    <div class="mode-card">
      <div class="mode-icon">&#x1F5B1;</div>
      <h3>Autopilot</h3>
      <p>Master rhythm - automatic cursor movement</p>
      {{ $a := rediget "ripple:submitted_scores_ap" }}
      <span class="mode-stat">{{ with atoi $a }}{{ humanize . }}{{ end }} scores</span>
    </div>
  </div>
</div>
```

---

### Phase 6: Personalized Section (Logged-in Only)

```html
{{ if .Context.User.ID }}
<div class="personalized-section">
  <h2 class="section-title">Welcome back, {{ .Context.User.Username }}</h2>
  <!-- Quick links to profile, settings, recent plays -->
  <div class="quick-actions">
    <a href="/u/{{ .Context.User.ID }}" class="quick-action">My Profile</a>
    <a href="/settings" class="quick-action">Settings</a>
    <a href="/leaderboard" class="quick-action">Leaderboards</a>
  </div>
</div>
{{ end }}
```

---

### Phase 7: CSS Styling

**File: `hanayo/web/src/css/pages/home.css`**

Add styles for new sections (activity feed, mode cards, personalized section). Key patterns:
- Use existing brand colors from `akatsuki.css`
- Grid layouts for responsiveness
- Card-based design with subtle shadows
- Tab switching via JS class toggling

---

### Phase 8: JavaScript

**File: `hanayo/web/src/js/pages/home.js`** (new file)

Simple tab switching:
```javascript
document.querySelectorAll('.activity-tab').forEach(tab => {
  tab.addEventListener('click', () => {
    document.querySelectorAll('.activity-tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.activity-content').forEach(c => c.classList.remove('active'));
    tab.classList.add('active');
    document.getElementById(tab.dataset.tab).classList.add('active');
  });
});
```

---

## Files to Modify

| File | Changes |
|------|---------|
| `new-cron/main.py` | Add `update_homepage_cache()` function |
| `hanayo/app/usecases/funcmap/funcmap.go` | Add `redigetJSON` helper |
| `hanayo/web/templates/homepage.html` | Complete redesign with new sections |
| `hanayo/web/src/css/pages/home.css` | Styles for activity feed, mode cards, etc. |
| `hanayo/web/src/js/pages/home.js` | Tab switching (new file) |
| `hanayo/web/templates/base.html` | Include home.js on homepage |

---

## Verification

1. **Run new-cron manually**: `cd new-cron && python main.py`
2. **Verify Redis keys**: `redis-cli GET akatsuki:recent_first_places`
3. **Run hanayo**: `cd hanayo && ./run-server.sh`
4. **Test homepage**: Check all sections render, tabs work
5. **Test logged-out view**: Verify personalized section hidden
6. **Test mobile**: Check responsive layout in DevTools
7. **Run gulp build**: `cd hanayo/web && gulp`

---

## Performance Notes

- All expensive queries run in new-cron (every 5-30 min), not at render time
- `rediget` and `redigetJSON` are O(1) Redis GETs
- Images use `loading="lazy"` for deferred loading
- Beatmap images from ppy.sh CDN (fast, cached)
