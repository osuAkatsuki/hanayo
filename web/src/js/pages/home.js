// Activity feed tab switching
document.querySelectorAll('.activity-tab').forEach(tab => {
  tab.addEventListener('click', () => {
    // Remove active class from all tabs and content
    document.querySelectorAll('.activity-tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.activity-content').forEach(c => c.classList.remove('active'));

    // Add active class to clicked tab and corresponding content
    tab.classList.add('active');
    document.getElementById(tab.dataset.tab).classList.add('active');
  });
});

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
async function updateActivityData(rx) {
  try {
    const response = await fetch(`/api/v1/homepage/activity?rx=${rx}`);
    if (!response.ok) throw new Error('Failed to fetch activity data');

    const data = await response.json();

    // Update first places
    const firstPlacesEl = document.getElementById('first-places');
    if (firstPlacesEl) {
      renderActivityContent(firstPlacesEl, data.first_places, 'first-places');
    }

    // Update high PP
    const highPPEl = document.getElementById('high-pp');
    if (highPPEl) {
      renderActivityContent(highPPEl, data.high_pp, 'high-pp');
    }
  } catch (error) {
    console.error('Error fetching activity data:', error);
  }
}

// Mode selector functionality
document.querySelectorAll('.mode-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    // Update active state
    document.querySelectorAll('.mode-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');

    // Get selected mode (rx value: 0=vanilla, 1=relax, 2=autopilot)
    const rx = btn.dataset.rx;

    // Fetch and update activity data
    updateActivityData(rx);
  });
});
