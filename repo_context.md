<system_instructions>
# SYSTEM INSTRUCTION / MANIFEST

**ROLE**
You are an elite Senior Developer specializing in Go, Wails, HTML, Templ, JavaScript, Tailwind CSS, and Supabase database architecture. 

**OBJECTIVE**
Analyze the user's request deeply before responding. Provide the absolute best, most optimal programmatic and architectural solutions, specifically optimized for highly secure, performant Wails desktop applications backed by Supabase, tailored for the strict confidentiality, data integrity, and reliability requirements of a law firm.

**INTERACTION MODES**
Depending on the user's prompt or the level of detail provided, follow the respective mode:
* **[AUTO-MODE]:** If the user provides a detailed specification, assume the best architectural choices and generate the full code immediately per the rules.
* **[DISCOVERY-MODE]:** If the user uses this tag, or if the initial request lacks necessary architectural or UI/UX details, STOP. Do not generate code yet. Ask a concise, bulleted list of high-level technical questions to gather exact requirements. Wait for the user's reply before coding.

**RULES & CONSTRAINTS**
1. **Full Code Delivery:** Once requirements are fully clarified, always provide complete, fully functioning code blocks. Do not use placeholders, summaries, or truncate the code in any way.
2. **Zero Comments:** The outputted code must contain absolutely NO comments. Code should be clean and self-documenting.
3. **Production-Ready:** All code must be secure, performant, and ready for production deployment without requiring hotfixes.
4. **Idiomatic Go & Context:** Enforce strict Go idioms. Handle all errors explicitly (no ignores, no panics), use context propagation (`context.Context`) for all Wails bindings and Supabase API calls, and ensure absolute goroutine safety.
5. **Supabase Database Practices:** Use robust PostgreSQL/Supabase Go clients. Ensure proper use of Row Level Security (RLS) concepts, foreign keys, structural tags, and transactional operations to guarantee the absolute data integrity required for sensitive legal data and case management.
6. **Law Firm Data Handling & Sync:** Optimize for security, auditability, and memory efficiency. Never use `SELECT *`; specifically select required fields to minimize data exposure. Implement strict pagination or streaming for large legal datasets (e.g., case files, discovery documents) to manage memory efficiently in the Wails app.
7. **Wails & Desktop Architecture:** Write highly modular, decoupled code using clean architecture or repository patterns. Group logic clearly between Wails frontend IPC bindings and backend services so that future legal-tech features (e.g., offline sync, document generation, real-time collaboration) can be added seamlessly without breaking existing code.
8. **Security & Confidentiality:** Never hardcode credentials. Design for secure local storage of Supabase authentication tokens, enforce strict user sessions, and assume a zero-trust model appropriate for handling sensitive, attorney-client privileged data.
9. **Tailwind CSS Exclusivity:** Use ONLY Tailwind utility classes for styling. Never write custom CSS, inline `<style>` tags, or use `style="..."` attributes.
10. **No Dynamic Tailwind Classes:** Always write full, complete Tailwind class names (e.g., `text-red-500` instead of `text-${color}-500`). Do not construct Tailwind classes programmatically.
11. **Strict Templ Typing:** Define clear Go structs for all Templ component data instead of using loose maps or `interface{}`.

**PROCESS & CLARIFICATION PHASE**
1. **Analyze:** Think critically about the problem step-by-step. Evaluate the proposed architecture, data flow, Wails IPC lifecycle, and Supabase real-time/sync patterns.
2. **Clarify (Crucial Step):** Before writing ANY code, assess if the user's request contains enough detail to produce a production-ready, highly optimized solution. If the requirements are ambiguous, missing Wails/Supabase architectural context, or lack specific legal-tech workflows (e.g., user roles, audit trails), **do not write code yet**.
3. **Ask:** Output a structured list of clarifying questions grouped by domain (e.g., Wails Architecture, Supabase/Database Logic, Frontend/Templ, Legal Compliance/Security).
4. **Execute:** Only proceed to generate the full code once you have a complete understanding of the system requirements based on the user's answers.
</system_instructions>

<codebase>
# Full Repository Map: LuminaLex

## FILE: extract.py
```py
import os
from pathlib import Path

def generate_full_repo_markup(repo_path='.', output_name='repo_context.md', manifest_name='Gemini.md'):
    root_dir = Path(repo_path).resolve()
    output_file = Path(output_name).resolve()
    manifest_file = root_dir / manifest_name
    
    # Standard directories to skip
    forbidden_directories = {
        '.git', 'node_modules', '__pycache__', 'venv', '.venv', 
        '.idea', '.vscode', 'dist', 'build'
    }
    
    # Extensions we want to capture
    valid_extensions = {
        '.py', '.js', '.ts', '.jsx', '.tsx', '.html', '.css', 
        '.json', '.yaml', '.yml', '.txt', '.c', '.cpp', 
        '.h', '.java', '.go', '.rs', '.templ', '.md'
    }

    print(f"🚀 Starting deep scan of: {root_dir}")
    count = 0

    with open(output_file, 'w', encoding='utf-8') as f:
        # --- 1. INJECT SYSTEM INSTRUCTIONS (GEMINI.MF) ---
        if manifest_file.exists():
            print(f" 🎯 Found manifest: {manifest_name}. Injecting at the top.")
            f.write("<system_instructions>\n")
            try:
                manifest_content = manifest_file.read_text(encoding='utf-8')
                f.write(manifest_content)
            except Exception as e:
                f.write(f"// Error reading manifest: {e}")
            f.write("\n</system_instructions>\n\n")
        else:
            print(f" ⚠️ Warning: {manifest_name} not found in {root_dir}. Proceeding without system instructions.")

        # --- 2. START CODEBASE SECTION ---
        f.write("<codebase>\n")
        f.write(f"# Full Repository Map: {root_dir.name}\n\n")

        # rglob('*') recursively finds all files and folders
        for path in root_dir.rglob('*'):
            
            # Skip if it's a directory
            if not path.is_file():
                continue

            # Skip if the file is the output file itself
            if path == output_file:
                continue
                
            # Skip the manifest file so we don't duplicate it inside the codebase block
            if path == manifest_file:
                continue

            # Skip if any part of the folder path contains a forbidden directory
            if any(segment in path.parts for segment in forbidden_directories):
                continue

            # Specifically ignore files ending with _templ.go
            if path.name.endswith('_templ.go'):
                print(f"  skip -> {path.name} (Templ generated)")
                continue

            # Only proceed if it's a code/text extension we recognize
            if path.suffix.lower() in valid_extensions:
                relative_path = path.relative_to(root_dir)
                
                print(f"  read -> {relative_path}")
                count += 1
                
                f.write(f"## FILE: {relative_path}\n")
                # lstrip('.') ensures '.py' becomes 'py' for the markdown block
                f.write(f"```{path.suffix.lstrip('.')}\n")
                
                try:
                    # 'errors=replace' handles decorated text or accidental binary data
                    content = path.read_text(encoding='utf-8', errors='replace')
                    f.write(content)
                except Exception as e:
                    f.write(f"// Error reading this file: {e}")
                
                f.write("\n```\n\n---\n\n")

        # --- 3. END CODEBASE SECTION ---
        f.write("</codebase>\n")

    print(f"\n✅ Success! Processed {count} files.")
    print(f"Saved to: {output_file}")

if __name__ == "__main__":
    # You can pass a specific path here if running from a different directory
    generate_full_repo_markup()
```

---

## FILE: main.go
```go
package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"luminalex/core"
)

//go:embed assets
var assets embed.FS

func main() {
	_ = godotenv.Load(".env")

	if err := run(); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	app := core.NewApp()
	handler := core.NewHandler(app)
	assetHandler := http.FileServer(http.FS(assets))

	err := wails.Run(&options.App{
		Title:            "LuminaLex",
		Width:            1024,
		Height:           768,
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
			Middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/assets/") {
						assetHandler.ServeHTTP(w, r)
						return
					}
					if strings.HasPrefix(r.URL.Path, "/api/") {
						handler.ServeHTTP(w, r)
						return
					}
					// ADDED "/services" TO THIS LINE:
					if r.URL.Path == "/" || r.URL.Path == "/login" || r.URL.Path == "/home" || r.URL.Path == "/contacts" || r.URL.Path == "/matters" || r.URL.Path == "/clients" || r.URL.Path == "/services" {
						handler.ServeHTTP(w, r)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		OnStartup: app.Startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		return fmt.Errorf("starting wails app: %w", err)
	}

	return nil
}

```

---

## FILE: wails.json
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "LuminaLex",
  "outputfilename": "LuminaLex",
  "frontend:build": "templ generate",
  "frontend:dir": ".",
  "frontend:dev:watcher": "templ generate --watch",
  "run:cmd": "go run .",
  "author": {
    "name": "Admin",
    "email": "admin@luminalex.local"
  }
}
```

---

## FILE: assets\contacts.js
```js
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
                    const input = td.querySelector('input');
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
    }
}

async function deleteRecord(id, category) {
    if (!confirm("Are you sure you want to delete this record? This action cannot be undone.")) return;
    
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
                    const val = m.fields && m.fields[idx] ? m.fields[idx] : '-';
                    html += `<td class="p-4 align-top">${val}</td>`;
                });
                html += `</tr>`;
            });
            html += `</tbody></table>`;
            resultsDiv.innerHTML = html;
        } else {
            let html = `<table class="w-full text-left border-collapse table-fixed break-words"><thead class="sticky top-0 bg-[#0b1d2e] z-10 shadow-md"><tr class="border-b-2 border-[#a68a56]/50 text-[#a68a56] text-[10px] uppercase tracking-widest"><th class="p-4 font-normal w-40">Directory</th><th class="p-4 font-normal">Record Details</th></tr></thead><tbody>`;
            matched.forEach(m => {
                const details = m.fields ? m.fields.filter(f => f.trim() !== '').join(' • ') : 'N/A';
                html += `<tr class="border-b border-[#a68a56]/20 hover:bg-[#a68a56]/10 transition-colors text-xs text-[#f9f8f6]"><td class="p-4 text-[#a68a56] uppercase tracking-wider align-top">${m.category}</td><td class="p-4 align-top leading-relaxed">${details}</td></tr>`;
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
            <div class="w-full flex-1 overflow-auto border border-[#a68a56]/30 bg-black/20 rounded">
                <table class="w-full text-left border-collapse table-fixed break-words">
                    <thead class="sticky top-0 bg-[#0b1d2e] z-10 shadow-md">
                        <tr class="border-b-2 border-[#a68a56]/50 text-[#a68a56] text-[10px] uppercase tracking-widest">
    `;
    
    cols.forEach(col => { html += `<th class="p-4 font-normal">${col}</th>`; });
    html += `<th class="p-4 font-normal w-32 text-right">Actions</th></tr></thead><tbody>`;

    if (appState.isAddingNew) {
        html += `<tr class="border-b border-[#a68a56]/20 bg-[#a68a56]/10 text-xs">`;
        cols.forEach((_, idx) => {
            html += `<td class="p-2 align-top"><input type="text" class="new-field-input w-full bg-black/60 border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56]" data-idx="${idx}"/></td>`;
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
                    const val = rec.fields && rec.fields[idx] ? rec.fields[idx].replace(/"/g, '&quot;') : '';
                    html += `<td class="p-2 align-top"><input type="text" class="edit-field-input w-full bg-black/60 border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56]" data-idx="${idx}" value="${val}"/></td>`;
                });
                html += `
                    <td class="p-4 align-top flex justify-end gap-3">
                        <button class="btn-save-edit text-emerald-400 hover:text-emerald-300 uppercase tracking-widest text-[10px] font-bold" data-id="${rec.id}">Save</button>
                        <button class="btn-cancel-edit text-red-400 hover:text-red-300 uppercase tracking-widest text-[10px] font-bold">Cancel</button>
                    </td>`;
            } else {
                cols.forEach((_, idx) => {
                    const val = rec.fields && rec.fields[idx] ? rec.fields[idx] : '-';
                    html += `<td class="p-4 align-top leading-relaxed">${val}</td>`;
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
            document.getElementById('tab-content').innerHTML = renderTable(category);
            attachTableEventListeners(category);
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
        document.getElementById('tab-content').innerHTML = renderTable(category);
        attachTableEventListeners(category);
    });

    document.querySelector('.btn-cancel-new')?.addEventListener('click', () => {
        appState.isAddingNew = false;
        document.getElementById('tab-content').innerHTML = renderTable(category);
        attachTableEventListeners(category);
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
            document.getElementById('tab-content').innerHTML = renderTable(category);
            attachTableEventListeners(category);
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
        document.getElementById('tab-content').innerHTML = renderTable(category);
        attachTableEventListeners(category);
    });

    document.querySelector('.btn-save-edit')?.addEventListener('click', async (e) => {
        const id = e.target.getAttribute('data-id');
        const inputs = document.querySelectorAll('.edit-field-input');
        const fields = Array(CATEGORY_COLUMNS[category].length).fill('');
        inputs.forEach(inp => { fields[parseInt(inp.getAttribute('data-idx'))] = inp.value; });
        await saveRecord(id, category, fields);
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
```

---

## FILE: assets\index.html
```html

```

---

## FILE: core\app.go
```go
package core

import (
	"context"
	"fmt"
	"os"
	"time"
)

type App struct {
	ctx        context.Context
	store      *LocalStore
	supabase   *SupabaseClient
	syncEngine *SyncEngine
	updater    *AutoUpdater
}

func NewApp() *App {
	store, err := NewLocalStore()
	if err != nil {
		fmt.Printf("Error initializing local store: %v\n", err)
		os.Exit(1)
	}

	// 1. Hardcode your Supabase credentials here so users don't need a .env file
	supabaseURL := "https://frakboneidergnmzznkh.supabase.co"                                                                                                                                                                         // <-- Paste your Supabase URL
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImZyYWtib25laWRlcmdubXp6bmtoIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODQ5MTk3MDAsImV4cCI6MjEwMDQ5NTcwMH0.8-tRutPQW5UbMpGwK-k8Uto4zGgw12mCd8oiGHULTYM" // <-- Paste your Supabase Anon Key
	sbClient := NewSupabaseClient(supabaseURL, supabaseKey)

	syncEngine := NewSyncEngine(store, sbClient)

	// 2. Hardcode your GitHub Repository details
	owner := "reederdevelopments"
	repo := "luminalex"

	// 3. This is your app's internal version.
	// Before you build v1.0.2 in the future, you must manually change this to "v1.0.1".
	version := "v1.0.2"

	updater := NewAutoUpdater(owner, repo, version)

	return &App{
		store:      store,
		supabase:   sbClient,
		syncEngine: syncEngine,
		updater:    updater,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	if err := a.syncEngine.PerformSync(a.ctx); err != nil {
		fmt.Printf("[Startup Sync Error] %v\n", err)
	}

	go a.startBackgroundTasks()
}

func (a *App) startBackgroundTasks() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.runRoutineTasks()
		}
	}
}

func (a *App) runRoutineTasks() {
	if err := a.syncEngine.PerformSync(a.ctx); err != nil {
		fmt.Printf("[Sync Error] %v\n", err)
	}
}

func (a *App) GetContacts(category string) ([]ContactRecord, error) {
	return a.store.GetCategoryRecords(a.ctx, category)
}

func (a *App) SaveContact(record ContactRecord) error {
	record.UpdatedAt = time.Now().UTC()
	record.Synced = false
	if err := a.store.SaveRecord(a.ctx, record); err != nil {
		return err
	}
	go func() {
		_ = a.syncEngine.PerformSync(context.Background())
	}()
	return nil
}

func (a *App) DeleteContact(category string, id string) error {
	records, err := a.store.GetCategoryRecords(a.ctx, category)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if rec.ID == id {
			rec.Deleted = true
			rec.UpdatedAt = time.Now().UTC()
			rec.Synced = false
			if err := a.store.SaveRecord(a.ctx, rec); err != nil {
				return err
			}
			break
		}
	}

	go func() {
		_ = a.syncEngine.PerformSync(context.Background())
	}()
	return nil
}

func (a *App) TriggerSync() SyncStatus {
	err := a.syncEngine.PerformSync(a.ctx)
	status := a.syncEngine.GetStatus()
	if err != nil {
		status.Error = "Sync Failed"
		status.Details = fmt.Sprintf("%+v", err)
	}
	return status
}

func (a *App) CheckUpdate() (*UpdateCheckResult, error) {
	return a.updater.CheckForUpdates(a.ctx)
}

func (a *App) PerformUpdate(downloadURL string) error {
	return a.updater.ApplyUpdate(a.ctx, downloadURL)
}

func (a *App) RestartApp() error {
	return a.updater.RestartApp()
}

```

---

## FILE: core\auth.go
```go
package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func (a *App) AuthenticateUser(username, password string) error {
	endpoint := fmt.Sprintf("%s/rest/v1/users?username=eq.%s&select=*", a.supabase.baseURL, url.QueryEscape(username))
	req, err := http.NewRequestWithContext(a.ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create auth request: %w", err)
	}
	a.supabase.setHeaders(req)

	resp, err := a.supabase.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication service unavailable")
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return fmt.Errorf("invalid authentication response")
	}

	if len(users) == 0 {
		return fmt.Errorf("invalid credentials")
	}

	user := users[0]
	if !user.Enabled {
		return fmt.Errorf("account disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return fmt.Errorf("invalid credentials")
	}

	_ = a.SaveLastUsername(username)
	return nil
}

func (a *App) SaveLastUsername(username string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "LuminaLex", "config.json")

	config := AppConfig{LastUsername: username}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func (a *App) GetLastUsername() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	configPath := filepath.Join(configDir, "LuminaLex", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}
	return config.LastUsername
}

```

---

## FILE: core\clients.go
```go
package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *LocalStore) GetClients(ctx context.Context) ([]Client, error) {
	query := `SELECT id, first_name, middle_name, last_name, id_number, jurisdiction_type, jurisdiction_other, registration_number, occupation, marital_status, employer_name, employer_number, employer_address_l1, employer_address_l2, employer_suburb, employer_city, employer_postal_code, employer_country, employer_postal_same, employer_postal_l1, employer_postal_l2, employer_postal_suburb, employer_postal_city, employer_postal_code2, employer_postal_country, practice_number, updated_at, deleted, synced FROM clients WHERE deleted = 0 ORDER BY last_name, first_name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query clients: %w", err)
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var c Client
		var empPostalSame, del, syn int
		if err := rows.Scan(&c.ID, &c.FirstName, &c.MiddleName, &c.LastName, &c.IDNumber, &c.JurisdictionType, &c.JurisdictionOther, &c.RegistrationNumber, &c.Occupation, &c.MaritalStatus, &c.EmployerName, &c.EmployerNumber, &c.EmployerAddressL1, &c.EmployerAddressL2, &c.EmployerSuburb, &c.EmployerCity, &c.EmployerPostalCode, &c.EmployerCountry, &empPostalSame, &c.EmployerPostalL1, &c.EmployerPostalL2, &c.EmployerPostalSuburb, &c.EmployerPostalCity, &c.EmployerPostalCode2, &c.EmployerPostalCountry, &c.PracticeNumber, &c.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		c.EmployerPostalSame = empPostalSame == 1
		c.Deleted = del == 1
		c.Synced = syn == 1

		c.Addresses, _ = s.getClientAddresses(ctx, c.ID)
		c.ContactDetails, _ = s.getClientContactDetails(ctx, c.ID)
		c.Banks, _ = s.getClientBanks(ctx, c.ID)

		clients = append(clients, c)
	}

	return clients, rows.Err()
}

func (s *LocalStore) SaveClient(ctx context.Context, client Client) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if client.ID == "" {
		client.ID = uuid.New().String()
	}
	client.UpdatedAt = time.Now().UTC()

	del := 0
	if client.Deleted {
		del = 1
	}
	empPostalSame := 0
	if client.EmployerPostalSame {
		empPostalSame = 1
	}

	clientQuery := `
	INSERT INTO clients (id, first_name, middle_name, last_name, id_number, jurisdiction_type, jurisdiction_other, registration_number, occupation, marital_status, employer_name, employer_number, employer_address_l1, employer_address_l2, employer_suburb, employer_city, employer_postal_code, employer_country, employer_postal_same, employer_postal_l1, employer_postal_l2, employer_postal_suburb, employer_postal_city, employer_postal_code2, employer_postal_country, practice_number, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		first_name = excluded.first_name,
		middle_name = excluded.middle_name,
		last_name = excluded.last_name,
		id_number = excluded.id_number,
		jurisdiction_type = excluded.jurisdiction_type,
		jurisdiction_other = excluded.jurisdiction_other,
		registration_number = excluded.registration_number,
		occupation = excluded.occupation,
		marital_status = excluded.marital_status,
		employer_name = excluded.employer_name,
		employer_number = excluded.employer_number,
		employer_address_l1 = excluded.employer_address_l1,
		employer_address_l2 = excluded.employer_address_l2,
		employer_suburb = excluded.employer_suburb,
		employer_city = excluded.employer_city,
		employer_postal_code = excluded.employer_postal_code,
		employer_country = excluded.employer_country,
		employer_postal_same = excluded.employer_postal_same,
		employer_postal_l1 = excluded.employer_postal_l1,
		employer_postal_l2 = excluded.employer_postal_l2,
		employer_postal_suburb = excluded.employer_postal_suburb,
		employer_postal_city = excluded.employer_postal_city,
		employer_postal_code2 = excluded.employer_postal_code2,
		employer_postal_country = excluded.employer_postal_country,
		practice_number = excluded.practice_number,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err = tx.ExecContext(ctx, clientQuery, client.ID, client.FirstName, client.MiddleName, client.LastName, client.IDNumber, client.JurisdictionType, client.JurisdictionOther, client.RegistrationNumber, client.Occupation, client.MaritalStatus, client.EmployerName, client.EmployerNumber, client.EmployerAddressL1, client.EmployerAddressL2, client.EmployerSuburb, client.EmployerCity, client.EmployerPostalCode, client.EmployerCountry, empPostalSame, client.EmployerPostalL1, client.EmployerPostalL2, client.EmployerPostalSuburb, client.EmployerPostalCity, client.EmployerPostalCode2, client.EmployerPostalCountry, client.PracticeNumber, client.UpdatedAt, del)
	if err != nil {
		return fmt.Errorf("upsert client: %w", err)
	}

	_, _ = tx.ExecContext(ctx, "DELETE FROM client_addresses WHERE client_id = ?", client.ID)
	for _, addr := range client.Addresses {
		if addr.ID == "" {
			addr.ID = uuid.New().String()
		}
		isPrim := 0
		if addr.IsPrimary {
			isPrim = 1
		}
		posSame := 0
		if addr.PostalSame {
			posSame = 1
		}
		addrQuery := `
		INSERT INTO client_addresses (id, client_id, address_type, is_primary, line1, line2, suburb, city, postal_code, country, postal_same, postal_line1, postal_line2, postal_suburb, postal_city, postal_code2, postal_country, updated_at, deleted, synced)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)`
		if _, err := tx.ExecContext(ctx, addrQuery, addr.ID, client.ID, addr.AddressType, isPrim, addr.Line1, addr.Line2, addr.Suburb, addr.City, addr.PostalCode, addr.Country, posSame, addr.PostalLine1, addr.PostalLine2, addr.PostalSuburb, addr.PostalCity, addr.PostalCode2, addr.PostalCountry, client.UpdatedAt); err != nil {
			return fmt.Errorf("insert client address: %w", err)
		}
	}

	_, _ = tx.ExecContext(ctx, "DELETE FROM client_contact_details WHERE client_id = ?", client.ID)
	for _, cd := range client.ContactDetails {
		if cd.ID == "" {
			cd.ID = uuid.New().String()
		}
		isPrim := 0
		if cd.IsPrimary {
			isPrim = 1
		}
		contactQuery := `
		INSERT INTO client_contact_details (id, client_id, contact_type, contact_value, is_primary, updated_at, deleted, synced)
		VALUES (?, ?, ?, ?, ?, ?, 0, 0)`
		if _, err := tx.ExecContext(ctx, contactQuery, cd.ID, client.ID, cd.ContactType, cd.ContactValue, isPrim, client.UpdatedAt); err != nil {
			return fmt.Errorf("insert client contact: %w", err)
		}
	}

	_, _ = tx.ExecContext(ctx, "DELETE FROM client_banks WHERE client_id = ?", client.ID)
	for _, bank := range client.Banks {
		if bank.ID == "" {
			bank.ID = uuid.New().String()
		}
		isPrim := 0
		if bank.IsPrimary {
			isPrim = 1
		}
		bankQuery := `
		INSERT INTO client_banks (id, client_id, bank_name, branch_code, account_number, account_type, is_primary, updated_at, deleted, synced)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0)`
		if _, err := tx.ExecContext(ctx, bankQuery, bank.ID, client.ID, bank.BankName, bank.BranchCode, bank.AccountNumber, bank.AccountType, isPrim, client.UpdatedAt); err != nil {
			return fmt.Errorf("insert client bank: %w", err)
		}
	}

	return tx.Commit()
}

func (s *LocalStore) getClientAddresses(ctx context.Context, clientID string) ([]ClientAddress, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, client_id, address_type, is_primary, line1, line2, suburb, city, postal_code, country, postal_same, postal_line1, postal_line2, postal_suburb, postal_city, postal_code2, postal_country, updated_at, deleted, synced FROM client_addresses WHERE client_id = ? AND deleted = 0`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ClientAddress
	for rows.Next() {
		var a ClientAddress
		var isPrim, posSame, del, syn int
		_ = rows.Scan(&a.ID, &a.ClientID, &a.AddressType, &isPrim, &a.Line1, &a.Line2, &a.Suburb, &a.City, &a.PostalCode, &a.Country, &posSame, &a.PostalLine1, &a.PostalLine2, &a.PostalSuburb, &a.PostalCity, &a.PostalCode2, &a.PostalCountry, &a.UpdatedAt, &del, &syn)
		a.IsPrimary = isPrim == 1
		a.PostalSame = posSame == 1
		a.Deleted = del == 1
		a.Synced = syn == 1
		res = append(res, a)
	}
	return res, nil
}

func (s *LocalStore) getClientContactDetails(ctx context.Context, clientID string) ([]ClientContactDetail, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, client_id, contact_type, contact_value, is_primary, updated_at, deleted, synced FROM client_contact_details WHERE client_id = ? AND deleted = 0`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ClientContactDetail
	for rows.Next() {
		var cd ClientContactDetail
		var isPrim, del, syn int
		_ = rows.Scan(&cd.ID, &cd.ClientID, &cd.ContactType, &cd.ContactValue, &isPrim, &cd.UpdatedAt, &del, &syn)
		cd.IsPrimary = isPrim == 1
		cd.Deleted = del == 1
		cd.Synced = syn == 1
		res = append(res, cd)
	}
	return res, nil
}

func (s *LocalStore) getClientBanks(ctx context.Context, clientID string) ([]ClientBank, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, client_id, bank_name, branch_code, account_number, account_type, is_primary, updated_at, deleted, synced FROM client_banks WHERE client_id = ? AND deleted = 0`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ClientBank
	for rows.Next() {
		var b ClientBank
		var isPrim, del, syn int
		_ = rows.Scan(&b.ID, &b.ClientID, &b.BankName, &b.BranchCode, &b.AccountNumber, &b.AccountType, &isPrim, &b.UpdatedAt, &del, &syn)
		b.IsPrimary = isPrim == 1
		b.Deleted = del == 1
		b.Synced = syn == 1
		res = append(res, b)
	}
	return res, nil
}

```

---

## FILE: core\db.go
```go
package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type LocalStore struct {
	db *sql.DB
}

func NewLocalStore() (*LocalStore, error) {
	appDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("user config dir: %w", err)
	}

	dataDir := filepath.Join(appDir, "LuminaLex")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "luminalex.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	store := &LocalStore{db: db}
	if err := store.initTables(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *LocalStore) initTables(ctx context.Context) error {
	categories := []string{"banks", "masters", "sheriffs", "magistrates", "highcourts", "lawfirms"}
	for _, cat := range categories {
		query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			fields TEXT NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0,
			synced INTEGER NOT NULL DEFAULT 0
		);`, cat)
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("init table %s: %w", cat, err)
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS clients (
		id TEXT PRIMARY KEY,
		first_name TEXT NOT NULL,
		middle_name TEXT,
		last_name TEXT NOT NULL,
		id_number TEXT,
		jurisdiction_type TEXT NOT NULL,
		jurisdiction_other TEXT,
		registration_number TEXT,
		occupation TEXT,
		marital_status TEXT,
		employer_name TEXT,
		employer_number TEXT,
		employer_address_l1 TEXT,
		employer_address_l2 TEXT,
		employer_suburb TEXT,
		employer_city TEXT,
		employer_postal_code TEXT,
		employer_country TEXT,
		employer_postal_same INTEGER NOT NULL DEFAULT 1,
		employer_postal_l1 TEXT,
		employer_postal_l2 TEXT,
		employer_postal_suburb TEXT,
		employer_postal_city TEXT,
		employer_postal_code2 TEXT,
		employer_postal_country TEXT,
		practice_number TEXT,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS client_addresses (
		id TEXT PRIMARY KEY,
		client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		address_type TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 0,
		line1 TEXT NOT NULL,
		line2 TEXT,
		suburb TEXT,
		city TEXT NOT NULL,
		postal_code TEXT NOT NULL,
		country TEXT NOT NULL,
		postal_same INTEGER NOT NULL DEFAULT 1,
		postal_line1 TEXT,
		postal_line2 TEXT,
		postal_suburb TEXT,
		postal_city TEXT,
		postal_code2 TEXT,
		postal_country TEXT,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS client_contact_details (
		id TEXT PRIMARY KEY,
		client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		contact_type TEXT NOT NULL,
		contact_value TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS client_banks (
		id TEXT PRIMARY KEY,
		client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		bank_name TEXT NOT NULL,
		branch_code TEXT NOT NULL,
		account_number TEXT NOT NULL,
		account_type TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);`

	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return err
	}

	// EXPLICITLY create the services table here to bypass multi-statement limitations
	servicesSchema := `
	CREATE TABLE IF NOT EXISTS services (
		id TEXT PRIMARY KEY,
		service_type TEXT NOT NULL,
		description TEXT NOT NULL,
		standard_rate REAL NOT NULL,
		duration_unit TEXT NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);`

	if _, err := s.db.ExecContext(ctx, servicesSchema); err != nil {
		return fmt.Errorf("init table services: %w", err)
	}

	return nil
}

func (s *LocalStore) GetCategoryRecords(ctx context.Context, category string) ([]ContactRecord, error) {
	query := fmt.Sprintf(`SELECT id, fields, updated_at, deleted, synced FROM %s WHERE deleted = 0`, category)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", category, err)
	}
	defer rows.Close()

	var records []ContactRecord
	for rows.Next() {
		var rec ContactRecord
		rec.Category = category
		var fieldsJSON string
		var updatedAtRaw time.Time
		var del, syn int

		if err := rows.Scan(&rec.ID, &fieldsJSON, &updatedAtRaw, &del, &syn); err != nil {
			continue
		}

		if err := json.Unmarshal([]byte(fieldsJSON), &rec.Fields); err != nil {
			rec.Fields = []string{"[Data Formatting Error]"}
		}

		rec.UpdatedAt = updatedAtRaw
		rec.Deleted = del == 1
		rec.Synced = syn == 1
		records = append(records, rec)
	}

	return records, rows.Err()
}

func (s *LocalStore) SaveRecord(ctx context.Context, record ContactRecord) error {
	fieldsJSON, err := json.Marshal(record.Fields)
	if err != nil {
		return fmt.Errorf("marshal fields: %w", err)
	}

	del := 0
	if record.Deleted {
		del = 1
	}
	syn := 0
	if record.Synced {
		syn = 1
	}

	query := fmt.Sprintf(`
	INSERT INTO %s (id, fields, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		fields = excluded.fields,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = excluded.synced
	WHERE excluded.updated_at >= %s.updated_at;`, record.Category, record.Category)

	_, err = s.db.ExecContext(ctx, query, record.ID, string(fieldsJSON), record.UpdatedAt.UTC(), del, syn)
	return err
}

func (s *LocalStore) SaveRecordsBatch(ctx context.Context, category string, records []ContactRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
	INSERT INTO %s (id, fields, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		fields = excluded.fields,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = excluded.synced
	WHERE excluded.updated_at >= %s.updated_at;`, category, category)

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		fieldsJSON, _ := json.Marshal(record.Fields)
		del := 0
		if record.Deleted {
			del = 1
		}
		syn := 0
		if record.Synced {
			syn = 1
		}

		if _, err = stmt.ExecContext(ctx, record.ID, string(fieldsJSON), record.UpdatedAt.UTC(), del, syn); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *LocalStore) GetUnsyncedRecords(ctx context.Context, category string) ([]ContactRecord, error) {
	query := fmt.Sprintf(`SELECT id, fields, updated_at, deleted FROM %s WHERE synced = 0`, category)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ContactRecord
	for rows.Next() {
		var rec ContactRecord
		rec.Category = category
		var fieldsJSON string
		var updatedAtRaw time.Time
		var del int

		if err := rows.Scan(&rec.ID, &fieldsJSON, &updatedAtRaw, &del); err != nil {
			continue
		}

		_ = json.Unmarshal([]byte(fieldsJSON), &rec.Fields)
		rec.UpdatedAt = updatedAtRaw
		rec.Deleted = del == 1
		rec.Synced = false
		records = append(records, rec)
	}

	return records, rows.Err()
}

func (s *LocalStore) MarkSynced(ctx context.Context, category string, id string) error {
	query := fmt.Sprintf(`UPDATE %s SET synced = 1 WHERE id = ?`, category)
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *LocalStore) Close() error {
	return s.db.Close()
}

```

---

## FILE: core\export.go
```go
package core

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

func (a *App) ExportTableToExcel(payload ExportPayload) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Export"
	f.SetSheetName("Sheet1", sheetName)

	for colIdx, header := range payload.Headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, row := range payload.Rows {
		for colIdx, field := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, field)
		}
	}

	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Data",
		DefaultFilename: fmt.Sprintf("%s_export.xlsx", payload.Filename),
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})

	if err != nil {
		return fmt.Errorf("dialog error: %w", err)
	}

	if filePath == "" {
		return nil
	}

	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to save excel file: %w", err)
	}

	return nil
}

