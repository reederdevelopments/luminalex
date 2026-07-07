// --- GLOBAL UI STATE ---
const loaderOverlay = document.getElementById('iframe-loading');
const mainIframe = document.getElementById('main-view');
const nativeToolView = document.getElementById('native-tool-view');
const toolTitle = document.getElementById('tool-title');

const POST_LOAD_BUFFER_MS = 2500;
const MAX_WAIT_MS = 15000;

let hideTimer = null;
let maxTimer = null;

function showLoader(url) {
    if (!loaderOverlay) return;
    
    // Only show the loading animation for Looker/DataStudio dashboards
    if (url && !(url.includes('lookerstudio.google.com') || url.includes('datastudio.google.com'))) {
        return;
    }
    
    loaderOverlay.classList.remove('hidden');
    // Request animation frame ensures display block triggers before fading in
    requestAnimationFrame(() => {
        loaderOverlay.style.opacity = '1';
    });
    
    clearTimeout(hideTimer);
    clearTimeout(maxTimer);
    maxTimer = setTimeout(hideLoader, MAX_WAIT_MS);
}

function hideLoader() {
    if (!loaderOverlay) return;
    clearTimeout(hideTimer);
    clearTimeout(maxTimer);
    
    loaderOverlay.style.opacity = '0';
    hideTimer = setTimeout(() => { 
        loaderOverlay.classList.add('hidden'); 
    }, 300); // Wait for the transition to finish
}

// --- USER ENVIRONMENT INITIALIZATION ---
document.addEventListener('DOMContentLoaded', () => {
    // 1. Safely extract Go variables from the DOM bridge
    const userEnv = document.getElementById('user-env');
    if (userEnv) {
        window.userConfig = { name: userEnv.getAttribute('data-name') };
        window.currentCountryCode = userEnv.getAttribute('data-country');
        window.currentCountryName = userEnv.getAttribute('data-country-name');
    }
    
    // 2. Set default sidebar state
    document.body.setAttribute('data-sidebar', 'closed');

    // 3. Initialize Search
    initSearch();

    // 4. Initialize Favorites
    initFavorites();

    // 5. Initialize Iframe Loader Logic
    if (mainIframe && loaderOverlay) {
        if (mainIframe.getAttribute('src')) {
            showLoader(mainIframe.getAttribute('src'));
        } else {
            hideLoader();
        }

        mainIframe.addEventListener('load', () => {
            if (!mainIframe.getAttribute('src')) {
                hideLoader();
                return;
            }
            clearTimeout(hideTimer);
            hideTimer = setTimeout(hideLoader, POST_LOAD_BUFFER_MS);
        });

        // Watch for source swaps via JS
        const obs = new MutationObserver((mutations) => {
            for (const m of mutations) {
                if (m.attributeName !== 'src') continue;
                const src = mainIframe.getAttribute('src');
                if (src) {
                    showLoader(src);
                } else {
                    hideLoader();
                }
            }
        });
        obs.observe(mainIframe, { attributes: true, attributeFilter: ['src'] });
    }
});

// --- CORE UI FUNCTIONS ---
function setSidebarState(state) {
    document.body.setAttribute('data-sidebar', state);
}

function changeIframeUrl(url) {
    if(nativeToolView) nativeToolView.style.display = 'none'; 
    if(mainIframe) {
        mainIframe.style.display = 'block';    
        mainIframe.src = url;
    }
    showLoader(url);
    if (window.innerWidth <= 768) { setSidebarState('closed'); }
}

function loadNativeTool(title) {
    if(mainIframe) mainIframe.style.display = 'none';     
    if(nativeToolView) {
        nativeToolView.style.display = 'flex'; 
        if(toolTitle) toolTitle.innerText = title; 
    }
    if (window.innerWidth <= 768) { setSidebarState('closed'); }
}

// --- APP & MENU INTERACTIONS ---
window.logClick = function(name, url) {
    if(!name || !url) return;
    
    // Switch to JSON to guarantee backend parsing
    fetch('/api/track-click', { 
        method: 'POST', 
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name, url: url })
    }).catch(e => console.error("Logging failed", e));
};

window.handleAppBtnClick = function(e, el) {
    if(e) e.stopPropagation();
    const name = el.getAttribute('data-name');
    let url = el.getAttribute('data-url');
    
    if (window.searchDatabase) {
        const match = window.searchDatabase.find(item => item.name === name);
        if (match && !match.isTool) {
            url = window.buildLookerUrl(match.url, match.params, match.code);
        }
    }
    
    window.logClick(name, url);
    changeIframeUrl(url);
};

