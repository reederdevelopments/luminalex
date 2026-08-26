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

	// --- MATTERS ---
	mux.HandleFunc("/matters", func(w http.ResponseWriter, r *http.Request) {
		currentUser := app.GetLastUsername()
		coreClients, _ := app.store.GetClients(r.Context())
		coreServices, _ := app.store.GetAllServices(r.Context())

		var viewClients []views.Client
		for _, c := range coreClients {
			viewClients = append(viewClients, views.Client{
				ID:                c.ID,
				FirstName:         c.FirstName,
				MiddleName:        c.MiddleName,
				LastName:          c.LastName,
				IDNumber:          c.IDNumber,
				JurisdictionType:  c.JurisdictionType,
				Occupation:        c.Occupation,
				EmployerName:      c.EmployerName,
				EmployerAddressL1: c.EmployerAddressL1,
			})
		}

		var viewServices []views.Service
		for _, s := range coreServices {
			viewServices = append(viewServices, views.Service{
				ID:           s.ID,
				ServiceType:  s.ServiceType,
				Description:  s.Description,
				StandardRate: s.StandardRate,
				DurationUnit: s.DurationUnit,
			})
		}

		err := views.Matters(viewClients, viewServices, currentUser).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "rendering matters page failed", http.StatusInternalServerError)
		}
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

	mux.HandleFunc("/api/matters/save", func(w http.ResponseWriter, r *http.Request) {
		var m Matter
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if m.Reference == "" {
			m.Reference = app.GenerateMatterReference(r.Context())
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

	// --- MATTER NOTES ---
	mux.HandleFunc("/api/matters/notes", func(w http.ResponseWriter, r *http.Request) {
		matterID := r.URL.Query().Get("matter_id")
		notes, err := app.store.GetMatterNotes(r.Context(), matterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(notes)
	})

	mux.HandleFunc("/api/matters/notes/save", func(w http.ResponseWriter, r *http.Request) {
		var n MatterNote
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.store.SaveMatterNote(r.Context(), n); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() { _ = app.syncEngine.PerformSync(context.Background()) }()
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/matters/notes/delete", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = app.store.DeleteMatterNote(r.Context(), req.ID)
		go func() { _ = app.syncEngine.PerformSync(context.Background()) }()
		w.WriteHeader(http.StatusOK)
	})

	// --- MATTER SERVICES ---
	mux.HandleFunc("/api/matters/services", func(w http.ResponseWriter, r *http.Request) {
		matterID := r.URL.Query().Get("matter_id")
		svcs, err := app.store.GetMatterServices(r.Context(), matterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(svcs)
	})

	mux.HandleFunc("/api/matters/services/save", func(w http.ResponseWriter, r *http.Request) {
		var svc MatterService
		if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.store.SaveMatterService(r.Context(), svc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() { _ = app.syncEngine.PerformSync(context.Background()) }()
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/matters/services/delete", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = app.store.DeleteMatterService(r.Context(), req.ID)
		go func() { _ = app.syncEngine.PerformSync(context.Background()) }()
		w.WriteHeader(http.StatusOK)
	})

	return mux
}