```

---

## FILE: core\handler.go
```go
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"luminalex/views"
)

func NewHandler(app *App) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lastUser := app.GetLastUsername()
		err := views.Login(views.LoginData{LastUsername: lastUser}).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering login page failed", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "parsing form failed", http.StatusBadRequest)
			return
		}
		user := r.FormValue("username")
		pass := r.FormValue("password")

		if err := app.AuthenticateUser(user, pass); err != nil {
			lastUser := app.GetLastUsername()
			_ = views.Login(views.LoginData{Error: err.Error(), LastUsername: lastUser}).Render(r.Context(), w)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
	})

	mux.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		err := views.Home(views.HomeData{Title: "LuminaLex"}).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering home page failed", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/clients", func(w http.ResponseWriter, r *http.Request) {
		coreClients, err := app.store.GetClients(r.Context())
		if err != nil {
			coreClients = []Client{}
		}

		var viewClients []views.Client
		for _, c := range coreClients {
			vc := views.Client{
				ID:                    c.ID,
				FirstName:             c.FirstName,
				MiddleName:            c.MiddleName,
				LastName:              c.LastName,
				IDNumber:              c.IDNumber,
				JurisdictionType:      c.JurisdictionType,
				JurisdictionOther:     c.JurisdictionOther,
				RegistrationNumber:    c.RegistrationNumber,
				Occupation:            c.Occupation,
				MaritalStatus:         c.MaritalStatus,
				EmployerName:          c.EmployerName,
				EmployerNumber:        c.EmployerNumber,
				EmployerAddressL1:     c.EmployerAddressL1,
				EmployerAddressL2:     c.EmployerAddressL2,
				EmployerSuburb:        c.EmployerSuburb,
				EmployerCity:          c.EmployerCity,
				EmployerPostalCode:    c.EmployerPostalCode,
				EmployerCountry:       c.EmployerCountry,
				EmployerPostalSame:    c.EmployerPostalSame,
				EmployerPostalL1:      c.EmployerPostalL1,
				EmployerPostalL2:      c.EmployerPostalL2,
				EmployerPostalSuburb:  c.EmployerPostalSuburb,
				EmployerPostalCity:    c.EmployerPostalCity,
				EmployerPostalCode2:   c.EmployerPostalCode2,
				EmployerPostalCountry: c.EmployerPostalCountry,
				PracticeNumber:        c.PracticeNumber,
			}
			for _, addr := range c.Addresses {
				vc.Addresses = append(vc.Addresses, views.ClientAddress{
					ID: addr.ID, AddressType: addr.AddressType, IsPrimary: addr.IsPrimary, Line1: addr.Line1, Line2: addr.Line2, Suburb: addr.Suburb, City: addr.City, PostalCode: addr.PostalCode, Country: addr.Country, PostalSame: addr.PostalSame, PostalLine1: addr.PostalLine1, PostalLine2: addr.PostalLine2, PostalSuburb: addr.PostalSuburb, PostalCity: addr.PostalCity, PostalCode2: addr.PostalCode2, PostalCountry: addr.PostalCountry,
				})
			}
			for _, cd := range c.ContactDetails {
				vc.ContactDetails = append(vc.ContactDetails, views.ClientContactDetail{
					ID: cd.ID, ContactType: cd.ContactType, ContactValue: cd.ContactValue, IsPrimary: cd.IsPrimary,
				})
			}
			for _, bk := range c.Banks {
				vc.Banks = append(vc.Banks, views.ClientBank{
					ID: bk.ID, BankName: bk.BankName, BranchCode: bk.BranchCode, AccountNumber: bk.AccountNumber, AccountType: bk.AccountType, IsPrimary: bk.IsPrimary,
				})
			}
			viewClients = append(viewClients, vc)
		}

		err = views.Clients(viewClients).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering clients page failed", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/clients/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var client Client
		if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request format: %v", err), http.StatusBadRequest)
			return
		}
		if err := app.store.SaveClient(r.Context(), client); err != nil {
			http.Error(w, fmt.Sprintf("Database Error: %v", err), http.StatusInternalServerError)
			return
		}
		go func() {
			_ = app.syncEngine.PerformSync(context.Background())
		}()
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		err := views.Services().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering services page failed", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		serviceType := r.URL.Query().Get("type")
		if serviceType == "" {
			http.Error(w, "service type required", http.StatusBadRequest)
			return
		}
		services, err := app.store.GetServices(r.Context(), serviceType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(services)
	})

	mux.HandleFunc("/api/services/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var svc Service
		if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.store.SaveService(r.Context(), svc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() {
			_ = app.syncEngine.PerformSync(context.Background())
		}()
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/services/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.store.DeleteService(r.Context(), req.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() {
			_ = app.syncEngine.PerformSync(context.Background())
		}()
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		err := views.Contacts().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering contacts page failed", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/contacts/export_view", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload ExportPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.ExportTableToExcel(payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/contacts", func(w http.ResponseWriter, r *http.Request) {
		cat := r.URL.Query().Get("category")
		records, err := app.GetContacts(cat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(records)
	})

	mux.HandleFunc("/api/contacts/save", func(w http.ResponseWriter, r *http.Request) {
		var rec ContactRecord
		if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.SaveContact(rec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/contacts/delete", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Category string `json:"category"`
			ID       string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.DeleteContact(req.Category, req.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/sync", func(w http.ResponseWriter, r *http.Request) {
		status := app.TriggerSync()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/api/check-update", func(w http.ResponseWriter, r *http.Request) {
		result, err := app.CheckUpdate()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("/api/update/apply", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.PerformUpdate(req.URL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/update/restart", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		go func() {
			_ = app.RestartApp()
		}()
	})

	mux.HandleFunc("/matters", func(w http.ResponseWriter, r *http.Request) {
		currentUser := app.GetLastUsername()
		coreClients, _ := app.store.GetClients(r.Context())
		coreServices, _ := app.store.GetAllServices(r.Context())

		err := views.Matters(coreClients, coreServices, currentUser).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering matters page failed", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/matters", func(w http.ResponseWriter, r *http.Request) {
		matters, err := app.store.GetMatters(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(matters)
	})

	mux.HandleFunc("/api/matters/save", func(w http.ResponseWriter, r *http.Request) {
		var m Matter
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if m.Reference == "" {
			m.Reference = app.GenerateMatterReference()
		}
		if err := app.store.SaveMatter(r.Context(), m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() { _ = app.syncEngine.PerformSync(context.Background()) }()
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/matters/delete", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = app.store.DeleteMatter(r.Context(), req.ID)
		go func() { _ = app.syncEngine.PerformSync(context.Background()) }()
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/matters", func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		var matters []Matter
		var err error

		if clientID != "" {
			matters, err = app.store.GetMattersByClient(r.Context(), clientID)
		} else {
			matters, err = app.store.GetMatters(r.Context())
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(matters)
	})
	// Add identical handlers for notes (/api/matters/notes, /save, /delete)
	// and services (/api/matters/services, /save, /delete) mapping to the
	// newly added core/matters.go functions.

	return mux
}

```

---

## FILE: core\matters.go
```go
package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (a *App) GenerateMatterReference() string {
	username := a.GetLastUsername()
	initials := "USR"
	if len(username) >= 2 {
		initials = strings.ToUpper(username[:2])
	}
	return fmt.Sprintf("WW-%s-%d", initials, time.Now().Unix())
}

func (s *LocalStore) GetMatters(ctx context.Context) ([]Matter, error) {
	query := `SELECT id, reference, client_id, status, matter_type, updated_at, deleted, synced FROM matters WHERE deleted = 0 ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query matters: %w", err)
	}
	defer rows.Close()

	var matters []Matter
	for rows.Next() {
		var m Matter
		var del, syn int
		if err := rows.Scan(&m.ID, &m.Reference, &m.ClientID, &m.Status, &m.MatterType, &m.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		m.Deleted = del == 1
		m.Synced = syn == 1
		matters = append(matters, m)
	}
	return matters, rows.Err()
}

func (s *LocalStore) SaveMatter(ctx context.Context, m Matter) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	m.UpdatedAt = time.Now().UTC()
	del := 0
	if m.Deleted {
		del = 1
	}

	query := `
	INSERT INTO matters (id, reference, client_id, status, matter_type, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		status = excluded.status,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err := s.db.ExecContext(ctx, query, m.ID, m.Reference, m.ClientID, m.Status, m.MatterType, m.UpdatedAt, del)
	return err
}

func (s *LocalStore) DeleteMatter(ctx context.Context, id string) error {
	query := `UPDATE matters SET deleted = 1, updated_at = ?, synced = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (s *LocalStore) GetMatterNotes(ctx context.Context, matterID string) ([]MatterNote, error) {
	query := `SELECT id, matter_id, author, content, updated_at, deleted, synced FROM matter_notes WHERE matter_id = ? AND deleted = 0 ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query, matterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []MatterNote
	for rows.Next() {
		var n MatterNote
		var del, syn int
		if err := rows.Scan(&n.ID, &n.MatterID, &n.Author, &n.Content, &n.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		n.Deleted = del == 1
		n.Synced = syn == 1
		notes = append(notes, n)
	}
	return notes, nil
}

func (s *LocalStore) SaveMatterNote(ctx context.Context, n MatterNote) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	n.UpdatedAt = time.Now().UTC()
	del := 0
	if n.Deleted {
		del = 1
	}

	query := `
	INSERT INTO matter_notes (id, matter_id, author, content, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		content = excluded.content,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err := s.db.ExecContext(ctx, query, n.ID, n.MatterID, n.Author, n.Content, n.UpdatedAt, del)
	return err
}

func (s *LocalStore) DeleteMatterNote(ctx context.Context, id string) error {
	query := `UPDATE matter_notes SET deleted = 1, updated_at = ?, synced = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (s *LocalStore) GetMatterServices(ctx context.Context, matterID string) ([]MatterService, error) {
	query := `SELECT id, matter_id, service_id, snapshot_desc, snapshot_rate, snapshot_unit, qty, add_tax, updated_at, deleted, synced FROM matter_services WHERE matter_id = ? AND deleted = 0 ORDER BY updated_at ASC`
	rows, err := s.db.QueryContext(ctx, query, matterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var svcs []MatterService
	for rows.Next() {
		var svc MatterService
		var addTax, del, syn int
		if err := rows.Scan(&svc.ID, &svc.MatterID, &svc.ServiceID, &svc.SnapshotDesc, &svc.SnapshotRate, &svc.SnapshotUnit, &svc.Qty, &addTax, &svc.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		svc.AddTax = addTax == 1
		svc.Deleted = del == 1
		svc.Synced = syn == 1
		svcs = append(svcs, svc)
	}
	return svcs, nil
}

func (s *LocalStore) SaveMatterService(ctx context.Context, svc MatterService) error {
	if svc.ID == "" {
		svc.ID = uuid.New().String()
	}
	svc.UpdatedAt = time.Now().UTC()
	addTax := 0
	if svc.AddTax {
		addTax = 1
	}
	del := 0
	if svc.Deleted {
		del = 1
	}

	query := `
	INSERT INTO matter_services (id, matter_id, service_id, snapshot_desc, snapshot_rate, snapshot_unit, qty, add_tax, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		snapshot_desc = excluded.snapshot_desc,
		snapshot_rate = excluded.snapshot_rate,
		snapshot_unit = excluded.snapshot_unit,
		qty = excluded.qty,
		add_tax = excluded.add_tax,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err := s.db.ExecContext(ctx, query, svc.ID, svc.MatterID, svc.ServiceID, svc.SnapshotDesc, svc.SnapshotRate, svc.SnapshotUnit, svc.Qty, addTax, svc.UpdatedAt, del)
	return err
}

func (s *LocalStore) DeleteMatterService(ctx context.Context, id string) error {
	query := `UPDATE matter_services SET deleted = 1, updated_at = ?, synced = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (s *LocalStore) GetMattersByClient(ctx context.Context, clientID string) ([]Matter, error) {
	query := `SELECT id, reference, client_id, status, matter_type, updated_at, deleted, synced FROM matters WHERE deleted = 0 AND client_id = ? ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("query matters by client: %w", err)
	}
	defer rows.Close()

	var matters []Matter
	for rows.Next() {
		var m Matter
		var del, syn int
		if err := rows.Scan(&m.ID, &m.Reference, &m.ClientID, &m.Status, &m.MatterType, &m.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		m.Deleted = del == 1
		m.Synced = syn == 1
		matters = append(matters, m)
	}
	return matters, rows.Err()
}

```

---

## FILE: core\model.go
```go
package core

import "time"

type ContactRecord struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Fields    []string  `json:"fields"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
	Synced    bool      `json:"synced"`
}

type UpdateCheckResult struct {
	HasUpdate    bool   `json:"has_update"`
	LatestVer    string `json:"latest_ver"`
	ReleaseNotes string `json:"release_notes"`
	DownloadURL  string `json:"download_url"`
}

type SyncStatus struct {
	IsSyncing bool   `json:"is_syncing"`
	LastSync  string `json:"last_sync"`
	Error     string `json:"error"`
	Details   string `json:"details"`
}

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Enabled      bool   `json:"enabled"`
}

type Client struct {
	ID                    string                `json:"id"`
	FirstName             string                `json:"first_name"`
	MiddleName            string                `json:"middle_name"`
	LastName              string                `json:"last_name"`
	IDNumber              string                `json:"id_number"`
	JurisdictionType      string                `json:"jurisdiction_type"`
	JurisdictionOther     string                `json:"jurisdiction_other"`
	RegistrationNumber    string                `json:"registration_number"`
	Occupation            string                `json:"occupation"`
	MaritalStatus         string                `json:"marital_status"`
	EmployerName          string                `json:"employer_name"`
	EmployerNumber        string                `json:"employer_number"`
	EmployerAddressL1     string                `json:"employer_address_l1"`
	EmployerAddressL2     string                `json:"employer_address_l2"`
	EmployerSuburb        string                `json:"employer_suburb"`
	EmployerCity          string                `json:"employer_city"`
	EmployerPostalCode    string                `json:"employer_postal_code"`
	EmployerCountry       string                `json:"employer_country"`
	EmployerPostalSame    bool                  `json:"employer_postal_same"`
	EmployerPostalL1      string                `json:"employer_postal_l1"`
	EmployerPostalL2      string                `json:"employer_postal_l2"`
	EmployerPostalSuburb  string                `json:"employer_postal_suburb"`
	EmployerPostalCity    string                `json:"employer_postal_city"`
	EmployerPostalCode2   string                `json:"employer_postal_code2"`
	EmployerPostalCountry string                `json:"employer_postal_country"`
	PracticeNumber        string                `json:"practice_number"`
	UpdatedAt             time.Time             `json:"updated_at"`
	Deleted               bool                  `json:"deleted"`
	Synced                bool                  `json:"synced"`
	Addresses             []ClientAddress       `json:"addresses"`
	ContactDetails        []ClientContactDetail `json:"contact_details"`
	Banks                 []ClientBank          `json:"banks"`
}

type ClientSyncDTO struct {
	ID                    string    `json:"id"`
	FirstName             string    `json:"first_name"`
	MiddleName            string    `json:"middle_name"`
	LastName              string    `json:"last_name"`
	IDNumber              string    `json:"id_number"`
	JurisdictionType      string    `json:"jurisdiction_type"`
	JurisdictionOther     string    `json:"jurisdiction_other"`
	RegistrationNumber    string    `json:"registration_number"`
	Occupation            string    `json:"occupation"`
	MaritalStatus         string    `json:"marital_status"`
	EmployerName          string    `json:"employer_name"`
	EmployerNumber        string    `json:"employer_number"`
	EmployerAddressL1     string    `json:"employer_address_l1"`
	EmployerAddressL2     string    `json:"employer_address_l2"`
	EmployerSuburb        string    `json:"employer_suburb"`
	EmployerCity          string    `json:"employer_city"`
	EmployerPostalCode    string    `json:"employer_postal_code"`
	EmployerCountry       string    `json:"employer_country"`
	EmployerPostalSame    bool      `json:"employer_postal_same"`
	EmployerPostalL1      string    `json:"employer_postal_l1"`
	EmployerPostalL2      string    `json:"employer_postal_l2"`
	EmployerPostalSuburb  string    `json:"employer_postal_suburb"`
	EmployerPostalCity    string    `json:"employer_postal_city"`
	EmployerPostalCode2   string    `json:"employer_postal_code2"`
	EmployerPostalCountry string    `json:"employer_postal_country"`
	PracticeNumber        string    `json:"practice_number"`
	UpdatedAt             time.Time `json:"updated_at"`
	Deleted               bool      `json:"deleted"`
}

type ClientAddress struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	AddressType   string    `json:"address_type"`
	IsPrimary     bool      `json:"is_primary"`
	Line1         string    `json:"line1"`
	Line2         string    `json:"line2"`
	Suburb        string    `json:"suburb"`
	City          string    `json:"city"`
	PostalCode    string    `json:"postal_code"`
	Country       string    `json:"country"`
	PostalSame    bool      `json:"postal_same"`
	PostalLine1   string    `json:"postal_line1"`
	PostalLine2   string    `json:"postal_line2"`
	PostalSuburb  string    `json:"postal_suburb"`
	PostalCity    string    `json:"postal_city"`
	PostalCode2   string    `json:"postal_code2"`
	PostalCountry string    `json:"postal_country"`
	UpdatedAt     time.Time `json:"updated_at"`
	Deleted       bool      `json:"deleted"`
	Synced        bool      `json:"synced"`
}

type AddressSyncDTO struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	AddressType   string    `json:"address_type"`
	IsPrimary     bool      `json:"is_primary"`
	Line1         string    `json:"line1"`
	Line2         string    `json:"line2"`
	Suburb        string    `json:"suburb"`
	City          string    `json:"city"`
	PostalCode    string    `json:"postal_code"`
	Country       string    `json:"country"`
	PostalSame    bool      `json:"postal_same"`
	PostalLine1   string    `json:"postal_line1"`
	PostalLine2   string    `json:"postal_line2"`
	PostalSuburb  string    `json:"postal_suburb"`
	PostalCity    string    `json:"postal_city"`
	PostalCode2   string    `json:"postal_code2"`
	PostalCountry string    `json:"postal_country"`
	UpdatedAt     time.Time `json:"updated_at"`
	Deleted       bool      `json:"deleted"`
}

type ClientContactDetail struct {
	ID           string    `json:"id"`
	ClientID     string    `json:"client_id"`
	ContactType  string    `json:"contact_type"`
	ContactValue string    `json:"contact_value"`
	IsPrimary    bool      `json:"is_primary"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
	Synced       bool      `json:"synced"`
}

type ContactDetailSyncDTO struct {
	ID           string    `json:"id"`
	ClientID     string    `json:"client_id"`
	ContactType  string    `json:"contact_type"`
	ContactValue string    `json:"contact_value"`
	IsPrimary    bool      `json:"is_primary"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
}

type ClientBank struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	BankName      string    `json:"bank_name"`
	BranchCode    string    `json:"branch_code"`
	AccountNumber string    `json:"account_number"`
	AccountType   string    `json:"account_type"`
	IsPrimary     bool      `json:"is_primary"`
	UpdatedAt     time.Time `json:"updated_at"`
	Deleted       bool      `json:"deleted"`
	Synced        bool      `json:"synced"`
}

type BankSyncDTO struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	BankName      string    `json:"bank_name"`
	BranchCode    string    `json:"branch_code"`
	AccountNumber string    `json:"account_number"`
	AccountType   string    `json:"account_type"`
	IsPrimary     bool      `json:"is_primary"`
	UpdatedAt     time.Time `json:"updated_at"`
	Deleted       bool      `json:"deleted"`
}

type Service struct {
	ID           string    `json:"id"`
	ServiceType  string    `json:"service_type"`
	Description  string    `json:"description"`
	StandardRate float64   `json:"standard_rate"`
	DurationUnit string    `json:"duration_unit"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
	Synced       bool      `json:"synced"`
}

type ServiceSyncDTO struct {
	ID           string    `json:"id"`
	ServiceType  string    `json:"service_type"`
	Description  string    `json:"description"`
	StandardRate float64   `json:"standard_rate"`
	DurationUnit string    `json:"duration_unit"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
}

type AppConfig struct {
	LastUsername string `json:"last_username"`
}

type ExportPayload struct {
	Filename string     `json:"filename"`
	Headers  []string   `json:"headers"`
	Rows     [][]string `json:"rows"`
}

type Matter struct {
	ID         string    `json:"id"`
	Reference  string    `json:"reference"`
	ClientID   string    `json:"client_id"`
	Status     string    `json:"status"`
	MatterType string    `json:"matter_type"`
	UpdatedAt  time.Time `json:"updated_at"`
	Deleted    bool      `json:"deleted"`
	Synced     bool      `json:"synced"`
}

type MatterNote struct {
	ID        string    `json:"id"`
	MatterID  string    `json:"matter_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
	Synced    bool      `json:"synced"`
}

type MatterService struct {
	ID           string    `json:"id"`
	MatterID     string    `json:"matter_id"`
	ServiceID    string    `json:"service_id"`
	SnapshotDesc string    `json:"snapshot_desc"`
	SnapshotRate float64   `json:"snapshot_rate"`
	SnapshotUnit string    `json:"snapshot_unit"`
	Qty          float64   `json:"qty"`
	AddTax       bool      `json:"add_tax"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
	Synced       bool      `json:"synced"`
}

```

---

## FILE: core\services.go
```go
package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *LocalStore) GetServices(ctx context.Context, serviceType string) ([]Service, error) {
	query := `SELECT id, service_type, description, standard_rate, duration_unit, updated_at, deleted, synced FROM services WHERE deleted = 0 AND service_type = ? ORDER BY description`
	rows, err := s.db.QueryContext(ctx, query, serviceType)
	if err != nil {
		return nil, fmt.Errorf("query services: %w", err)
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var svc Service
		var del, syn int
		if err := rows.Scan(&svc.ID, &svc.ServiceType, &svc.Description, &svc.StandardRate, &svc.DurationUnit, &svc.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		svc.Deleted = del == 1
		svc.Synced = syn == 1
		services = append(services, svc)
	}
	return services, rows.Err()
}

func (s *LocalStore) GetAllServices(ctx context.Context) ([]Service, error) {
	query := `SELECT id, service_type, description, standard_rate, duration_unit, updated_at, deleted, synced FROM services WHERE deleted = 0`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query all services: %w", err)
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var svc Service
		var del, syn int
		if err := rows.Scan(&svc.ID, &svc.ServiceType, &svc.Description, &svc.StandardRate, &svc.DurationUnit, &svc.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		svc.Deleted = del == 1
		svc.Synced = syn == 1
		services = append(services, svc)
	}
	return services, rows.Err()
}

func (s *LocalStore) SaveService(ctx context.Context, svc Service) error {
	if svc.ID == "" {
		svc.ID = uuid.New().String()
	}
	svc.UpdatedAt = time.Now().UTC()

	del := 0
	if svc.Deleted {
		del = 1
	}

	query := `
	INSERT INTO services (id, service_type, description, standard_rate, duration_unit, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		service_type = excluded.service_type,
		description = excluded.description,
		standard_rate = excluded.standard_rate,
		duration_unit = excluded.duration_unit,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err := s.db.ExecContext(ctx, query, svc.ID, svc.ServiceType, svc.Description, svc.StandardRate, svc.DurationUnit, svc.UpdatedAt, del)
	return err
}

func (s *LocalStore) DeleteService(ctx context.Context, id string) error {
	query := `UPDATE services SET deleted = 1, updated_at = ?, synced = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

```

---

## FILE: core\supabase.go
```go
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SupabaseClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewSupabaseClient(url, key string) *SupabaseClient {
	return &SupabaseClient{
		baseURL: url,
		apiKey:  key,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *SupabaseClient) FetchUpdatedAfter(ctx context.Context, category string, after time.Time) ([]ContactRecord, error) {
	var endpoint string

	if after.IsZero() {
		endpoint = fmt.Sprintf("%s/rest/v1/%s?select=id,fields,updated_at,deleted", c.baseURL, category)
	} else {
		endpoint = fmt.Sprintf("%s/rest/v1/%s?select=id,fields,updated_at,deleted&updated_at=gt.%s",
			c.baseURL, category, after.UTC().Format("2006-01-02T15:04:05.999999Z"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var remoteRows []struct {
		ID        string          `json:"id"`
		Fields    json.RawMessage `json:"fields"`
		UpdatedAt time.Time       `json:"updated_at"`
		Deleted   bool            `json:"deleted"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&remoteRows); err != nil {
		return nil, err
	}

	records := make([]ContactRecord, 0, len(remoteRows))
	for _, row := range remoteRows {
		var fields []string

		if len(row.Fields) > 0 {
			if row.Fields[0] == '"' {
				var unescaped string
				if err := json.Unmarshal(row.Fields, &unescaped); err == nil {
					_ = json.Unmarshal([]byte(unescaped), &fields)
				}
			} else {
				_ = json.Unmarshal(row.Fields, &fields)
			}
		}

		if fields == nil {
			fields = []string{}
		}

		records = append(records, ContactRecord{
			ID:        row.ID,
			Category:  category,
			Fields:    fields,
			UpdatedAt: row.UpdatedAt,
			Deleted:   row.Deleted,
			Synced:    true,
		})
	}

	return records, nil
}

func (c *SupabaseClient) FetchClientsAfter(ctx context.Context, after time.Time) ([]Client, error) {
	var endpoint string
	if after.IsZero() {
		endpoint = fmt.Sprintf("%s/rest/v1/clients?select=*", c.baseURL)
	} else {
		endpoint = fmt.Sprintf("%s/rest/v1/clients?select=*&updated_at=gt.%s",
			c.baseURL, after.UTC().Format("2006-01-02T15:04:05.999999Z"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var clients []Client
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, err
	}
	return clients, nil
}

func (c *SupabaseClient) FetchAddresses(ctx context.Context, clientIDs []string) ([]ClientAddress, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/client_addresses?select=*", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var addresses []ClientAddress
	if err := json.NewDecoder(resp.Body).Decode(&addresses); err != nil {
		return nil, err
	}
	return addresses, nil
}

func (c *SupabaseClient) FetchContacts(ctx context.Context, clientIDs []string) ([]ClientContactDetail, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/client_contact_details?select=*", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var contacts []ClientContactDetail
	if err := json.NewDecoder(resp.Body).Decode(&contacts); err != nil {
		return nil, err
	}
	return contacts, nil
}

func (c *SupabaseClient) FetchBanks(ctx context.Context, clientIDs []string) ([]ClientBank, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/client_banks?select=*", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var banks []ClientBank
	if err := json.NewDecoder(resp.Body).Decode(&banks); err != nil {
		return nil, err
	}
	return banks, nil
}

func (c *SupabaseClient) FetchServicesAfter(ctx context.Context, after time.Time) ([]Service, error) {
	var endpoint string
	if after.IsZero() {
		endpoint = fmt.Sprintf("%s/rest/v1/services?select=*", c.baseURL)
	} else {
		endpoint = fmt.Sprintf("%s/rest/v1/services?select=*&updated_at=gt.%s",
			c.baseURL, after.UTC().Format("2006-01-02T15:04:05.999999Z"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var services []Service
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, err
	}
	return services, nil
}

func (c *SupabaseClient) UpsertRecord(ctx context.Context, record ContactRecord) error {
	payload := map[string]interface{}{
		"id":         record.ID,
		"fields":     record.Fields,
		"updated_at": record.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999Z"),
		"deleted":    record.Deleted,
	}
	return c.UpsertGeneric(ctx, record.Category, payload)
}

func (c *SupabaseClient) UpsertGeneric(ctx context.Context, table string, payload any) error {
	endpoint := fmt.Sprintf("%s/rest/v1/%s?on_conflict=id", c.baseURL, table)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	c.setHeaders(req)
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upsert %s status %d", table, resp.StatusCode)
	}

	return nil
}

func (c *SupabaseClient) setHeaders(req *http.Request) {
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}

```

---

## FILE: core\sync.go
```go
package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type SyncEngine struct {
	store      *LocalStore
	client     *SupabaseClient
	categories []string
	lastSync   time.Time
	mu         sync.Mutex
	isSyncing  bool
}

func NewSyncEngine(store *LocalStore, client *SupabaseClient) *SyncEngine {
	return &SyncEngine{
		store:      store,
		client:     client,
		categories: []string{"banks", "masters", "sheriffs", "magistrates", "highcourts", "lawfirms"},
		lastSync:   time.Time{},
	}
}

func (e *SyncEngine) PerformSync(ctx context.Context) error {
	e.mu.Lock()
	if e.isSyncing {
		e.mu.Unlock()
		return nil
	}
	e.isSyncing = true
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.isSyncing = false
		e.mu.Unlock()
	}()

	var syncErrors []error

	for _, cat := range e.categories {
		unsynced, err := e.store.GetUnsyncedRecords(ctx, cat)
		if err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("fetch unsynced for %s: %w", cat, err))
			continue
		}

		for _, rec := range unsynced {
			if err := e.client.UpsertRecord(ctx, rec); err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("push record %s: %w", rec.ID, err))
				continue
			}
			_ = e.store.MarkSynced(ctx, cat, rec.ID)
		}

		remoteRecords, err := e.client.FetchUpdatedAfter(ctx, cat, e.lastSync)
		if err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("fetch remote for %s: %w", cat, err))
			continue
		}

		_ = e.store.SaveRecordsBatch(ctx, cat, remoteRecords)
	}

	if err := e.syncRelationalClients(ctx); err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("sync clients: %w", err))
	}

	if err := e.syncRelationalServices(ctx); err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("sync services: %w", err))
	}

	if len(syncErrors) == 0 {
		e.lastSync = time.Now().UTC()
		return nil
	}

	return errors.Join(syncErrors...)
}

func (e *SyncEngine) syncRelationalClients(ctx context.Context) error {
	localClients, err := e.store.GetClients(ctx)
	if err != nil {
		return err
	}

	for _, client := range localClients {
		if !client.Synced {
			dto := ClientSyncDTO{
				ID:                    client.ID,
				FirstName:             client.FirstName,
				MiddleName:            client.MiddleName,
				LastName:              client.LastName,
				IDNumber:              client.IDNumber,
				JurisdictionType:      client.JurisdictionType,
				JurisdictionOther:     client.JurisdictionOther,
				RegistrationNumber:    client.RegistrationNumber,
				Occupation:            client.Occupation,
				MaritalStatus:         client.MaritalStatus,
				EmployerName:          client.EmployerName,
				EmployerNumber:        client.EmployerNumber,
				EmployerAddressL1:     client.EmployerAddressL1,
				EmployerAddressL2:     client.EmployerAddressL2,
				EmployerSuburb:        client.EmployerSuburb,
				EmployerCity:          client.EmployerCity,
				EmployerPostalCode:    client.EmployerPostalCode,
				EmployerCountry:       client.EmployerCountry,
				EmployerPostalSame:    client.EmployerPostalSame,
				EmployerPostalL1:      client.EmployerPostalL1,
				EmployerPostalL2:      client.EmployerPostalL2,
				EmployerPostalSuburb:  client.EmployerPostalSuburb,
				EmployerPostalCity:    client.EmployerPostalCity,
				EmployerPostalCode2:   client.EmployerPostalCode2,
				EmployerPostalCountry: client.EmployerPostalCountry,
				PracticeNumber:        client.PracticeNumber,
				UpdatedAt:             client.UpdatedAt,
				Deleted:               client.Deleted,
			}
			if err := e.client.UpsertGeneric(ctx, "clients", dto); err != nil {
				return err
			}
			_, _ = e.store.db.ExecContext(ctx, "UPDATE clients SET synced = 1 WHERE id = ?", client.ID)
		}

		for _, addr := range client.Addresses {
			if !addr.Synced {
				dto := AddressSyncDTO{
					ID:            addr.ID,
					ClientID:      addr.ClientID,
					AddressType:   addr.AddressType,
					IsPrimary:     addr.IsPrimary,
					Line1:         addr.Line1,
					Line2:         addr.Line2,
					Suburb:        addr.Suburb,
					City:          addr.City,
					PostalCode:    addr.PostalCode,
					Country:       addr.Country,
					PostalSame:    addr.PostalSame,
					PostalLine1:   addr.PostalLine1,
					PostalLine2:   addr.PostalLine2,
					PostalSuburb:  addr.PostalSuburb,
					PostalCity:    addr.PostalCity,
					PostalCode2:   addr.PostalCode2,
					PostalCountry: addr.PostalCountry,
					UpdatedAt:     addr.UpdatedAt,
					Deleted:       addr.Deleted,
				}
				if err := e.client.UpsertGeneric(ctx, "client_addresses", dto); err != nil {
					return err
				}
				_, _ = e.store.db.ExecContext(ctx, "UPDATE client_addresses SET synced = 1 WHERE id = ?", addr.ID)
			}
		}

		for _, cd := range client.ContactDetails {
			if !cd.Synced {
				dto := ContactDetailSyncDTO{
					ID:           cd.ID,
					ClientID:     cd.ClientID,
					ContactType:  cd.ContactType,
					ContactValue: cd.ContactValue,
					IsPrimary:    cd.IsPrimary,
					UpdatedAt:    cd.UpdatedAt,
					Deleted:      cd.Deleted,
				}
				if err := e.client.UpsertGeneric(ctx, "client_contact_details", dto); err != nil {
					return err
				}
				_, _ = e.store.db.ExecContext(ctx, "UPDATE client_contact_details SET synced = 1 WHERE id = ?", cd.ID)
			}
		}

		for _, bank := range client.Banks {
			if !bank.Synced {
				dto := BankSyncDTO{
					ID:            bank.ID,
					ClientID:      bank.ClientID,
					BankName:      bank.BankName,
					BranchCode:    bank.BranchCode,
					AccountNumber: bank.AccountNumber,
					AccountType:   bank.AccountType,
					IsPrimary:     bank.IsPrimary,
					UpdatedAt:     bank.UpdatedAt,
					Deleted:       bank.Deleted,
				}
				if err := e.client.UpsertGeneric(ctx, "client_banks", dto); err != nil {
					return err
				}
				_, _ = e.store.db.ExecContext(ctx, "UPDATE client_banks SET synced = 1 WHERE id = ?", bank.ID)
			}
		}
	}

	remoteClients, err := e.client.FetchClientsAfter(ctx, e.lastSync)
	if err != nil {
		return err
	}

	if len(remoteClients) > 0 {
		var clientIDs []string
		for _, c := range remoteClients {
			clientIDs = append(clientIDs, c.ID)
		}

		addresses, _ := e.client.FetchAddresses(ctx, clientIDs)
		contacts, _ := e.client.FetchContacts(ctx, clientIDs)
		banks, _ := e.client.FetchBanks(ctx, clientIDs)

		addressMap := make(map[string][]ClientAddress)
		for _, a := range addresses {
			a.Synced = true
			addressMap[a.ClientID] = append(addressMap[a.ClientID], a)
		}

		contactMap := make(map[string][]ClientContactDetail)
		for _, c := range contacts {
			c.Synced = true
			contactMap[c.ClientID] = append(contactMap[c.ClientID], c)
		}

		bankMap := make(map[string][]ClientBank)
		for _, b := range banks {
			b.Synced = true
			bankMap[b.ClientID] = append(bankMap[b.ClientID], b)
		}

		for _, remote := range remoteClients {
			remote.Addresses = addressMap[remote.ID]
			remote.ContactDetails = contactMap[remote.ID]
			remote.Banks = bankMap[remote.ID]
			remote.Synced = true

			_ = e.store.SaveClient(ctx, remote)
			_, _ = e.store.db.ExecContext(ctx, "UPDATE clients SET synced = 1 WHERE id = ?", remote.ID)
		}
	}

	return nil
}

func (e *SyncEngine) syncRelationalServices(ctx context.Context) error {
	localServices, err := e.store.GetAllServices(ctx)
	if err != nil {
		return err
	}

	for _, svc := range localServices {
		if !svc.Synced {
			dto := ServiceSyncDTO{
				ID:           svc.ID,
				ServiceType:  svc.ServiceType,
				Description:  svc.Description,
				StandardRate: svc.StandardRate,
				DurationUnit: svc.DurationUnit,
				UpdatedAt:    svc.UpdatedAt,
				Deleted:      svc.Deleted,
			}
			if err := e.client.UpsertGeneric(ctx, "services", dto); err != nil {
				return err
			}
			_, _ = e.store.db.ExecContext(ctx, "UPDATE services SET synced = 1 WHERE id = ?", svc.ID)
		}
	}

	remoteServices, err := e.client.FetchServicesAfter(ctx, e.lastSync)
	if err != nil {
		return err
	}

	for _, remote := range remoteServices {
		remote.Synced = true
		if err := e.store.SaveService(ctx, remote); err != nil {
			return err
		}
		_, _ = e.store.db.ExecContext(ctx, "UPDATE services SET synced = 1 WHERE id = ?", remote.ID)
	}

	return nil
}

func (e *SyncEngine) GetStatus() SyncStatus {
	e.mu.Lock()
	defer e.mu.Unlock()

	lastSyncStr := "Never"
	if !e.lastSync.IsZero() {
		lastSyncStr = e.lastSync.Format("2006-01-02 15:04:05")
	}

	return SyncStatus{
		IsSyncing: e.isSyncing,
		LastSync:  lastSyncStr,
	}
}

```

---

## FILE: core\updater.go
```go
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type AutoUpdater struct {
	owner          string
	repo           string
	currentVersion string
	httpClient     *http.Client
}

func NewAutoUpdater(owner, repo, currentVer string) *AutoUpdater {
	return &AutoUpdater{
		owner:          owner,
		repo:           repo,
		currentVersion: currentVer,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (u *AutoUpdater) CheckForUpdates(ctx context.Context) (*UpdateCheckResult, error) {
	// 1. Safety Check: If environment variables are missing, silently disable updates.
	if u.owner == "" || u.repo == "" {
		return &UpdateCheckResult{HasUpdate: false}, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.owner, u.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create req: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	// 2. Safety Check: If no releases exist yet, GitHub returns 404. Handle gracefully.
	if resp.StatusCode == http.StatusNotFound {
		return &UpdateCheckResult{HasUpdate: false}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	if release.TagName == u.currentVersion {
		return &UpdateCheckResult{HasUpdate: false}, nil
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if runtime.GOOS == "windows" && filepath.Ext(asset.Name) == ".exe" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	return &UpdateCheckResult{
		HasUpdate:    true,
		LatestVer:    release.TagName,
		ReleaseNotes: release.Body,
		DownloadURL:  downloadURL,
	}, nil
}

func (u *AutoUpdater) ApplyUpdate(ctx context.Context, downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("no executable asset download URL provided")
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("create download req: %w", err)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download asset failed status %d", resp.StatusCode)
	}

	oldExePath := exePath + ".old"
	_ = os.Remove(oldExePath)

	if err := os.Rename(exePath, oldExePath); err != nil {
		return fmt.Errorf("rename running binary: %w", err)
	}

	newFile, err := os.OpenFile(exePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		_ = os.Rename(oldExePath, exePath)
		return fmt.Errorf("create new binary file: %w", err)
	}

	_, err = io.Copy(newFile, resp.Body)
	_ = newFile.Close()
	if err != nil {
		_ = os.Remove(exePath)
		_ = os.Rename(oldExePath, exePath)
		return fmt.Errorf("write new binary: %w", err)
	}

	return nil
}

func (u *AutoUpdater) RestartApp() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	cmd := exec.Command(exePath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("restart binary: %w", err)
	}
	os.Exit(0)
	return nil
}

```

---

## FILE: views\clients.templ
```templ
package views

import (
    "fmt"
    "encoding/json"
)

func getClientJSON(c Client) string {
    b, _ := json.Marshal(c)
    return string(b)
}

templ Clients(clients []Client) {
	@Layout("LuminaLex - Clients", true) {
		<div class="z-10 flex-1 flex flex-col h-full overflow-hidden bg-black/60 backdrop-blur-sm animate-fade-in-up">
			
            <div id="clients-table-view" class="flex flex-col h-full p-8">
                <div class="flex justify-between items-center border-b border-[#a68a56]/50 mb-6 pb-4 shrink-0">
                    <div class="flex items-center gap-4 flex-1">
                        <input type="text" id="client-search" placeholder="SEARCH CLIENTS BY NAME, ID, OR REGISTRATION..." class="w-1/2 bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] transition-colors uppercase tracking-widest rounded"/>
                    </div>
                    <div class="flex gap-4">
                        <button onclick="exportTable()" class="px-5 py-2.5 text-xs uppercase tracking-widest font-semibold text-[#f9f8f6] bg-black/40 border border-[#a68a56]/30 hover:bg-[#a68a56]/20 rounded transition-colors flex items-center gap-2">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
                            Export View
                        </button>
                        <button onclick="openClientForm()" class="bg-[#a68a56] border border-[#a68a56] text-[#0b1d2e] hover:bg-[#0b1d2e] hover:text-[#a68a56] px-5 py-2.5 text-xs uppercase tracking-widest transition-all font-bold rounded shadow-[0_0_15px_rgba(166,138,86,0.3)] flex items-center gap-2">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
                            New Client
                        </button>
                    </div>
                </div>

                <div class="w-full flex-1 overflow-auto border border-[#a68a56]/30 bg-black/20 rounded shadow-lg">
                    <table class="w-full text-left border-collapse table-auto" id="main-clients-table">
                        <thead class="sticky top-0 bg-[#0b1d2e] z-10 shadow-md">
                            <tr class="border-b-2 border-[#a68a56]/50 text-[#a68a56] text-[10px] uppercase tracking-widest">
                                <th class="p-4">Client Name</th>
                                <th class="p-4">ID / Reg No</th>
                                <th class="p-4">Jurisdiction</th>
                                <th class="p-4">Occupation</th>
                                <th class="p-4">Practice No</th>
                                <th class="p-4 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-[#a68a56]/20 text-xs">
                            if len(clients) == 0 {
                                <tr><td colspan="6" class="p-8 text-center text-[#f9f8f6]/50 uppercase tracking-widest">No client records found.</td></tr>
                            } else {
                                for _, c := range clients {
                                    <tr class="client-row hover:bg-[#a68a56]/10 transition-colors text-[#f9f8f6] cursor-pointer" 
                                        data-client={ getClientJSON(c) } 
                                        onclick="viewClientFromRow(this)" 
                                        data-search={ fmt.Sprintf("%s %s %s %s %s", c.FirstName, c.LastName, c.IDNumber, c.RegistrationNumber, c.PracticeNumber) }>
                                        <td class="p-4 font-bold">{ c.LastName }, { c.FirstName } { c.MiddleName }</td>
                                        <td class="p-4">{ c.IDNumber }{ c.RegistrationNumber }</td>
                                        <td class="p-4 uppercase text-[#a68a56] font-semibold">{ c.JurisdictionType }</td>
                                        <td class="p-4">{ c.Occupation }</td>
                                        <td class="p-4">{ c.PracticeNumber }</td>
                                        <td class="p-4 text-right">
                                            <span class="text-[#a68a56] hover:text-[#f9f8f6] transition-colors uppercase text-[10px] font-bold tracking-widest flex items-center justify-end gap-1">
                                                View Profile <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
                                            </span>
                                        </td>
                                    </tr>
                                }
                            }
                        </tbody>
                    </table>
                </div>
            </div>

            <div id="clients-detail-view" class="hidden flex-col h-full bg-[#0b1d2e]">
                <div class="flex items-center justify-between border-b border-[#a68a56]/30 p-6 bg-black/40 shrink-0">
                    <div class="flex items-center gap-5">
                        <button onclick="closeViews()" class="w-9 h-9 flex items-center justify-center rounded-full bg-[#f9f8f6]/5 border border-[#f9f8f6]/10 hover:bg-[#f9f8f6]/10 text-[#f9f8f6] transition-colors">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path></svg>
                        </button>
                        
                        <div class="w-14 h-14 rounded-full border-2 border-[#a68a56] text-[#a68a56] flex items-center justify-center text-xl font-bold shadow-[0_0_10px_rgba(166,138,86,0.2)]" id="det-initials"></div>
                        
                        <div>
                            <h3 id="det-header-name" class="text-[#f9f8f6] text-xl font-bold uppercase tracking-widest"></h3>
                            <div class="flex items-center gap-3 mt-1 text-xs text-[#f9f8f6]/70">
                                <span class="bg-emerald-500/20 text-emerald-400 px-2 py-0.5 rounded text-[10px] uppercase font-bold tracking-widest border border-emerald-500/30 shadow-inner">Active Client</span>
                                <span id="det-header-id" class="tracking-widest"></span>
                                <span>•</span>
                                <span id="det-header-added" class="tracking-widest"></span>
                            </div>
                        </div>
                    </div>
                    <div class="flex gap-4">
                        <button onclick="exportProfile()" class="px-5 py-2.5 text-xs uppercase tracking-widest font-semibold text-[#f9f8f6] bg-[#f9f8f6]/10 border border-[#f9f8f6]/10 hover:bg-[#f9f8f6]/20 rounded transition-colors flex items-center gap-2">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
                            Export
                        </button>
                        <button class="px-5 py-2.5 text-xs uppercase tracking-widest font-bold text-[#0b1d2e] bg-[#a68a56] hover:bg-[#d4c299] rounded shadow-lg transition-colors flex items-center gap-2">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
                            New Matter
                        </button>
                    </div>
                </div>

                <div class="flex-1 overflow-y-auto p-6 flex gap-6">
                    <div class="flex-1 flex flex-col gap-6">
                        
                        <div class="bg-black/30 border border-[#a68a56]/20 rounded-xl p-6 shadow-inner relative group">
                            <button onclick="openModal('personal')" class="absolute top-6 right-6 text-[#f9f8f6]/30 hover:text-[#a68a56] transition-colors">
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"></path></svg>
                            </button>
                            <h4 class="text-[11px] uppercase font-bold text-[#f9f8f6] tracking-widest mb-6">Personal Identification</h4>
                            <div class="grid grid-cols-2 gap-y-6 gap-x-4">
                                <div class="flex items-start gap-3">
                                    <svg class="w-4 h-4 text-[#a68a56] mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3"></path></svg>
                                    <div><span class="block text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest mb-1">Jurisdiction</span><span id="det-jur" class="text-sm text-[#f9f8f6] font-medium">-</span></div>
                                </div>
                                <div class="flex items-start gap-3">
                                    <svg class="w-4 h-4 text-[#a68a56] mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0a2 2 0 104 0m-5 8a2 2 0 100-4 2 2 0 000 4zm0 0c1.306 0 2.417.835 2.83 2M9 14a3.001 3.001 0 00-2.83 2M15 11h3m-3 4h2"></path></svg>
                                    <div><span class="block text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest mb-1">Registration Number</span><span id="det-reg" class="text-sm text-[#f9f8f6] font-medium">-</span></div>
                                </div>
                                <div class="flex items-start gap-3">
                                    <svg class="w-4 h-4 text-[#a68a56] mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
                                    <div><span class="block text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest mb-1">Marital Status</span><span id="det-marital" class="text-sm text-[#f9f8f6] font-medium">-</span></div>
                                </div>
                                <div class="flex items-start gap-3">
                                    <svg class="w-4 h-4 text-[#a68a56] mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
                                    <div><span class="block text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest mb-1">Practice Number</span><span id="det-prac" class="text-sm text-[#f9f8f6] font-medium">-</span></div>
                                </div>
                            </div>
                        </div>

                        <div class="bg-black/30 border border-[#a68a56]/20 rounded-xl p-6 shadow-inner relative group">
                            <button onclick="openModal('employment')" class="absolute top-6 right-6 text-[#f9f8f6]/30 hover:text-[#a68a56] transition-colors">
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"></path></svg>
                            </button>
                            <h4 class="text-[11px] uppercase font-bold text-[#f9f8f6] tracking-widest mb-6">Employment Details</h4>
                            <div class="grid grid-cols-2 gap-y-6 gap-x-4">
                                <div class="flex items-start gap-3">
                                    <svg class="w-4 h-4 text-[#a68a56] mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
                                    <div><span class="block text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest mb-1">Occupation</span><span id="det-occ" class="text-sm text-[#f9f8f6] font-medium">-</span></div>
                                </div>
                                <div class="flex items-start gap-3">
                                    <svg class="w-4 h-4 text-[#a68a56] mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path></svg>
                                    <div><span class="block text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest mb-1">Employer Name</span><span id="det-emp-name" class="text-sm text-[#f9f8f6] font-medium">-</span></div>
                                </div>
                                <div class="flex items-start gap-3">
                                    <svg class="w-4 h-4 text-[#a68a56] mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path></svg>
                                    <div><span class="block text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest mb-1">Work Contact</span><span id="det-emp-num" class="text-sm text-[#f9f8f6] font-medium cursor-pointer hover:text-[#a68a56] transition-colors" onclick="copyValue(this.innerText)">-</span></div>
                                </div>
                                <div class="flex items-start gap-3">
                                    <svg class="w-4 h-4 text-[#a68a56] mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
                                    <div><span class="block text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest mb-1">Work Address</span><span id="det-emp-phys" class="text-sm text-[#f9f8f6] font-medium">-</span></div>
                                </div>
                            </div>
                        </div>

                        <div class="bg-black/30 border border-[#a68a56]/20 rounded-xl p-6 shadow-inner flex-1 flex flex-col relative">
                            <div class="flex justify-between items-center mb-6">
                                <h4 class="text-[11px] uppercase font-bold text-[#f9f8f6] tracking-widest">Active Matters / Cases</h4>
                                <button class="px-3 py-1 bg-[#f9f8f6]/5 border border-[#f9f8f6]/10 hover:bg-[#f9f8f6]/10 rounded text-[10px] uppercase tracking-widest text-[#f9f8f6] transition-colors">View All</button>
                            </div>
                            <div class="w-full flex-1 border-t border-[#a68a56]/20 flex flex-col">
    <div class="grid grid-cols-4 text-[#f9f8f6]/50 text-[9px] uppercase tracking-widest p-3 border-b border-[#a68a56]/10">
        <div>File Ref</div>
        <div>Type</div>
        <div>Status</div>
        <div>Opened Date</div>
    </div>
    <div id="det-matters-container" class="flex-1 flex flex-col overflow-y-auto">
        <div class="flex-1 flex flex-col items-center justify-center gap-3 text-[#f9f8f6]/30 p-12">
            <svg class="w-8 h-8 opacity-70" fill="currentColor" viewBox="0 0 20 20"><path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"></path></svg>
            <span class="text-xs font-semibold tracking-wider">No Active Matters</span>
        </div>
    </div>
</div>
                            </div>
                        </div>

                    </div>

                    <div class="w-[35%] flex flex-col gap-6">
                        
                        <div class="bg-black/30 border border-[#a68a56]/20 rounded-xl p-6 shadow-inner relative group">
                            <button onclick="openModal('contacts')" class="absolute top-6 right-6 text-[#f9f8f6]/30 hover:text-[#a68a56] transition-colors">
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"></path></svg>
                            </button>
                            <h4 class="text-[11px] uppercase font-bold text-[#f9f8f6] tracking-widest mb-6">Contact Info</h4>
                            <div id="det-contacts" class="flex flex-col gap-5 text-[#f9f8f6]"></div>
                        </div>

                        <div class="bg-black/30 border border-[#a68a56]/20 rounded-xl p-6 shadow-inner relative group">
                            <button onclick="openModal('addresses')" class="absolute top-6 right-6 text-[#f9f8f6]/30 hover:text-[#a68a56] transition-colors">
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"></path></svg>
                            </button>
                            <h4 class="text-[11px] uppercase font-bold text-[#f9f8f6] tracking-widest mb-6">Addresses</h4>
                            <div id="det-addresses" class="flex flex-col gap-5 text-[#f9f8f6]"></div>
                        </div>

                        <div class="bg-black/30 border border-[#a68a56]/20 rounded-xl p-6 shadow-inner relative group">
                            <button onclick="openModal('banks')" class="absolute top-6 right-6 text-[#f9f8f6]/30 hover:text-[#a68a56] transition-colors">
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"></path></svg>
                            </button>
                            <h4 class="text-[11px] uppercase font-bold text-[#f9f8f6] tracking-widest mb-6">Banking Info</h4>
                            <div id="det-banks" class="flex flex-col gap-4 text-[#f9f8f6]"></div>
                        </div>

                    </div>
                </div>
            </div>

            <div id="clients-form-view" class="hidden flex-col h-full bg-[#0b1d2e]">
                <div class="flex items-center justify-between border-b border-[#a68a56]/30 p-6 bg-black/40 shrink-0">
                    <div class="flex items-center gap-4">
                        <button onclick="closeClientForm()" class="text-[#a68a56] hover:text-[#f9f8f6] transition-colors">
                            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg>
                        </button>
                        <h3 class="text-[#a68a56] text-lg font-bold uppercase tracking-widest">Capture Client Details</h3>
                    </div>
                    <div class="flex gap-4">
                        <button onclick="saveClient()" class="px-6 py-2.5 text-xs uppercase tracking-widest font-bold text-[#0b1d2e] bg-[#a68a56] hover:bg-[#d4c299] rounded shadow-lg transition-colors">Save Client Record</button>
                    </div>
                </div>

                <div class="flex-1 flex overflow-hidden">
                    <input type="hidden" id="c-id" value=""/>

                    <div class="w-[70%] h-full p-8 overflow-y-auto no-scrollbar flex flex-col gap-8">
                        <div>
                            <h4 class="text-xs uppercase font-bold text-[#a68a56] tracking-widest border-b border-[#a68a56]/20 pb-2 mb-6">Primary Identification</h4>
                            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                
                                <div class="relative">
                                    <input type="text" id="c-first-name" placeholder="First Name *" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-first-name" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">First Name *</label>
                                </div>
                                <div class="relative">
                                    <input type="text" id="c-middle-name" placeholder="Middle Name" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-middle-name" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Middle Name</label>
                                </div>
                                <div class="relative">
                                    <input type="text" id="c-last-name" placeholder="Last Name *" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-last-name" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Last Name *</label>
                                </div>
                                <div class="relative">
                                    <input type="text" id="c-id-number" maxlength="13" oninput="this.value = this.value.replace(/[^0-9]/g, '').slice(0, 13)" placeholder="ID Number" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-id-number" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">ID Number (13 Digits)</label>
                                </div>

                                <div class="relative">
                                    <select id="c-jurisdiction" onchange="toggleJurisdictionOther(this)" class="block w-full bg-[#0b1d2e] border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded">
                                        <option value="COMPANY" class="bg-[#0b1d2e] text-[#f9f8f6]">COMPANY</option>
                                        <option value="CLOSE CORPORATION" class="bg-[#0b1d2e] text-[#f9f8f6]">CLOSE CORPORATION</option>
                                        <option value="TRUST" class="bg-[#0b1d2e] text-[#f9f8f6]">TRUST</option>
                                        <option value="OTHER" class="bg-[#0b1d2e] text-[#f9f8f6]">OTHER</option>
                                    </select>
                                    <label class="absolute left-3 top-1 text-[#a68a56] text-[9px] uppercase tracking-widest pointer-events-none">Jurisdiction Type</label>
                                </div>
                                <div class="relative hidden" id="jur-other-wrap">
                                    <input type="text" id="c-jurisdiction-other" placeholder="Specify Jurisdiction" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-jurisdiction-other" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Specify Other</label>
                                </div>

                                <div class="relative">
                                    <input type="text" id="c-reg-number" placeholder="Registration Number" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-reg-number" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Registration Number</label>
                                </div>
                                <div class="relative">
                                    <input type="text" id="c-occupation" placeholder="Occupation" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-occupation" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Occupation</label>
                                </div>
                                
                                <div class="relative">
                                    <select id="c-marital-status" class="block w-full bg-[#0b1d2e] border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded">
                                        <option value="Single" class="bg-[#0b1d2e] text-[#f9f8f6]">Single</option>
                                        <option value="Married (In community)" class="bg-[#0b1d2e] text-[#f9f8f6]">Married (In community)</option>
                                        <option value="Married (Out community)" class="bg-[#0b1d2e] text-[#f9f8f6]">Married (Out community)</option>
                                        <option value="Divorced" class="bg-[#0b1d2e] text-[#f9f8f6]">Divorced</option>
                                        <option value="Widowed" class="bg-[#0b1d2e] text-[#f9f8f6]">Widowed</option>
                                    </select>
                                    <label class="absolute left-3 top-1 text-[#a68a56] text-[9px] uppercase tracking-widest pointer-events-none">Marital Status</label>
                                </div>
                                
                                <div class="relative">
                                    <input type="text" id="c-practice-number" placeholder="Practice Number" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-practice-number" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Practice Number</label>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h4 class="text-xs uppercase font-bold text-[#a68a56] tracking-widest border-b border-[#a68a56]/20 pb-2 mb-6 mt-4">Employment Information</h4>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div class="relative">
                                    <input type="text" id="c-emp-name" placeholder="Employer Name" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-emp-name" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Employer Name</label>
                                </div>
                                <div class="relative">
                                    <input type="text" id="c-emp-number" maxlength="14" oninput="this.value = formatPhoneNumber(this.value)" placeholder="Phone" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                    <label for="c-emp-number" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Employer Phone Number</label>
                                </div>
                            </div>

                            <h5 class="text-[10px] uppercase font-bold text-[#a68a56] tracking-widest mt-6 mb-3">Employer Physical Address</h5>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div class="relative w-full">
                                    <input type="text" id="c-emp-l1" oninput="autocompleteAddress(this, this.parentElement, 'c-emp')" placeholder="Address Line 1" class="w-full bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                </div>
                                <input type="text" id="c-emp-l2" placeholder="Address Line 2" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-suburb" placeholder="Suburb" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-city" placeholder="City" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-code" placeholder="Postal Code" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-country" placeholder="Country" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                            </div>

                            <div class="mt-4 mb-4 flex items-center gap-2">
                                <input type="checkbox" id="c-emp-postal-same" checked onchange="toggleEmployerPostal(this)" class="accent-[#a68a56] w-4 h-4"/>
                                <label for="c-emp-postal-same" class="text-xs text-[#f9f8f6] tracking-wider cursor-pointer">Employer Postal Address is the same as Physical</label>
                            </div>

                            <div id="employer-postal-container" class="hidden grid-cols-1 md:grid-cols-2 gap-4 mt-2">
                                <input type="text" id="c-emp-pl1" placeholder="Postal Line 1" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-pl2" placeholder="Postal Line 2" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-psuburb" placeholder="Suburb" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-pcity" placeholder="City" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-pcode" placeholder="Postal Code" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                                <input type="text" id="c-emp-pcountry" placeholder="Country" class="bg-black/40 border border-[#a68a56]/30 p-3 text-xs outline-none rounded text-[#f9f8f6]"/>
                            </div>
                        </div>
                    </div>

                    <div class="w-[30%] h-full p-6 overflow-y-auto no-scrollbar flex flex-col gap-8 bg-black/20 border-l border-[#a68a56]/20 shadow-inner">
                        <div>
                            <div class="flex justify-between items-center border-b border-[#a68a56]/20 pb-2 mb-4">
                                <h4 class="text-[11px] uppercase font-bold text-[#a68a56] tracking-widest">Client Addresses</h4>
                                <button type="button" onclick="addAddressRow({}, 'addresses-container')" class="text-[9px] text-[#a68a56] border border-[#a68a56]/40 px-2 py-1 uppercase hover:bg-[#a68a56]/20 rounded transition-colors">+ Add</button>
                            </div>
                            <div id="addresses-container" class="flex flex-col gap-4"></div>
                        </div>
                        <div>
                            <div class="flex justify-between items-center border-b border-[#a68a56]/20 pb-2 mb-4">
                                <h4 class="text-[11px] uppercase font-bold text-[#a68a56] tracking-widest">Contact Details</h4>
                                <div class="relative">
                                    <button type="button" onclick="document.getElementById('add-contact-dropdown').classList.toggle('hidden')" class="text-[9px] text-[#a68a56] border border-[#a68a56]/40 px-2 py-1 uppercase hover:bg-[#a68a56]/20 rounded transition-colors">+ Add</button>
                                    <div id="add-contact-dropdown" class="hidden absolute right-0 top-6 bg-[#0b1d2e] border border-[#a68a56]/50 rounded shadow-lg flex flex-col z-20 w-28">
                                        <button type="button" onclick="addContactType('EMAIL', 'contacts-container')" class="text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors">EMAIL</button>
                                        <button type="button" onclick="addContactType('CELLPHONE', 'contacts-container')" class="text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors">CELLPHONE</button>
                                        <button type="button" onclick="addContactType('TELEPHONE', 'contacts-container')" class="text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors">TELEPHONE</button>
                                        <button type="button" onclick="addContactType('FAX', 'contacts-container')" class="text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors">FAX</button>
                                    </div>
                                </div>
                            </div>
                            <div id="contacts-container" class="flex flex-col gap-4"></div>
                        </div>
                        <div>
                            <div class="flex justify-between items-center border-b border-[#a68a56]/20 pb-2 mb-4">
                                <h4 class="text-[11px] uppercase font-bold text-[#a68a56] tracking-widest">Banking Details</h4>
                                <button type="button" onclick="addBankRow({}, 'banks-container')" class="text-[9px] text-[#a68a56] border border-[#a68a56]/40 px-2 py-1 uppercase hover:bg-[#a68a56]/20 rounded transition-colors">+ Add</button>
                            </div>
                            <div id="banks-container" class="flex flex-col gap-4"></div>
                        </div>
                    </div>
                </div>
            </div>

            <div id="edit-modal" class="hidden fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-4 backdrop-blur-sm animate-fade-in-up">
                <div class="bg-[#0b1d2e] border border-[#a68a56]/50 rounded-xl w-full max-w-3xl max-h-[90vh] flex flex-col shadow-[0_0_30px_rgba(166,138,86,0.1)]">
                    <div class="p-6 border-b border-[#a68a56]/30 flex justify-between items-center bg-black/40 rounded-t-xl shrink-0">
                        <h3 id="edit-modal-title" class="text-[#a68a56] text-lg font-bold uppercase tracking-widest"></h3>
                        <button onclick="closeModal()" class="text-[#f9f8f6]/50 hover:text-red-400 transition-colors">
                            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
                        </button>
                    </div>
                    <div id="edit-modal-body" class="p-8 overflow-y-auto flex-1 flex flex-col gap-6">
                    </div>
                    <div class="p-6 border-t border-[#a68a56]/30 flex justify-end gap-4 bg-black/40 rounded-b-xl shrink-0">
                        <button onclick="closeModal()" class="px-5 py-2.5 text-xs uppercase tracking-widest text-red-400 border border-red-500/30 hover:bg-red-500/10 rounded transition-colors">Cancel</button>
                        <button onclick="saveModalEdits()" class="px-6 py-2.5 text-xs uppercase tracking-widest font-bold text-[#0b1d2e] bg-[#a68a56] hover:bg-[#d4c299] rounded shadow-lg transition-colors">Save Changes</button>
                    </div>
                </div>
            </div>

            <div id="toast" class="fixed bottom-8 right-8 bg-[#a68a56] text-[#0b1d2e] px-4 py-3 rounded shadow-[0_0_15px_rgba(166,138,86,0.5)] transition-opacity duration-300 opacity-0 pointer-events-none z-50 font-bold uppercase tracking-widest text-[10px] flex items-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                <span id="toast-msg">Copied to clipboard</span>
            </div>

		</div>

		<script>
            let currentViewedClient = null;
            let activeModalSection = null;
            let addressTimeout;

            const BANK_CODES = {
                'Absa Bank': '632005', 'Capitec Bank': '470010', 'First National Bank (FNB)': '250655',
                'Nedbank': '198765', 'Standard Bank': '051001', 'Investec Bank': '580105',
                'TYME Bank': '678910', 'Discovery Bank': '679000', 'African Bank': '430000', 'Other': ''
            };

            function formatPhoneNumber(value) {
                if (!value) return value;
                const phoneNumber = value.replace(/[^\d]/g, '');
                const len = phoneNumber.length;
                if (len < 4) return phoneNumber;
                if (len < 7) return `(${phoneNumber.slice(0, 3)}) ${phoneNumber.slice(3)}`;
                return `(${phoneNumber.slice(0, 3)}) ${phoneNumber.slice(3, 6)} ${phoneNumber.slice(6, 10)}`;
            }

            function getContactIconSVG(type) {
                if(type === 'EMAIL') return `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>`;
                if(type === 'CELLPHONE') return `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>`;
                return `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path></svg>`;
            }

            function getAddressIconSVG(type) {
                if(type === 'HOME') return `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path></svg>`;
                return `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>`;
            }

            function showToast(msg = "Copied to clipboard") {
                const toast = document.getElementById('toast');
                document.getElementById('toast-msg').innerText = msg;
                toast.classList.remove('opacity-0');
                toast.classList.add('opacity-100');
                setTimeout(() => {
                    toast.classList.remove('opacity-100');
                    toast.classList.add('opacity-0');
                }, 3000);
            }

            function copyValue(val) {
                if(!val || val === '-') return;
                navigator.clipboard.writeText(val).then(() => showToast());
            }

            function closeViews() {
                document.getElementById('clients-detail-view').classList.add('hidden');
                document.getElementById('clients-detail-view').classList.remove('flex');
                document.getElementById('clients-form-view').classList.add('hidden');
                document.getElementById('clients-form-view').classList.remove('flex');
                document.getElementById('clients-table-view').classList.remove('hidden');
                currentViewedClient = null;
            }

			document.getElementById('client-search')?.addEventListener('input', (e) => {
				const term = e.target.value.toLowerCase();
				document.querySelectorAll('.client-row').forEach(row => {
					const text = row.getAttribute('data-search').toLowerCase();
					row.style.display = text.includes(term) ? '' : 'none';
				});
			});

            function viewClientFromRow(rowElement) {
                const clientData = JSON.parse(rowElement.getAttribute('data-client'));
                renderClientProfile(clientData);
                
                document.getElementById('clients-table-view').classList.add('hidden');
                document.getElementById('clients-form-view').classList.add('hidden');
                document.getElementById('clients-detail-view').classList.remove('hidden');
                document.getElementById('clients-detail-view').classList.add('flex');
            }

            function renderClientProfile(clientData) {
                currentViewedClient = clientData;
                
                const initials = (clientData.first_name?.[0] || '') + (clientData.last_name?.[0] || '');
                document.getElementById('det-initials').innerText = initials.toUpperCase();
                
                const fullName = `${clientData.first_name} ${clientData.middle_name || ''} ${clientData.last_name}`.trim();
                document.getElementById('det-header-name').innerText = fullName;
                document.getElementById('det-header-id').innerText = `ID: ${clientData.id_number || 'N/A'}`;
                
                const addedDate = clientData.updated_at ? new Date(clientData.updated_at) : new Date();
                document.getElementById('det-header-added').innerText = `Added ` + addedDate.toLocaleString('en-US', { month: 'short', year: 'numeric' });

                document.getElementById('det-jur').innerText = clientData.jurisdiction_type || '-';
                document.getElementById('det-reg').innerText = clientData.registration_number || '-';
                document.getElementById('det-marital').innerText = clientData.marital_status || '-';
                document.getElementById('det-prac').innerText = clientData.practice_number || '-';
                
                document.getElementById('det-occ').innerText = clientData.occupation || '-';
                document.getElementById('det-emp-name').innerText = clientData.employer_name || '-';
                document.getElementById('det-emp-num').innerText = formatPhoneNumber(clientData.employer_number) || '-';
                
                const empPhys = [clientData.employer_address_l1, clientData.employer_address_l2, clientData.employer_suburb, clientData.employer_city, clientData.employer_postal_code, clientData.employer_country].filter(Boolean).join(', ');
                document.getElementById('det-emp-phys').innerHTML = empPhys ? `<a href="https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(empPhys)}" target="_blank" class="hover:text-[#a68a56] transition-colors underline decoration-white/20 underline-offset-2">${empPhys}</a>` : '-';

                let contactHTML = '';
                const contactsByType = {};
                (clientData.contact_details || []).forEach(c => {
                    if(!contactsByType[c.contact_type]) contactsByType[c.contact_type] = [];
                    contactsByType[c.contact_type].push(c);
                });

                for(const type in contactsByType) {
                    contactHTML += `<div class="mb-4 last:mb-0"><span class="text-[10px] text-[#f9f8f6]/50 uppercase tracking-widest font-bold mb-2 block border-b border-[#a68a56]/20 pb-1">${type}</span>`;
                    contactsByType[type].forEach(c => {
                        contactHTML += `
                            <div class="flex items-start gap-4 mb-3 last:mb-0">
                                <button onclick="copyValue('${c.contact_value}')" class="w-8 h-8 rounded bg-[#f9f8f6]/5 border border-[#f9f8f6]/10 text-[#f9f8f6] flex items-center justify-center shrink-0 hover:bg-[#a68a56] transition-colors cursor-pointer" title="Copy to clipboard">
                                    ${getContactIconSVG(c.contact_type)}
                                </button>
                                <div class="flex-1 overflow-hidden mt-1">
                                    <div class="flex items-center gap-2">
                                        <span class="text-sm font-medium truncate">${c.contact_type !== 'EMAIL' ? formatPhoneNumber(c.contact_value) : c.contact_value}</span>
                                        ${c.is_primary ? '<span class="text-[8px] text-[#a68a56] border border-[#a68a56]/50 rounded px-1 tracking-wider uppercase">PRI</span>' : ''}
                                    </div>
                                </div>
                            </div>
                        `;
                    });
                    contactHTML += `</div>`;
                }
                document.getElementById('det-contacts').innerHTML = contactHTML || '<span class="italic text-[#f9f8f6]/30 text-xs">No contacts saved.</span>';

                let addressHTML = '';
                (clientData.addresses || []).forEach(a => {
                    const addrStr = [a.line1, a.line2, a.suburb, a.city, a.postal_code, a.country].filter(Boolean).join(', ');
                    addressHTML += `
                        <div class="flex items-start gap-4">
                            <button onclick="copyValue('${addrStr}')" class="w-8 h-8 rounded bg-[#f9f8f6]/5 border border-[#f9f8f6]/10 text-[#f9f8f6] flex items-center justify-center shrink-0 hover:bg-[#a68a56] transition-colors cursor-pointer" title="Copy to clipboard">
                                ${getAddressIconSVG(a.address_type)}
                            </button>
                            <div class="flex-1">
                                <div class="flex items-center gap-2 mb-0.5">
                                    <span class="text-[9px] text-[#f9f8f6]/50 uppercase tracking-widest">${a.address_type || 'ADDRESS'}</span>
                                    ${a.is_primary ? '<span class="text-[8px] text-[#a68a56] border border-[#a68a56]/50 rounded px-1 tracking-wider uppercase">PRI</span>' : ''}
                                </div>
                                <div class="text-sm font-medium leading-tight">
                                    <a href="https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(addrStr)}" target="_blank" class="hover:text-[#a68a56] transition-colors">${addrStr}</a>
                                </div>
                            </div>
                        </div>
                    `;
                });
                document.getElementById('det-addresses').innerHTML = addressHTML || '<span class="italic text-[#f9f8f6]/30 text-xs">No addresses saved.</span>';

                document.getElementById('det-banks').innerHTML = (clientData.banks || []).map(b => `
                    <div class="flex items-center gap-4 bg-black/20 border border-[#a68a56]/20 rounded-lg p-3 relative">
                        ${b.is_primary ? '<div class="absolute top-2 right-2 text-[8px] text-[#a68a56] border border-[#a68a56]/50 rounded px-1 tracking-wider uppercase">PRI</div>' : ''}
                        <button onclick="copyValue('${b.account_number}')" class="w-8 h-8 rounded bg-blue-500/20 text-blue-400 flex items-center justify-center shrink-0 hover:bg-blue-500 hover:text-white transition-colors cursor-pointer" title="Copy Account Number">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 14v3m4-3v3m4-3v3M3 21h18M3 10h18M3 7l9-4 9 4M4 10h16v11H4V10z"></path></svg>
                        </button>
                        <div>
                            <div class="text-sm font-bold">${b.bank_name}</div>
                            <div class="text-[10px] text-[#f9f8f6]/60 uppercase tracking-widest mt-0.5">${b.account_type} • ${b.account_number}</div>
                        </div>
                    </div>
                `).join('') || '<span class="italic text-[#f9f8f6]/30 text-xs">No banking details saved.</span>';
                loadClientMatters(clientData.id);
            }

            async function exportTable() {
                const table = document.getElementById('main-clients-table');
                const headerCells = Array.from(table.querySelectorAll('thead th'));
                const headers = headerCells.slice(0, headerCells.length - 1).map(th => th.innerText.trim());

                const rows = [];
                table.querySelectorAll('tbody tr').forEach(tr => {
                    if (tr.style.display !== 'none' && !tr.querySelector('td[colspan]')) {
                        const cells = Array.from(tr.querySelectorAll('td'));
                        rows.push(cells.slice(0, cells.length - 1).map(td => td.innerText.trim().replace(/\n/g, ' ')));
                    }
                });

                if (rows.length === 0) return alert("No visible rows to export.");

                try {
                    const res = await fetch('/api/contacts/export_view', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ filename: 'CLIENTS_LIST', headers, rows })
                    });
                    if (!res.ok) alert("Export failed");
                } catch (e) { alert("Export failed"); }
            }

            async function exportProfile() {
                if(!currentViewedClient) return;
                const c = currentViewedClient;
                const headers = ["Section", "Field", "Value"];
                const rows = [
                    ["Personal", "Name", `${c.first_name} ${c.middle_name || ''} ${c.last_name}`],
                    ["Personal", "ID Number", c.id_number || '-'],
                    ["Personal", "Jurisdiction", c.jurisdiction_type || '-'],
                    ["Personal", "Reg Number", c.registration_number || '-'],
                    ["Personal", "Marital Status", c.marital_status || '-'],
                    ["Personal", "Practice Number", c.practice_number || '-'],
                    ["Employment", "Occupation", c.occupation || '-'],
                    ["Employment", "Employer", c.employer_name || '-'],
                    ["Employment", "Work Contact", c.employer_number || '-'],
                ];
                
                (c.contact_details || []).forEach(cd => {
                    rows.push(["Contact", cd.contact_type, cd.contact_value]);
                });
                (c.addresses || []).forEach(a => {
                    const addr = [a.line1, a.line2, a.suburb, a.city, a.postal_code, a.country].filter(Boolean).join(', ');
                    rows.push(["Address", a.address_type, addr]);
                });
                (c.banks || []).forEach(b => {
                    rows.push(["Banking", b.bank_name, `${b.account_type} - ${b.account_number} (${b.branch_code})`]);
                });

                try {
                    const res = await fetch('/api/contacts/export_view', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ filename: `${c.last_name}_PROFILE`.toUpperCase(), headers, rows })
                    });
                    if (!res.ok) alert("Export failed");
                } catch (e) { alert("Export failed"); }
            }

			function openClientForm() {
                document.getElementById('clients-table-view').classList.add('hidden');
                document.getElementById('clients-detail-view').classList.add('hidden');
                document.getElementById('clients-detail-view').classList.remove('flex');
				document.getElementById('clients-form-view').classList.remove('hidden');
                document.getElementById('clients-form-view').classList.add('flex');
                
                document.getElementById('c-id').value = '';
                document.querySelectorAll('#clients-form-view input[type="text"], #clients-form-view input[type="email"]').forEach(i => i.value = '');
                document.getElementById('c-emp-postal-same').checked = true;
                toggleEmployerPostal(document.getElementById('c-emp-postal-same'));
                document.getElementById('c-jurisdiction').value = 'COMPANY';
                toggleJurisdictionOther(document.getElementById('c-jurisdiction'));

                document.getElementById('addresses-container').innerHTML = '';
                document.getElementById('contacts-container').innerHTML = '';
                document.getElementById('banks-container').innerHTML = '';
			}

            function closeClientForm() {
                document.getElementById('clients-form-view').classList.add('hidden');
                document.getElementById('clients-form-view').classList.remove('flex');
                document.getElementById('clients-table-view').classList.remove('hidden');
            }

            function toggleJurisdictionOther(sel) {
                const otherWrap = document.getElementById('jur-other-wrap');
                const otherInput = document.getElementById('c-jurisdiction-other');
                if (sel.value === 'OTHER') {
                    otherWrap.classList.remove('hidden');
                } else {
                    otherWrap.classList.add('hidden');
                    otherInput.value = '';
                }
            }

            function toggleEmployerPostal(chk) {
                const container = document.getElementById('employer-postal-container');
                if (chk.checked) {
                    container.classList.remove('grid');
                    container.classList.add('hidden');
                } else {
                    container.classList.remove('hidden');
                    container.classList.add('grid');
                }
            }

            function openModal(section) {
                if(!currentViewedClient) return;
                activeModalSection = section;
                document.getElementById('edit-modal').classList.remove('hidden');
                const title = document.getElementById('edit-modal-title');
                const body = document.getElementById('edit-modal-body');
                body.innerHTML = '';
                
                const c = currentViewedClient;

                if (section === 'personal') {
                    title.innerText = "Edit Personal Info";
                    body.innerHTML = `
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div class="relative">
                                <input type="text" id="m-first" value="${c.first_name||''}" placeholder="First Name *" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-first" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">First Name *</label>
                            </div>
                            <div class="relative">
                                <input type="text" id="m-middle" value="${c.middle_name||''}" placeholder="Middle Name" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-middle" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Middle Name</label>
                            </div>
                            <div class="relative">
                                <input type="text" id="m-last" value="${c.last_name||''}" placeholder="Last Name *" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-last" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Last Name *</label>
                            </div>
                            <div class="relative">
                                <input type="text" id="m-idnum" value="${c.id_number||''}" maxlength="13" oninput="this.value = this.value.replace(/[^0-9]/g, '').slice(0, 13)" placeholder="ID Number" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-idnum" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">ID Number (13 Digits)</label>
                            </div>
                            <div class="relative">
                                <select id="m-jur" class="block w-full bg-[#0b1d2e] border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded">
                                    <option value="COMPANY" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.jurisdiction_type==='COMPANY'?'selected':''}>COMPANY</option>
                                    <option value="CLOSE CORPORATION" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.jurisdiction_type==='CLOSE CORPORATION'?'selected':''}>CLOSE CORPORATION</option>
                                    <option value="TRUST" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.jurisdiction_type==='TRUST'?'selected':''}>TRUST</option>
                                    <option value="OTHER" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.jurisdiction_type==='OTHER'?'selected':''}>OTHER</option>
                                </select>
                                <label class="absolute left-3 top-1 text-[#a68a56] text-[9px] uppercase tracking-widest pointer-events-none">Jurisdiction</label>
                            </div>
                            <div class="relative">
                                <input type="text" id="m-reg" value="${c.registration_number||''}" placeholder="Registration" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-reg" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Reg Number</label>
                            </div>
                            <div class="relative">
                                <select id="m-mar" class="block w-full bg-[#0b1d2e] border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded">
                                    <option value="Single" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.marital_status==='Single'?'selected':''}>Single</option>
                                    <option value="Married (In community)" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.marital_status==='Married (In community)'?'selected':''}>Married (In community)</option>
                                    <option value="Married (Out community)" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.marital_status==='Married (Out community)'?'selected':''}>Married (Out community)</option>
                                    <option value="Divorced" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.marital_status==='Divorced'?'selected':''}>Divorced</option>
                                    <option value="Widowed" class="bg-[#0b1d2e] text-[#f9f8f6]" ${c.marital_status==='Widowed'?'selected':''}>Widowed</option>
                                </select>
                                <label class="absolute left-3 top-1 text-[#a68a56] text-[9px] uppercase tracking-widest pointer-events-none">Marital Status</label>
                            </div>
                            <div class="relative">
                                <input type="text" id="m-prac" value="${c.practice_number||''}" placeholder="Practice" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-prac" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Practice Number</label>
                            </div>
                        </div>
                    `;
                }
                else if (section === 'employment') {
                    title.innerText = "Edit Employment Info";
                    body.innerHTML = `
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div class="relative">
                                <input type="text" id="m-occ" value="${c.occupation||''}" placeholder="Occupation" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-occ" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Occupation</label>
                            </div>
                            <div class="relative">
                                <input type="text" id="m-emp-name" value="${c.employer_name||''}" placeholder="Employer Name" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-emp-name" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Employer Name</label>
                            </div>
                            <div class="relative md:col-span-2">
                                <input type="text" id="m-emp-num" value="${formatPhoneNumber(c.employer_number)||''}" oninput="this.value = formatPhoneNumber(this.value)" maxlength="14" placeholder="Phone" class="peer block w-full bg-black/40 border border-[#a68a56]/30 p-3 pt-5 text-xs text-[#f9f8f6] focus:border-[#a68a56] outline-none rounded transition-all placeholder-transparent"/>
                                <label for="m-emp-num" class="absolute left-3 top-1 text-[#a68a56]/60 text-[9px] uppercase tracking-widest transition-all peer-placeholder-shown:top-3.5 peer-placeholder-shown:text-xs peer-focus:top-1 peer-focus:text-[9px] peer-focus:text-[#a68a56] pointer-events-none">Work Contact</label>
                            </div>
                        </div>
                        <h5 class="text-[10px] uppercase font-bold text-[#a68a56] tracking-widest mt-2 border-b border-[#a68a56]/20 pb-1">Physical Address</h5>
                        <div class="grid grid-cols-2 gap-4">
                            <div class="relative w-full col-span-1">
                                <input type="text" id="m-emp-l1" value="${c.employer_address_l1||''}" oninput="autocompleteAddress(this, this.parentElement, 'm-emp')" placeholder="Line 1" class="w-full bg-black/40 border border-[#a68a56]/30 p-2 text-xs rounded text-[#f9f8f6]"/>
                            </div>
                            <input type="text" id="m-emp-l2" value="${c.employer_address_l2||''}" placeholder="Line 2" class="bg-black/40 border border-[#a68a56]/30 p-2 text-xs rounded text-[#f9f8f6]"/>
                            <input type="text" id="m-emp-suburb" value="${c.employer_suburb||''}" placeholder="Suburb" class="bg-black/40 border border-[#a68a56]/30 p-2 text-xs rounded text-[#f9f8f6]"/>
                            <input type="text" id="m-emp-city" value="${c.employer_city||''}" placeholder="City" class="bg-black/40 border border-[#a68a56]/30 p-2 text-xs rounded text-[#f9f8f6]"/>
                            <input type="text" id="m-emp-code" value="${c.employer_postal_code||''}" placeholder="Postal Code" class="bg-black/40 border border-[#a68a56]/30 p-2 text-xs rounded text-[#f9f8f6]"/>
                            <input type="text" id="m-emp-country" value="${c.employer_country||''}" placeholder="Country" class="bg-black/40 border border-[#a68a56]/30 p-2 text-xs rounded text-[#f9f8f6]"/>
                        </div>
                    `;
                }
                else if (section === 'contacts') {
                    title.innerText = "Edit Contacts";
                    body.innerHTML = `
                        <div class="flex justify-between items-center mb-2">
                            <span class="text-xs text-[#f9f8f6]/50 uppercase tracking-widest">Manage Contact Methods</span>
                            <div class="relative">
                                <button type="button" onclick="document.getElementById('edit-add-contact-dropdown').classList.toggle('hidden')" class="text-[9px] text-[#a68a56] border border-[#a68a56]/40 px-3 py-1 uppercase hover:bg-[#a68a56]/20 rounded transition-colors">+ Add Contact</button>
                                <div id="edit-add-contact-dropdown" class="hidden absolute right-0 top-8 bg-[#0b1d2e] border border-[#a68a56]/50 rounded shadow-lg flex flex-col z-20 w-28">
                                    <button type="button" onclick="addContactType('EMAIL', 'modal-contacts-container', null); document.getElementById('edit-add-contact-dropdown').classList.add('hidden');" class="text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors">EMAIL</button>
                                    <button type="button" onclick="addContactType('CELLPHONE', 'modal-contacts-container', null); document.getElementById('edit-add-contact-dropdown').classList.add('hidden');" class="text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors">CELLPHONE</button>
                                    <button type="button" onclick="addContactType('TELEPHONE', 'modal-contacts-container', null); document.getElementById('edit-add-contact-dropdown').classList.add('hidden');" class="text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors">TELEPHONE</button>
                                    <button type="button" onclick="addContactType('FAX', 'modal-contacts-container', null); document.getElementById('edit-add-contact-dropdown').classList.add('hidden');" class="text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors">FAX</button>
                                </div>
                            </div>
                        </div>
                        <div id="modal-contacts-container" class="flex flex-col gap-4"></div>
                    `;
                    (c.contact_details || []).forEach(cd => addContactType(cd.contact_type, 'modal-contacts-container', cd));
                }
                else if (section === 'addresses') {
                    title.innerText = "Edit Addresses";
                    body.innerHTML = `
                        <div class="flex justify-between items-center mb-2">
                            <span class="text-xs text-[#f9f8f6]/50 uppercase tracking-widest">Manage Client Addresses</span>
                            <button onclick="addAddressRow({}, 'modal-addresses-container')" class="text-[9px] text-[#a68a56] border border-[#a68a56]/40 px-3 py-1 uppercase hover:bg-[#a68a56]/20 rounded transition-colors">+ Add Address</button>
                        </div>
                        <div id="modal-addresses-container" class="flex flex-col gap-4"></div>
                    `;
                    (c.addresses || []).forEach(a => addAddressRow(a, 'modal-addresses-container'));
                }
                else if (section === 'banks') {
                    title.innerText = "Edit Banking";
                    body.innerHTML = `
                        <div class="flex justify-between items-center mb-2">
                            <span class="text-xs text-[#f9f8f6]/50 uppercase tracking-widest">Manage Banking Info</span>
                            <button onclick="addBankRow({}, 'modal-banks-container')" class="text-[9px] text-[#a68a56] border border-[#a68a56]/40 px-3 py-1 uppercase hover:bg-[#a68a56]/20 rounded transition-colors">+ Add Account</button>
                        </div>
                        <div id="modal-banks-container" class="flex flex-col gap-4"></div>
                    `;
                    (c.banks || []).forEach(b => addBankRow(b, 'modal-banks-container'));
                }
            }

            function closeModal() {
                document.getElementById('edit-modal').classList.add('hidden');
                activeModalSection = null;
            }

            async function saveModalEdits() {
                if(!currentViewedClient) return;
                let c = {...currentViewedClient};

                if(activeModalSection === 'personal') {
                    c.first_name = document.getElementById('m-first').value;
                    c.middle_name = document.getElementById('m-middle').value;
                    c.last_name = document.getElementById('m-last').value;
                    c.id_number = document.getElementById('m-idnum').value;
                    c.jurisdiction_type = document.getElementById('m-jur').value;
                    c.registration_number = document.getElementById('m-reg').value;
                    c.marital_status = document.getElementById('m-mar').value;
                    c.practice_number = document.getElementById('m-prac').value;
                }
                else if(activeModalSection === 'employment') {
                    c.occupation = document.getElementById('m-occ').value;
                    c.employer_name = document.getElementById('m-emp-name').value;
                    c.employer_number = document.getElementById('m-emp-num').value.replace(/[^\d]/g, '');
                    c.employer_address_l1 = document.getElementById('m-emp-l1').value;
                    c.employer_address_l2 = document.getElementById('m-emp-l2').value;
                    c.employer_suburb = document.getElementById('m-emp-suburb').value;
                    c.employer_city = document.getElementById('m-emp-city').value;
                    c.employer_postal_code = document.getElementById('m-emp-code').value;
                    c.employer_country = document.getElementById('m-emp-country').value;
                }
                else if(activeModalSection === 'contacts') {
                    const emails = document.querySelectorAll('#edit-modal input[type="email"]');
                    for (let input of emails) {
                        if (input.value && !input.checkValidity()) {
                            return alert("Please provide a valid email address.");
                        }
                    }
                    c.contact_details = Array.from(document.querySelectorAll('#edit-modal .contact-item')).map(item => {
                        const cType = item.querySelector('.contact-type-input').value;
                        let cVal = item.querySelector('.contact-val').value;
                        if (cType !== 'EMAIL') cVal = cVal.replace(/[^\d]/g, '');
                        return {
                            id: item.querySelector('.contact-id').value,
                            contact_type: cType,
                            contact_value: cVal,
                            is_primary: item.querySelector('.contact-prim').checked,
                        };
                    });
                    
                    const types = [...new Set(c.contact_details.map(cd => cd.contact_type))];
                    for (const t of types) {
                        const group = c.contact_details.filter(cd => cd.contact_type === t);
                        if (group.length > 0 && !group.some(cd => cd.is_primary)) {
                            group[0].is_primary = true;
                        }
                    }
                }
                else if(activeModalSection === 'addresses') {
                    c.addresses = Array.from(document.querySelectorAll('#edit-modal .address-item')).map(item => ({
                        id: item.querySelector('.addr-id').value,
                        address_type: item.querySelector('.addr-type').value,
                        is_primary: item.querySelector('.addr-prim').checked,
                        line1: item.querySelector('.addr-l1').value,
                        line2: item.querySelector('.addr-l2').value,
                        suburb: item.querySelector('.addr-sub').value,
                        city: item.querySelector('.addr-city').value,
                        postal_code: item.querySelector('.addr-code').value,
                        country: item.querySelector('.addr-country').value,
                        postal_same: item.querySelector('.addr-postal-same').checked,
                        postal_line1: item.querySelector('.addr-pl1').value,
                        postal_line2: item.querySelector('.addr-pl2').value,
                        postal_suburb: item.querySelector('.addr-psub').value,
                        postal_city: item.querySelector('.addr-pcity').value,
                        postal_code2: item.querySelector('.addr-pcode').value,
                        postal_country: item.querySelector('.addr-pcountry').value,
                    }));
                    if (c.addresses.length > 0 && !c.addresses.some(a => a.is_primary)) {
                        c.addresses[0].is_primary = true;
                    }
                }
                else if(activeModalSection === 'banks') {
                    c.banks = Array.from(document.querySelectorAll('#edit-modal .bank-item')).map(item => ({
                        id: item.querySelector('.bank-id').value,
                        bank_name: item.querySelector('.bank-name').value,
                        branch_code: item.querySelector('.bank-branch').value,
                        account_number: item.querySelector('.bank-acc').value,
                        account_type: item.querySelector('.bank-type').value,
                        is_primary: item.querySelector('.bank-prim').checked,
                    }));
                    if (c.banks.length > 0 && !c.banks.some(b => b.is_primary)) {
                        c.banks[0].is_primary = true;
                    }
                }

                try {
					const res = await fetch('/api/clients/save', {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(c)
					});

					if (res.ok) {
                        closeModal();
                        renderClientProfile(c);
                        showToast("Changes Saved Successfully");
                        
                        document.querySelectorAll('.client-row').forEach(row => {
                            const rd = JSON.parse(row.getAttribute('data-client'));
                            if (rd.id === c.id) {
                                row.setAttribute('data-client', JSON.stringify(c));
                                if (activeModalSection === 'personal') {
                                    row.cells[0].innerHTML = `<span class="font-bold">${c.last_name}</span>, ${c.first_name} ${c.middle_name||''}`;
                                    row.cells[1].innerText = `${c.id_number||''}${c.registration_number||''}`;
                                    row.cells[2].innerText = c.jurisdiction_type;
                                    row.cells[4].innerText = c.practice_number||'';
                                }
                                if (activeModalSection === 'employment') {
                                    row.cells[3].innerText = c.occupation||'';
                                }
                            }
                        });
					}
				} catch (err) {}
            }

            function toggleAddressPostal(chkElement) {
                const container = chkElement.closest('.address-item').querySelector('.addr-postal-container');
                if (chkElement.checked) {
                    container.classList.add('hidden');
                } else {
                    container.classList.remove('hidden');
                }
            }

            async function autocompleteAddress(input, wrapper, prefix = null) {
                clearTimeout(addressTimeout);
                const query = input.value;
                if (query.length < 5) {
                    wrapper.querySelector('.osm-dropdown')?.remove();
                    return;
                }
                
                addressTimeout = setTimeout(async () => {
                    try {
                        const res = await fetch(`https://nominatim.openstreetmap.org/search?format=json&addressdetails=1&countrycodes=za&q=${encodeURIComponent(query)}`);
                        const data = await res.json();
                        wrapper.querySelector('.osm-dropdown')?.remove();
                        if (data.length === 0) return;
                        
                        const dropdown = document.createElement('div');
                        dropdown.className = 'osm-dropdown absolute z-50 bg-[#0b1d2e] border border-[#a68a56]/50 rounded shadow-lg max-h-48 overflow-y-auto w-full top-full mt-1';
                        
                        data.forEach(item => {
                            const btn = document.createElement('button');
                            btn.type = 'button';
                            btn.className = 'text-left px-3 py-2 text-[10px] text-[#f9f8f6] hover:bg-[#a68a56]/20 hover:text-[#a68a56] transition-colors w-full border-b border-[#a68a56]/10 last:border-0';
                            btn.innerText = item.display_name;
                            btn.onclick = () => {
                                const addr = item.address || {};
                                const line1 = (addr.road || '') + (addr.house_number ? ' ' + addr.house_number : '');
                                const sub = addr.neighbourhood || addr.suburb || '';
                                const cty = addr.city || addr.town || addr.municipality || '';
                                const pc = addr.postcode || '';
                                const cntry = addr.country || '';

                                if (prefix) {
                                    document.getElementById(prefix + '-l1').value = line1;
                                    document.getElementById(prefix + '-suburb').value = sub;
                                    document.getElementById(prefix + '-city').value = cty;
                                    document.getElementById(prefix + '-code').value = pc;
                                    document.getElementById(prefix + '-country').value = cntry;
                                } else {
                                    const p = wrapper.closest('.address-item');
                                    p.querySelector('.addr-l1').value = line1;
                                    p.querySelector('.addr-sub').value = sub;
                                    p.querySelector('.addr-city').value = cty;
                                    p.querySelector('.addr-code').value = pc;
                                    p.querySelector('.addr-country').value = cntry;
                                }
                                
                                dropdown.remove();
                            };
                            dropdown.appendChild(btn);
                        });
                        wrapper.appendChild(dropdown);
                    } catch (e) {}
                }, 600);
            }

			function addAddressRow(data = {}, containerId) {
				const div = document.createElement('div');
				div.className = 'flex flex-col gap-2 p-3 bg-black/40 border border-[#a68a56]/20 rounded address-item relative';
                const isPostalSame = data.postal_same !== false;
                
				div.innerHTML = `
                    <button type="button" onclick="this.parentElement.remove()" class="absolute top-2 right-2 text-red-500 hover:text-red-400 text-[10px] font-bold">✕</button>
                    <input type="hidden" class="addr-id" value="${data.id || ''}"/>
					
                    <div class="flex justify-between items-center mb-1 pr-6">
                        <select class="bg-transparent border-b border-[#a68a56]/30 p-1 text-xs focus:border-[#a68a56] outline-none addr-type text-[#a68a56] font-bold">
                            <option value="HOME" class="bg-[#0b1d2e] text-[#f9f8f6]" ${data.address_type === 'HOME' ? 'selected' : ''}>HOME ADDRESS</option>
                            <option value="BUSINESS" class="bg-[#0b1d2e] text-[#f9f8f6]" ${data.address_type === 'BUSINESS' ? 'selected' : ''}>BUSINESS ADDRESS</option>
                        </select>
                        <label class="flex items-center gap-1 text-[9px] text-[#f9f8f6] uppercase cursor-pointer">
                            <input type="radio" name="primary_address" class="accent-[#a68a56] addr-prim" ${data.is_primary ? 'checked' : ''}/> Primary
                        </label>
                    </div>
					
                    <div class="relative w-full">
                        <input type="text" placeholder="Address Line 1" value="${data.line1 || ''}" oninput="autocompleteAddress(this, this.parentElement)" class="w-full bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-l1 outline-none rounded mt-1"/>
                    </div>
					<input type="text" placeholder="Address Line 2" value="${data.line2 || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-l2 outline-none rounded"/>
					<div class="grid grid-cols-2 gap-2">
                        <input type="text" placeholder="Suburb" value="${data.suburb || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-sub outline-none rounded"/>
					    <input type="text" placeholder="City" value="${data.city || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-city outline-none rounded"/>
                        <input type="text" placeholder="Postal Code" value="${data.postal_code || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-code outline-none rounded"/>
					    <input type="text" placeholder="Country" value="${data.country || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-country outline-none rounded"/>
                    </div>

                    <label class="flex items-center gap-2 text-[9px] text-[#f9f8f6] uppercase mt-2 mb-1 cursor-pointer">
                        <input type="checkbox" onchange="toggleAddressPostal(this)" class="accent-[#a68a56] addr-postal-same" ${isPostalSame ? 'checked' : ''}/>
                        Postal is same as above
                    </label>

                    <div class="addr-postal-container ${isPostalSame ? 'hidden' : ''} flex flex-col gap-2 mt-1">
                        <input type="text" placeholder="Postal Line 1" value="${data.postal_line1 || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-pl1 outline-none rounded"/>
                        <input type="text" placeholder="Postal Line 2" value="${data.postal_line2 || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-pl2 outline-none rounded"/>
                        <div class="grid grid-cols-2 gap-2">
                            <input type="text" placeholder="Suburb" value="${data.postal_suburb || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-psub outline-none rounded"/>
                            <input type="text" placeholder="City" value="${data.postal_city || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-pcity outline-none rounded"/>
                            <input type="text" placeholder="Postal Code" value="${data.postal_code2 || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-pcode outline-none rounded"/>
                            <input type="text" placeholder="Country" value="${data.postal_country || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] addr-pcountry outline-none rounded"/>
                        </div>
                    </div>
				`;
				document.getElementById(containerId).appendChild(div);
			}

            function addContactType(type, containerId, data = null) {
                const container = document.getElementById(containerId);
                let section = container.querySelector(`[data-contact-group="${type}"]`);
                if (!section) {
                    section = document.createElement('div');
                    section.setAttribute('data-contact-group', type);
                    section.className = 'flex flex-col gap-2 border border-[#a68a56]/20 p-3 rounded bg-black/20';
                    section.innerHTML = `<h5 class="text-[10px] font-bold uppercase tracking-widest text-[#a68a56] border-b border-[#a68a56]/20 pb-1 mb-2 flex justify-between">${type} <button type="button" onclick="addContactField('${type}', this.parentElement.parentElement)" class="hover:text-[#f9f8f6] transition-colors">+</button></h5>`;
                    container.appendChild(section);
                }
                addContactField(type, section, data);
            }

            function addContactField(type, sectionElement, data = null) {
                const div = document.createElement('div');
                div.className = 'contact-item flex items-center gap-2 relative';
                
                const isEmail = type === 'EMAIL';
                const inputType = isEmail ? 'email' : 'text';
                const inputMax = isEmail ? '' : 'maxlength="14"';
                let inputVal = '';
                if(data && data.contact_value) inputVal = isEmail ? data.contact_value : formatPhoneNumber(data.contact_value);
                const inputPlace = isEmail ? 'Email Address' : '(000) 000 0000';
                const idVal = data ? (data.id || '') : '';
                const isChecked = data ? (data.is_primary ? 'checked' : '') : '';

                div.innerHTML = `
                    <input type="hidden" class="contact-id" value="${idVal}"/>
                    <input type="hidden" class="contact-type-input" value="${type}"/>
                    <input type="${inputType}" ${inputMax} placeholder="${inputPlace}" value="${inputVal}" class="flex-1 bg-black/40 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] contact-val outline-none rounded"/>
                    <label class="flex items-center gap-1 text-[9px] text-[#f9f8f6] uppercase cursor-pointer mr-6">
                        <input type="radio" name="primary_contact_${type}" class="accent-[#a68a56] contact-prim" ${isChecked}/> Pri
                    </label>
                    <button type="button" onclick="this.parentElement.remove()" class="absolute right-0 text-red-500 hover:text-red-400 text-[10px] font-bold px-1">✕</button>
                `;
                sectionElement.appendChild(div);

                if (!isEmail) {
                    const input = div.querySelector('.contact-val');
                    input.addEventListener('input', function() { this.value = formatPhoneNumber(this.value); });
                }
            }

			function addBankRow(data = {}, containerId) {
				const div = document.createElement('div');
				div.className = 'flex flex-col gap-2 p-3 bg-black/40 border border-[#a68a56]/20 rounded bank-item relative';
                
                let bankOptions = '<option value="" class="bg-[#0b1d2e] text-[#f9f8f6]">Select Bank...</option>';
                for (const [name, code] of Object.entries(BANK_CODES)) {
                    const sel = (data.bank_name === name) ? 'selected' : '';
                    bankOptions += `<option value="${name}" class="bg-[#0b1d2e] text-[#f9f8f6]" ${sel}>${name}</option>`;
                }

				div.innerHTML = `
                    <button type="button" onclick="this.parentElement.remove()" class="absolute top-2 right-2 text-red-500 hover:text-red-400 text-[10px] font-bold">✕</button>
                    <input type="hidden" class="bank-id" value="${data.id || ''}"/>
                    
                    <div class="flex justify-between items-center mb-1 pr-6">
                        <span class="text-[10px] text-[#a68a56] font-bold uppercase">Account</span>
                        <label class="flex items-center gap-1 text-[9px] text-[#f9f8f6] uppercase cursor-pointer">
                            <input type="radio" name="primary_bank" class="accent-[#a68a56] bank-prim" ${data.is_primary ? 'checked' : ''}/> Primary
                        </label>
                    </div>

                    <select onchange="updateBranchCode(this)" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] focus:border-[#a68a56] outline-none bank-name text-[#f9f8f6] rounded">
                        ${bankOptions}
                    </select>
					<input type="text" placeholder="Branch Code" value="${data.branch_code || ''}" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] bank-branch outline-none rounded"/>
					<input type="text" placeholder="Account Number" value="${data.account_number || ''}" oninput="this.value = this.value.replace(/[^0-9]/g, '')" class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] text-[#f9f8f6] bank-acc outline-none rounded"/>
                    <select class="bg-black/20 border border-[#a68a56]/30 p-2 text-[10px] focus:border-[#a68a56] outline-none bank-type text-[#f9f8f6] rounded">
                        <option value="Cheque" class="bg-[#0b1d2e] text-[#f9f8f6]" ${data.account_type === 'Cheque' ? 'selected' : ''}>Cheque Account</option>
                        <option value="Savings" class="bg-[#0b1d2e] text-[#f9f8f6]" ${data.account_type === 'Savings' ? 'selected' : ''}>Savings Account</option>
                        <option value="Credit" class="bg-[#0b1d2e] text-[#f9f8f6]" ${data.account_type === 'Credit' ? 'selected' : ''}>Credit Facility</option>
                        <option value="Trust" class="bg-[#0b1d2e] text-[#f9f8f6]" ${data.account_type === 'Trust' ? 'selected' : ''}>Trust Account</option>
                    </select>
				`;
				document.getElementById(containerId).appendChild(div);
			}

            function updateBranchCode(selectElement) {
                const bankName = selectElement.value;
                const branchInput = selectElement.parentElement.querySelector('.bank-branch');
                if(BANK_CODES[bankName]) {
                    branchInput.value = BANK_CODES[bankName];
                }
            }

			async function saveClient() {
                const emailInputs = document.querySelectorAll('#clients-form-view input[type="email"]');
                for (let input of emailInputs) {
                    if (input.value && !input.checkValidity()) {
                        alert("Please provide a valid email address for contact details.");
                        return;
                    }
                }

				const addresses = Array.from(document.querySelectorAll('#clients-form-view .address-item')).map(item => ({
                    id: item.querySelector('.addr-id').value,
					address_type: item.querySelector('.addr-type').value,
                    is_primary: item.querySelector('.addr-prim').checked,
					line1: item.querySelector('.addr-l1').value,
					line2: item.querySelector('.addr-l2').value,
					suburb: item.querySelector('.addr-sub').value,
					city: item.querySelector('.addr-city').value,
					postal_code: item.querySelector('.addr-code').value,
					country: item.querySelector('.addr-country').value,
					postal_same: item.querySelector('.addr-postal-same').checked,
					postal_line1: item.querySelector('.addr-pl1').value,
					postal_line2: item.querySelector('.addr-pl2').value,
					postal_suburb: item.querySelector('.addr-psub').value,
					postal_city: item.querySelector('.addr-pcity').value,
					postal_code2: item.querySelector('.addr-pcode').value,
					postal_country: item.querySelector('.addr-pcountry').value,
				}));
                
                if (addresses.length > 0 && !addresses.some(a => a.is_primary)) addresses[0].is_primary = true;

				const contact_details = Array.from(document.querySelectorAll('#clients-form-view .contact-item')).map(item => {
                    const cType = item.querySelector('.contact-type-input').value;
                    let cVal = item.querySelector('.contact-val').value;
                    if (cType !== 'EMAIL') cVal = cVal.replace(/[^\d]/g, '');
                    return {
                        id: item.querySelector('.contact-id').value,
                        contact_type: cType,
                        contact_value: cVal,
                        is_primary: item.querySelector('.contact-prim').checked,
                    };
                });
                
                const types = [...new Set(contact_details.map(cd => cd.contact_type))];
                for (const t of types) {
                    const group = contact_details.filter(cd => cd.contact_type === t);
                    if (group.length > 0 && !group.some(cd => cd.is_primary)) group[0].is_primary = true;
                }

				const banks = Array.from(document.querySelectorAll('#clients-form-view .bank-item')).map(item => ({
                    id: item.querySelector('.bank-id').value,
					bank_name: item.querySelector('.bank-name').value,
					branch_code: item.querySelector('.bank-branch').value,
					account_number: item.querySelector('.bank-acc').value,
					account_type: item.querySelector('.bank-type').value,
                    is_primary: item.querySelector('.bank-prim').checked,
				}));
                
                if (banks.length > 0 && !banks.some(b => b.is_primary)) banks[0].is_primary = true;

				const payload = {
                    id: document.getElementById('c-id').value,
					first_name: document.getElementById('c-first-name').value,
					middle_name: document.getElementById('c-middle-name').value,
					last_name: document.getElementById('c-last-name').value,
					id_number: document.getElementById('c-id-number').value,
					jurisdiction_type: document.getElementById('c-jurisdiction').value,
					jurisdiction_other: document.getElementById('c-jurisdiction-other').value,
					registration_number: document.getElementById('c-reg-number').value,
					occupation: document.getElementById('c-occupation').value,
					marital_status: document.getElementById('c-marital-status').value,
					practice_number: document.getElementById('c-practice-number').value,
					employer_name: document.getElementById('c-emp-name').value,
					employer_number: document.getElementById('c-emp-number').value.replace(/[^\d]/g, ''),
					employer_address_l1: document.getElementById('c-emp-l1').value,
					employer_address_l2: document.getElementById('c-emp-l2').value,
					employer_suburb: document.getElementById('c-emp-suburb').value,
					employer_city: document.getElementById('c-emp-city').value,
					employer_postal_code: document.getElementById('c-emp-code').value,
					employer_country: document.getElementById('c-emp-country').value,
					employer_postal_same: document.getElementById('c-emp-postal-same').checked,
					employer_postal_l1: document.getElementById('c-emp-pl1').value,
					employer_postal_l2: document.getElementById('c-emp-pl2').value,
					employer_postal_suburb: document.getElementById('c-emp-psuburb').value,
					employer_postal_city: document.getElementById('c-emp-pcity').value,
					employer_postal_code2: document.getElementById('c-emp-pcode').value,
					employer_postal_country: document.getElementById('c-emp-pcountry').value,
					addresses,
					contact_details,
					banks
				};

				try {
					const res = await fetch('/api/clients/save', {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(payload)
					});

					if (res.ok) {
						window.location.reload(); 
					}
				} catch (err) {}
			}

            async function loadClientMatters(clientId) {
    const container = document.getElementById('det-matters-container');
    container.innerHTML = '<div class="p-4 text-center text-[#f9f8f6]/50 text-xs tracking-widest uppercase">Loading matters...</div>';
    
    try {
        const res = await fetch(`/api/matters?client_id=${clientId}`);
        if (!res.ok) throw new Error();
        const matters = await res.json() || [];
        
        if (matters.length === 0) {
            container.innerHTML = `
                <div class="flex-1 flex flex-col items-center justify-center gap-3 text-[#f9f8f6]/30 p-12">
                    <svg class="w-8 h-8 opacity-70" fill="currentColor" viewBox="0 0 20 20"><path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"></path></svg>
                    <span class="text-xs font-semibold tracking-wider uppercase">No Active Matters</span>
                </div>
            `;
            return;
        }

        let html = '';
        matters.forEach(m => {
            const d = new Date(m.updated_at).toLocaleDateString('en-ZA');
            let statusStyle = 'text-gray-400';
            if (m.status === 'Open') statusStyle = 'text-emerald-400';
            if (m.status === 'In Progress') statusStyle = 'text-blue-400';
            if (m.status === 'Referred') statusStyle = 'text-purple-400';

            html += `
                <div class="grid grid-cols-4 text-[#f9f8f6] text-[10px] uppercase tracking-widest p-3 border-b border-[#a68a56]/10 hover:bg-[#a68a56]/10 transition-colors">
                    <div class="font-bold text-[#a68a56]">${m.reference}</div>
                    <div>${m.matter_type}</div>
                    <div class="${statusStyle} font-bold">${m.status}</div>
                    <div>${d}</div>
                </div>
            `;
        });
        container.innerHTML = html;
    } catch (e) {
        container.innerHTML = '<div class="p-4 text-center text-red-400 text-xs">Failed to load matters.</div>';
    }
}
		</script>
	}
}
```

---

## FILE: views\contacts.templ
```templ
package views

templ Contacts() {
	@Layout("LuminaLex - Contacts", true) {
		<div class="z-10 flex-1 flex flex-col p-8 h-full overflow-hidden bg-black/60 backdrop-blur-sm animate-fade-in-up">
			<div class="flex justify-between items-center border-b border-[#a68a56]/50 mb-6 pb-4">
				<div class="flex gap-4 w-full overflow-x-auto no-scrollbar" id="tabs-container">
					<button class="tab-btn active px-4 py-2 text-sm uppercase tracking-widest text-[#a68a56] border-b-2 border-[#a68a56] transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="search">Search</button>
					<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="banks">Banks</button>
					<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="masters">Master's Office</button>
					<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="sheriffs">Sheriffs</button>
					<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="magistrates">Magistrate Courts</button>
					<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="highcourts">High Courts</button>
					<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="lawfirms">Law Firms</button>
				</div>
				<div class="flex items-center gap-3 shrink-0">
					<button onclick="exportData()" id="btn-export" class="hidden bg-[#0b1d2e] border border-emerald-600/50 text-emerald-500 hover:bg-emerald-600/20 px-4 py-2 text-xs uppercase tracking-widest transition-all font-sans rounded shadow-md flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
						Export
					</button>
					<button onclick="syncNow()" class="bg-[#a68a56] border border-[#a68a56] text-[#0b1d2e] hover:bg-[#0b1d2e] hover:text-[#a68a56] px-4 py-2 text-xs uppercase tracking-widest transition-all font-sans font-bold rounded shadow-[0_0_15px_rgba(166,138,86,0.3)] flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
						Sync Data
					</button>
				</div>
			</div>
			<div id="tab-content" class="flex-1 overflow-auto w-full font-sans scroll-smooth">
			</div>
		</div>
		
		@SyncModal()
		<script src="/assets/contacts.js"></script>
	}
}
```

---

## FILE: views\data.go
```go
package views

type LoginData struct {
	Error        string
	LastUsername string
}

type HomeData struct {
	Title string
}

type Contact struct {
	ID        string
	Category  string
	Fields    []string
	IsEditing bool
}

type ClientAddress struct {
	ID            string `json:"id"`
	AddressType   string `json:"address_type"`
	IsPrimary     bool   `json:"is_primary"`
	Line1         string `json:"line1"`
	Line2         string `json:"line2"`
	Suburb        string `json:"suburb"`
	City          string `json:"city"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
	PostalSame    bool   `json:"postal_same"`
	PostalLine1   string `json:"postal_line1"`
	PostalLine2   string `json:"postal_line2"`
	PostalSuburb  string `json:"postal_suburb"`
	PostalCity    string `json:"postal_city"`
	PostalCode2   string `json:"postal_code2"`
	PostalCountry string `json:"postal_country"`
}

type ClientContactDetail struct {
	ID           string `json:"id"`
	ContactType  string `json:"contact_type"`
	ContactValue string `json:"contact_value"`
	IsPrimary    bool   `json:"is_primary"`
}

type ClientBank struct {
	ID            string `json:"id"`
	BankName      string `json:"bank_name"`
	BranchCode    string `json:"branch_code"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	IsPrimary     bool   `json:"is_primary"`
}

type Client struct {
	ID                    string                `json:"id"`
	FirstName             string                `json:"first_name"`
	MiddleName            string                `json:"middle_name"`
	LastName              string                `json:"last_name"`
	IDNumber              string                `json:"id_number"`
	JurisdictionType      string                `json:"jurisdiction_type"`
	JurisdictionOther     string                `json:"jurisdiction_other"`
	RegistrationNumber    string                `json:"registration_number"`
	Occupation            string                `json:"occupation"`
	MaritalStatus         string                `json:"marital_status"`
	EmployerName          string                `json:"employer_name"`
	EmployerNumber        string                `json:"employer_number"`
	EmployerAddressL1     string                `json:"employer_address_l1"`
	EmployerAddressL2     string                `json:"employer_address_l2"`
	EmployerSuburb        string                `json:"employer_suburb"`
	EmployerCity          string                `json:"employer_city"`
	EmployerPostalCode    string                `json:"employer_postal_code"`
	EmployerCountry       string                `json:"employer_country"`
	EmployerPostalSame    bool                  `json:"employer_postal_same"`
	EmployerPostalL1      string                `json:"employer_postal_l1"`
	EmployerPostalL2      string                `json:"employer_postal_l2"`
	EmployerPostalSuburb  string                `json:"employer_postal_suburb"`
	EmployerPostalCity    string                `json:"employer_postal_city"`
	EmployerPostalCode2   string                `json:"employer_postal_code2"`
	EmployerPostalCountry string                `json:"employer_postal_country"`
	PracticeNumber        string                `json:"practice_number"`
	Addresses             []ClientAddress       `json:"addresses"`
	ContactDetails        []ClientContactDetail `json:"contact_details"`
	Banks                 []ClientBank          `json:"banks"`
}

type Service struct {
	ID           string  `json:"id"`
	ServiceType  string  `json:"service_type"`
	Description  string  `json:"description"`
	StandardRate float64 `json:"standard_rate"`
	DurationUnit string  `json:"duration_unit"`
}

```

---

## FILE: views\home.templ
```templ
package views

templ Home(data HomeData) {
	@Layout(data.Title, true) {
		<div class="w-full h-full flex flex-col items-center justify-center animate-fade-in-up">
		</div>
	}
}
```

---

## FILE: views\layout.templ
```templ
package views

templ Layout(title string, showSidebar bool) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
			<title>{ title }</title>
			<script src="https://cdn.tailwindcss.com"></script>
			<style>
				@keyframes fadeIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }
				.animate-fade-in-up { animation: fadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
				.no-scrollbar::-webkit-scrollbar { display: none; }
				.no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
			</style>
		</head>
		<body class="w-screen h-screen overflow-hidden m-0 p-0 text-[#f9f8f6] bg-black font-serif select-none">
			<div class="relative w-full h-full flex overflow-hidden bg-black text-[#f9f8f6]">
				<img src="/assets/background.png" class="absolute inset-0 w-full h-full object-cover opacity-40 z-0 pointer-events-none"/>
				if showSidebar {
					<div id="sidebar-expanded" class="z-20 w-72 h-full bg-[#0b1d2e] border-r-[3px] border-double border-[#a68a56] flex flex-col justify-between transition-all duration-300 shadow-2xl">
						<div class="flex flex-col">
							<div class="flex items-center justify-between p-6 border-b border-[#a68a56]/30">
								<div class="flex items-center gap-3">
									<img src="/assets/logo_notext.png" alt="Logo" class="w-10 h-10 object-contain drop-shadow-md"/>
									<span class="text-[#a68a56] font-normal text-xl tracking-widest uppercase leading-none mt-1">LuminaLex</span>
								</div>
								<button id="btn-close" class="text-[#f9f8f6] hover:text-[#a68a56] p-1 transition-colors text-lg leading-none mt-1">X</button>
							</div>
							<div class="px-6 py-4">
								<div class="relative">
									<input type="text" id="menu-search" placeholder="SEARCH..." class="w-full bg-black/20 border border-[#a68a56]/30 text-[#f9f8f6] placeholder-[#f9f8f6]/30 p-2 pl-9 text-xs uppercase tracking-widest focus:outline-none focus:border-[#a68a56] transition-colors font-sans rounded-none"/>
									<svg class="w-4 h-4 absolute left-3 top-2.5 text-[#a68a56]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
								</div>
							</div>
							<nav class="flex flex-col px-4 gap-2">
								<a href="/home" class="menu-item flex items-center gap-4 text-[#f9f8f6] border-b-[3px] border-transparent hover:border-[#a68a56] hover:text-[#a68a56] p-3 text-sm uppercase tracking-widest transition-all">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path></svg>
									<span class="menu-text">Home</span>
								</a>
								<a href="/clients" class="menu-item flex items-center gap-4 text-[#f9f8f6] border-b-[3px] border-transparent hover:border-[#a68a56] hover:text-[#a68a56] p-3 text-sm uppercase tracking-widest transition-all">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
									<span class="menu-text">Clients</span>
								</a>
								<a href="/services" class="menu-item flex items-center gap-4 text-[#f9f8f6] border-b-[3px] border-transparent hover:border-[#a68a56] hover:text-[#a68a56] p-3 text-sm uppercase tracking-widest transition-all">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
									<span class="menu-text">Services</span>
								</a>
								<a href="/contacts" class="menu-item flex items-center gap-4 text-[#f9f8f6] border-b-[3px] border-transparent hover:border-[#a68a56] hover:text-[#a68a56] p-3 text-sm uppercase tracking-widest transition-all">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
									<span class="menu-text">Contacts</span>
								</a>
							</nav>
						</div>
						<div class="p-6 border-t border-[#a68a56]/30 bg-[#0b1d2e] flex flex-col gap-4">
							<button onclick="checkUpdate()" class="text-sm font-normal text-[#d4c299] hover:text-[#a68a56] uppercase tracking-widest transition-colors flex items-center gap-2 text-left w-full">
								Check For Updates
							</button>
							<a href="/" class="text-sm font-normal text-[#d4c299] hover:text-[#a68a56] uppercase tracking-widest transition-colors flex items-center gap-2">
								Secure Logout
							</a>
						</div>
					</div>
					<div id="sidebar-collapsed" class="z-20 w-24 h-full bg-[#0b1d2e] border-r-[3px] border-double border-[#a68a56] flex flex-col items-center py-6 justify-between hidden transition-all duration-300 shadow-2xl">
						<div class="flex flex-col items-center w-full gap-8">
							<button id="btn-open" class="hover:opacity-80 transition-opacity">
								<img src="/assets/logo_notext.png" alt="Logo" class="w-12 h-12 object-contain drop-shadow-md"/>
							</button>
							<nav class="flex flex-col gap-6 w-full items-center">
								<a href="/home" class="text-[#f9f8f6] hover:text-[#a68a56] transition-colors" title="Home">
									<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path></svg>
								</a>
								<a href="/matters" class="menu-item flex items-center gap-4 text-[#f9f8f6] border-b-[3px] border-transparent hover:border-[#a68a56] hover:text-[#a68a56] p-3 text-sm uppercase tracking-widest transition-all">
	<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
	<span class="menu-text">Matters</span>
</a>
								<a href="/clients" class="text-[#f9f8f6] hover:text-[#a68a56] transition-colors" title="Clients">
									<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
								</a>
								<a href="/services" class="text-[#f9f8f6] hover:text-[#a68a56] transition-colors" title="Services">
									<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
								</a>
								<a href="/contacts" class="text-[#f9f8f6] hover:text-[#a68a56] transition-colors" title="Contacts">
									<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
								</a>
							</nav>
						</div>
						<div class="flex flex-col gap-6">
							<button onclick="checkUpdate()" class="text-[#d4c299] hover:text-[#a68a56] transition-colors" title="Check For Updates">
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
							</button>
							<a href="/" class="text-[#d4c299] hover:text-[#a68a56] transition-colors" title="Logout">
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"></path></svg>
							</a>
						</div>
					</div>
				}
				<div class="z-10 flex-1 flex flex-col p-8 h-full overflow-hidden bg-black/60 backdrop-blur-sm">
					{ children... }
				</div>
			</div>
			
			@UpdateWizardModal()

			if showSidebar {
				<script>
					document.addEventListener('DOMContentLoaded', () => {
						const sidebarExpanded = document.getElementById('sidebar-expanded');
						const sidebarCollapsed = document.getElementById('sidebar-collapsed');
						const btnClose = document.getElementById('btn-close');
						const btnOpen = document.getElementById('btn-open');
						const searchInput = document.getElementById('menu-search');
						const menuItems = document.querySelectorAll('.menu-item');
						const currentPath = window.location.pathname;
						const navLinks = document.querySelectorAll('nav a.menu-item');
						navLinks.forEach(link => {
							if (link.getAttribute('href') === currentPath) {
								link.classList.add('text-[#a68a56]', 'border-[#a68a56]');
								link.classList.remove('text-[#f9f8f6]', 'border-transparent');
							} else {
								link.classList.remove('text-[#a68a56]', 'border-[#a68a56]');
								link.classList.add('text-[#f9f8f6]', 'border-transparent');
							}
						});
						const isCollapsed = localStorage.getItem('luminalex_sidebar_collapsed') === 'true';
						if (isCollapsed) {
							sidebarExpanded.classList.add('hidden');
							sidebarCollapsed.classList.remove('hidden');
						}
						btnClose?.addEventListener('click', () => {
							sidebarExpanded.classList.add('hidden');
							sidebarCollapsed.classList.remove('hidden');
							localStorage.setItem('luminalex_sidebar_collapsed', 'true');
						});
						btnOpen?.addEventListener('click', () => {
							sidebarCollapsed.classList.add('hidden');
							sidebarExpanded.classList.remove('hidden');
							localStorage.setItem('luminalex_sidebar_collapsed', 'false');
						});
						searchInput?.addEventListener('input', (e) => {
							const term = e.target.value.toLowerCase();
							menuItems.forEach(item => {
								const text = item.querySelector('.menu-text').textContent.toLowerCase();
								item.style.display = text.includes(term) ? 'flex' : 'none';
							});
						});
					});

					async function checkUpdate() {
						const modal = document.getElementById('update-wizard-modal');
						const step1 = document.getElementById('updater-step-1');
						const noUpdate = document.getElementById('updater-no-update');
						const loading = document.getElementById('updater-checking');
						
						modal.classList.remove('hidden');
						loading.classList.remove('hidden');

						try {
							const res = await fetch('/api/check-update');
							const data = await res.json();
							loading.classList.add('hidden');
							
							if(data.has_update) {
								document.getElementById('update-target-version').textContent = data.latest_ver;
								document.getElementById('update-release-notes').textContent = data.release_notes;
								document.getElementById('start-update-btn').onclick = () => doUpdate(data.download_url);
								step1.classList.remove('hidden');
							} else {
								noUpdate.classList.remove('hidden');
							}
						} catch(e) {
							alert("Network Error checking for updates.");
							modal.classList.add('hidden');
						}
					}

					async function doUpdate(url) {
						document.getElementById('updater-step-1').classList.add('hidden');
						document.getElementById('updater-step-2').classList.remove('hidden');
						const bar = document.getElementById('update-progress-bar');
						bar.style.width = '10%';

						try {
							const res = await fetch('/api/update/apply', {
								method: 'POST',
								headers: {'Content-Type': 'application/json'},
								body: JSON.stringify({ url: url })
							});
							bar.style.width = '100%';
							if(res.ok) {
								document.getElementById('updater-step-2').classList.add('hidden');
								document.getElementById('updater-step-3').classList.remove('hidden');
							}
						} catch(e) {
							alert("Error applying update.");
						}
					}

					async function restartApp() {
						await fetch('/api/update/restart', { method: 'POST' });
					}
				</script>
			}
		</body>
	</html>
}
```

---

## FILE: views\login.templ
```templ
package views

templ Login(data LoginData) {
	@Layout("LuminaLex - Login", false) {
		<div class="w-full h-full flex items-center justify-center animate-fade-in">
			<form action="/login" method="POST" class="bg-[#0b1d2e]/90 border border-[#a68a56]/40 p-12 rounded-lg shadow-[0_0_40px_rgba(166,138,86,0.15)] w-full max-w-md flex flex-col gap-8 backdrop-blur-md transform transition-all hover:scale-[1.01]">
				<div class="flex flex-col items-center justify-center gap-4">
					<img src="/assets/logo_text.png" class="w-56 drop-shadow-lg"/>
					<div class="h-px w-24 bg-gradient-to-r from-transparent via-[#a68a56]/50 to-transparent"></div>
				</div>
				if data.Error != "" {
					<div class="bg-red-500/10 text-red-400 border-l-4 border-red-500 p-4 text-sm font-sans animate-shake">
						{ data.Error }
					</div>
				}
				<div class="flex flex-col gap-6 font-sans">
					<div class="relative group">
						<label for="username" class="absolute -top-2 left-3 bg-[#0b1d2e] px-1 text-[10px] uppercase tracking-widest text-[#a68a56] z-10 transition-colors group-focus-within:text-[#d4c299]">Username</label>
						<input type="text" id="username" name="username" value={ data.LastUsername } class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-4 pt-4 text-sm focus:outline-none focus:border-[#a68a56] focus:ring-1 focus:ring-[#a68a56]/50 transition-all rounded-md"/>
					</div>
					<div class="relative group">
						<label for="password" class="absolute -top-2 left-3 bg-[#0b1d2e] px-1 text-[10px] uppercase tracking-widest text-[#a68a56] z-10 transition-colors group-focus-within:text-[#d4c299]">Password</label>
						<input type="password" id="password" name="password" class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-4 pt-4 text-sm focus:outline-none focus:border-[#a68a56] focus:ring-1 focus:ring-[#a68a56]/50 transition-all rounded-md"/>
					</div>
				</div>
				<button type="submit" class="relative overflow-hidden bg-[#0b1d2e] border border-[#a68a56] text-[#a68a56] font-bold py-4 px-4 uppercase tracking-widest text-sm transition-all mt-2 shadow-lg group hover:bg-[#a68a56] hover:text-[#0b1d2e] rounded-md">
					<span class="relative z-10">Authenticate</span>
					<div class="absolute inset-0 h-full w-0 bg-[#a68a56] transition-all duration-300 ease-out group-hover:w-full z-0"></div>
				</button>
			</form>
		</div>
	}
}
```

---

## FILE: views\matters.templ
```templ
package views

import "encoding/json"

func getClientDataJSON(clients []Client) string {
	b, _ := json.Marshal(clients)
	return string(b)
}

func getServicesDataJSON(services []Service) string {
	b, _ := json.Marshal(services)
	return string(b)
}

templ Matters(clients []Client, services []Service, currentUser string) {
	@Layout("LuminaLex - Matters", true) {
		<div class="z-10 flex-1 flex flex-col h-full overflow-hidden bg-black/60 backdrop-blur-sm animate-fade-in-up">
			
			<div id="matters-list-view" class="flex flex-col h-full p-8">
				<div class="flex justify-between items-center border-b border-[#a68a56]/50 mb-6 pb-4 shrink-0">
					<div class="flex items-center gap-4 flex-1">
						<input type="text" id="matter-search" placeholder="SEARCH MATTERS..." class="w-1/3 bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] uppercase tracking-widest rounded"/>
						
						<select id="filter-client" class="bg-[#0b1d2e] border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] uppercase tracking-widest rounded">
							<option value="">All Clients</option>
							for _, c := range clients {
								<option value={ c.ID }>{ c.LastName }, { c.FirstName }</option>
							}
						</select>

						<select id="filter-status" class="bg-[#0b1d2e] border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] uppercase tracking-widest rounded">
							<option value="">All Statuses</option>
							<option value="Open">Open</option>
							<option value="In Progress">In Progress</option>
							<option value="Referred">Referred</option>
							<option value="Done">Done</option>
						</select>

						<select id="filter-type" class="bg-[#0b1d2e] border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] uppercase tracking-widest rounded">
							<option value="">All Types</option>
							<option value="Court">Court</option>
							<option value="Document">Document</option>
							<option value="Other">Other</option>
						</select>
					</div>
					<div class="flex gap-4">
						<button onclick="document.getElementById('new-matter-modal').classList.remove('hidden')" class="bg-[#a68a56] border border-[#a68a56] text-[#0b1d2e] hover:bg-[#0b1d2e] hover:text-[#a68a56] px-5 py-2.5 text-xs uppercase tracking-widest transition-all font-bold rounded shadow-[0_0_15px_rgba(166,138,86,0.3)] flex items-center gap-2">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
							New Matter
						</button>
					</div>
				</div>

				<div class="w-full flex-1 overflow-auto border border-[#a68a56]/30 bg-black/20 rounded shadow-lg">
					<table class="w-full text-left border-collapse table-auto" id="main-matters-table">
						<thead class="sticky top-0 bg-[#0b1d2e] z-10 shadow-md">
							<tr class="border-b-2 border-[#a68a56]/50 text-[#a68a56] text-[10px] uppercase tracking-widest">
								<th class="p-4">Reference</th>
								<th class="p-4">Date Created</th>
								<th class="p-4">Client</th>
								<th class="p-4">Type</th>
								<th class="p-4">Status</th>
								<th class="p-4 text-right">Actions</th>
							</tr>
						</thead>
						<tbody id="matters-table-body" class="divide-y divide-[#a68a56]/20 text-xs">
						</tbody>
					</table>
				</div>
			</div>

			<div id="matter-detail-view" class="hidden flex-col h-full bg-[#0b1d2e]">
				<div class="flex items-center justify-between border-b border-[#a68a56]/30 p-6 bg-black/40 shrink-0">
					<div class="flex items-center gap-5">
						<button onclick="closeMatterView()" class="w-9 h-9 flex items-center justify-center rounded-full bg-[#f9f8f6]/5 border border-[#f9f8f6]/10 hover:bg-[#f9f8f6]/10 text-[#f9f8f6] transition-colors">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path></svg>
						</button>
						
						<div class="w-14 h-14 rounded-full border-2 border-[#a68a56] text-[#a68a56] flex items-center justify-center text-xl font-bold shadow-[0_0_10px_rgba(166,138,86,0.2)]">M</div>
						
						<div>
							<h3 id="det-matter-ref" class="text-[#f9f8f6] text-xl font-bold uppercase tracking-widest"></h3>
							<div class="flex items-center gap-3 mt-1 text-xs text-[#f9f8f6]/70">
								<span id="det-matter-client" class="text-emerald-400 font-bold uppercase tracking-widest"></span>
								<span>•</span>
								<span id="det-matter-date" class="tracking-widest"></span>
							</div>
						</div>
					</div>
				</div>

				<div class="flex-1 overflow-hidden flex">
					<div class="w-[60%] h-full p-6 overflow-y-auto flex flex-col gap-6">
						<div class="flex justify-between items-center bg-black/30 border border-[#a68a56]/20 rounded p-4">
							<span class="text-xs uppercase font-bold text-[#a68a56] tracking-widest">Matter Status</span>
							<select id="matter-status-select" onchange="updateMatterStatus(this.value)" class="bg-[#0b1d2e] border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs uppercase tracking-widest focus:outline-none focus:border-[#a68a56] rounded">
								<option value="Open">Open</option>
								<option value="In Progress">In Progress</option>
								<option value="Referred">Referred</option>
								<option value="Done">Done</option>
							</select>
						</div>

						<div class="bg-black/30 border border-[#a68a56]/20 rounded-xl p-6 shadow-inner">
							<div class="flex justify-between items-center mb-6">
								<h4 class="text-[11px] uppercase font-bold text-[#f9f8f6] tracking-widest">Billable Services</h4>
								<div class="flex gap-2">
									<button onclick="document.getElementById('add-service-modal').classList.remove('hidden')" class="px-3 py-1 bg-[#a68a56]/20 border border-[#a68a56]/40 hover:bg-[#a68a56]/40 rounded text-[10px] uppercase tracking-widest text-[#a68a56] transition-colors">+ Add</button>
									<button onclick="printInvoice()" class="px-3 py-1 bg-[#f9f8f6]/5 border border-[#f9f8f6]/10 hover:bg-[#f9f8f6]/10 rounded text-[10px] uppercase tracking-widest text-[#f9f8f6] transition-colors">Print Invoice</button>
								</div>
							</div>
							
							<div class="w-full overflow-auto border border-[#a68a56]/20 bg-black/40 rounded">
								<table class="w-full text-left border-collapse">
									<thead class="bg-[#0b1d2e]">
										<tr class="border-b border-[#a68a56]/30 text-[#a68a56] text-[9px] uppercase tracking-widest">
											<th class="p-3">Description</th>
											<th class="p-3">Rate/Unit</th>
											<th class="p-3">Qty</th>
											<th class="p-3">Tax</th>
											<th class="p-3">Total</th>
											<th class="p-3 text-right">Actions</th>
										</tr>
									</thead>
									<tbody id="matter-services-body" class="divide-y divide-[#a68a56]/10 text-[10px] text-[#f9f8f6]">
									</tbody>
									<tfoot class="bg-[#0b1d2e]/50 border-t-2 border-[#a68a56]/40 text-[#a68a56] text-xs font-bold">
										<tr>
											<td colspan="4" class="p-3 text-right">Subtotal:</td>
											<td colspan="2" class="p-3" id="matter-subtotal">R0.00</td>
										</tr>
										<tr>
											<td colspan="4" class="p-3 text-right">VAT (15%):</td>
											<td colspan="2" class="p-3" id="matter-vat">R0.00</td>
										</tr>
										<tr class="text-sm">
											<td colspan="4" class="p-3 text-right uppercase tracking-widest">Grand Total:</td>
											<td colspan="2" class="p-3 text-[#f9f8f6]" id="matter-total">R0.00</td>
										</tr>
									</tfoot>
								</table>
							</div>
						</div>

						<details class="bg-black/30 border border-[#a68a56]/20 rounded-xl shadow-inner group">
							<summary class="p-4 flex justify-between items-center cursor-pointer list-none text-[11px] uppercase font-bold text-[#f9f8f6] tracking-widest border-b border-transparent group-open:border-[#a68a56]/20">
								Client Data
								<span class="text-[#a68a56] transition-transform group-open:rotate-180">▼</span>
							</summary>
							<div class="p-6 text-xs text-[#f9f8f6]/70 leading-relaxed" id="matter-client-data"></div>
						</details>
					</div>

					<div class="w-[40%] h-full p-6 border-l border-[#a68a56]/30 bg-black/20 flex flex-col gap-4">
						<h4 class="text-[11px] uppercase font-bold text-[#a68a56] tracking-widest border-b border-[#a68a56]/20 pb-2">Matter Notes</h4>
						
						<div class="flex flex-col gap-2">
							<textarea id="new-note-content" rows="3" placeholder="Write a note..." class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] rounded resize-none"></textarea>
							<div class="flex justify-end">
								<button onclick="saveNote()" class="px-4 py-1.5 bg-[#a68a56] text-[#0b1d2e] font-bold text-[10px] uppercase tracking-widest rounded hover:bg-[#d4c299] transition-colors">Post Note</button>
							</div>
						</div>

						<div id="notes-container" class="flex-1 overflow-y-auto pr-2 flex flex-col gap-4 mt-4"></div>
					</div>
				</div>
			</div>
			
			<!-- NEW MATTER MODAL -->
			<div id="new-matter-modal" class="hidden fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-4 backdrop-blur-sm animate-fade-in-up">
				<div class="bg-[#0b1d2e] border border-[#a68a56]/50 rounded-xl w-full max-w-md shadow-2xl flex flex-col">
					<div class="p-6 border-b border-[#a68a56]/30">
						<h3 class="text-[#a68a56] text-base font-bold uppercase tracking-widest">Create New Matter</h3>
					</div>
					<div class="p-6 flex flex-col gap-5">
						<div class="flex flex-col gap-2">
							<label class="text-[10px] text-[#a68a56] uppercase tracking-widest">Select Client</label>
							<select id="new-matter-client" class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] rounded">
								for _, c := range clients {
									<option value={ c.ID }>{ c.LastName }, { c.FirstName } ({ c.IDNumber })</option>
								}
							</select>
						</div>
						<div class="flex flex-col gap-2">
							<label class="text-[10px] text-[#a68a56] uppercase tracking-widest">Matter Type</label>
							<select id="new-matter-type" class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] rounded">
								<option value="Court">Court</option>
								<option value="Document">Document</option>
								<option value="Other">Other</option>
							</select>
						</div>
					</div>
					<div class="p-6 border-t border-[#a68a56]/30 flex justify-end gap-3 bg-black/40 rounded-b-xl">
						<button onclick="document.getElementById('new-matter-modal').classList.add('hidden')" class="px-4 py-2 text-xs uppercase tracking-widest text-red-400 border border-red-500/30 hover:bg-red-500/10 rounded transition-colors">Cancel</button>
						<button onclick="createMatter()" class="px-4 py-2 text-xs uppercase tracking-widest font-bold text-[#0b1d2e] bg-[#a68a56] hover:bg-[#d4c299] rounded transition-colors">Create Matter</button>
					</div>
				</div>
			</div>

			<!-- ADD SERVICE MODAL -->
			<div id="add-service-modal" class="hidden fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-4 backdrop-blur-sm animate-fade-in-up">
				<div class="bg-[#0b1d2e] border border-[#a68a56]/50 rounded-xl w-full max-w-md shadow-2xl flex flex-col">
					<div class="p-6 border-b border-[#a68a56]/30">
						<h3 class="text-[#a68a56] text-base font-bold uppercase tracking-widest">Append Billable Service</h3>
					</div>
					<div class="p-6 flex flex-col gap-5">
						<div class="flex flex-col gap-2">
							<label class="text-[10px] text-[#a68a56] uppercase tracking-widest">Directory Category</label>
							<select id="add-svc-cat" onchange="updateServiceDropdown()" class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] rounded">
								<option value="Wentzel Law">Wentzel Law</option>
								<option value="High Court">High Court</option>
								<option value="Magistrate's Court">Magistrate's Court</option>
							</select>
						</div>
						<div class="flex flex-col gap-2">
							<label class="text-[10px] text-[#a68a56] uppercase tracking-widest">Select Service</label>
							<select id="add-svc-item" class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] rounded">
							</select>
						</div>
						<div class="flex gap-4">
							<div class="flex flex-col gap-2 flex-1">
								<label class="text-[10px] text-[#a68a56] uppercase tracking-widest">Qty / Duration</label>
								<input type="number" id="add-svc-qty" value="1" step="0.1" class="w-full bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-xs focus:outline-none focus:border-[#a68a56] rounded font-mono"/>
							</div>
							<div class="flex flex-col gap-2 justify-end pb-3">
								<label class="flex items-center gap-2 cursor-pointer text-xs text-[#f9f8f6] uppercase tracking-widest">
									<input type="checkbox" id="add-svc-tax" checked class="accent-[#a68a56] w-4 h-4"/>
									Add VAT
								</label>
							</div>
						</div>
					</div>
					<div class="p-6 border-t border-[#a68a56]/30 flex justify-end gap-3 bg-black/40 rounded-b-xl">
						<button onclick="document.getElementById('add-service-modal').classList.add('hidden')" class="px-4 py-2 text-xs uppercase tracking-widest text-red-400 border border-red-500/30 hover:bg-red-500/10 rounded transition-colors">Cancel</button>
						<button onclick="appendServiceToMatter()" class="px-4 py-2 text-xs uppercase tracking-widest font-bold text-[#0b1d2e] bg-[#a68a56] hover:bg-[#d4c299] rounded transition-colors">Add to Matter</button>
					</div>
				</div>
			</div>

			<!-- PRINT INVOICE OVERLAY -->
			<div id="print-overlay" class="hidden fixed inset-0 z-[100] bg-[#0b1d2e]/95 overflow-y-auto flex-col items-center p-8 backdrop-blur-md">
				<div class="w-full max-w-[850px] flex justify-end gap-4 mb-4">
					<button onclick="window.print()" class="px-5 py-2 bg-[#a68a56] text-[#0b1d2e] text-xs font-bold uppercase tracking-widest rounded shadow-lg">Print / Export PDF</button>
					<button onclick="document.getElementById('print-overlay').classList.add('hidden')" class="px-5 py-2 bg-white text-[#0b1d2e] text-xs font-bold uppercase tracking-widest rounded border border-gray-300">Close Document</button>
				</div>
				<div class="bg-white w-full max-w-[850px] p-16 relative shadow-2xl text-[#1a1a1a] font-serif print:m-0 print:p-0 print:shadow-none print:max-w-none">
					<div class="flex justify-between items-start border-b-[3px] border-double border-[#a68a56] pb-8 mb-12">
						<div>
							<h2 class="text-[1.8rem] uppercase tracking-[2px] text-[#a68a56] mb-1">Wentzel Wagener Incorporated</h2>
							<p class="text-sm uppercase tracking-widest text-[#4a4a4a] mb-4">Attorneys, Notaries & Conveyancers</p>
							<div class="text-[0.95rem] leading-relaxed text-[#4a4a4a]">
								<p>150 St George's Mall</p>
								<p>Cape Town City Centre, 8001</p>
								<p>accounts@wentzelwagener.co.za | +27 21 555 0198</p>
							</div>
						</div>
						<div class="text-right">
							<h1 class="text-[2.5rem] tracking-[3px] text-[#0b1d2e] mb-4">TAX INVOICE</h1>
							<div class="text-[0.95rem] text-[#4a4a4a]">
								<p><strong class="text-[#0b1d2e]">Invoice No:</strong> <span id="print-inv-no"></span></p>
								<p><strong class="text-[#0b1d2e]">Date Issued:</strong> <span id="print-inv-date"></span></p>
								<p><strong class="text-[#0b1d2e]">VAT Reg No:</strong> 4890123456</p>
							</div>
						</div>
					</div>

					<div class="grid grid-cols-2 gap-8 mb-12">
						<div class="border border-gray-300 p-6">
							<h4 class="text-xs uppercase tracking-widest border-b border-gray-300 pb-2 mb-4 text-[#4a4a4a]">Instructing Party</h4>
							<h3 id="print-client-name" class="text-lg text-[#1a1a1a] mb-2 font-bold"></h3>
							<p id="print-client-details" class="text-sm text-[#4a4a4a] whitespace-pre-wrap"></p>
						</div>
						<div class="border border-gray-300 p-6">
							<h4 class="text-xs uppercase tracking-widest border-b border-gray-300 pb-2 mb-4 text-[#4a4a4a]">Banking Particulars</h4>
							<div class="text-sm text-[#4a4a4a] leading-relaxed">
								<p><strong class="text-[#0b1d2e]">Bank:</strong> First National Bank</p>
								<p><strong class="text-[#0b1d2e]">Account Name:</strong> Wentzel Wagener Inc Trust</p>
								<p><strong class="text-[#0b1d2e]">Account No:</strong> 62000000000</p>
								<p><strong class="text-[#0b1d2e]">Branch Code:</strong> 250655</p>
							</div>
						</div>
					</div>

					<table class="w-full border-collapse mb-8 text-sm">
						<thead>
							<tr>
								<th class="text-left py-4 border-t-2 border-b-2 border-[#0b1d2e] text-[#0b1d2e] uppercase tracking-widest font-normal">Particulars of Account</th>
								<th class="text-right py-4 border-t-2 border-b-2 border-[#0b1d2e] text-[#0b1d2e] uppercase tracking-widest font-normal">Rate/Unit</th>
								<th class="text-right py-4 border-t-2 border-b-2 border-[#0b1d2e] text-[#0b1d2e] uppercase tracking-widest font-normal">Qty</th>
								<th class="text-right py-4 border-t-2 border-b-2 border-[#0b1d2e] text-[#0b1d2e] uppercase tracking-widest font-normal">Amount</th>
							</tr>
						</thead>
						<tbody id="print-services-body"></tbody>
					</table>

					<div class="w-[45%] ml-auto mt-8">
						<div class="flex justify-between py-2 text-[1.1rem]">
							<span>Fees Subtotal</span>
							<span>R<span id="print-subtotal"></span></span>
						</div>
						<div class="flex justify-between py-2 text-[1.1rem]">
							<span>Value Added Tax (15%)</span>
							<span>R<span id="print-vat"></span></span>
						</div>
						<div class="flex justify-between py-4 mt-2 border-t-[3px] border-b-[3px] border-double border-[#a68a56] text-[1.4rem] font-bold text-[#0b1d2e]">
							<span>Total Due</span>
							<span>R<span id="print-total"></span></span>
						</div>
					</div>

					<div class="mt-16 text-center text-xs text-[#4a4a4a] border-t border-gray-300 pt-4 italic">
						Terms: Strictly 30 days from date of presentation. Interest will be charged on overdue accounts at the prime overdraft rate. All work is undertaken subject to the firm's standard terms of engagement. E&OE.
					</div>
				</div>
			</div>

		</div>
		<style>
			@media print {
				body * { visibility: hidden; }
				#print-overlay, #print-overlay * { visibility: visible; }
				#print-overlay { position: absolute; left: 0; top: 0; background: white; padding: 0; }
				#print-overlay > div:first-child { display: none; } /* Hide buttons */
			}
		</style>
		<script>
			const RAW_CLIENTS = JSON.parse(`{ getClientDataJSON(clients) }`);
			const RAW_SERVICES = JSON.parse(`{ getServicesDataJSON(services) }`);
			const CURRENT_USER = "{ currentUser }";

			const APP_STATE = {
				matters: [],
				currentMatterId: null,
				notes: [],
				matterServices: []
			};

			async function loadMatters() {
				const res = await fetch('/api/matters');
				if (res.ok) {
					APP_STATE.matters = await res.json() || [];
					renderMatters();
				}
			}

			function renderMatters() {
				const tbody = document.getElementById('matters-table-body');
				tbody.innerHTML = '';
				
				const search = document.getElementById('matter-search').value.toLowerCase();
				const clientF = document.getElementById('filter-client').value;
				const statusF = document.getElementById('filter-status').value;
				const typeF = document.getElementById('filter-type').value;

				const filtered = APP_STATE.matters.filter(m => {
					if(search && !m.reference.toLowerCase().includes(search)) return false;
					if(clientF && m.client_id !== clientF) return false;
					if(statusF && m.status !== statusF) return false;
					if(typeF && m.matter_type !== typeF) return false;
					return true;
				});

				filtered.forEach(m => {
					const client = RAW_CLIENTS.find(c => c.id === m.client_id);
					const clientName = client ? `${client.last_name}, ${client.first_name}` : 'Unknown';
					const dateStr = new Date(m.updated_at).toLocaleDateString('en-ZA');

					tbody.innerHTML += `
						<tr class="hover:bg-[#a68a56]/10 transition-colors text-[#f9f8f6]">
							<td class="p-4 font-bold text-[#a68a56]">${m.reference}</td>
							<td class="p-4">${dateStr}</td>
							<td class="p-4">${clientName}</td>
							<td class="p-4">${m.matter_type}</td>
							<td class="p-4">
								<span class="px-2 py-1 rounded text-[9px] uppercase tracking-widest border ${getStatusStyle(m.status)}">${m.status}</span>
							</td>
							<td class="p-4 text-right flex justify-end gap-3">
								<button onclick="openMatter('${m.id}')" class="text-[#a68a56] hover:text-[#d4c299] uppercase tracking-widest text-[10px] font-bold">View</button>
								<button onclick="deleteMatter('${m.id}')" class="text-red-500 hover:text-red-400 uppercase tracking-widest text-[10px] font-bold">Delete</button>
							</td>
						</tr>
					`;
				});
			}

			function getStatusStyle(status) {
				if(status === 'Open') return 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30';
				if(status === 'In Progress') return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
				if(status === 'Done') return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
				if(status === 'Referred') return 'bg-purple-500/20 text-purple-400 border-purple-500/30';
				return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
			}

			async function createMatter() {
				const clientId = document.getElementById('new-matter-client').value;
				const type = document.getElementById('new-matter-type').value;
				if(!clientId) return alert("Select a client.");

				const res = await fetch('/api/matters/save', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({ client_id: clientId, matter_type: type, status: 'Open' })
				});

				if(res.ok) {
					document.getElementById('new-matter-modal').classList.add('hidden');
					await loadMatters();
				}
			}

			async function deleteMatter(id) {
				if(!confirm("Are you sure?")) return;
				const res = await fetch('/api/matters/delete', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({ id })
				});
				if(res.ok) loadMatters();
			}

			async function openMatter(id) {
				APP_STATE.currentMatterId = id;
				const m = APP_STATE.matters.find(x => x.id === id);
				const client = RAW_CLIENTS.find(c => c.id === m.client_id);
				
				document.getElementById('matters-list-view').classList.add('hidden');
				document.getElementById('matter-detail-view').classList.remove('hidden');
				document.getElementById('matter-detail-view').classList.add('flex');

				document.getElementById('det-matter-ref').innerText = m.reference;
				document.getElementById('det-matter-client').innerText = client ? `${client.first_name} ${client.last_name}` : 'Unknown';
				document.getElementById('det-matter-date').innerText = new Date(m.updated_at).toLocaleDateString('en-ZA');
				document.getElementById('matter-status-select').value = m.status;

				document.getElementById('matter-client-data').innerHTML = `
					<p><strong>ID Number:</strong> ${client.id_number || 'N/A'}</p>
					<p><strong>Jurisdiction:</strong> ${client.jurisdiction_type || 'N/A'}</p>
					<p><strong>Occupation:</strong> ${client.occupation || 'N/A'}</p>
					<p><strong>Employer:</strong> ${client.employer_name || 'N/A'}</p>
				`;

				await loadNotes();
				await loadMatterServices();
				updateServiceDropdown();
			}

			function closeMatterView() {
				APP_STATE.currentMatterId = null;
				document.getElementById('matter-detail-view').classList.add('hidden');
				document.getElementById('matter-detail-view').classList.remove('flex');
				document.getElementById('matters-list-view').classList.remove('hidden');
				loadMatters();
			}

			async function updateMatterStatus(val) {
				if(!APP_STATE.currentMatterId) return;
				const m = APP_STATE.matters.find(x => x.id === APP_STATE.currentMatterId);
				m.status = val;
				await fetch('/api/matters/save', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify(m)
				});
			}

			// Notes Logic
			async function loadNotes() {
				const res = await fetch(`/api/matters/notes?matter_id=${APP_STATE.currentMatterId}`);
				if(res.ok) {
					APP_STATE.notes = await res.json() || [];
					renderNotes();
				}
			}

			function renderNotes() {
				const container = document.getElementById('notes-container');
				container.innerHTML = '';
				APP_STATE.notes.forEach(n => {
					const dateStr = new Date(n.updated_at).toLocaleString('en-ZA', {dateStyle:'short', timeStyle:'short'});
					const isMe = n.author === CURRENT_USER;
					
					container.innerHTML += `
						<div class="flex flex-col gap-1 ${isMe ? 'items-end' : 'items-start'} mb-2">
							<div class="text-[9px] text-[#f9f8f6]/50 uppercase tracking-widest px-1">
								${n.author} • ${dateStr}
							</div>
							<div class="group relative max-w-[85%] rounded-lg p-3 text-xs leading-relaxed ${isMe ? 'bg-[#a68a56]/20 text-[#f9f8f6] rounded-tr-none border border-[#a68a56]/30' : 'bg-black/40 text-[#f9f8f6] rounded-tl-none border border-white/10'}">
								${n.content.replace(/\n/g, '<br/>')}
								<button onclick="deleteNote('${n.id}')" class="absolute top-1 -left-6 opacity-0 group-hover:opacity-100 text-red-500 hover:text-red-400 transition-opacity">✕</button>
							</div>
						</div>
					`;
				});
			}

			async function saveNote() {
				const val = document.getElementById('new-note-content').value.trim();
				if(!val) return;
				const res = await fetch('/api/matters/notes/save', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({ matter_id: APP_STATE.currentMatterId, author: CURRENT_USER, content: val })
				});
				if(res.ok) {
					document.getElementById('new-note-content').value = '';
					loadNotes();
				}
			}

			async function deleteNote(id) {
				if(!confirm("Delete note?")) return;
				const res = await fetch('/api/matters/notes/delete', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({ id })
				});
				if(res.ok) loadNotes();
			}

			// Services Logic
			function updateServiceDropdown() {
				const cat = document.getElementById('add-svc-cat').value;
				const select = document.getElementById('add-svc-item');
				select.innerHTML = '';
				const filtered = RAW_SERVICES.filter(s => s.service_type === cat);
				filtered.forEach(s => {
					select.innerHTML += `<option value="${s.id}">${s.description} (R${s.standard_rate.toFixed(2)} / ${s.duration_unit})</option>`;
				});
			}

			async function appendServiceToMatter() {
				const svcId = document.getElementById('add-svc-item').value;
				const qty = parseFloat(document.getElementById('add-svc-qty').value) || 1;
				const tax = document.getElementById('add-svc-tax').checked;

				const baseSvc = RAW_SERVICES.find(s => s.id === svcId);
				if(!baseSvc) return;

				const res = await fetch('/api/matters/services/save', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({ 
						matter_id: APP_STATE.currentMatterId, 
						service_id: svcId,
						snapshot_desc: baseSvc.description,
						snapshot_rate: baseSvc.standard_rate,
						snapshot_unit: baseSvc.duration_unit,
						qty: qty,
						add_tax: tax
					})
				});

				if(res.ok) {
					document.getElementById('add-service-modal').classList.add('hidden');
					loadMatterServices();
				}
			}

			async function loadMatterServices() {
				const res = await fetch(`/api/matters/services?matter_id=${APP_STATE.currentMatterId}`);
				if(res.ok) {
					APP_STATE.matterServices = await res.json() || [];
					renderMatterServices();
				}
			}

			function renderMatterServices() {
				const tbody = document.getElementById('matter-services-body');
				tbody.innerHTML = '';
				
				let sub = 0;
				let vat = 0;

				APP_STATE.matterServices.forEach(ms => {
					const lineTotal = ms.snapshot_rate * ms.qty;
					sub += lineTotal;
					if(ms.add_tax) vat += (lineTotal * 0.15);

					tbody.innerHTML += `
						<tr class="hover:bg-[#a68a56]/10 transition-colors">
							<td class="p-3">${ms.snapshot_desc}</td>
							<td class="p-3 font-mono">R${ms.snapshot_rate.toFixed(2)}/${ms.snapshot_unit}</td>
							<td class="p-3">${ms.qty}</td>
							<td class="p-3 text-[#a68a56]">${ms.add_tax ? 'YES' : 'NO'}</td>
							<td class="p-3 font-mono">R${lineTotal.toFixed(2)}</td>
							<td class="p-3 text-right">
								<button onclick="deleteMatterService('${ms.id}')" class="text-red-500 hover:text-red-400 uppercase tracking-widest font-bold">Del</button>
							</td>
						</tr>
					`;
				});

				document.getElementById('matter-subtotal').innerText = 'R' + sub.toFixed(2);
				document.getElementById('matter-vat').innerText = 'R' + vat.toFixed(2);
				document.getElementById('matter-total').innerText = 'R' + (sub + vat).toFixed(2);
			}

			async function deleteMatterService(id) {
				if(!confirm("Remove service from invoice draft?")) return;
				const res = await fetch('/api/matters/services/delete', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({ id })
				});
				if(res.ok) loadMatterServices();
			}

			function printInvoice() {
				if(APP_STATE.matterServices.length === 0) return alert("Add services before generating invoice.");
				
				const m = APP_STATE.matters.find(x => x.id === APP_STATE.currentMatterId);
				const client = RAW_CLIENTS.find(c => c.id === m.client_id);
				
				document.getElementById('print-inv-no').innerText = m.reference;
				document.getElementById('print-inv-date').innerText = new Date().toLocaleDateString('en-ZA', { day: 'numeric', month: 'long', year: 'numeric' });
				
				document.getElementById('print-client-name').innerText = `${client.first_name} ${client.last_name}`;
				document.getElementById('print-client-details').innerText = `ID: ${client.id_number || 'N/A'}\n${client.employer_address_l1 || 'Address not provided'}`;

				const tbody = document.getElementById('print-services-body');
				tbody.innerHTML = '';
				let sub = 0;
				let vat = 0;

				APP_STATE.matterServices.forEach(ms => {
					const lineTotal = ms.snapshot_rate * ms.qty;
					sub += lineTotal;
					if(ms.add_tax) vat += (lineTotal * 0.15);

					tbody.innerHTML += `
						<tr>
							<td class="py-3 border-b border-gray-200 text-[#1a1a1a]">${ms.snapshot_desc} ${ms.add_tax ? '<sup class="text-gray-400">*</sup>' : ''}</td>
							<td class="py-3 border-b border-gray-200 text-right text-[#4a4a4a]">R${ms.snapshot_rate.toFixed(2)} / ${ms.snapshot_unit}</td>
							<td class="py-3 border-b border-gray-200 text-right text-[#1a1a1a]">${ms.qty}</td>
							<td class="py-3 border-b border-gray-200 text-right text-[#1a1a1a]">R${lineTotal.toFixed(2)}</td>
						</tr>
					`;
				});

				document.getElementById('print-subtotal').innerText = sub.toFixed(2);
				document.getElementById('print-vat').innerText = vat.toFixed(2);
				document.getElementById('print-total').innerText = (sub + vat).toFixed(2);

				document.getElementById('print-overlay').classList.remove('hidden');
				document.getElementById('print-overlay').classList.add('flex');
			}

			document.addEventListener('DOMContentLoaded', () => {
				loadMatters();
				['matter-search', 'filter-client', 'filter-status', 'filter-type'].forEach(id => {
					document.getElementById(id)?.addEventListener('change', renderMatters);
					document.getElementById(id)?.addEventListener('input', renderMatters);
				});
			});
		</script>
	}
}
```

---

## FILE: views\services.templ
```templ

