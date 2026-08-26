const CATEGORY_COLUMNS = {
    banks: ["Bank Name", "Branch", "Physical Address", "Contact Name", "Title/Rank", "Email", "Phone", "Fax", "Notes"],
    masters: ["Masters Office", "Physical Address", "Department/Division", "Contact Name", "Rank/Title", "File Number Allocation", "Email", "Phone", "Notes"],
    sheriffs: ["Sheriff Office", "Sheriff Name", "Deputy Name", "Physical Address", "Jurisdiction", "Email", "Phone", "Fax", "Notes"],
    magistrates: ["Court Name", "Physical Address", "Contact Name", "Title/Rank", "Email", "Phone", "Fax", "Court Citation", "Notes"],
    highcourts: ["Court Name", "Physical Address", "Contact Name", "Title/Rank", "Email", "Phone", "Fax", "Court Citation", "Notes"],
    lawfirms: ["Firm Name", "Physical Address", "Contact Name", "Title/Rank", "Email", "Phone", "Fax", "Notes"]
};

const appState = {
    data: {},
    editingId: null,
    isAddingNew: false,
    activeTab: 'search',
    localSearchTerm: ''
};

function generateUUID() {
    return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, c =>
        (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
    );
}

function refreshTable(category) {
    const wrapper = document.getElementById('table-scroll-wrapper');
    const scrollTop = wrapper ? wrapper.scrollTop : 0;
    
    document.getElementById('tab-content').innerHTML = renderTable(category);
    attachTableEventListeners(category);
    
    const newWrapper = document.getElementById('table-scroll-wrapper');
    if (newWrapper) {
        newWrapper.scrollTop = scrollTop;
    }
}

async function loadDataFromBackend() {
    const categories = Object.keys(CATEGORY_COLUMNS);
    for (const cat of categories) {
        try {
            const response = await fetch(`/api/contacts?category=${cat}`);
            if (!response.ok) throw new Error(`HTTP status ${response.status}`);
            const data = await response.json();
            appState.data[cat] = Array.isArray(data) ? data : [];
        } catch (err) {
            appState.data[cat] = [];
        }
    }
}

async function exportData() {
    if (appState.activeTab === 'search') return;
    
    const table = document.querySelector('#tab-content table');
    if (!table) {
        alert("No table data found to export.");
        return;
    }

    const headerCells = Array.from(table.querySelectorAll('thead th'));
    const headers = headerCells.slice(0, headerCells.length - 1).map(th => th.innerText.trim());

    const rows = [];
    const trs = table.querySelectorAll('tbody tr');
    
    trs.forEach(tr => {
        if (tr.style.display !== 'none') {
            const cells = Array.from(tr.querySelectorAll('td'));
            if (cells.length === headerCells.length) {
                const rowData = cells.slice(0, cells.length - 1).map(td => {
                    const input = td.querySelector('textarea, input');
                    return input ? input.value.trim() : td.innerText.trim();
                });
                rows.push(rowData);
            }
        }
    });

    if (rows.length === 0) {
        alert("No visible rows to export based on your current filter.");
        return;
    }

    try {
        const payload = {
            filename: appState.activeTab.toUpperCase(),
            headers: headers,
            rows: rows
        };

        const response = await fetch('/api/contacts/export_view', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errText = await response.text();
            alert("Export failed: " + errText);
        }
    } catch (e) {
        alert("Export failed: " + e.message);
    }
}

async function saveRecord(id, category, fields) {
    const record = {
        id: id || generateUUID(),
        category: category,
        fields: fields
    };

    const wrapper = document.getElementById('table-scroll-wrapper');
    const scrollTop = wrapper ? wrapper.scrollTop : 0;

    const res = await fetch('/api/contacts/save', {
        method: 'POST',
        body: JSON.stringify(record),
        headers: { 'Content-Type': 'application/json' }
    });

    if (!res.ok) throw new Error("Failed to save record");

    appState.editingId = null;
    appState.isAddingNew = false;
    await loadDataFromBackend();
    
    if (appState.activeTab !== 'search') {
        document.getElementById('tab-content').innerHTML = renderTable(appState.activeTab);
        attachTableEventListeners(appState.activeTab);
        
        const newWrapper = document.getElementById('table-scroll-wrapper');
        if (newWrapper) {
            newWrapper.scrollTop = scrollTop;
        }
    }
}

