const schemas = {
    search: [],
    banks: ['Bank Name', 'Branch', 'Physical Address', 'Contact Name', 'Title/Rank', 'Email', 'Phone', 'Fax', 'Notes'],
    masters: ['Masters Office', 'Physical Address', 'Department/Division', 'Contact Name', 'Rank/Title', 'File Number Allocation', 'Email', 'Phone', 'Notes'],
    sheriffs: ['Sheriff Office', 'Sheriff Name', 'Deputy Name', 'Physical Address', 'Jurisdiction', 'Email', 'Phone', 'Fax', 'Notes'],
    magistrates: ['Court Name', 'Physical Address', 'Contact Name', 'Title/Rank', 'Email', 'Phone', 'Fax', 'Court Citation', 'Notes'],
    highcourts: ['Court Name', 'Physical Address', 'Contact Name', 'Title/Rank', 'Email', 'Phone', 'Fax', 'Court Citation', 'Notes'],
    lawfirms: ['Firm Name', 'Physical Address', 'Contact Name', 'Email', 'Phone', 'Notes']
};

let db = {
    banks: [],
    masters: [],
    sheriffs: [],
    magistrates: [],
    highcourts: [],
    lawfirms: []
};

let activeDownloadUrl = "";

async function loadDataFromBackend() {
    for (const cat of Object.keys(db)) {
        if (window.go && window.go.core && window.go.core.App) {
            try {
                const records = await window.go.core.App.GetContacts(cat);
                db[cat] = records || [];
            } catch (e) {
                db[cat] = [];
            }
        }
    }
}

async function syncNow() {
    if (window.go && window.go.core && window.go.core.App) {
        await window.go.core.App.TriggerSync();
        await loadDataFromBackend();
        const activeTab = document.querySelector('.tab-btn.active')?.getAttribute('data-target') || 'search';
        if (activeTab === 'search') {
            document.getElementById('tab-content').innerHTML = renderSearch();
            executeSearch();
        } else {
            document.getElementById('tab-content').innerHTML = renderTable(activeTab);
        }
    }
}

async function checkSoftwareUpdate() {
    if (window.go && window.go.core && window.go.core.App) {
        try {
            const res = await window.go.core.App.CheckUpdate();
            if (res && res.has_update) {
                activeDownloadUrl = res.download_url;
                document.getElementById('update-target-version').textContent = res.latest_ver;
                document.getElementById('update-release-notes').textContent = res.release_notes || "No release notes available.";
                document.getElementById('update-wizard-modal').classList.remove('hidden');
            } else {
                alert("You are running the latest version.");
            }
        } catch (e) {
            alert("Error checking for updates: " + e);
        }
    }
}

function renderTable(type) {
    const cols = schemas[type];
    const data = db[type];

    let tableHtml = `<div class="overflow-x-auto"><table class="w-full text-left border-collapse text-sm">
        <thead><tr class="bg-[#a68a56]/20 border-b-2 border-[#a68a56]">`;
    cols.forEach(col => {
        tableHtml += `<th class="p-3 text-[#a68a56] uppercase tracking-wider text-xs whitespace-nowrap">${col}</th>`;
    });
    tableHtml += `<th class="p-3 text-[#a68a56] uppercase tracking-wider text-xs text-right">Actions</th></tr></thead><tbody>`;

    data.forEach((row, rowIdx) => {
        tableHtml += `<tr class="border-b border-[#a68a56]/20 hover:bg-[#a68a56]/10 transition-colors">`;
        row.fields.forEach(cell => {
            tableHtml += `<td class="p-3 text-[#f9f8f6]/90 text-xs">${cell}</td>`;
        });
        tableHtml += `<td class="p-3 text-right whitespace-nowrap">
            <button onclick="editRecord('${type}', ${rowIdx})" class="text-blue-400 hover:text-blue-300 text-xs uppercase tracking-widest mr-4">Edit</button>
            <button onclick="deleteRecord('${type}', ${rowIdx})" class="text-red-400 hover:text-red-300 text-xs uppercase tracking-widest">Delete</button>
        </td></tr>`;
    });

    tableHtml += `</tbody></table></div>`;

    let formHtml = `<div id="add-form-container-${type}" class="hidden bg-[#0b1d2e]/80 border border-[#a68a56]/30 p-4 mb-4">
    <h3 class="text-[#a68a56] uppercase tracking-widest text-sm mb-4">Add New Record</h3>
    <form id="add-form-${type}" class="grid grid-cols-3 gap-4">`;

    cols.forEach((col, idx) => {
        formHtml += `<input type="text" placeholder="${col}" class="bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] w-full" id="input-${type}-${idx}" />`;
    });
    formHtml += `<div class="col-span-3 flex justify-end"><button type="button" onclick="addRecord('${type}')" class="bg-[#0b1d2e] border border-[#a68a56] text-[#a68a56] hover:bg-[#a68a56] hover:text-[#0b1d2e] px-6 py-2 text-xs uppercase tracking-widest transition-colors">Save Record</button></div></form></div>`;

    let searchBarHtml = `<div class="bg-[#0b1d2e]/80 border border-[#a68a56]/30 p-2 mb-4 flex gap-4 items-end">
        <input type="text" id="tab-search-${type}" onkeyup="renderTable('${type}')" placeholder="Search in ${type}..." class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-2 text-sm focus:outline-none focus:border-[#a68a56]"/>
        <button onclick="toggleAddForm('${type}')" class="bg-[#0b1d2e] border border-[#a68a56] text-[#a68a56] hover:bg-[#a68a56] hover:text-[#0b1d2e] px-4 py-2 text-xs uppercase tracking-widest transition-colors">Add New</button>
    </div>`;

    return searchBarHtml + formHtml + tableHtml;
}