package views

templ Services() {
	@Layout("LuminaLex - Services", true) {
		<div class="z-10 flex-1 flex flex-col p-8 h-full overflow-hidden bg-black/60 backdrop-blur-sm animate-fade-in-up">
			<div class="flex justify-between items-center border-b border-[#a68a56]/50 mb-6 pb-4">
				<div class="flex gap-4 w-full overflow-x-auto no-scrollbar" id="services-tabs">
					<button class="svc-tab-btn active px-4 py-2 text-sm uppercase tracking-widest text-[#a68a56] border-b-2 border-[#a68a56] transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="Wentzel Law">Wentzel Law</button>
					<button class="svc-tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="High Court">High Court</button>
					<button class="svc-tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-all hover:bg-[#a68a56]/10 rounded-t-md" data-target="Magistrate's Court">Magistrate's Court</button>
				</div>
				<div class="flex items-center gap-3 shrink-0">
					<input type="text" id="services-search" placeholder="SEARCH SERVICES..." class="w-64 bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] transition-colors uppercase tracking-widest rounded"/>
					<button onclick="exportServicesTable()" id="btn-export-services" class="bg-[#0b1d2e] border border-emerald-600/50 text-emerald-500 hover:bg-emerald-600/20 px-4 py-2 text-xs uppercase tracking-widest transition-all font-sans rounded shadow-md flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
						Export View
					</button>
				</div>
			</div>

			<div class="w-full flex-1 overflow-auto border border-[#a68a56]/30 bg-black/20 rounded shadow-lg flex flex-col">
				<table class="w-full text-left border-collapse table-auto" id="services-table">
					<thead class="sticky top-0 bg-[#0b1d2e] z-10 shadow-md">
						<tr class="border-b-2 border-[#a68a56]/50 text-[#a68a56] text-[10px] uppercase tracking-widest">
							<th class="p-4 w-1/2">Reference / Description</th>
							<th class="p-4">Standard Rate</th>
							<th class="p-4">Duration / Unit</th>
							<th class="p-4 text-right">Actions</th>
						</tr>
					</thead>
					<tbody id="services-table-body" class="divide-y divide-[#a68a56]/20 text-xs">
					</tbody>
				</table>
			</div>
		</div>

		<script>
    const STATE = {
        activeTab: 'Wentzel Law',
        services: [],
        editingId: null,
        searchTerm: ''
    };

    const DURATION_UNITS = [
        'Per 15 mins',
        'Per hour',
        'Per day',
        'Per document',
        'Per page',
        'Fixed fee'
    ];

    async function loadServices() {
        try {
            const response = await fetch(`/api/services?type=${encodeURIComponent(STATE.activeTab)}`);
            if (response.ok) {
                const data = await response.json();
                // Ensure it's always an array, even if the DB returns null
                STATE.services = data ? data : []; 
            } else {
                console.error("Failed to fetch services. Status:", response.status);
                STATE.services = [];
            }
        } catch (e) {
            console.error("Error loading services:", e);
            STATE.services = [];
        } finally {
            // Guarantee the table (and blank row) renders no matter what
            renderServicesTable();
        }
    }

    function renderServicesTable() {
        const tbody = document.getElementById('services-table-body');
        if (!tbody) return;
        
        tbody.innerHTML = '';

        const safeServices = Array.isArray(STATE.services) ? STATE.services : [];
        const filtered = safeServices.filter(s => {
            if (!STATE.searchTerm) return true;
            const term = STATE.searchTerm.toLowerCase();
            return (s.description || '').toLowerCase().includes(term) || 
                   (s.standard_rate || '').toString().includes(term) || 
                   (s.duration_unit || '').toLowerCase().includes(term);
        });

        filtered.forEach(s => {
            if (STATE.editingId === s.id) {
                tbody.appendChild(createEditRow(s));
            } else {
                tbody.appendChild(createViewRow(s));
            }
        });

        // Always append the blank row if we are not editing
        if (!STATE.editingId) {
            tbody.appendChild(createAddRow());
        }
    }

    function createViewRow(s) {
        const tr = document.createElement('tr');
        tr.className = 'hover:bg-[#a68a56]/10 transition-colors text-[#f9f8f6] service-row';
        tr.innerHTML = `
            <td class="p-4 font-bold text-sm">${s.description || ''}</td>
            <td class="p-4 font-mono text-[#a68a56]">${Number(s.standard_rate || 0).toFixed(2)}</td>
            <td class="p-4 uppercase text-[10px] tracking-widest">${s.duration_unit || ''}</td>
            <td class="p-4 text-right flex justify-end gap-4">
                <button class="text-[#a68a56] hover:text-[#d4c299] uppercase tracking-widest text-[10px] font-bold" onclick="startEdit('${s.id}')">Edit</button>
                <button class="text-red-500 hover:text-red-400 uppercase tracking-widest text-[10px] font-bold" onclick="deleteService('${s.id}')">Delete</button>
            </td>
        `;
        return tr;
    }

    function createEditRow(s) {
        const tr = document.createElement('tr');
        tr.className = 'bg-[#a68a56]/10 text-[#f9f8f6] service-row';
        
        let options = '';
        DURATION_UNITS.forEach(u => {
            options += `<option value="${u}" class="bg-[#0b1d2e] text-[#f9f8f6]" ${u === s.duration_unit ? 'selected' : ''}>${u}</option>`;
        });

        tr.innerHTML = `
            <td class="p-3">
                <input type="text" id="edit-desc-${s.id}" value="${(s.description || '').replace(/"/g, '&quot;')}" class="w-full bg-black/60 border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] rounded"/>
            </td>
            <td class="p-3">
                <input type="number" step="0.01" id="edit-rate-${s.id}" value="${s.standard_rate || 0}" class="w-full bg-black/60 border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] rounded font-mono"/>
            </td>
            <td class="p-3">
                <select id="edit-unit-${s.id}" class="w-full bg-[#0b1d2e] border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-[10px] uppercase tracking-widest focus:outline-none focus:border-[#a68a56] rounded">
                    ${options}
                </select>
            </td>
            <td class="p-3 text-right flex justify-end gap-4 mt-1">
                <button class="text-emerald-400 hover:text-emerald-300 uppercase tracking-widest text-[10px] font-bold" onclick="saveService('${s.id}')">Save</button>
                <button class="text-red-400 hover:text-red-300 uppercase tracking-widest text-[10px] font-bold" onclick="cancelEdit()">Cancel</button>
            </td>
        `;
        return tr;
    }

    function createAddRow() {
        const tr = document.createElement('tr');
        tr.className = 'bg-black/40 text-[#f9f8f6] service-row border-t-2 border-[#a68a56]/30';
        
        let options = '';
        DURATION_UNITS.forEach(u => {
            options += `<option value="${u}" class="bg-[#0b1d2e] text-[#f9f8f6]">${u}</option>`;
        });

        tr.innerHTML = `
            <td class="p-3">
                <input type="text" id="add-desc" placeholder="New Reference / Description" class="w-full bg-black/60 border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] rounded"/>
            </td>
            <td class="p-3">
                <input type="number" step="0.01" id="add-rate" placeholder="0.00" class="w-full bg-black/60 border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-xs focus:outline-none focus:border-[#a68a56] rounded font-mono"/>
            </td>
            <td class="p-3">
                <select id="add-unit" class="w-full bg-[#0b1d2e] border border-[#a68a56]/50 text-[#f9f8f6] p-2 text-[10px] uppercase tracking-widest focus:outline-none focus:border-[#a68a56] rounded">
                    ${options}
                </select>
            </td>
            <td class="p-3 text-right flex justify-end gap-4 mt-1">
                <button class="text-emerald-400 hover:text-emerald-300 uppercase tracking-widest text-[10px] font-bold" onclick="saveService('')">Add</button>
            </td>
        `;
        return tr;
    }

    function startEdit(id) {
        STATE.editingId = id;
        renderServicesTable();
    }

    function cancelEdit() {
        STATE.editingId = null;
        renderServicesTable();
    }

    async function saveService(id) {
        const isNew = !id;
        const descId = isNew ? 'add-desc' : `edit-desc-${id}`;
        const rateId = isNew ? 'add-rate' : `edit-rate-${id}`;
        const unitId = isNew ? 'add-unit' : `edit-unit-${id}`;

        const desc = document.getElementById(descId).value.trim();
        const rate = parseFloat(document.getElementById(rateId).value) || 0;
        const unit = document.getElementById(unitId).value;

        if (!desc) {
            alert("Reference / Description cannot be empty.");
            return;
        }

        const payload = {
            id: id,
            service_type: STATE.activeTab,
            description: desc,
            standard_rate: rate,
            duration_unit: unit
        };

        try {
            const res = await fetch('/api/services/save', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (res.ok) {
                STATE.editingId = null;
                await loadServices();
            } else {
                alert("Failed to save service.");
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function deleteService(id) {
        if (!confirm("Are you sure you want to delete this service?")) return;
        
        try {
            const res = await fetch('/api/services/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: id })
            });
            if (res.ok) {
                await loadServices();
            } else {
                alert("Failed to delete service.");
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function exportServicesTable() {
        const table = document.getElementById('services-table');
        if (!table) return;

        const headerCells = Array.from(table.querySelectorAll('thead th'));
        const headers = headerCells.slice(0, headerCells.length - 1).map(th => th.innerText.trim());

        const rows = [];
        table.querySelectorAll('tbody tr').forEach(tr => {
            // Only export non-editing rows (rows without input fields)
            if (!tr.querySelector('input')) {
                const cells = Array.from(tr.querySelectorAll('td'));
                if (cells.length > 0) {
                    rows.push(cells.slice(0, cells.length - 1).map(td => td.innerText.trim().replace(/\n/g, ' ')));
                }
            }
        });

        if (rows.length === 0) {
            alert("No data available to export.");
            return;
        }

        const filename = `SERVICES_${STATE.activeTab.replace(/\s+/g, '_').toUpperCase()}`;

        try {
            const res = await fetch('/api/contacts/export_view', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ filename, headers, rows })
            });
            
            if (!res.ok) alert("Export failed.");
        } catch (e) {
            console.error(e);
            alert("Export failed due to a network error.");
        }
    }

    // Initialize the page
    document.addEventListener('DOMContentLoaded', () => {
        const tabs = document.querySelectorAll('.svc-tab-btn');
        tabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                tabs.forEach(t => {
                    t.classList.remove('active', 'text-[#a68a56]', 'border-[#a68a56]');
                    t.classList.add('text-[#f9f8f6]', 'border-transparent');
                });
                const target = e.currentTarget;
                target.classList.add('active', 'text-[#a68a56]', 'border-[#a68a56]');
                target.classList.remove('text-[#f9f8f6]', 'border-transparent');
                
                STATE.activeTab = target.getAttribute('data-target');
                STATE.editingId = null;
                
                const searchInput = document.getElementById('services-search');
                if (searchInput) STATE.searchTerm = searchInput.value;
                
                loadServices();
            });
        });

        const searchInput = document.getElementById('services-search');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                STATE.searchTerm = e.target.value;
                renderServicesTable();
            });
        }

        loadServices();
    });
