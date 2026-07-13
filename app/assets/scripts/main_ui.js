// --- GLOBAL UI STATE ---
document.addEventListener('DOMContentLoaded', () => {
    // 2. Set default sidebar state
    document.body.setAttribute('data-sidebar', 'closed');
});

// --- CORE UI FUNCTIONS ---
function setSidebarState(state) {
    document.body.setAttribute('data-sidebar', state);
}
