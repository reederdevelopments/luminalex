// --- GLOBAL UI STATE ---
let loaderTimer;
const loaderOverlay = document.getElementById('loader-overlay');
const mainIframe = document.getElementById('main-view');
const nativeToolView = document.getElementById('native-tool-view');
const toolTitle = document.getElementById('tool-title');

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
});

// --- CORE UI FUNCTIONS ---
function setSidebarState(state) {
    document.body.setAttribute('data-sidebar', state);
}

function triggerLoader() {
    if (loaderOverlay) loaderOverlay.style.display = 'flex'; 
    clearTimeout(loaderTimer); 
    loaderTimer = setTimeout(() => { if (loaderOverlay) loaderOverlay.style.display = 'none'; }, 5000);
}

function changeIframeUrl(url) {
    if(nativeToolView) nativeToolView.style.display = 'none'; 
    if(mainIframe) {
        mainIframe.style.display = 'block';    
        mainIframe.src = url;
    }
    triggerLoader();
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
    const formData = new FormData();
    formData.append("name", name);
    formData.append("url", url);
    fetch('/api/track-click', { method: 'POST', body: formData })
        .catch(e => console.error("Logging failed", e));
};

window.handleAppBtnClick = function(e, el) {
    if(e) e.stopPropagation();
    const name = el.getAttribute('data-name');
    const url = el.getAttribute('data-url');
    window.logClick(name, url);
    changeIframeUrl(url);
};

window.handleToolClick = function(e, el) {
    if(e) e.stopPropagation();
    const toolName = el.getAttribute('data-tool');
    window.logClick(toolName, toolName);
    loadNativeTool(toolName);
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
        try { searchDatabase = JSON.parse(searchDbEl.textContent); } 
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