function toggleAddForm(type) {
    const form = document.getElementById(`add-form-container-${type}`);
    const button = event.target;
    form.classList.toggle('hidden');
    if (form.classList.contains('hidden')) {
        button.textContent = 'Add New';
    } else {
        button.textContent = 'Close Add New';
    }
}

function renderSearch() {
    let html = `<div class="bg-[#0b1d2e]/80 border border-[#a68a56]/30 p-6 mb-6 flex gap-4 items-end">
        <div class="flex-1">
            <label class="text-[#a68a56] text-xs uppercase tracking-widest mb-2 block">Search Query</label>
            <input type="text" id="global-search" onkeyup="executeSearch()" placeholder="Type to filter..." class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-sm focus:outline-none focus:border-[#a68a56]"/>
        </div>
        <div class="w-64">
            <label class="text-[#a68a56] text-xs uppercase tracking-widest mb-2 block">Category</label>
            <select id="search-category" onchange="executeSearch()" class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-sm focus:outline-none focus:border-[#a68a56] appearance-none">
                <option value="all">All Directories</option>
                <option value="banks">Banks</option>
                <option value="masters">Master's Office</option>
                <option value="sheriffs">Sheriffs</option>
                <option value="magistrates">Magistrate Courts</option>
                <option value="highcourts">High Courts</option>
                <option value="lawfirms">Law Firms</option>
            </select>
        </div>
    </div>
    <div id="search-results"></div>`;
    return html;
}

window.executeSearch = function() {
    const query = document.getElementById('global-search').value.toLowerCase();
    const cat = document.getElementById('search-category').value;
    const resultsContainer = document.getElementById('search-results');

    let html = '';
    const typesToSearch = cat === 'all' ? ['banks', 'masters', 'sheriffs', 'magistrates', 'highcourts', 'lawfirms'] : [cat];

    typesToSearch.forEach(type => {
        const cols = schemas[type];
        const data = db[type] || [];
        const filtered = data.filter(row => row.fields && row.fields.some(cell => cell.toLowerCase().includes(query)));

        if (filtered.length > 0) {
            html += `<h4 class="text-[#a68a56] uppercase tracking-widest mb-2 mt-6 border-b border-[#a68a56]/30 pb-2">${type.toUpperCase()}</h4>`;
            html += `<div class="overflow-x-auto mb-6"><table class="w-full text-left border-collapse text-sm">
                <thead><tr class="bg-[#a68a56]/20 border-b border-[#a68a56]">`;
            cols.forEach(col => html += `<th class="p-2 text-[#a68a56] uppercase tracking-wider text-xs whitespace-nowrap">${col}</th>`);
            html += `<th class="p-2 text-[#a68a56] uppercase tracking-wider text-xs text-right">Actions</th></tr></thead><tbody>`;

            filtered.forEach(row => {
                const originalIndex = db[type].findIndex(item => item.id === row.id);
                html += `<tr class="border-b border-[#a68a56]/20 hover:bg-[#a68a56]/10">`;
                row.fields.forEach(cell => html += `<td class="p-2 text-[#f9f8f6]/90 text-xs">${cell}</td>`);
                html += `<td class="p-2 text-right whitespace-nowrap">
                    <button onclick="editRecord('${type}', ${originalIndex})" class="text-blue-400 hover:text-blue-300 text-xs uppercase tracking-widest mr-4">Edit</button>
                    <button onclick="deleteRecord('${type}', ${originalIndex})" class="text-red-400 hover:text-red-300 text-xs uppercase tracking-widest">Delete</button>
                </td></tr>`;
            });
            html += `</tbody></table></div>`;
        }
    });

    if (html === '') {
        html = `<p class="text-[#f9f8f6]/50 italic text-sm">No records found matching your criteria.</p>`;
    }
    resultsContainer.innerHTML = html;
}

