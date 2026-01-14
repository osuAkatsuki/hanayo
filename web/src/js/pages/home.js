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

// Mode selector functionality
document.querySelectorAll('.mode-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    // Update active state
    document.querySelectorAll('.mode-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');

    // Get selected mode (rx value: 0=vanilla, 1=relax, 2=autopilot)
    const rx = btn.dataset.rx;

    // TODO: Fetch mode-filtered activity data via API
    // For now, mode selector just updates UI state
    // Future: fetch from /api/v1/homepage/activity?rx={rx}
    console.log('Mode selected:', rx);
  });
});
