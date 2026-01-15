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

  const playerLink = document.createElement('a');
  playerLink.href = `/u/${userid}`;
  playerLink.className = 'activity-player';
  playerLink.textContent = item.username;
  details.appendChild(playerLink);

  const actionSpan = document.createElement('span');
  actionSpan.className = 'activity-action';
  actionSpan.textContent = ` ${action} `;
  details.appendChild(actionSpan);

  const mapLink = document.createElement('a');
  mapLink.href = `/b/${beatmapId}`;
  mapLink.className = 'activity-map';
  mapLink.textContent = item.song_name;
  details.appendChild(mapLink);

  div.appendChild(details);

  const ppSpan = document.createElement('span');
  ppSpan.className = 'activity-pp';
  ppSpan.textContent = `${pp}pp`;
  div.appendChild(ppSpan);

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