async function deleteRecord(id, category) {
    if (!confirm("Are you sure you want to delete this record? This action cannot be undone.")) return;
    
    const wrapper = document.getElementById('table-scroll-wrapper');
    const scrollTop = wrapper ? wrapper.scrollTop : 0;

    const res = await fetch('/api/contacts/delete', {
        method: 'POST',
        body: JSON.stringify({ id, category }),
        headers: { 'Content-Type': 'application/json' }
    });

    if (!res.ok) {
        alert("Failed to delete record");
        return;
    }

    await loadDataFromBackend();
    if (appState.activeTab !== 'search') {
        document.getElementById('tab-content').innerHTML = renderTable(appState.activeTab);
        attachTableEventListeners(appState.activeTab);
        
        const newWrapper = document.getElementById('table-scroll-wrapper');
        if (newWrapper) {
            newWrapper.scrollTop = scrollTop;
        }
    }
}

function renderSearch() {
    const catOptions = [
        { val: '', label: 'All Directories' },
        { val: 'banks', label: 'Banks' },
        { val: 'masters', label: 'Master\'s Office' },
        { val: 'sheriffs', label: 'Sheriffs' },
        { val: 'magistrates', label: 'Magistrate Courts' },
        { val: 'highcourts', label: 'High Courts' },
        { val: 'lawfirms', label: 'Law Firms' }
    ].map(o => `<option value="${o.val}" class="bg-black text-[#f9f8f6]">${o.label}</option>`).join('');

    return `
        <div class="flex flex-col gap-6 p-4 w-full h-full">
            <div class="flex flex-col gap-4 bg-black/40 border border-[#a68a56]/30 p-4 rounded">
                <div class="flex gap-4">
                    <input type="text" id="search-global" placeholder="SEARCH ALL DIRECTORIES..." class="flex-1 bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-sm focus:outline-none focus:border-[#a68a56] transition-colors uppercase tracking-widest font-sans rounded" />
                    <select id="search-category" class="w-64 bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-sm focus:outline-none focus:border-[#a68a56] transition-colors uppercase tracking-widest font-sans cursor-pointer rounded">
                        ${catOptions}
                    </select>
                </div>
                <div id="search-column-filters" class="hidden grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 pt-4 border-t border-[#a68a56]/20"></div>
            </div>
            <div id="search-results" class="w-full flex-1 overflow-auto border border-[#a68a56]/30 bg-black/20 rounded"></div>
        </div>
    `;
}

