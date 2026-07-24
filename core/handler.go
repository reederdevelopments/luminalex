package core

import (
	"encoding/json"
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

	return mux
}