window.handleToolClick = function(e, el) {
    if(e) e.stopPropagation();
    const toolName = el.getAttribute('data-tool');
    const toolUrl = el.getAttribute('data-url');
    window.logClick(toolName, toolUrl);
    changeIframeUrl(toolUrl);
};

// Enforcing inline styles alongside Tailwind classes for bulletproof dropdowns
window.toggleMenu = function(e, el) {
    if(e) e.stopPropagation();
    const content = el.nextElementSibling;
    if (!content) return;

    const isHidden = content.classList.contains('hidden') || content.style.display === 'none';
    
    document.querySelectorAll('.menu-dropdown-content').forEach(dropdown => {
        dropdown.classList.add('hidden');
        dropdown.classList.remove('flex');
        dropdown.style.display = 'none'; 
    });
    
    if (isHidden) {
        content.classList.remove('hidden');
        content.classList.add('flex');
        content.style.display = 'flex'; 
    }
};

window.togglePages = function(e, el) {
    if(e) e.stopPropagation();
    const content = el.parentElement.nextElementSibling;
    if (!content) return;

    if (content.classList.contains('hidden') || content.style.display === 'none') {
        content.classList.remove('hidden');
        content.classList.add('flex');
        content.style.display = 'flex';
    } else {
        content.classList.add('hidden');
        content.classList.remove('flex');
        content.style.display = 'none';
    }
};

// --- DYNAMIC URL BUILDER ---
window.buildLookerUrl = function(baseUrl, p, pageCode) {
    if (!baseUrl) return "";
    let finalUrl = baseUrl;
    
    if (pageCode) {
        if (finalUrl.includes('/page/')) {
            finalUrl = finalUrl.replace(/\/page\/[^\/\?]+/, '/page/' + pageCode);
        } else {
            let separator = finalUrl.endsWith('/') ? '' : '/';
            finalUrl = finalUrl + separator + 'page/' + pageCode;
        }
    }

    if (!p) return finalUrl;

    let reportParams = {};
    const sep = '\uE000';
    const prefix = `include${sep}0${sep}IN${sep}`;
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = now.getDate();

    const pCountry = p.CountryCode || p.countryCode;
    const pConsultant = p.Consultant || p.consultant;
    const pPeriod = p.Period || p.period;
    const pDate = p.Date || p.date;
    const pCycle = p.Cycle || p.cycle;
    const pPrev = p.PrevCycle || p.prevCycle;

    if (pCountry) reportParams[pCountry] = prefix + (window.currentCountryName || '');
    if (pConsultant) reportParams[pConsultant] = prefix + (window.userConfig?.name || '');
    if (pPeriod) reportParams[pPeriod] = prefix + `${year}${month}`;
    if (pDate) reportParams[pDate] = prefix + `${year}-${month}-${String(day).padStart(2, '0')}`;

    if (pCycle || pPrev) {
        const cc = (window.currentCountryCode || 'za').toLowerCase();
        const cutoff = (cc === 'za') ? 4 : 10;
        
        let currentCycleTime = new Date(now);
        if (day <= cutoff) {
            currentCycleTime.setMonth(currentCycleTime.getMonth() - 1);
        }
        
        if (pCycle) {
            const cYear = currentCycleTime.getFullYear();
            const cMonth = String(currentCycleTime.getMonth() + 1).padStart(2, '0');
            reportParams[pCycle] = prefix + `${cYear}${cMonth}`;
        }

        if (pPrev) {
            let prevCycleTime = new Date(currentCycleTime);
            prevCycleTime.setMonth(prevCycleTime.getMonth() - 1);
            const pYear = prevCycleTime.getFullYear();
            const pMonth = String(prevCycleTime.getMonth() + 1).padStart(2, '0');
            reportParams[pPrev] = prefix + `${pYear}${pMonth}`;
        }
    }

    if (Object.keys(reportParams).length > 0) {
        const encodedParams = encodeURIComponent(JSON.stringify(reportParams));
        const char = finalUrl.includes('?') ? '&' : '?';
        finalUrl += char + 'params=' + encodedParams;
    }

    return finalUrl;
};

window.handleDashboardClick = function(e, el, pageCode = null) {
    if(e) e.stopPropagation();
    const baseUrl = el.getAttribute('data-url');
    const dashName = el.innerText;
    
    const p = {
        CountryCode: el.getAttribute('data-p-country'),
        Branch: el.getAttribute('data-p-branch'),
        Consultant: el.getAttribute('data-p-consultant'),
        Cycle: el.getAttribute('data-p-cycle'),
        PrevCycle: el.getAttribute('data-p-prev'),
        Date: el.getAttribute('data-p-date'),
        Period: el.getAttribute('data-p-period')
    };

    const finalUrl = window.buildLookerUrl(baseUrl, p, pageCode);
    window.logClick(dashName, finalUrl);
    changeIframeUrl(finalUrl);
};

