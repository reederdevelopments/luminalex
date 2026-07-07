// Setup Cron Translator Logic (Gracefully handle App Engine crons)
window.updateCronDesc = function(inputElement) {
    const val = inputElement.value;
    const descEl = inputElement.nextElementSibling;
    if(descEl && descEl.classList.contains('cron-desc')) {
        try {
            const parsed = window.cronstrue.toString(val, { throwExceptionOnParseError: true });
            descEl.innerText = parsed;
        } catch(e) {
            if (val.trim() === '') {
                descEl.innerText = '';
            } else {
                descEl.innerText = 'App Engine / Custom Schedule';
            }
        }
    }
}

window.switchEngTab = function(tab) {
    ['reload', 'dataform', 'batch'].forEach(t => {
        document.getElementById('tab-' + t).style.display = t === tab ? 'flex' : 'none';
        const btn = document.getElementById('tab-btn-' + t);
        if(t === tab) {
            btn.className = "pb-3 text-sm font-bold tracking-widest uppercase border-b-2 border-hl text-hl transition-colors cursor-pointer";
        } else {
            btn.className = "pb-3 text-sm font-bold tracking-widest uppercase border-b-2 border-transparent text-gray-400 hover:text-gray-600 transition-colors cursor-pointer";
        }
    });
}

// Modal popover handler for Reload Tab
window.openReloadAction = function(table, database, cron, key) {
    const panel = document.getElementById('reload-action-panel');
    panel.innerHTML = `
        <div class="bg-white border border-gray-200 rounded-xl p-8 shadow-sm">
            <h2 class="text-xl font-black tracking-widest text-green-x-dark uppercase mb-2">${table}</h2>
            <p class="text-xs text-gray-400 font-bold mb-6 border-b border-gray-100 pb-4">${database}</p>
            
            <div id="reload-action-response"></div>

            <form hx-post="/admin/api/engineering/reload/action" hx-target="#reload-action-response" class="flex flex-col gap-4">
                <input type="hidden" name="table" value="${table}">
                <input type="hidden" name="database" value="${database}">
                <input type="hidden" name="key" value="${key}">
                
                <input type="hidden" name="cc" value="${document.querySelector('input[name="cc"]:checked')?.value || 'za'}">
                
                <div class="flex gap-4">
                    <div class="flex-1">
                        <label class="block text-[10px] font-bold text-gray-500 uppercase tracking-wider mb-1">Update Cron</label>
                        <input type="text" name="cron" value="${cron}" oninput="window.updateCronDesc(this)" class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:border-hl outline-none font-mono"/>
                        <span class="cron-desc text-[10px] font-bold text-green-x-light mt-1 block"></span>
                    </div>
                    <div class="flex items-start mt-[1.3rem]">
                        <button type="submit" name="action" value="update_cron" class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-bold text-xs uppercase tracking-widest px-4 py-2.5 rounded-lg transition-colors">Update</button>
                    </div>
                </div>
                
                <div class="grid grid-cols-2 gap-4 mt-6">
                    <button type="submit" name="action" value="refresh_table" class="bg-hl hover:bg-green-x-dark text-white font-bold text-xs uppercase tracking-widest px-4 py-3 rounded-lg transition-colors shadow-sm">Refresh Table</button>
                    <button type="button" class="bg-gray-900 hover:bg-gray-700 text-white font-bold text-xs uppercase tracking-widest px-4 py-3 rounded-lg transition-colors shadow-sm">Refresh All Countries</button>
                </div>
            </form>
        </div>
    `;
    htmx.process(panel);
    const cronInput = panel.querySelector('input[name="cron"]');
    if (cronInput && cron) window.updateCronDesc(cronInput);
}