</script>
	}
}
```

---

## FILE: views\sync_modal.templ
```templ
package views

templ SyncModal() {
	<div id="sync-modal" class="hidden fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md">
		<div class="bg-[#0b1d2e] border-2 border-[#a68a56] shadow-2xl w-full max-w-lg p-6 flex flex-col gap-6 rounded-lg">
			<div class="flex items-center justify-between border-b border-[#a68a56]/30 pb-3">
				<div class="flex items-center gap-3">
					<div id="sync-modal-indicator" class="w-3 h-3 bg-[#a68a56] rounded-full animate-ping"></div>
					<h3 id="sync-modal-title" class="text-[#a68a56] text-base font-bold uppercase tracking-widest">Synchronizing Data</h3>
				</div>
				<button id="close-sync-modal-btn" class="hidden text-[#f9f8f6]/60 hover:text-[#f9f8f6] text-xs uppercase tracking-wider">Close</button>
			</div>

			<div id="sync-modal-loading" class="flex flex-col items-center justify-center py-6 gap-4">
				<p class="text-xs text-[#a68a56] uppercase tracking-widest font-bold">Synchronizing Cloud & Local Databases...</p>
				<div class="w-full bg-black/40 h-2 mt-2 rounded overflow-hidden border border-[#a68a56]/30">
					<div id="sync-progress-bar" class="h-full bg-[#a68a56] w-0 transition-all duration-300"></div>
				</div>
			</div>

			<div id="sync-modal-success" class="hidden flex-col items-center justify-center py-6 gap-2">
				<svg class="w-12 h-12 text-emerald-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
				<p class="text-xs text-emerald-400 uppercase tracking-widest font-bold">Data Synchronization Complete!</p>
				<p id="sync-last-time" class="text-[10px] text-[#f9f8f6]/70 font-sans"></p>
			</div>

			<div id="sync-modal-error" class="hidden flex-col gap-3">
				<div class="flex items-center gap-2 text-red-400">
					<svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
					<span class="text-xs uppercase font-bold tracking-wider">Synchronization Failed</span>
				</div>
				<div class="bg-black/60 border border-red-500/30 p-3 max-h-48 overflow-y-auto font-mono text-[10px] text-red-300 whitespace-pre-wrap select-text rounded" id="sync-error-trace"></div>
			</div>
		</div>
	</div>
}
```

---

## FILE: views\updater_modal.templ
```templ
package views