// --- SYSTEM SEARCH ---
function initSearch() {
    const searchInputMenu = document.getElementById('menu-search-input');
    const searchResultsMenu = document.getElementById('menu-search-results');
    const searchDbEl = document.getElementById('search-db');
    
    let searchDatabase = [];
    if (searchDbEl) {
        try { 
            searchDatabase = JSON.parse(searchDbEl.textContent); 
            window.searchDatabase = searchDatabase; 
        } 
        catch(e) { console.error("Failed to load search data:", e); }
    }

    if (searchInputMenu && searchResultsMenu) {
        searchInputMenu.addEventListener('input', function() {
            const query = this.value.toLowerCase().trim();
            searchResultsMenu.innerHTML = ''; 
            
            if (query === '') {
                searchResultsMenu.style.display = 'none';
                return;
            }

            const results = searchDatabase.filter(item => item.name.toLowerCase().includes(query));

            if (results.length > 0) {
                results.forEach(item => {
                    const resultDiv = document.createElement('div');
                    resultDiv.className = 'search-result-item';
                    resultDiv.innerHTML = `${item.name} <span class="search-type-tag">[${item.type}]</span>`;
                    
                    resultDiv.addEventListener('click', () => {
                        let finalUrl = item.url;
                        if (item.type === "Page" || item.type === "Dashboard") {
                            finalUrl = window.buildLookerUrl(item.url, item.params, item.code);
                        }
                        
                        window.logClick(item.name, finalUrl);

                        if (item.isTool) {
                            loadNativeTool(finalUrl); 
                        } else {
                            changeIframeUrl(finalUrl);
                        }
                        searchInputMenu.value = '';
                        searchResultsMenu.style.display = 'none';
                    });

                    searchResultsMenu.appendChild(resultDiv);
                });
                searchResultsMenu.style.display = 'block';
            } else {
                searchResultsMenu.style.display = 'none';
            }
        });

        document.addEventListener('click', function(e) {
            if (!searchInputMenu.contains(e.target) && !searchResultsMenu.contains(e.target)) {
                searchResultsMenu.style.display = 'none';
            }
        });
    }
}

// --- FAVORITES & SORTING ---
function initFavorites() {
    window.userFavorites = [];
    const favDbEl = document.getElementById('user-favorites-data');
    if (favDbEl) {
        try { 
            const parsed = JSON.parse(favDbEl.textContent); 
            if (Array.isArray(parsed)) {
                window.userFavorites = parsed;
            }
        } 
        catch(e) { console.error("Failed to load favorites data:", e); }
    }

    const favList = document.getElementById('my-favorites-list');
    if (favList && typeof Sortable !== 'undefined') {
        Sortable.create(favList, {
            animation: 150,
            draggable: '.fav-item',
            onEnd: function() {
                const newOrder = [];
                favList.querySelectorAll('.fav-item').forEach(el => {
                    newOrder.push(el.getAttribute('data-id'));
                });
                window.userFavorites = newOrder;
                saveFavorites();
            }
        });
    }
}