function executeSearch() {
    const globalInput = document.getElementById('search-global');
    const catSelect = document.getElementById('search-category');
    const colFiltersDiv = document.getElementById('search-column-filters');
    const resultsDiv = document.getElementById('search-results');

    const renderResults = () => {
        const globalTerm = globalInput.value.toLowerCase();
        const selectedCat = catSelect.value;
        const colInputs = document.querySelectorAll('.col-filter-input');
        
        let colFilters = [];
        colInputs.forEach(inp => {
            const val = inp.value.toLowerCase();
            if (val) colFilters.push({ idx: parseInt(inp.getAttribute('data-col-idx')), term: val });
        });

        let matched = [];
        const categoriesToSearch = selectedCat ? [selectedCat] : Object.keys(appState.data);

        categoriesToSearch.forEach(cat => {
            const records = appState.data[cat] || [];
            records.forEach(rec => {
                let isMatch = true;
                if (globalTerm) {
                    const rowText = (rec.fields || []).join(' ').toLowerCase();
                    if (!rowText.includes(globalTerm)) isMatch = false;
                }
                if (isMatch && colFilters.length > 0) {
                    colFilters.forEach(cf => {
                        const fieldVal = (rec.fields && rec.fields[cf.idx]) ? rec.fields[cf.idx].toLowerCase() : '';
                        if (!fieldVal.includes(cf.term)) isMatch = false;
                    });
                }
                if (isMatch) matched.push({ ...rec, category: cat });
            });
        });

        if (matched.length === 0) {
            resultsDiv.innerHTML = '<div class="p-8 text-center text-[#f9f8f6]/50 text-sm uppercase tracking-widest">No matching records found.</div>';
            return;
        }

        if (selectedCat) {
            const cols = CATEGORY_COLUMNS[selectedCat];
            let html = `<table class="w-full text-left border-collapse table-fixed break-words"><thead class="sticky top-0 bg-[#0b1d2e] z-10 shadow-md"><tr class="border-b-2 border-[#a68a56]/50 text-[#a68a56] text-[10px] uppercase tracking-widest">`;
            cols.forEach(col => { html += `<th class="p-4 font-normal">${col}</th>`; });
            html += `</tr></thead><tbody>`;
            matched.forEach(m => {
                html += `<tr class="border-b border-[#a68a56]/20 hover:bg-[#a68a56]/10 transition-colors text-xs text-[#f9f8f6]">`;
                cols.forEach((_, idx) => {
                    let val = m.fields && m.fields[idx] ? m.fields[idx] : '-';
                    val = val.replace(/</g, '&lt;').replace(/>/g, '&gt;');
                    html += `<td class="p-4 align-top whitespace-pre-wrap">${val}</td>`;
                });
                html += `</tr>`;
            });
            html += `</tbody></table>`;
            resultsDiv.innerHTML = html;
        } else {
            let html = `<table class="w-full text-left border-collapse table-fixed break-words"><thead class="sticky top-0 bg-[#0b1d2e] z-10 shadow-md"><tr class="border-b-2 border-[#a68a56]/50 text-[#a68a56] text-[10px] uppercase tracking-widest"><th class="p-4 font-normal w-40">Directory</th><th class="p-4 font-normal">Record Details</th></tr></thead><tbody>`;
            matched.forEach(m => {
                const details = m.fields ? m.fields.filter(f => f.trim() !== '').map(f => f.replace(/</g, '&lt;').replace(/>/g, '&gt;')).join(' • ') : 'N/A';
                html += `<tr class="border-b border-[#a68a56]/20 hover:bg-[#a68a56]/10 transition-colors text-xs text-[#f9f8f6]"><td class="p-4 text-[#a68a56] uppercase tracking-wider align-top">${m.category}</td><td class="p-4 align-top leading-relaxed whitespace-pre-wrap">${details}</td></tr>`;
            });
            html += `</tbody></table>`;
            resultsDiv.innerHTML = html;
        }
    };

    catSelect.addEventListener('change', (e) => {
        const cat = e.target.value;
        if (cat && CATEGORY_COLUMNS[cat]) {
            colFiltersDiv.classList.remove('hidden');
            colFiltersDiv.innerHTML = CATEGORY_COLUMNS[cat].map((col, idx) => `
                <input type="text" data-col-idx="${idx}" placeholder="${col.toUpperCase()}..." class="col-filter-input bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-2 text-[10px] focus:outline-none focus:border-[#a68a56] transition-colors uppercase tracking-widest font-sans rounded" />
            `).join('');
            document.querySelectorAll('.col-filter-input').forEach(inp => inp.addEventListener('input', renderResults));
        } else {
            colFiltersDiv.classList.add('hidden');
            colFiltersDiv.innerHTML = '';
        }
        renderResults();
    });

    globalInput.addEventListener('input', renderResults);
    renderResults();
}