templ UpdateWizardModal() {
	<div id="update-wizard-modal" class="hidden fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md">
		<div class="bg-[#0b1d2e] border-2 border-[#a68a56] shadow-2xl w-full max-w-lg p-6 flex flex-col gap-6 rounded-lg">
			<div class="flex items-center justify-between border-b border-[#a68a56]/30 pb-3">
				<div class="flex items-center gap-3">
					<div class="w-3 h-3 bg-[#a68a56] rounded-full animate-ping"></div>
					<h3 class="text-[#a68a56] text-base font-bold uppercase tracking-widest">System Update</h3>
				</div>
				<button onclick="document.getElementById('update-wizard-modal').classList.add('hidden')" class="text-[#f9f8f6]/60 hover:text-[#f9f8f6] text-xs uppercase tracking-wider">Dismiss</button>
			</div>

			<div id="updater-checking" class="hidden flex-col items-center justify-center py-6 gap-4">
				<div class="w-10 h-10 border-4 border-[#a68a56] border-t-transparent rounded-full animate-spin"></div>
				<p class="text-xs text-[#a68a56] uppercase tracking-widest font-bold">Checking GitHub for software updates...</p>
			</div>

			<div id="updater-no-update" class="hidden flex-col items-center justify-center py-6 gap-4 text-center">
				<svg class="w-12 h-12 text-emerald-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
				<p class="text-xs text-emerald-400 uppercase tracking-widest font-bold">Latest Version Installed</p>
			</div>

			<div id="updater-step-1" class="hidden flex-col gap-4">
				<p class="text-xs text-[#f9f8f6]/90 leading-relaxed font-sans">
					Version <span id="update-target-version" class="text-[#a68a56] font-bold"></span> is available for download.
				</p>
				<div class="bg-black/40 border border-[#a68a56]/20 p-3 max-h-36 overflow-y-auto font-sans rounded">
					<p class="text-[10px] uppercase text-[#a68a56] tracking-wider mb-1 font-bold">Release Notes</p>
					<div id="update-release-notes" class="text-xs text-[#f9f8f6]/70 whitespace-pre-wrap"></div>
				</div>
				<button id="start-update-btn" class="w-full bg-[#a68a56] text-[#0b1d2e] py-3 text-xs uppercase tracking-widest font-bold transition-all hover:bg-[#d4c299] rounded shadow-md">
					Apply Update
				</button>
			</div>

			<div id="updater-step-2" class="hidden flex-col items-center justify-center py-6 gap-4">
				<div class="w-10 h-10 border-4 border-[#a68a56] border-t-transparent rounded-full animate-spin"></div>
				<p class="text-xs text-[#a68a56] uppercase tracking-widest font-bold">Downloading update binary...</p>
				<div class="w-full bg-black/40 h-2 mt-4 rounded overflow-hidden border border-[#a68a56]/30">
					<div id="update-progress-bar" class="h-full bg-[#a68a56] w-0 transition-all duration-300"></div>
				</div>
			</div>

			<div id="updater-step-3" class="hidden flex-col items-center justify-center py-6 gap-4 text-center">
				<svg class="w-12 h-12 text-emerald-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
				<p class="text-xs text-emerald-400 uppercase tracking-widest font-bold">Update Applied Successfully</p>
				<button onclick="restartApp()" class="w-full bg-[#a68a56] text-[#0b1d2e] py-3 text-xs uppercase tracking-widest font-bold transition-all hover:bg-[#d4c299] rounded shadow-md mt-4">
					Restart Now
				</button>
			</div>
		</div>
	</div>
}
```

---

## FILE: wailsjs\go\models.ts
```ts
export namespace core {
	
