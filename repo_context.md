<system_instructions>
# SYSTEM INSTRUCTION / MANIFEST

**ROLE**
You are an elite Senior Developer specializing in Go, HTML, Templ, JavaScript, Tailwind CSS, and a certified Google Cloud (GCP) Enterprise Architect. 

**OBJECTIVE**
Analyze the user's request deeply before responding. Provide the absolute best, most optimal programmatic and architectural solutions, specifically optimized for scalable, cloud-native GCP applications.

**INTERACTION MODES**
Depending on the user's prompt or the level of detail provided, follow the respective mode:
* **[AUTO-MODE]:** If the user provides a detailed specification, assume the best architectural choices and generate the full code immediately per the rules.
* **[DISCOVERY-MODE]:** If the user uses this tag, or if the initial request lacks necessary architectural or UI/UX details, STOP. Do not generate code yet. Ask a concise, bulleted list of high-level technical questions to gather exact requirements. Wait for the user's reply before coding.

**RULES & CONSTRAINTS**
1. **Full Code Delivery:** Once requirements are fully clarified, always provide complete, fully functioning code blocks. Do not use placeholders, summaries, or truncate the code in any way.
2. **Zero Comments:** The outputted code must contain absolutely NO comments. Code should be clean and self-documenting.
3. **Production-Ready:** All code must be secure, performant, and ready for production deployment without requiring hotfixes.
4. **Idiomatic Go & Context:** Enforce strict Go idioms. Handle all errors explicitly (no ignores, no panics), use context propagation (`context.Context`) for all GCP API clients, and ensure absolute goroutine safety.
5. **GCP Datastore Practices:** Use the official `cloud.google.com/go/datastore` client. Ensure proper use of keys, structural tags, and transactional operations where data integrity is required.
6. **BigQuery Optimization:** Use the official `cloud.google.com/go/bigquery` client. Optimize for cost and memory: never use `SELECT *`, use parameterized queries to prevent injection, and always fetch rows via Iterators to stream data rather than loading large datasets entirely into memory.
7. **Future-Proof Monitoring Architecture:** Write highly modular, decoupled code using clean architecture or repository patterns. Group logic by GCP service domain so that future monitoring features (e.g., Cloud Functions, Datastream, Cloud Logging) can be added seamlessly without breaking existing code.
8. **GCP Security & IAM:** Never hardcode credentials. Design for GCP Secret Manager, environment variables, and work under the assumption of Least-Privilege IAM service account roles.
9. **Tailwind CSS Exclusivity:** Use ONLY Tailwind utility classes for styling. Never write custom CSS, inline `<style>` tags, or use `style="..."` attributes.
10. **No Dynamic Tailwind Classes:** Always write full, complete Tailwind class names (e.g., `text-red-500` instead of `text-${color}-500`). Do not construct Tailwind classes programmatically.
11. **Strict Templ Typing:** Define clear Go structs for all Templ component data instead of using loose maps or `interface{}`.

**PROCESS & CLARIFICATION PHASE**
1. **Analyze:** Think critically about the problem step-by-step. Evaluate the proposed architecture, data flow, and GCP client lifecycle patterns.
2. **Clarify (Crucial Step):** Before writing ANY code, assess if the user's request contains enough detail to produce a production-ready, highly optimized solution. If the requirements are ambiguous, missing GCP architectural context, or lack specific UI/UX details, **do not write code yet**.
3. **Ask:** Output a structured list of clarifying questions grouped by domain (e.g., GCP Architecture, Go Backend Logic, Frontend/Templ).
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

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"luminalex/core"
)

//go:embed assets
var assets embed.FS

