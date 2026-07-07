package user_management

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"ujuzi_reloaded/app/backend/auth"
	"ujuzi_reloaded/app/backend/collection"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Legacy Admin Identity for Directory Delegation
const AssumeIdentity = "rohan@unifi.credit"

type Module struct {
	l            *log.Logger
	sessionStore auth.Store
	coreDBs      map[string]*sql.DB
}

func NewModule(l *log.Logger, sessionStore auth.Store, coreDBs map[string]*sql.DB) *Module {
	return &Module{
		l:            l,
		sessionStore: sessionStore,
		coreDBs:      coreDBs,
	}
}

// getAvailableUnibosGroups dynamically fetches deduplicated Unibos group names across all countries
func (m *Module) getAvailableUnibosGroups(ctx context.Context) []string {
	groupSet := make(map[string]struct{})
	iter := m.sessionStore.Db().Collection("usergroups").Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			m.l.Printf("ERROR fetching unibos groups: %v", err)
			break
		}
		if name, ok := doc.Data()["Name"].(string); ok && name != "" {
			groupSet[name] = struct{}{}
		}
	}

	var groups []string
	for g := range groupSet {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	return groups
}

func (m *Module) AdminLoader(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var users []auth.User
	iter := m.sessionStore.Db().Collection(collection.Users).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		var u auth.User
		doc.DataTo(&u)
		u.ID = doc.Ref.ID
		users = append(users, u)
	}

	groups := m.getAllGroupsSorted(r)
	dashboards := m.getAllDashboards(r)
	unibosGroups := m.getAvailableUnibosGroups(ctx)

	return adminManagementPage(users, groups, dashboards, unibosGroups).Render(ctx, w)
}

func (m *Module) AdminUserDetails(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	doc, err := m.sessionStore.Db().Collection(collection.Users).Doc(userID).Get(ctx)
	if err != nil {
		return err
	}
	var u auth.User
	doc.DataTo(&u)
	u.ID = doc.Ref.ID

	groups := m.getAllGroupsSorted(r)
	return userDetailsPanel(u, groups).Render(ctx, w)
}

func (m *Module) AdminSaveUserGroups(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	r.ParseForm()
	selectedGroups := r.Form["groups"]
	selectedTools := r.Form["tools"]

	if selectedGroups == nil {
		selectedGroups = []string{}
	}
	if selectedTools == nil {
		selectedTools = []string{}
	}

	_, err := m.sessionStore.Db().Collection(collection.Users).Doc(userID).Update(ctx, []firestore.Update{
		{Path: "Groups", Value: selectedGroups},
		{Path: "Tools", Value: selectedTools},
	})
	if err != nil {
		return err
	}

	m.sessionStore.InvalidateUserCache(userID)

	w.Header().Set("HX-Trigger", "user-saved")
	w.Write([]byte(`<div class="p-4 bg-green-50 text-green-700 rounded mb-4 font-bold text-sm">Assignments updated successfully!</div>`))
	return nil
}