window.toggleFavorite = function(e, btn) {
    if(e) e.stopPropagation();
    const id = btn.getAttribute('data-id');
    const idx = window.userFavorites.indexOf(id);
    const favList = document.getElementById('my-favorites-list');
    const emptyMsg = document.getElementById('my-favorites-empty');

    if (idx === -1) {
        // ADD FAVORITE
        window.userFavorites.push(id);
        btn.innerHTML = `<svg class="w-3 h-3 text-yellow-400" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>`;
        
        // Build list item to inject
        if (favList) {
            const name = btn.getAttribute('data-name');
            const url = btn.getAttribute('data-url') || btn.getAttribute('data-tool');
            const type = url ? (url.includes('/tools/') ? 'Tool' : (id.startsWith('Page:') ? 'Page' : 'Dashboard')) : 'Dashboard';
            const code = btn.getAttribute('data-code') || '';
            
            const pCountry = btn.getAttribute('data-p-country') || '';
            const pBranch = btn.getAttribute('data-p-branch') || '';
            const pConsultant = btn.getAttribute('data-p-consultant') || '';
            const pCycle = btn.getAttribute('data-p-cycle') || '';
            const pPrev = btn.getAttribute('data-p-prev') || '';
            const pDate = btn.getAttribute('data-p-date') || '';
            const pPeriod = btn.getAttribute('data-p-period') || '';
            
            const itemHtml = `
                <div class="fav-item flex items-center justify-between px-4 py-3 border-b border-gray-100 hover:bg-gray-50 cursor-grab group/favitem transition-colors bg-white" data-id="${id}">
                    <div class="flex items-center gap-3 overflow-hidden flex-1 cursor-pointer"
                        data-name="${name}" data-url="${url}" data-type="${type}" data-code="${code}"
                        data-p-country="${pCountry}" data-p-branch="${pBranch}" data-p-consultant="${pConsultant}"
                        data-p-cycle="${pCycle}" data-p-prev="${pPrev}" data-p-date="${pDate}" data-p-period="${pPeriod}"
                        onclick="window.handleFavBtnClick(event, this)">
                        <span class="text-xs font-bold text-gray-700 truncate">${name}</span>
                    </div>
                    <button class="text-gray-300 hover:text-red-500 transition-colors w-5 h-5 flex items-center justify-center shrink-0"
                        onclick="window.removeFavorite(event, this, '${id}')">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                </div>
            `;
            favList.insertAdjacentHTML('beforeend', itemHtml);
            if (emptyMsg) emptyMsg.style.display = 'none';
        }
    } else {
        // REMOVE FAVORITE
        window.userFavorites.splice(idx, 1);
        btn.innerHTML = `<svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>`;
        
        // Remove from list if present
        if (favList) {
            const item = favList.querySelector(`[data-id="${id}"]`);
            if (item) item.remove();
            if (window.userFavorites.length === 0 && emptyMsg) emptyMsg.style.display = 'block';
        }
    }
    
    saveFavorites();
};

window.removeFavorite = function(e, btn, id) {
    if (e) e.stopPropagation();
    const idx = window.userFavorites.indexOf(id);
    const emptyMsg = document.getElementById('my-favorites-empty');

    if (idx !== -1) {
        window.userFavorites.splice(idx, 1);
        saveFavorites();
        
        // Remove from list
        const item = btn.closest('.fav-item');
        if (item) item.remove();
        
        if (window.userFavorites.length === 0 && emptyMsg) emptyMsg.style.display = 'block';

        // Reset star in sidebar
        const sidebarStar = document.querySelector(`button[data-id="${id}"]`);
        if (sidebarStar) {
            sidebarStar.innerHTML = `<svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>`;
        }
    }
};

window.handleFavBtnClick = function(e, btn) {
    if(e) e.stopPropagation();
    const type = btn.getAttribute('data-type');
    let url = btn.getAttribute('data-url');
    const name = btn.getAttribute('data-name');
    
    if (window.searchDatabase && type !== 'Tool') {
        const match = window.searchDatabase.find(item => item.name === name);
        if (match) {
            url = window.buildLookerUrl(match.url, match.params, match.code);
        } else {
             const p = {
                CountryCode: btn.getAttribute('data-p-country'),
                Branch: btn.getAttribute('data-p-branch'),
                Consultant: btn.getAttribute('data-p-consultant'),
                Cycle: btn.getAttribute('data-p-cycle'),
                PrevCycle: btn.getAttribute('data-p-prev'),
                Date: btn.getAttribute('data-p-date'),
                Period: btn.getAttribute('data-p-period')
            };
            const code = btn.getAttribute('data-code');
            url = window.buildLookerUrl(url, p, code);
        }
    }
    
    window.logClick(name, url);

    if (type === 'Tool') {
        changeIframeUrl(url); 
    } else {
        changeIframeUrl(url);
    }
}

function saveFavorites() {
    return fetch('/api/favorites', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ favorites: window.userFavorites })
    }).catch(e => console.error("Saving favorites failed:", e));
}

// --- FLAG DROPDOWN LOGIC ---
const flagBtn = document.getElementById('flag-btn');
const countryDropdown = document.getElementById('country-dropdown');
const currentFlagImg = document.getElementById('current-flag');

if (flagBtn) {
    flagBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        if(countryDropdown) countryDropdown.classList.toggle('hidden');
    });
}

document.querySelectorAll('.country-option').forEach(option => {
    option.addEventListener('click', function() {
        const country = this.getAttribute('data-country');
        const flagUrl = this.getAttribute('data-flag');
        const countryName = this.querySelector('span').innerText;
        
        window.currentCountryCode = country;
        window.currentCountryName = countryName;

        if(currentFlagImg) currentFlagImg.src = flagUrl;
        if(countryDropdown) countryDropdown.classList.add('hidden');
    });
});

document.addEventListener('click', function(e) {
    if (flagBtn && !flagBtn.contains(e.target) && countryDropdown && !countryDropdown.contains(e.target)) {
        countryDropdown.classList.add('hidden');
    }
});