function renderTable(category) {
    const cols = CATEGORY_COLUMNS[category] || [];
    let records = appState.data[category] || [];
    
    if (appState.localSearchTerm) {
        const term = appState.localSearchTerm.toLowerCase();
        records = records.filter(rec => rec.fields && rec.fields.some(f => f.toLowerCase().includes(term)));
    }
    
    let html = `
        <div class="flex flex-col h-full p-4 gap-4">
            <div class="flex justify-between items-center">
                <input type="text" id="local-search-input" value="${appState.localSearchTerm || ''}" placeholder="SEARCH THIS DIRECTORY..." class="w-1/3 bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] transition-colors uppercase tracking-widest font-sans rounded" />
                <button id="btn-add-record" class="bg-[#0b1d2e] border border-[#a68a56] text-[#a68a56] hover:bg-[#a68a56] hover:text-[#0b1d2e] px-4 py-2 text-xs uppercase tracking-widest transition-colors font-sans rounded shadow-md">+ Add New Record</button>
            </div>
            <div id="table-scroll-wrapper" class="w-full flex-1 overflow-auto border border-[#a68a56]/30 bg-black/20 rounded">
                <table class="w-full text-left border-collapse table-fixed break-words">
                    <thead class="sticky top-0 bg-[#0b1d2e] z-10 shadow-md">
                        <tr class="border-b-2 border-[#a68a56]/50 text-[#a68a56] text-[10px] uppercase tracking-widest">
    `;
    
    cols.forEach(col => { html += `<th class="p-4 font-normal">${col}</th>`; });
    html += `<th class="p-4 font-normal w-32 text-right">Actions</th></tr></thead><tbody>`;

    if (appState.isAddingNew) {
        html += `<tr class="border-b border-[#a68a56]/20 bg-[#a68a56]/10 text-xs">`;
        cols.forEach((_, idx) => {
            html += `<td class="p-2 align-top"><textarea rows="1" class="new-field-input w-full bg-black/60 border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] resize-y whitespace-pre-wrap" data-idx="${idx}"></textarea></td>`;
        });
        html += `
            <td class="p-4 align-top flex justify-end gap-3">
                <button class="btn-save-new text-emerald-400 hover:text-emerald-300 uppercase tracking-widest text-[10px] font-bold">Save</button>
                <button class="btn-cancel-new text-red-400 hover:text-red-300 uppercase tracking-widest text-[10px] font-bold">Cancel</button>
            </td>
        </tr>`;
    }

    if (records.length === 0 && !appState.isAddingNew) {
        html += `<tr><td colspan="${cols.length + 1}" class="p-8 text-center text-[#f9f8f6]/50 text-sm uppercase tracking-widest">No records available.</td></tr>`;
    } else {
        records.forEach(rec => {
            const isEditing = appState.editingId === rec.id;
            html += `<tr class="border-b border-[#a68a56]/20 hover:bg-[#a68a56]/5 transition-colors text-xs text-[#f9f8f6]">`;
            if (isEditing) {
                cols.forEach((_, idx) => {
                    const val = rec.fields && rec.fields[idx] ? rec.fields[idx].replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;') : '';
                    html += `<td class="p-2 align-top"><textarea rows="2" class="edit-field-input w-full bg-black/60 border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] resize-y whitespace-pre-wrap" data-idx="${idx}">${val}</textarea></td>`;
                });
                html += `
                    <td class="p-4 align-top flex justify-end gap-3">
                        <button class="btn-save-edit text-emerald-400 hover:text-emerald-300 uppercase tracking-widest text-[10px] font-bold" data-id="${rec.id}">Save</button>
                        <button class="btn-cancel-edit text-red-400 hover:text-red-300 uppercase tracking-widest text-[10px] font-bold">Cancel</button>
                    </td>`;
            } else {
                cols.forEach((_, idx) => {
                    let val = rec.fields && rec.fields[idx] ? rec.fields[idx] : '-';
                    val = val.replace(/</g, '&lt;').replace(/>/g, '&gt;');
                    html += `<td class="p-4 align-top leading-relaxed whitespace-pre-wrap">${val}</td>`;
                });
                html += `
                    <td class="p-4 align-top flex justify-end gap-3">
                        <button class="btn-edit-row text-[#a68a56] hover:text-[#d4c299] uppercase tracking-widest text-[10px] font-bold" data-id="${rec.id}">Edit</button>
                        <button class="btn-delete-row text-red-500 hover:text-red-400 uppercase tracking-widest text-[10px] font-bold" data-id="${rec.id}">Delete</button>
                    </td>`;
            }
            html += `</tr>`;
        });
    }

    html += `</tbody></table></div></div>`;
    return html;
}