func (m *Module) AdminCreateGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	r.ParseForm()

	level := r.FormValue("level")
	if !m.isLevelUnique(r, level, "") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="text-red-500 text-xs font-bold mt-2">Error: Level must be unique!</div>`))
		return nil
	}

	newGroup := auth.Group{
		Name:         r.FormValue("name"),
		Level:        level,
		Dashboards:   []string{},
		Tools:        []string{},
		UnibosGroups: []string{},
	}

	ref, _, err := m.sessionStore.Db().Collection("groups").Add(ctx, newGroup)
	if err != nil {
		return err
	}
	newGroup.ID = ref.ID

	w.Header().Set("HX-Refresh", "true")
	return nil
}

func (m *Module) AdminGroupDetails(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	groupID := chi.URLParam(r, "id")

	doc, err := m.sessionStore.Db().Collection("groups").Doc(groupID).Get(ctx)
	if err != nil {
		return err
	}
	var g auth.Group
	doc.DataTo(&g)
	g.ID = doc.Ref.ID

	dashboards := m.getAllDashboards(r)
	unibosGroups := m.getAvailableUnibosGroups(ctx)
	return groupDetailsPanel(g, dashboards, unibosGroups).Render(ctx, w)
}

func (m *Module) AdminUpdateGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	groupID := chi.URLParam(r, "id")
	r.ParseForm()

	level := r.FormValue("level")
	if !m.isLevelUnique(r, level, groupID) {
		w.Write([]byte(`<div class="p-4 bg-red-50 text-red-700 rounded mb-4 font-bold text-sm">Error: Level must be unique!</div>`))
		return nil
	}

	dashboards := r.Form["dashboards"]
	if dashboards == nil {
		dashboards = []string{}
	}

	tools := r.Form["tools"]
	if tools == nil {
		tools = []string{}
	}

	unibosGroups := r.Form["unibosGroups"]
	if unibosGroups == nil {
		unibosGroups = []string{}
	}

	updates := []firestore.Update{
		{Path: "name", Value: r.FormValue("name")},
		{Path: "level", Value: level},
		{Path: "dashboards", Value: dashboards},
		{Path: "tools", Value: tools},
		{Path: "unibosGroups", Value: unibosGroups},
	}

	_, err := m.sessionStore.Db().Collection("groups").Doc(groupID).Update(ctx, updates)
	if err != nil {
		return err
	}

	m.sessionStore.ClearAllUserCache()

	w.Header().Set("HX-Refresh", "true")
	return nil
}

func (m *Module) AdminDeleteGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	groupID := chi.URLParam(r, "id")
	_, err := m.sessionStore.Db().Collection("groups").Doc(groupID).Delete(ctx)
	if err != nil {
		return err
	}

	m.sessionStore.ClearAllUserCache()

	w.Header().Set("HX-Refresh", "true")
	return nil
}

func (m *Module) AdminSyncUsers(w http.ResponseWriter, r *http.Request) error {
	m.l.Printf("Admin manual trigger: Starting directory user sync protocol")

	go func() {
		bgCtx := context.Background()
		if err := m.runGoogleWorkspaceSync(bgCtx); err != nil {
			m.l.Printf("ERROR: Google Workspace Sync Failed: %v", err)
		}
	}()

	w.Write([]byte(`<div class="p-3 bg-green-100 text-green-800 rounded text-xs font-bold uppercase tracking-wider animate-pulse">User sync pipeline launched in background...</div>`))
	return nil
}

func (m *Module) AdminSyncUserGroups(w http.ResponseWriter, r *http.Request) error {
	m.l.Printf("Admin manual trigger: Starting internal relational mapping sync across core DB nodes")

	go func() {
		bgCtx := context.Background()
		countries := []string{"za", "ke", "ug", "tz", "zm"}

		for _, cc := range countries {
			if err := m.syncUserGroupsForCountry(bgCtx, cc); err != nil {
				m.l.Printf("ERROR syncing user groups for %s: %v", cc, err)
			}
		}
		m.l.Println("Database group replication routine cleanly shut down.")
	}()

	w.Write([]byte(`<div class="p-3 bg-green-100 text-green-800 rounded text-xs font-bold uppercase tracking-wider animate-pulse">Group synchronization task queued and running in background...</div>`))
	return nil
}

func (m *Module) syncUserGroupsForCountry(ctx context.Context, cc string) error {
	m.l.Printf("Starting user group sync for country: %s", cc)

	countryDb, ok := m.coreDBs[cc]
	if !ok || countryDb == nil {
		return fmt.Errorf("database connection for country %s not found in module injection", cc)
	}

	const getGroupsQuery = `SELECT id, group_name FROM sec_groups;`
	rows, err := countryDb.QueryContext(ctx, getGroupsQuery)
	if err != nil {
		return fmt.Errorf("querying sec_groups for %s failed: %w", cc, err)
	}
	defer rows.Close()

	groupsFromDB := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return fmt.Errorf("scanning sec_group row failed: %w", err)
		}
		groupsFromDB[id] = name
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error during rows iteration for sec_groups: %w", err)
	}
	m.l.Printf("Found %d groups in MySQL sec_groups for %s", len(groupsFromDB), cc)

	existingGroupsFirestore := make(map[string]string)
	iter := m.sessionStore.Db().Collection("usergroups").Where("CC", "==", cc).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to retrieve existing usergroups from Firestore: %w", err)
		}
		unibosID, _ := doc.Data()["UnibosID"].(string)
		if unibosID != "" {
			existingGroupsFirestore[unibosID] = doc.Ref.ID
		}
	}

	batch := m.sessionStore.Db().Batch()
	newGroupCount := 0
	deletedGroupCount := 0

	for unibosID, name := range groupsFromDB {
		if _, exists := existingGroupsFirestore[unibosID]; !exists {
			newID := m.sessionStore.Db().Collection("temp").NewDoc().ID
			docRef := m.sessionStore.Db().Collection("usergroups").Doc(newID)

			batch.Set(docRef, map[string]interface{}{
				"ID":              newID,
				"Name":            name,
				"UnibosID":        unibosID,
				"CC":              cc,
				"FunctionalAreas": map[string]bool{},
			})
			newGroupCount++
		}
	}

	for unibosID, docID := range existingGroupsFirestore {
		if _, exists := groupsFromDB[unibosID]; !exists {
			docRef := m.sessionStore.Db().Collection("usergroups").Doc(docID)
			batch.Delete(docRef)
			deletedGroupCount++
		}
	}

	if newGroupCount > 0 || deletedGroupCount > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit usergroup changes to Firestore: %w", err)
		}
		m.l.Printf("Successfully created %d and deleted %d usergroups for %s", newGroupCount, deletedGroupCount, cc)
	} else {
		m.l.Printf("No usergroups to create or delete for %s. Database is in sync.", cc)
	}

	return nil
}

func (m *Module) runGoogleWorkspaceSync(ctx context.Context) error {
	serviceAccB64 := os.Getenv("SERVICE_ACC")
	if serviceAccB64 == "" {
		return errors.New("SERVICE_ACC environment variable is missing for directory sync")
	}

	saBytes, err := base64.URLEncoding.DecodeString(serviceAccB64)
	if err != nil {
		return fmt.Errorf("failed to decode service account: %w", err)
	}

	config, err := google.JWTConfigFromJSON(
		saBytes,
		admin.AdminDirectoryUserReadonlyScope,
	)
	if err != nil {
		return fmt.Errorf("jwt config error: %w", err)
	}

	config.Subject = AssumeIdentity
	srv, err := admin.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx)))
	if err != nil {
		return fmt.Errorf("admin service error: %w", err)
	}

	m.l.Println("Syncing Google Workspace users and processing Ujuzi group mappings...")
	userCol := m.sessionStore.Db().Collection(collection.Users)
	now := time.Now().Unix()
	googleUserIds := map[string]struct{}{}
	userCount := 0

	// 1. Fetch Ujuzi Groups to build the Unibos mapping & identify the Administrator group
	unibosToUjuzi := make(map[string][]string)
	adminGroupIDs := make(map[string]bool)

	gIter := m.sessionStore.Db().Collection("groups").Documents(ctx)
	for {
		gDoc, err := gIter.Next()
		if err != nil {
			break
		}
		var g auth.Group
		if err := gDoc.DataTo(&g); err == nil {
			gID := gDoc.Ref.ID
			if strings.EqualFold(strings.TrimSpace(g.Name), "administrator") {
				adminGroupIDs[gID] = true
			}
			for _, ubGrp := range g.UnibosGroups {
				unibosToUjuzi[ubGrp] = append(unibosToUjuzi[ubGrp], gID)
			}
		}
	}

	// 2. Fetch all Unibos users and their groups from Core DBs
	userUnibosGroupsMap := make(map[string][]string)
	query := `
		SELECT sa.email_address, sg.group_name
		FROM sec_tenant_users stu
		INNER JOIN sec_accounts sa ON sa.id = stu.realm_account_id
		INNER JOIN sec_group_members sgm ON stu.id = sgm.tenant_user_id 
		INNER JOIN sec_groups sg ON sgm.user_group_id = sg.id 
		WHERE stu.active = 1 AND sa.disabled = 0;
	`
	for cc, db := range m.coreDBs {
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			m.l.Printf("Warning: failed to fetch user groups for %s: %v", cc, err)
			continue
		}
		for rows.Next() {
			var email, groupName string
			if err := rows.Scan(&email, &groupName); err == nil {
				email = strings.ToLower(strings.TrimSpace(email))
				userUnibosGroupsMap[email] = append(userUnibosGroupsMap[email], groupName)
			}
		}
		rows.Close()
	}

	err = srv.Users.List().Customer("my_customer").Pages(ctx, func(u *admin.Users) error {
		userCount += len(u.Users)
		for _, gu := range u.Users {
			googleUserIds[gu.Id] = struct{}{}

			var department, office string
			var defaultCountryCode string
			var countryCodes []string
			countryCodeSet := make(map[string]bool)

			// Safely parse custom org fields (Department)
			if orgs, ok := gu.Organizations.([]interface{}); ok && len(orgs) > 0 {
				if orgMap, ok := orgs[0].(map[string]interface{}); ok {
					if d, ok := orgMap["department"].(string); ok {
						department = d
					}
				}
			}

			// Safely parse Addresses to extract multiple countries & primary office
			if addrs, ok := gu.Addresses.([]interface{}); ok && len(addrs) > 0 {
				for _, addrIf := range addrs {
					if addrMap, ok := addrIf.(map[string]interface{}); ok {
						if a, ok := addrMap["formatted"].(string); ok {
							if office == "" {
								office = a // Keep the first formatted address as primary office
							}
							parts := strings.Split(a, ",")
							if len(parts) > 1 {
								cName := strings.TrimSpace(parts[len(parts)-1])
								cCode := mapCountryNameToCode(cName)
								if cCode != "" {
									if defaultCountryCode == "" {
										defaultCountryCode = cCode // Set first matched country as default
									}
									if !countryCodeSet[cCode] {
										countryCodeSet[cCode] = true
										countryCodes = append(countryCodes, cCode)
									}
								}
							}
						}
					}
				}
			}

			email := strings.ToLower(strings.TrimSpace(gu.PrimaryEmail))

			// 3. Calculate Synced Groups based on mapping
			calculatedUjuziGroups := make(map[string]struct{})
			for _, ubGrp := range userUnibosGroupsMap[email] {
				if ujuziGroupIDs, ok := unibosToUjuzi[ubGrp]; ok {
					for _, uID := range ujuziGroupIDs {
						// Exclude Administrator group from automated processing
						if !adminGroupIDs[uID] {
							calculatedUjuziGroups[uID] = struct{}{}
						}
					}
				}
			}

			iter := userCol.Where("GoogleID", "==", gu.Id).Limit(1).Documents(ctx)
			doc, err := iter.Next()

			if err != nil && errors.Is(err, iterator.Done) {
				if gu.Archived || gu.Suspended {
					continue
				}

				var finalGroups []string
				for g := range calculatedUjuziGroups {
					finalGroups = append(finalGroups, g)
				}

				uid := m.sessionStore.Db().Collection("temp").NewDoc().ID
				_, err = userCol.Doc(uid).Set(ctx, map[string]interface{}{
					"ID":             uid,
					"FirstName":      gu.Name.GivenName,
					"LastName":       gu.Name.FamilyName,
					"Email":          gu.PrimaryEmail,
					"Name":           gu.Name.DisplayName,
					"GoogleID":       gu.Id,
					"OrgUnit":        department,
					"Address":        office,
					"Thumbnail":      gu.ThumbnailPhotoUrl,
					"LastSyncTime":   now,
					"IsAdmin":        false,
					"DefaultCountry": defaultCountryCode,
					"Countries":      countryCodes,
					"Groups":         finalGroups,
					"Tools":          []string{},
				})
				if err != nil {
					return err
				}
			} else if err == nil {
				uid := doc.Ref.ID
				if gu.Archived || gu.Suspended {
					m.l.Printf("Removing user %q (suspended or archived).", gu.PrimaryEmail)
					userCol.Doc(uid).Delete(ctx)
					continue
				}

				// Preserve the Admin group if the existing user manually has it
				if gArr, ok := doc.Data()["Groups"].([]interface{}); ok {
					for _, g := range gArr {
						if gStr, ok := g.(string); ok {
							if adminGroupIDs[gStr] {
								calculatedUjuziGroups[gStr] = struct{}{}
							}
						}
					}
				}

				var finalGroups []string
				for g := range calculatedUjuziGroups {
					finalGroups = append(finalGroups, g)
				}

				updates := map[string]interface{}{
					"FirstName":    gu.Name.GivenName,
					"LastName":     gu.Name.FamilyName,
					"Email":        gu.PrimaryEmail,
					"Name":         gu.Name.DisplayName,
					"OrgUnit":      department,
					"Address":      office,
					"Thumbnail":    gu.ThumbnailPhotoUrl,
					"LastSyncTime": now,
					"Groups":       finalGroups,
				}

				if defaultCountryCode != "" {
					updates["DefaultCountry"] = defaultCountryCode
				}

				if len(countryCodes) > 0 {
					var ccIfaces []interface{}
					for _, cc := range countryCodes {
						ccIfaces = append(ccIfaces, cc)
					}
					updates["Countries"] = firestore.ArrayUnion(ccIfaces...)
				}

				_, err = userCol.Doc(uid).Set(ctx, updates, firestore.MergeAll)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Clean up stale users not found in Google
	ujuziUsers, err := userCol.Documents(ctx).GetAll()
	if err == nil {
		for _, ujuziUser := range ujuziUsers {
			if gid, ok := ujuziUser.Data()["GoogleID"].(string); ok {
				if _, exists := googleUserIds[gid]; !exists {
					m.l.Printf("Removing user (no longer in Workspace)")
					userCol.Doc(ujuziUser.Ref.ID).Delete(ctx)
				}
			}
		}
	}

	m.l.Printf("Successfully synced %d users from Google Workspace.", userCount)
	return nil
}

func (m *Module) getAllGroupsSorted(r *http.Request) []auth.Group {
	ctx := r.Context()
	var groups []auth.Group
	gIter := m.sessionStore.Db().Collection("groups").Documents(ctx)
	for {
		doc, err := gIter.Next()
		if err != nil {
			break
		}
		var g auth.Group
		doc.DataTo(&g)
		g.ID = doc.Ref.ID
		groups = append(groups, g)
	}

	sort.Slice(groups, func(i, j int) bool {
		li, _ := strconv.Atoi(groups[i].Level)
		lj, _ := strconv.Atoi(groups[j].Level)
		return li < lj
	})
	return groups
}

func (m *Module) isLevelUnique(r *http.Request, level string, ignoreID string) bool {
	ctx := r.Context()
	iter := m.sessionStore.Db().Collection("groups").Where("level", "==", level).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		if doc.Ref.ID != ignoreID {
			return false
		}
	}
	return true
}

func (m *Module) getAllDashboards(r *http.Request) []auth.Dashboard {
	ctx := r.Context()
	var dashes []auth.Dashboard
	iter := m.sessionStore.Db().Collection("dashboards").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var d auth.Dashboard
		doc.DataTo(&d)
		d.ID = doc.Ref.ID
		dashes = append(dashes, d)
	}
	return dashes
}

// mapCountryNameToCode safely maps standard country names to the internal Ujuzi code schema
func mapCountryNameToCode(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "south africa":
		return "za"
	case "kenya":
		return "ke"
	case "uganda":
		return "ug"
	case "tanzania":
		return "tz"
	case "zambia":
		return "zm"
	default:
		return ""
	}
}
