// Activity feed tab switching
document.querySelectorAll('.activity-tab').forEach(tab => {
  tab.addEventListener('click', () => {
    document.querySelectorAll('.activity-tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.activity-content').forEach(c => c.classList.remove('active'));
    tab.classList.add('active');
    document.getElementById(tab.dataset.tab).classList.add('active');
  });
});

// Current selection state
let currentRx = 0;   // 0=vanilla, 1=relax, 2=autopilot
let currentMode = 0; // 0=std, 1=taiko, 2=ctb, 3=mania

// Mode availability per rx type
const availableModes = {
  0: [0, 1, 2, 3], // Vanilla: all modes
  1: [0, 1, 2],    // Relax: std, taiko, ctb (no mania)
  2: [0]           // Autopilot: std only
};

// Update which mode buttons are enabled based on rx
function updateModeAvailability() {
  const available = availableModes[currentRx];
  document.querySelectorAll('#mode-selector .mode-btn').forEach(btn => {
    const mode = parseInt(btn.dataset.mode);
    btn.disabled = !available.includes(mode);
    btn.classList.toggle('disabled', !available.includes(mode));
  });

  // If current mode is not available, switch to first available
  if (!available.includes(currentMode)) {
    currentMode = available[0];
    document.querySelectorAll('#mode-selector .mode-btn').forEach(btn => {
      btn.classList.toggle('active', parseInt(btn.dataset.mode) === currentMode);
    });
  }
}

// Convert mods integer to string (e.g., 24 -> "HDHR")
function modsToString(mods) {
  if (!mods || mods === 0) return '';

  const modMap = {
    1: 'NF', 2: 'EZ', 8: 'HD', 16: 'HR', 64: 'DT',
    256: 'HT', 512: 'NC', 1024: 'FL', 2048: 'SO'
  };

  const parts = [];
  for (const [flag, name] of Object.entries(modMap)) {
    if (mods & parseInt(flag)) {
      parts.push(name);
    }
  }
  return parts.join('');
}

// Convert ISO 8601 string to relative time (e.g., "2y ago", "3d ago")
function timeAgo(isoString) {
  const date = new Date(isoString);
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);

  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;
  if (seconds < 2592000) return `${Math.floor(seconds / 604800)}w ago`;
  if (seconds < 31536000) return `${Math.floor(seconds / 2592000)}mo ago`;
  return `${Math.floor(seconds / 31536000)}y ago`;
}

// Create activity item element safely (no innerHTML)
function createActivityItem(item, type) {
  const userid = Math.floor(item.userid);
  const beatmapId = Math.floor(item.beatmap_id);
  const pp = Math.floor(item.pp);
  const action = type === 'first-places' ? 'claimed #1 on' : 'set';

  const div = document.createElement('div');
  div.className = 'activity-item';

  const img = document.createElement('img');
  img.src = `https://a.akatsuki.gg/${userid}`;
  img.className = 'activity-avatar';
  img.loading = 'lazy';
  img.alt = '';
  div.appendChild(img);

  const details = document.createElement('div');
  details.className = 'activity-details';

  // First row: player/action/map and PP
  const row1 = document.createElement('div');
  row1.className = 'activity-row';

  const mainDiv = document.createElement('div');
  mainDiv.className = 'activity-main';

  const playerLink = document.createElement('a');
  playerLink.href = `/u/${userid}`;
  playerLink.className = 'activity-player';
  playerLink.textContent = item.username;
  mainDiv.appendChild(playerLink);

  const actionSpan = document.createElement('span');
  actionSpan.className = 'activity-action';
  actionSpan.textContent = ` ${action} `;
  mainDiv.appendChild(actionSpan);

  const mapLink = document.createElement('a');
  mapLink.href = `/b/${beatmapId}`;
  mapLink.className = 'activity-map';
  mapLink.textContent = item.song_name;
  mainDiv.appendChild(mapLink);

  row1.appendChild(mainDiv);

  const ppSpan = document.createElement('span');
  ppSpan.className = 'activity-pp';
  ppSpan.textContent = `${pp}pp`;
  row1.appendChild(ppSpan);

  details.appendChild(row1);

  // Second row: mods/time and accuracy
  const row2 = document.createElement('div');
  row2.className = 'activity-row';

  const metaDiv = document.createElement('div');
  metaDiv.className = 'activity-meta';

  // Add mods if present
  if (item.mods) {
    const modsStr = modsToString(item.mods);
    if (modsStr) {
      const modsSpan = document.createElement('span');
      modsSpan.className = 'activity-mods';
      modsSpan.textContent = `+${modsStr}`;
      metaDiv.appendChild(modsSpan);
    }
  }

  // Add time ago
  const isoString = type === 'first-places' ? item.score_time : item.time;
  if (isoString) {
    const timeEl = document.createElement('time');
    timeEl.className = 'activity-time';
    timeEl.dateTime = isoString; // Already in ISO format
    timeEl.textContent = timeAgo(isoString);
    // Add tooltip with full date/time
    const date = new Date(isoString);
    timeEl.title = date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      timeZoneName: 'short'
    });
    metaDiv.appendChild(timeEl);
  }

  row2.appendChild(metaDiv);

  // Add accuracy
  if (item.accuracy !== undefined) {
    const accSpan = document.createElement('span');
    accSpan.className = 'activity-accuracy';
    accSpan.textContent = `${item.accuracy.toFixed(2)}%`;
    row2.appendChild(accSpan);
  }

  details.appendChild(row2);

  div.appendChild(details);

  return div;
}

// Create no-activity message
function createNoActivityMessage(type) {
  const p = document.createElement('p');
  p.className = 'no-activity';
  p.textContent = type === 'first-places' ? 'No recent first places' : 'No high PP plays in the last 24h';
  return p;
}

// Render activity content into container
function renderActivityContent(container, items, type) {
  container.replaceChildren();
  if (!items || items.length === 0) {
    container.appendChild(createNoActivityMessage(type));
    return;
  }
  items.forEach(item => container.appendChild(createActivityItem(item, type)));
}

// Fetch and update activity data
async function updateActivityData() {
  try {
    const response = await fetch(`/api/v1/homepage/activity?mode=${currentMode}&rx=${currentRx}`);
    if (!response.ok) throw new Error('Failed to fetch activity data');

    const data = await response.json();

    const firstPlacesEl = document.getElementById('first-places');
    if (firstPlacesEl) {
      renderActivityContent(firstPlacesEl, data.first_places, 'first-places');
    }

    const highPPEl = document.getElementById('high-pp');
    if (highPPEl) {
      renderActivityContent(highPPEl, data.high_pp, 'high-pp');
    }
  } catch (error) {
    console.error('Error fetching activity data:', error);
  }
}

// RX selector (Vanilla/Relax/Autopilot)
document.querySelectorAll('#rx-selector .mode-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('#rx-selector .mode-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    currentRx = parseInt(btn.dataset.rx);
    updateModeAvailability();
    updateActivityData();
  });
});

// Mode selector (osu!/Taiko/Catch/Mania)
document.querySelectorAll('#mode-selector .mode-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    if (btn.disabled) return;
    document.querySelectorAll('#mode-selector .mode-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    currentMode = parseInt(btn.dataset.mode);
    updateActivityData();
  });
});

// Initialize mode availability on page load
updateModeAvailability();

// Fetch initial data for the default mode/rx
updateActivityData();
