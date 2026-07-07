// Complex Editor Initialization and Toggles
window.initJobConfigEditor = function() {
    // --- Field Toggles ---
    window.toggleJobType = function(val) {
        document.getElementById('jc-link-block').style.display = val === 'cloud-function' ? 'block' : 'none';
        document.getElementById('jc-repo-block').style.display = val === 'cloud-function' ? 'none' : 'grid';
        document.getElementById('jc-ustc-block').style.display = val === 'cloud-function' ? 'none' : 'block';
    };
    
    window.toggleSchedType = function(val) {
        const defInput = document.getElementById('jc-cron-default');
        if (val === 'cyclical') {
            defInput.disabled = true;
            defInput.classList.add('bg-gray-100', 'cursor-not-allowed', 'text-gray-400');
            document.getElementById('jc-cron-override-block').style.display = 'none';
            document.getElementById('jc-cron-disabled-msg').style.display = 'block';
        } else {
            defInput.disabled = false;
            defInput.classList.remove('bg-gray-100', 'cursor-not-allowed', 'text-gray-400');
            document.getElementById('jc-cron-override-block').style.display = 'block';
            document.getElementById('jc-cron-disabled-msg').style.display = 'none';
        }
    };

    const typeSelect = document.getElementById('jc-type');
    const schedSelect = document.getElementById('jc-schedule');
    if(typeSelect) window.toggleJobType(typeSelect.value);
    if(schedSelect) window.toggleSchedType(schedSelect.value);

    const safeParse = (str, fallback) => { try { return JSON.parse(str || fallback); } catch(e) { return JSON.parse(fallback); } };

    // 1. Countries Override
    const cJsonEl = document.getElementById('jc-countries-json');
    if(cJsonEl) {
        const cJsonVal = safeParse(cJsonEl.value, '[]');
        document.querySelectorAll('[data-country-cb]').forEach(cb => { if(cJsonVal.includes(cb.value)) cb.checked = true; });
    }

    // 2. Cron Override
    const cronContainer = document.getElementById('cron-rows-container');
    const cronJsonEl = document.getElementById('jc-cron-json');
    if (cronContainer && cronJsonEl) {
        const initCron = safeParse(cronJsonEl.value, '{}');
        window.addCronRow = (key = '', val = '') => {
            const div = document.createElement('div');
            div.className = 'flex gap-2 items-start';
            div.innerHTML = `
                <input type="text" placeholder="Country (e.g. za)" value="${key}" class="w-1/3 border border-gray-200 rounded px-2 py-1.5 text-xs focus:border-hl outline-none"/>
                <div class="flex-1">
                    <input type="text" placeholder="Cron (e.g. 0 8 * * *)" value="${val}" oninput="window.updateCronDesc(this)" class="w-full border border-gray-200 rounded px-2 py-1.5 text-xs font-mono focus:border-hl outline-none"/>
                    <span class="cron-desc text-[10px] font-bold text-green-x-light mt-1 block"></span>
                </div>
                <button type="button" onclick="this.parentElement.remove()" class="text-red-400 hover:text-red-600 font-bold px-2 pt-[0.35rem]">X</button>
            `;
            cronContainer.appendChild(div);
            const cInp = div.querySelector('input:nth-child(2)');
            if (cInp && val) window.updateCronDesc(cInp);
        };
        Object.entries(initCron).forEach(([k,v]) => window.addCronRow(k,v));
    }

    // --- Master Sync Function (Called on Submit) ---
    window.syncAllJSONs = () => {
        const cArr = [];
        document.querySelectorAll('[data-country-cb]:checked').forEach(cb => cArr.push(cb.value));
        if(document.getElementById('jc-countries-json')) document.getElementById('jc-countries-json').value = JSON.stringify(cArr);

        const cronMap = {};
        if(cronContainer) {
            cronContainer.querySelectorAll('div').forEach(row => {
                const inputs = row.querySelectorAll('input');
                if(inputs[0].value && inputs[1].value) cronMap[inputs[0].value.trim()] = inputs[1].value.trim();
            });
            document.getElementById('jc-cron-json').value = JSON.stringify(cronMap);
        }

        // Add USTC and Mailing list sync logic from previous block here...
    };
}

// Re-init editors when HTMX swaps the form into the DOM
document.body.addEventListener('htmx:afterSwap', function(evt) {
    if(evt.target.id === 'jobconfig-form-container') {
        window.initJobConfigEditor();
    }
});
