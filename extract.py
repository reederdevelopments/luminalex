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