function attachTableEventListeners(category) {
    const localSearchInput = document.getElementById('local-search-input');
    if (localSearchInput) {
        localSearchInput.addEventListener('input', (e) => {
            appState.localSearchTerm = e.target.value;
            refreshTable(category);
            const newSearchInput = document.getElementById('local-search-input');
            if (newSearchInput) {
                newSearchInput.focus();
                const len = appState.localSearchTerm.length;
                newSearchInput.setSelectionRange(len, len);
            }
        });
    }

    document.getElementById('btn-add-record')?.addEventListener('click', () => {
        appState.isAddingNew = true;
        appState.editingId = null;
        refreshTable(category);
        document.querySelector('.new-field-input')?.focus();
    });

    document.querySelector('.btn-cancel-new')?.addEventListener('click', () => {
        appState.isAddingNew = false;
        refreshTable(category);
    });

    document.querySelector('.btn-save-new')?.addEventListener('click', async () => {
        const inputs = document.querySelectorAll('.new-field-input');
        const fields = Array(CATEGORY_COLUMNS[category].length).fill('');
        inputs.forEach(inp => { fields[parseInt(inp.getAttribute('data-idx'))] = inp.value; });
        await saveRecord(null, category, fields);
    });

    document.querySelectorAll('.btn-edit-row').forEach(btn => {
        btn.addEventListener('click', (e) => {
            appState.editingId = e.target.getAttribute('data-id');
            appState.isAddingNew = false;
            refreshTable(category);
            document.querySelector('.edit-field-input')?.focus();
        });
    });

    document.querySelectorAll('.btn-delete-row').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const id = e.target.getAttribute('data-id');
            await deleteRecord(id, category);
        });
    });

    document.querySelector('.btn-cancel-edit')?.addEventListener('click', () => {
        appState.editingId = null;
        refreshTable(category);
    });

    document.querySelector('.btn-save-edit')?.addEventListener('click', async (e) => {
        const id = e.target.getAttribute('data-id');
        const inputs = document.querySelectorAll('.edit-field-input');
        const fields = Array(CATEGORY_COLUMNS[category].length).fill('');
        inputs.forEach(inp => { fields[parseInt(inp.getAttribute('data-idx'))] = inp.value; });
        await saveRecord(id, category, fields);
    });

    // Intercept Enter vs Shift+Enter
    document.querySelectorAll('.edit-field-input, .new-field-input').forEach(inp => {
        inp.addEventListener('keydown', (e) => {
            // Standard Enter saves the row. Shift+Enter creates a new line.
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                if (inp.classList.contains('edit-field-input')) {
                    inp.closest('tr').querySelector('.btn-save-edit').click();
                } else {
                    inp.closest('tr').querySelector('.btn-save-new').click();
                }
            }
        });
    });
}