window.addRecord = async function(type) {
    const cols = schemas[type];
    const newRecord = {
        id: `${type}-${Date.now()}`,
        category: type,
        fields: []
    };
    for (let i = 0; i < cols.length; i++) {
        const val = document.getElementById(`input-${type}-${i}`).value;
        newRecord.fields.push(val);
    }
    if (window.go && window.go.core && window.go.core.App) {
        await window.go.core.App.SaveContact(newRecord);
        await loadDataFromBackend();
    }
    document.getElementById('tab-content').innerHTML = renderTable(type);
}

window.deleteRecord = async function(type, idx) {
    if (confirm("Are you sure you want to delete this record?")) {
        const record = db[type][idx];
        if (window.go && window.go.core && window.go.core.App) {
            await window.go.core.App.DeleteContact(type, record.id);
            await loadDataFromBackend();
        }
        if (document.getElementById('search-results')) {
            executeSearch();
        } else {
            document.getElementById('tab-content').innerHTML = renderTable(type);
        }
    }
}

window.editRecord = function(type, idx) {
    const row = db[type][idx];
    const rowElement = event.target.closest('tr');
    let editHtml = '';
    row.fields.forEach((cell, cellIndex) => {
        editHtml += `<td class="p-2"><input type="text" value="${cell}" class="bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] w-full" id="edit-${type}-${idx}-${cellIndex}" /></td>`;
    });
    editHtml += `<td class="p-2 text-right whitespace-nowrap">
        <button onclick="saveRecord('${type}', ${idx})" class="text-green-400 hover:text-green-300 text-xs uppercase tracking-widest mr-4">Save</button>
        <button onclick="cancelEdit('${type}', ${idx})" class="text-gray-400 hover:text-gray-300 text-xs uppercase tracking-widest">Cancel</button>
    </td>`;
    rowElement.innerHTML = editHtml;
}

window.saveRecord = async function(type, idx) {
    const newFields = [];
    const cols = schemas[type];
    for (let i = 0; i < cols.length; i++) {
        const val = document.getElementById(`edit-${type}-${idx}-${i}`).value;
        newFields.push(val);
    }
    const record = db[type][idx];
    record.fields = newFields;

    if (window.go && window.go.core && window.go.core.App) {
        await window.go.core.App.SaveContact(record);
        await loadDataFromBackend();
    }

    if (document.getElementById('search-results')) {
        executeSearch();
    } else {
        document.getElementById('tab-content').innerHTML = renderTable(type);
    }
}

window.cancelEdit = function(type, idx) {
    if (document.getElementById('search-results')) {
        executeSearch();
    } else {
        document.getElementById('tab-content').innerHTML = renderTable(type);
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    await loadDataFromBackend();

    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContent = document.getElementById('tab-content');

    tabBtns.forEach(btn => {
        btn.addEventListener('click', (e) => {
            tabBtns.forEach(b => {
                b.classList.remove('active', 'text-[#a68a56]', 'border-[#a68a56]');
                b.classList.add('text-[#f9f8f6]', 'border-transparent');
            });
            e.target.classList.remove('text-[#f9f8f6]', 'border-transparent');
            e.target.classList.add('active', 'text-[#a68a56]', 'border-[#a68a56]');

            const target = e.target.getAttribute('data-target');
            if (target === 'search') {
                tabContent.innerHTML = renderSearch();
                executeSearch();
            } else {
                tabContent.innerHTML = renderTable(target);
            }
        });
    });

    if (tabContent) {
        tabContent.innerHTML = renderSearch();
        executeSearch();
    }

    document.getElementById('close-updater-btn')?.addEventListener('click', () => {
        document.getElementById('update-wizard-modal').classList.add('hidden');
    });

    document.getElementById('start-update-btn')?.addEventListener('click', async () => {
        document.getElementById('updater-step-1').classList.add('hidden');
        document.getElementById('updater-step-2').classList.remove('hidden');
        document.getElementById('updater-step-2').classList.add('flex');

        if (window.go && window.go.core && window.go.core.App) {
            try {
                await window.go.core.App.PerformUpdate(activeDownloadUrl);
            } catch (err) {
                alert("Update failed: " + err);
                document.getElementById('update-wizard-modal').classList.add('hidden');
            }
        }
    });

    setTimeout(() => {
        checkSoftwareUpdate();
    }, 2000);
});