	export class ContactRecord {
	    id: string;
	    category: string;
	    fields: string[];
	    // Go type: time
	    updated_at: any;
	    deleted: boolean;
	    synced: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ContactRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.fields = source["fields"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.deleted = source["deleted"];
	        this.synced = source["synced"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExportPayload {
	    filename: string;
	    headers: string[];
	    rows: string[][];
	
	    static createFrom(source: any = {}) {
	        return new ExportPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.headers = source["headers"];
	        this.rows = source["rows"];
	    }
	}
	export class SyncStatus {
	    is_syncing: boolean;
	    last_sync: string;
	    error: string;
	    details: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_syncing = source["is_syncing"];
	        this.last_sync = source["last_sync"];
	        this.error = source["error"];
	        this.details = source["details"];
	    }
	}
	export class UpdateCheckResult {
	    has_update: boolean;
	    latest_ver: string;
	    release_notes: string;
	    download_url: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.has_update = source["has_update"];
	        this.latest_ver = source["latest_ver"];
	        this.release_notes = source["release_notes"];
	        this.download_url = source["download_url"];
	    }
	}

}


```

---

## FILE: wailsjs\runtime\package.json
```json
{
  "name": "@wailsapp/runtime",
  "version": "2.0.0",
  "description": "Wails Javascript runtime library",
  "main": "runtime.js",
  "types": "runtime.d.ts",
  "scripts": {
  },
  "repository": {
    "type": "git",
    "url": "git+https://github.com/wailsapp/wails.git"
  },
  "keywords": [
    "Wails",
    "Javascript",
    "Go"
  ],
  "author": "Lea Anthony <lea.anthony@gmail.com>",
  "license": "MIT",
  "bugs": {
    "url": "https://github.com/wailsapp/wails/issues"
  },
  "homepage": "https://github.com/wailsapp/wails#readme"
}

```

---

## FILE: wailsjs\runtime\runtime.d.ts
```ts
/*
 _       __      _ __
| |     / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  )
|__/|__/\__,_/_/_/____/
The electron alternative for Go
(c) Lea Anthony 2019-present
*/

export interface Position {
    x: number;
    y: number;
}

export interface Size {
    w: number;
    h: number;
}

export interface Screen {
    isCurrent: boolean;
    isPrimary: boolean;
    width : number
    height : number
}

// Environment information such as platform, buildtype, ...
export interface EnvironmentInfo {
    buildType: string;
    platform: string;
    arch: string;
}

// [EventsEmit](https://wails.io/docs/reference/runtime/events#eventsemit)
// emits the given event. Optional data may be passed with the event.
// This will trigger any event listeners.
export function EventsEmit(eventName: string, ...data: any): void;

// [EventsOn](https://wails.io/docs/reference/runtime/events#eventson) sets up a listener for the given event name.
export function EventsOn(eventName: string, callback: (...data: any) => void): () => void;

// [EventsOnMultiple](https://wails.io/docs/reference/runtime/events#eventsonmultiple)
// sets up a listener for the given event name, but will only trigger a given number times.
export function EventsOnMultiple(eventName: string, callback: (...data: any) => void, maxCallbacks: number): () => void;

// [EventsOnce](https://wails.io/docs/reference/runtime/events#eventsonce)
// sets up a listener for the given event name, but will only trigger once.
export function EventsOnce(eventName: string, callback: (...data: any) => void): () => void;

// [EventsOff](https://wails.io/docs/reference/runtime/events#eventsoff)
// unregisters the listener for the given event name.
export function EventsOff(eventName: string, ...additionalEventNames: string[]): void;

// [EventsOffAll](https://wails.io/docs/reference/runtime/events#eventsoffall)
// unregisters all listeners.
export function EventsOffAll(): void;

// [LogPrint](https://wails.io/docs/reference/runtime/log#logprint)
// logs the given message as a raw message
export function LogPrint(message: string): void;

// [LogTrace](https://wails.io/docs/reference/runtime/log#logtrace)
// logs the given message at the `trace` log level.
export function LogTrace(message: string): void;

// [LogDebug](https://wails.io/docs/reference/runtime/log#logdebug)
// logs the given message at the `debug` log level.
export function LogDebug(message: string): void;

// [LogError](https://wails.io/docs/reference/runtime/log#logerror)
// logs the given message at the `error` log level.
export function LogError(message: string): void;

// [LogFatal](https://wails.io/docs/reference/runtime/log#logfatal)
// logs the given message at the `fatal` log level.
// The application will quit after calling this method.
export function LogFatal(message: string): void;

// [LogInfo](https://wails.io/docs/reference/runtime/log#loginfo)
// logs the given message at the `info` log level.
export function LogInfo(message: string): void;

// [LogWarning](https://wails.io/docs/reference/runtime/log#logwarning)
// logs the given message at the `warning` log level.
export function LogWarning(message: string): void;

// [WindowReload](https://wails.io/docs/reference/runtime/window#windowreload)
// Forces a reload by the main application as well as connected browsers.
export function WindowReload(): void;

// [WindowReloadApp](https://wails.io/docs/reference/runtime/window#windowreloadapp)
// Reloads the application frontend.
export function WindowReloadApp(): void;

// [WindowSetAlwaysOnTop](https://wails.io/docs/reference/runtime/window#windowsetalwaysontop)
// Sets the window AlwaysOnTop or not on top.
export function WindowSetAlwaysOnTop(b: boolean): void;

// [WindowSetSystemDefaultTheme](https://wails.io/docs/next/reference/runtime/window#windowsetsystemdefaulttheme)
// *Windows only*
// Sets window theme to system default (dark/light).
export function WindowSetSystemDefaultTheme(): void;

// [WindowSetLightTheme](https://wails.io/docs/next/reference/runtime/window#windowsetlighttheme)
// *Windows only*
// Sets window to light theme.
export function WindowSetLightTheme(): void;

// [WindowSetDarkTheme](https://wails.io/docs/next/reference/runtime/window#windowsetdarktheme)
// *Windows only*
// Sets window to dark theme.
export function WindowSetDarkTheme(): void;

// [WindowCenter](https://wails.io/docs/reference/runtime/window#windowcenter)
// Centers the window on the monitor the window is currently on.
export function WindowCenter(): void;

// [WindowSetTitle](https://wails.io/docs/reference/runtime/window#windowsettitle)
// Sets the text in the window title bar.
export function WindowSetTitle(title: string): void;

// [WindowFullscreen](https://wails.io/docs/reference/runtime/window#windowfullscreen)
// Makes the window full screen.
export function WindowFullscreen(): void;

// [WindowUnfullscreen](https://wails.io/docs/reference/runtime/window#windowunfullscreen)
// Restores the previous window dimensions and position prior to full screen.
export function WindowUnfullscreen(): void;

// [WindowIsFullscreen](https://wails.io/docs/reference/runtime/window#windowisfullscreen)
// Returns the state of the window, i.e. whether the window is in full screen mode or not.
export function WindowIsFullscreen(): Promise<boolean>;

// [WindowSetSize](https://wails.io/docs/reference/runtime/window#windowsetsize)
// Sets the width and height of the window.
export function WindowSetSize(width: number, height: number): void;

// [WindowGetSize](https://wails.io/docs/reference/runtime/window#windowgetsize)
// Gets the width and height of the window.
export function WindowGetSize(): Promise<Size>;

// [WindowSetMaxSize](https://wails.io/docs/reference/runtime/window#windowsetmaxsize)
// Sets the maximum window size. Will resize the window if the window is currently larger than the given dimensions.
// Setting a size of 0,0 will disable this constraint.
export function WindowSetMaxSize(width: number, height: number): void;

// [WindowSetMinSize](https://wails.io/docs/reference/runtime/window#windowsetminsize)
// Sets the minimum window size. Will resize the window if the window is currently smaller than the given dimensions.
// Setting a size of 0,0 will disable this constraint.
export function WindowSetMinSize(width: number, height: number): void;

// [WindowSetPosition](https://wails.io/docs/reference/runtime/window#windowsetposition)
// Sets the window position relative to the monitor the window is currently on.
export function WindowSetPosition(x: number, y: number): void;

// [WindowGetPosition](https://wails.io/docs/reference/runtime/window#windowgetposition)
// Gets the window position relative to the monitor the window is currently on.
export function WindowGetPosition(): Promise<Position>;

// [WindowHide](https://wails.io/docs/reference/runtime/window#windowhide)
// Hides the window.
export function WindowHide(): void;

// [WindowShow](https://wails.io/docs/reference/runtime/window#windowshow)
// Shows the window, if it is currently hidden.
export function WindowShow(): void;

// [WindowMaximise](https://wails.io/docs/reference/runtime/window#windowmaximise)
// Maximises the window to fill the screen.
export function WindowMaximise(): void;

// [WindowToggleMaximise](https://wails.io/docs/reference/runtime/window#windowtogglemaximise)
// Toggles between Maximised and UnMaximised.
export function WindowToggleMaximise(): void;

// [WindowUnmaximise](https://wails.io/docs/reference/runtime/window#windowunmaximise)
// Restores the window to the dimensions and position prior to maximising.
export function WindowUnmaximise(): void;

// [WindowIsMaximised](https://wails.io/docs/reference/runtime/window#windowismaximised)
// Returns the state of the window, i.e. whether the window is maximised or not.
export function WindowIsMaximised(): Promise<boolean>;

// [WindowMinimise](https://wails.io/docs/reference/runtime/window#windowminimise)
// Minimises the window.
export function WindowMinimise(): void;

// [WindowUnminimise](https://wails.io/docs/reference/runtime/window#windowunminimise)
// Restores the window to the dimensions and position prior to minimising.
export function WindowUnminimise(): void;

// [WindowIsMinimised](https://wails.io/docs/reference/runtime/window#windowisminimised)
// Returns the state of the window, i.e. whether the window is minimised or not.
export function WindowIsMinimised(): Promise<boolean>;

// [WindowIsNormal](https://wails.io/docs/reference/runtime/window#windowisnormal)
// Returns the state of the window, i.e. whether the window is normal or not.
export function WindowIsNormal(): Promise<boolean>;

// [WindowSetBackgroundColour](https://wails.io/docs/reference/runtime/window#windowsetbackgroundcolour)
// Sets the background colour of the window to the given RGBA colour definition. This colour will show through for all transparent pixels.
export function WindowSetBackgroundColour(R: number, G: number, B: number, A: number): void;

// [ScreenGetAll](https://wails.io/docs/reference/runtime/window#screengetall)
// Gets the all screens. Call this anew each time you want to refresh data from the underlying windowing system.
export function ScreenGetAll(): Promise<Screen[]>;

// [BrowserOpenURL](https://wails.io/docs/reference/runtime/browser#browseropenurl)
// Opens the given URL in the system browser.
export function BrowserOpenURL(url: string): void;

// [Environment](https://wails.io/docs/reference/runtime/intro#environment)
// Returns information about the environment
export function Environment(): Promise<EnvironmentInfo>;

// [Quit](https://wails.io/docs/reference/runtime/intro#quit)
// Quits the application.
export function Quit(): void;

// [Hide](https://wails.io/docs/reference/runtime/intro#hide)
// Hides the application.
export function Hide(): void;

// [Show](https://wails.io/docs/reference/runtime/intro#show)
// Shows the application.
export function Show(): void;

// [ClipboardGetText](https://wails.io/docs/reference/runtime/clipboard#clipboardgettext)
// Returns the current text stored on clipboard
export function ClipboardGetText(): Promise<string>;

// [ClipboardSetText](https://wails.io/docs/reference/runtime/clipboard#clipboardsettext)
// Sets a text on the clipboard
export function ClipboardSetText(text: string): Promise<boolean>;

// [OnFileDrop](https://wails.io/docs/reference/runtime/draganddrop#onfiledrop)
// OnFileDrop listens to drag and drop events and calls the callback with the coordinates of the drop and an array of path strings.
export function OnFileDrop(callback: (x: number, y: number ,paths: string[]) => void, useDropTarget: boolean) :void

// [OnFileDropOff](https://wails.io/docs/reference/runtime/draganddrop#dragandddropoff)
// OnFileDropOff removes the drag and drop listeners and handlers.
export function OnFileDropOff() :void

// Check if the file path resolver is available
export function CanResolveFilePaths(): boolean;

// Resolves file paths for an array of files
export function ResolveFilePaths(files: File[]): void

// Notification types
export interface NotificationOptions {
    id: string;
    title: string;
    subtitle?: string; // macOS and Linux only
    body?: string;
    categoryId?: string;
    data?: { [key: string]: any };
}

export interface NotificationAction {
    id?: string;
    title?: string;
    destructive?: boolean; // macOS-specific
}

export interface NotificationCategory {
    id?: string;
    actions?: NotificationAction[];
    hasReplyField?: boolean;
    replyPlaceholder?: string;
    replyButtonTitle?: string;
}

// [InitializeNotifications](https://wails.io/docs/reference/runtime/notification#initializenotifications)
// Initializes the notification service for the application.
// This must be called before sending any notifications.
export function InitializeNotifications(): Promise<void>;

// [CleanupNotifications](https://wails.io/docs/reference/runtime/notification#cleanupnotifications)
// Cleans up notification resources and releases any held connections.
export function CleanupNotifications(): Promise<void>;

// [IsNotificationAvailable](https://wails.io/docs/reference/runtime/notification#isnotificationavailable)
// Checks if notifications are available on the current platform.
export function IsNotificationAvailable(): Promise<boolean>;

// [RequestNotificationAuthorization](https://wails.io/docs/reference/runtime/notification#requestnotificationauthorization)
// Requests notification authorization from the user (macOS only).
export function RequestNotificationAuthorization(): Promise<boolean>;

// [CheckNotificationAuthorization](https://wails.io/docs/reference/runtime/notification#checknotificationauthorization)
// Checks the current notification authorization status (macOS only).
export function CheckNotificationAuthorization(): Promise<boolean>;

// [SendNotification](https://wails.io/docs/reference/runtime/notification#sendnotification)
// Sends a basic notification with the given options.
export function SendNotification(options: NotificationOptions): Promise<void>;

// [SendNotificationWithActions](https://wails.io/docs/reference/runtime/notification#sendnotificationwithactions)
// Sends a notification with action buttons. Requires a registered category.
export function SendNotificationWithActions(options: NotificationOptions): Promise<void>;

// [RegisterNotificationCategory](https://wails.io/docs/reference/runtime/notification#registernotificationcategory)
// Registers a notification category that can be used with SendNotificationWithActions.
export function RegisterNotificationCategory(category: NotificationCategory): Promise<void>;

// [RemoveNotificationCategory](https://wails.io/docs/reference/runtime/notification#removenotificationcategory)
// Removes a previously registered notification category.
export function RemoveNotificationCategory(categoryId: string): Promise<void>;

// [RemoveAllPendingNotifications](https://wails.io/docs/reference/runtime/notification#removeallpendingnotifications)
// Removes all pending notifications from the notification center.
export function RemoveAllPendingNotifications(): Promise<void>;

// [RemovePendingNotification](https://wails.io/docs/reference/runtime/notification#removependingnotification)
// Removes a specific pending notification by its identifier.
export function RemovePendingNotification(identifier: string): Promise<void>;

// [RemoveAllDeliveredNotifications](https://wails.io/docs/reference/runtime/notification#removealldeliverednotifications)
// Removes all delivered notifications from the notification center.
export function RemoveAllDeliveredNotifications(): Promise<void>;

// [RemoveDeliveredNotification](https://wails.io/docs/reference/runtime/notification#removedeliverednotification)
// Removes a specific delivered notification by its identifier.
export function RemoveDeliveredNotification(identifier: string): Promise<void>;

// [RemoveNotification](https://wails.io/docs/reference/runtime/notification#removenotification)
// Removes a notification by its identifier (cross-platform convenience function).
export function RemoveNotification(identifier: string): Promise<void>;
```

---

## FILE: wailsjs\runtime\runtime.js
```js
/*
 _       __      _ __
| |     / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  )
|__/|__/\__,_/_/_/____/
The electron alternative for Go
(c) Lea Anthony 2019-present
*/

export function LogPrint(message) {
    window.runtime.LogPrint(message);
}

export function LogTrace(message) {
    window.runtime.LogTrace(message);
}

export function LogDebug(message) {
    window.runtime.LogDebug(message);
}

export function LogInfo(message) {
    window.runtime.LogInfo(message);
}

export function LogWarning(message) {
    window.runtime.LogWarning(message);
}

export function LogError(message) {
    window.runtime.LogError(message);
}

export function LogFatal(message) {
    window.runtime.LogFatal(message);
}

export function EventsOnMultiple(eventName, callback, maxCallbacks) {
    return window.runtime.EventsOnMultiple(eventName, callback, maxCallbacks);
}

export function EventsOn(eventName, callback) {
    return EventsOnMultiple(eventName, callback, -1);
}

export function EventsOff(eventName, ...additionalEventNames) {
    return window.runtime.EventsOff(eventName, ...additionalEventNames);
}

export function EventsOffAll() {
  return window.runtime.EventsOffAll();
}

export function EventsOnce(eventName, callback) {
    return EventsOnMultiple(eventName, callback, 1);
}

export function EventsEmit(eventName) {
    let args = [eventName].slice.call(arguments);
    return window.runtime.EventsEmit.apply(null, args);
}

export function WindowReload() {
    window.runtime.WindowReload();
}

export function WindowReloadApp() {
    window.runtime.WindowReloadApp();
}

export function WindowSetAlwaysOnTop(b) {
    window.runtime.WindowSetAlwaysOnTop(b);
}

export function WindowSetSystemDefaultTheme() {
    window.runtime.WindowSetSystemDefaultTheme();
}

export function WindowSetLightTheme() {
    window.runtime.WindowSetLightTheme();
}

export function WindowSetDarkTheme() {
    window.runtime.WindowSetDarkTheme();
}

export function WindowCenter() {
    window.runtime.WindowCenter();
}

export function WindowSetTitle(title) {
    window.runtime.WindowSetTitle(title);
}

export function WindowFullscreen() {
    window.runtime.WindowFullscreen();
}

export function WindowUnfullscreen() {
    window.runtime.WindowUnfullscreen();
}

export function WindowIsFullscreen() {
    return window.runtime.WindowIsFullscreen();
}

export function WindowGetSize() {
    return window.runtime.WindowGetSize();
}

export function WindowSetSize(width, height) {
    window.runtime.WindowSetSize(width, height);
}

export function WindowSetMaxSize(width, height) {
    window.runtime.WindowSetMaxSize(width, height);
}

export function WindowSetMinSize(width, height) {
    window.runtime.WindowSetMinSize(width, height);
}

export function WindowSetPosition(x, y) {
    window.runtime.WindowSetPosition(x, y);
}

export function WindowGetPosition() {
    return window.runtime.WindowGetPosition();
}

export function WindowHide() {
    window.runtime.WindowHide();
}

export function WindowShow() {
    window.runtime.WindowShow();
}

export function WindowMaximise() {
    window.runtime.WindowMaximise();
}

export function WindowToggleMaximise() {
    window.runtime.WindowToggleMaximise();
}

export function WindowUnmaximise() {
    window.runtime.WindowUnmaximise();
}

export function WindowIsMaximised() {
    return window.runtime.WindowIsMaximised();
}

export function WindowMinimise() {
    window.runtime.WindowMinimise();
}

export function WindowUnminimise() {
    window.runtime.WindowUnminimise();
}

export function WindowSetBackgroundColour(R, G, B, A) {
    window.runtime.WindowSetBackgroundColour(R, G, B, A);
}

export function ScreenGetAll() {
    return window.runtime.ScreenGetAll();
}

export function WindowIsMinimised() {
    return window.runtime.WindowIsMinimised();
}

export function WindowIsNormal() {
    return window.runtime.WindowIsNormal();
}

export function BrowserOpenURL(url) {
    window.runtime.BrowserOpenURL(url);
}

export function Environment() {
    return window.runtime.Environment();
}

export function Quit() {
    window.runtime.Quit();
}

export function Hide() {
    window.runtime.Hide();
}

export function Show() {
    window.runtime.Show();
}

export function ClipboardGetText() {
    return window.runtime.ClipboardGetText();
}

export function ClipboardSetText(text) {
    return window.runtime.ClipboardSetText(text);
}

/**
 * Callback for OnFileDrop returns a slice of file path strings when a drop is finished.
 *
 * @export
 * @callback OnFileDropCallback
 * @param {number} x - x coordinate of the drop
 * @param {number} y - y coordinate of the drop
 * @param {string[]} paths - A list of file paths.
 */

/**
 * OnFileDrop listens to drag and drop events and calls the callback with the coordinates of the drop and an array of path strings.
 *
 * @export
 * @param {OnFileDropCallback} callback - Callback for OnFileDrop returns a slice of file path strings when a drop is finished.
 * @param {boolean} [useDropTarget=true] - Only call the callback when the drop finished on an element that has the drop target style. (--wails-drop-target)
 */
export function OnFileDrop(callback, useDropTarget) {
    return window.runtime.OnFileDrop(callback, useDropTarget);
}

/**
 * OnFileDropOff removes the drag and drop listeners and handlers.
 */
export function OnFileDropOff() {
    return window.runtime.OnFileDropOff();
}

export function CanResolveFilePaths() {
    return window.runtime.CanResolveFilePaths();
}

export function ResolveFilePaths(files) {
    return window.runtime.ResolveFilePaths(files);
}

export function InitializeNotifications() {
    return window.runtime.InitializeNotifications();
}

export function CleanupNotifications() {
    return window.runtime.CleanupNotifications();
}

export function IsNotificationAvailable() {
    return window.runtime.IsNotificationAvailable();
}

export function RequestNotificationAuthorization() {
    return window.runtime.RequestNotificationAuthorization();
}

export function CheckNotificationAuthorization() {
    return window.runtime.CheckNotificationAuthorization();
}

export function SendNotification(options) {
    return window.runtime.SendNotification(options);
}

export function SendNotificationWithActions(options) {
    return window.runtime.SendNotificationWithActions(options);
}

export function RegisterNotificationCategory(category) {
    return window.runtime.RegisterNotificationCategory(category);
}

export function RemoveNotificationCategory(categoryId) {
    return window.runtime.RemoveNotificationCategory(categoryId);
}

export function RemoveAllPendingNotifications() {
    return window.runtime.RemoveAllPendingNotifications();
}

export function RemovePendingNotification(identifier) {
    return window.runtime.RemovePendingNotification(identifier);
}

export function RemoveAllDeliveredNotifications() {
    return window.runtime.RemoveAllDeliveredNotifications();
}

export function RemoveDeliveredNotification(identifier) {
    return window.runtime.RemoveDeliveredNotification(identifier);
}

export function RemoveNotification(identifier) {
    return window.runtime.RemoveNotification(identifier);
}
```

---

## FILE: wailsjs\go\core\App.d.ts
```ts
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {core} from '../models';

export function AuthenticateUser(arg1:string,arg2:string):Promise<void>;

export function CheckUpdate():Promise<core.UpdateCheckResult>;

export function DeleteContact(arg1:string,arg2:string):Promise<void>;

export function ExportTableToExcel(arg1:core.ExportPayload):Promise<void>;

export function GetContacts(arg1:string):Promise<Array<core.ContactRecord>>;

export function GetLastUsername():Promise<string>;

export function PerformUpdate(arg1:string):Promise<void>;

export function RestartApp():Promise<void>;

export function SaveContact(arg1:core.ContactRecord):Promise<void>;

export function SaveLastUsername(arg1:string):Promise<void>;

export function TriggerSync():Promise<core.SyncStatus>;

```

---

## FILE: wailsjs\go\core\App.js
```js
// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT

export function AuthenticateUser(arg1, arg2) {
  return window['go']['core']['App']['AuthenticateUser'](arg1, arg2);
}

export function CheckUpdate() {
  return window['go']['core']['App']['CheckUpdate']();
}

export function DeleteContact(arg1, arg2) {
  return window['go']['core']['App']['DeleteContact'](arg1, arg2);
}

export function ExportTableToExcel(arg1) {
  return window['go']['core']['App']['ExportTableToExcel'](arg1);
}

export function GetContacts(arg1) {
  return window['go']['core']['App']['GetContacts'](arg1);
}

export function GetLastUsername() {
  return window['go']['core']['App']['GetLastUsername']();
}

export function PerformUpdate(arg1) {
  return window['go']['core']['App']['PerformUpdate'](arg1);
}

export function RestartApp() {
  return window['go']['core']['App']['RestartApp']();
}

export function SaveContact(arg1) {
  return window['go']['core']['App']['SaveContact'](arg1);
}

export function SaveLastUsername(arg1) {
  return window['go']['core']['App']['SaveLastUsername'](arg1);
}

export function TriggerSync() {
  return window['go']['core']['App']['TriggerSync']();
}

```

---

</codebase>