func main() {
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
		Title:  "LuminaLex",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
			Middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/assets/") {
						assetHandler.ServeHTTP(w, r)
						return
					}
					if r.URL.Path == "/" || r.URL.Path == "/login" || r.URL.Path == "/home" || r.URL.Path == "/contacts" {
						handler.ServeHTTP(w, r)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		OnStartup: app.Startup,
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

const schemas = {
    search: [],
    banks: ['Bank Name', 'Branch', 'Physical Address', 'Contact Name', 'Title/Rank', 'Email', 'Phone', 'Fax', 'Notes'],
    masters: ['Masters Office', 'Physical Address', 'Department/Division', 'Contact Name', 'Rank/Title', 'File Number Allocation', 'Email', 'Phone', 'Notes'],
    sheriffs: ['Sheriff Office', 'Sheriff Name', 'Deputy Name', 'Physical Address', 'Jurisdiction', 'Email', 'Phone', 'Fax', 'Notes'],
    magistrates: ['Court Name', 'Physical Address', 'Contact Name', 'Title/Rank', 'Email', 'Phone', 'Fax', 'Court Citation', 'Notes'],
    highcourts: ['Court Name', 'Physical Address', 'Contact Name', 'Title/Rank', 'Email', 'Phone', 'Fax', 'Court Citation', 'Notes'],
    lawfirms: ['Firm Name', 'Physical Address', 'Contact Name', 'Email', 'Phone', 'Notes']
};

const initialMagistrates = [
    ['Westonaria Magistrate Court', '36 President Steyn Street, Westonaria, Gauteng, South Africa, 1780', '', '', '', '(011) 753 2251', '(011) 753 3725', '', ''],
    ['Vereeniging Magistrate Court', 'Cnr Leslie & Beaconsfiels Ave., Vereeniging, Gauteng, South Africa, 1939', '', '', '', '(016) 422 0071', '(016) 422 0604', '', ''],
    ['Vanderbijlpark Magistrate Court', 'Cnr General Hertzog & FW Beyers street, Vanderbijlpark, Gauteng, South Africa, 1911', '', '', '', '(016) 933 4351', '(016) 933 3377', '', ''],
    ['Cape Town Magistrate Court', 'Cnr Claedon & Parade Street, Cape Town, Western Cape, South Africa, 8001', 'Mrs Macdelein W Cloete', 'Court Manager', 'GPetersen@justice.gov.za', '(021) 401 1544', '(021) 461 1948', '', 'Deals with Civil Trials'],
    ['Bellville Magistrate Court', 'Cnr Voortrekker Road & Landdros street, Bellville, Western Cape, South Africa, 7535', '', '', 'claster@justice.gov.za', '(021) 950 7700', '(021) 945 2255', 'IN THE MAGISTRATES’ COURT FOR THE DISTRICT OF THE CITY OF CAPE TOWN, SUB-DISTRICT BELLVILLE, HELD AT BELLVILLE', '']
];

let db = JSON.parse(localStorage.getItem('luminalex_db'));
if (!db || !db.magistrates.length || !db.magistrates[0].id) {
    db = {
        banks: (db && db.banks ? db.banks : []).map((row, index) => ({ id: `banks-${index}`, fields: row })),
        masters: (db && db.masters ? db.masters : []).map((row, index) => ({ id: `masters-${index}`, fields: row })),
        sheriffs: (db && db.sheriffs ? db.sheriffs : []).map((row, index) => ({ id: `sheriffs-${index}`, fields: row })),
        magistrates: initialMagistrates.map((row, index) => ({ id: `magistrates-${index}`, fields: row })),
        highcourts: (db && db.highcourts ? db.highcourts : []).map((row, index) => ({ id: `highcourts-${index}`, fields: row })),
        lawfirms: (db && db.lawfirms ? db.lawfirms : []).map((row, index) => ({ id: `lawfirms-${index}`, fields: row }))
    };
    localStorage.setItem('luminalex_db', JSON.stringify(db));
}


function saveDB() {
    localStorage.setItem('luminalex_db', JSON.stringify(db));
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
        const data = db[type];
        const filtered = data.filter(row => row.fields.some(cell => cell.toLowerCase().includes(query)));

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

window.addRecord = function(type) {
    const cols = schemas[type];
    const newRow = { id: `${type}-${db[type].length}`, fields: []};
    for(let i=0; i<cols.length; i++) {
        const val = document.getElementById(`input-${type}-${i}`).value;
        newRow.fields.push(val);
    }
    db[type].push(newRow);
    saveDB();
    document.getElementById('tab-content').innerHTML = renderTable(type);
}

window.deleteRecord = function(type, idx) {
    if(confirm("Are you sure you want to delete this record?")) {
        db[type].splice(idx, 1);
        saveDB();
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
    const cols = schemas[type];
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

window.saveRecord = function(type, idx) {
    const newFields = [];
    const cols = schemas[type];
    for(let i=0; i<cols.length; i++) {
        const val = document.getElementById(`edit-${type}-${idx}-${i}`).value;
        newFields.push(val);
    }
    db[type][idx].fields = newFields;
    saveDB();
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


document.addEventListener('DOMContentLoaded', () => {
    const sidebarExpanded = document.getElementById('sidebar-expanded');
    const sidebarCollapsed = document.getElementById('sidebar-collapsed');
    const btnClose = document.getElementById('btn-close');
    const btnOpen = document.getElementById('btn-open');
    const searchInput = document.getElementById('menu-search');
    const menuItems = document.querySelectorAll('.menu-item');
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContent = document.getElementById('tab-content');

    const isCollapsed = localStorage.getItem('luminalex_sidebar_collapsed') === 'true';
    if (isCollapsed) {
        sidebarExpanded.classList.add('hidden');
        sidebarCollapsed.classList.remove('hidden');
    }

    btnClose.addEventListener('click', () => {
        sidebarExpanded.classList.add('hidden');
        sidebarCollapsed.classList.remove('hidden');
        localStorage.setItem('luminalex_sidebar_collapsed', 'true');
    });

    btnOpen.addEventListener('click', () => {
        sidebarCollapsed.classList.add('hidden');
        sidebarExpanded.classList.remove('hidden');
        localStorage.setItem('luminalex_sidebar_collapsed', 'false');
    });

    searchInput.addEventListener('input', (e) => {
        const term = e.target.value.toLowerCase();
        menuItems.forEach(item => {
            const text = item.querySelector('.menu-text').textContent.toLowerCase();
            item.style.display = text.includes(term) ? 'flex' : 'none';
        });
    });

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

    if(tabContent) {
        tabContent.innerHTML = renderSearch();
        executeSearch();
    }
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

import "context"

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

```

---

## FILE: core\handler.go
```go
package core

import (
	"net/http"

	"luminalex/views"
)

func NewHandler(app *App) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		err := views.Login(views.LoginData{}).Render(r.Context(), w)
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

		if user == "admin" && pass == "admin" {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		err := views.Login(views.LoginData{Error: "Invalid credentials"}).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering login page failed", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		err := views.Home(views.HomeData{Title: "LuminaLex"}).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering home page failed", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		err := views.Contacts().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering contacts page failed", http.StatusInternalServerError)
		}
	})

	return mux
}

```

---

## FILE: views\contacts.templ
```templ
package views

templ Contacts() {
	@Layout("LuminaLex - Contacts", true) {
		<div class="z-10 flex-1 flex flex-col p-8 h-full overflow-hidden bg-black/60 backdrop-blur-sm">
			<div class="flex gap-4 border-b border-[#a68a56]/50 mb-6 w-full" id="tabs-container">
				<button class="tab-btn active px-4 py-2 text-sm uppercase tracking-widest text-[#a68a56] border-b-2 border-[#a68a56] transition-colors" data-target="search">Search</button>
				<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-colors" data-target="banks">Banks</button>
				<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-colors" data-target="masters">Master's Office</button>
				<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-colors" data-target="sheriffs">Sheriffs</button>
				<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-colors" data-target="magistrates">Magistrate Courts</button>
				<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-colors" data-target="highcourts">High Courts</button>
				<button class="tab-btn px-4 py-2 text-sm uppercase tracking-widest text-[#f9f8f6] hover:text-[#a68a56] border-b-2 border-transparent transition-colors" data-target="lawfirms">Law Firms</button>
			</div>
			<div id="tab-content" class="flex-1 overflow-auto w-full font-sans">
			</div>
		</div>
		<script src="/assets/contacts.js"></script>
	}
}
```

---

## FILE: views\data.go
```go
package views

type LoginData struct {
	Error string
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

```

---

## FILE: views\home.templ
```templ
package views

templ Home(data HomeData) {
	@Layout(data.Title, true) {
		<div class="flex flex-col items-center justify-center gap-6 text-center">
			<h1 class="text-6xl font-black text-[#a68a56] tracking-tighter drop-shadow-lg">LuminaLex</h1>
			<p class="text-xl text-[#f9f8f6]/80 font-medium">Welcome to the central command center.</p>
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
			<script src="https://unpkg.com/htmx.org@1.9.10" integrity="sha384-D1Kt99CQMDuVetoL1lrYwg5t+9QdHe7NLX/SoJYkXDFfX37iInKRy5xLSi8nO7UC" crossorigin="anonymous"></script>
		</head>
		<body class="w-screen h-screen overflow-hidden m-0 p-0 text-[#f9f8f6] bg-black font-serif">
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
									<svg class="w-4 h-4 absolute left-3 top-2.5 text-[#a68a56]" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
								</div>
							</div>
							<nav class="flex flex-col px-4 gap-2">
								<a href="/home" class="menu-item flex items-center gap-4 text-[#f9f8f6] border-b-[3px] border-transparent hover:border-[#a68a56] hover:text-[#a68a56] p-3 text-sm uppercase tracking-widest transition-all">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
									<span class="menu-text">Clients</span>
								</a>
								<a href="/home" class="menu-item flex items-center gap-4 text-[#f9f8f6] border-b-[3px] border-transparent hover:border-[#a68a56] hover:text-[#a68a56] p-3 text-sm uppercase tracking-widest transition-all">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
									<span class="menu-text">Services</span>
								</a>
								<a href="/contacts" class="menu-item flex items-center gap-4 text-[#f9f8f6] border-b-[3px] border-transparent hover:border-[#a68a56] hover:text-[#a68a56] p-3 text-sm uppercase tracking-widest transition-all">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
									<span class="menu-text">Contacts</span>
								</a>
							</nav>
						</div>
						<div class="p-6 border-t border-[#a68a56]/30 bg-[#0b1d2e]">
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
								<a href="/home" class="text-[#f9f8f6] hover:text-[#a68a56] transition-colors" title="Clients">
									<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
								</a>
								<a href="/home" class="text-[#f9f8f6] hover:text-[#a68a56] transition-colors" title="Services">
									<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
								</a>
								<a href="/contacts" class="text-[#f9f8f6] hover:text-[#a68a56] transition-colors" title="Contacts">
									<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>
								</a>
							</nav>
						</div>
						<a href="/" class="text-[#d4c299] hover:text-[#a68a56] transition-colors">
							<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"></path></svg>
						</a>
					</div>
				}
				<div class="z-10 flex-1 flex flex-col p-8 h-full overflow-hidden bg-black/60 backdrop-blur-sm">
					{ children... }
				</div>
			</div>
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
						btnClose.addEventListener('click', () => {
							sidebarExpanded.classList.add('hidden');
							sidebarCollapsed.classList.remove('hidden');
							localStorage.setItem('luminalex_sidebar_collapsed', 'true');
						});
						btnOpen.addEventListener('click', () => {
							sidebarCollapsed.classList.add('hidden');
							sidebarExpanded.classList.remove('hidden');
							localStorage.setItem('luminalex_sidebar_collapsed', 'false');
						});
						searchInput.addEventListener('input', (e) => {
							const term = e.target.value.toLowerCase();
							menuItems.forEach(item => {
								const text = item.querySelector('.menu-text').textContent.toLowerCase();
								item.style.display = text.includes(term) ? 'flex' : 'none';
							});
						});
					});
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
		<div class="w-full h-full flex items-center justify-center">
			<form action="/login" method="POST" class="bg-[#0b1d2e]/80 border border-[#a68a56]/30 p-10 rounded-none shadow-2xl w-full max-w-md flex flex-col gap-6 backdrop-blur-sm">
				<div class="flex flex-col items-center justify-center gap-2">
					<img src="/assets/logo_text.png" class="w-48"/>
				</div>
				if data.Error != "" {
					<div class="bg-red-500/10 text-red-400 border border-red-500/20 p-3 text-sm text-center font-sans">
						{ data.Error }
					</div>
				}
				<div class="flex flex-col gap-2 font-sans">
					<label for="username" class="text-xs uppercase tracking-widest text-[#a68a56]">Username</label>
					<input type="text" id="username" name="username" value="admin" class="bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-sm focus:outline-none focus:border-[#a68a56] transition-colors"/>
				</div>
				<div class="flex flex-col gap-2 font-sans">
					<label for="password" class="text-xs uppercase tracking-widest text-[#a68a56]">Password</label>
					<input type="password" id="password" name="password" value="admin" class="bg-black/40 border border-[#a68a56]/30 text-[#f9f8f6] p-3 text-sm focus:outline-none focus:border-[#a68a56] transition-colors"/>
				</div>
				<button type="submit" class="bg-[#0b1d2e] border border-[#a68a56] text-[#a68a56] hover:bg-[#a68a56] hover:text-[#0b1d2e] font-bold py-3 px-4 uppercase tracking-widest text-sm transition-colors mt-4 shadow-lg">
					Login
				</button>
			</form>
		</div>
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

</codebase>