async function syncNow() {
    const modal = document.getElementById('sync-modal');
    const loadingState = document.getElementById('sync-modal-loading');
    const successState = document.getElementById('sync-modal-success');
    const errorState = document.getElementById('sync-modal-error');
    const closeBtn = document.getElementById('close-sync-modal-btn');
    const indicator = document.getElementById('sync-modal-indicator');
    const title = document.getElementById('sync-modal-title');
    const trace = document.getElementById('sync-error-trace');
    const progressBar = document.getElementById('sync-progress-bar');

    modal.classList.remove('hidden');
    loadingState.classList.remove('hidden');
    successState.classList.add('hidden');
    errorState.classList.add('hidden');
    closeBtn.classList.add('hidden');
    indicator.className = "w-3 h-3 bg-[#a68a56] animate-ping rounded-full";
    title.textContent = "Synchronizing Data";

    progressBar.style.width = '10%';
    let progressInterval = setInterval(() => {
        let current = parseInt(progressBar.style.width);
        if(current < 90) progressBar.style.width = (current + 10) + '%';
    }, 300);

    try {
        const response = await fetch('/api/sync', { method: 'POST' });
        if (!response.ok) throw new Error(`HTTP status ${response.status}`);
        const status = await response.json();
        
        clearInterval(progressInterval);
        progressBar.style.width = '100%';

        setTimeout(async () => {
            loadingState.classList.add('hidden');

            if (status.error) {
                indicator.className = "w-3 h-3 bg-red-500 rounded-full";
                title.textContent = "Sync Failed";
                trace.textContent = status.details || status.error;
                errorState.classList.remove('hidden');
                errorState.classList.add('flex');
            } else {
                indicator.className = "w-3 h-3 bg-emerald-400 rounded-full";
                title.textContent = "Sync Complete";
                document.getElementById('sync-last-time').textContent = "Last synchronized: " + (status.last_sync || "Just now");
                successState.classList.remove('hidden');
                successState.classList.add('flex');

                await loadDataFromBackend();
                
                if (appState.activeTab === 'search') {
                    document.getElementById('tab-content').innerHTML = renderSearch();
                    executeSearch();
                } else {
                    document.getElementById('tab-content').innerHTML = renderTable(appState.activeTab);
                    attachTableEventListeners(appState.activeTab);
                }
            }
            closeBtn.classList.remove('hidden');
        }, 400);
        
    } catch (err) {
        clearInterval(progressInterval);
        loadingState.classList.add('hidden');
        indicator.className = "w-3 h-3 bg-red-500 rounded-full";
        title.textContent = "Fatal Exception";
        trace.textContent = err.stack || err.toString();
        errorState.classList.remove('hidden');
        errorState.classList.add('flex');
        closeBtn.classList.remove('hidden');
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    document.getElementById('close-sync-modal-btn')?.addEventListener('click', () => {
        document.getElementById('sync-modal').classList.add('hidden');
    });

    await loadDataFromBackend();

    const tabs = document.querySelectorAll('.tab-btn');
    const content = document.getElementById('tab-content');
    const exportBtn = document.getElementById('btn-export');

    function activateTab(target) {
        appState.activeTab = target;
        appState.editingId = null;
        appState.isAddingNew = false;
        appState.localSearchTerm = '';
        
        if (exportBtn) {
            if (target === 'search') exportBtn.classList.add('hidden');
            else exportBtn.classList.remove('hidden');
        }

        tabs.forEach(t => {
            if (t.getAttribute('data-target') === target) {
                t.classList.add('active', 'text-[#a68a56]', 'border-[#a68a56]');
                t.classList.remove('text-[#f9f8f6]', 'border-transparent');
            } else {
                t.classList.remove('active', 'text-[#a68a56]', 'border-[#a68a56]');
                t.classList.add('text-[#f9f8f6]', 'border-transparent');
            }
        });

        if (target === 'search') {
            content.innerHTML = renderSearch();
            executeSearch();
        } else {
            content.innerHTML = renderTable(target);
            attachTableEventListeners(target);
        }
    }

    tabs.forEach(tab => {
        tab.addEventListener('click', (e) => {
            const target = e.currentTarget.getAttribute('data-target');
            activateTab(target);
        });
    });

    activateTab('search');
});