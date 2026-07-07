package engineering

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/go-chi/chi/v5"
)

func (m *Module) EngineeringJobConfigList(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	search := strings.ToLower(r.URL.Query().Get("search"))

	client, err := getDatastoreClient(ctx, "df-fs-insights")
	if err != nil {
		return err
	}

	query := datastore.NewQuery("AUTOMATION_V2")
	var entities []datastore.PropertyList
	keys, err := client.GetAll(ctx, query, &entities)
	if err != nil {
		m.l.Printf("Failed fetching AUTOMATION_V2: %v", err)
		return jobConfigListPanel([]JobConfig{}).Render(ctx, w)
	}

	var jobs []JobConfig
	for i, entity := range entities {
		props := make(map[string]interface{})
		for _, prop := range entity {
			props[prop.Name] = prop.Value
		}

		nameStr := ""
		if name, ok := props["name"].(string); ok {
			nameStr = name
		}
		if search != "" && !strings.Contains(strings.ToLower(nameStr), search) {
			continue
		}

		active, _ := props["active"].(bool)
		jtype, _ := props["type"].(string)

		schedule := ""
		if val, ok := props["schedule_tags"].(string); ok {
			schedule = val
		} else if arr, ok := props["schedule_tags"].([]interface{}); ok && len(arr) > 0 {
			if strVal, ok := arr[0].(string); ok {
				schedule = strVal
			}
		}

		jobs = append(jobs, JobConfig{
			EncodedKey: keys[i].Encode(),
			Name:       nameStr,
			Active:     active,
			Type:       jtype,
			Schedule:   schedule,
		})
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Name < jobs[j].Name
	})

	return jobConfigListPanel(jobs).Render(ctx, w)
}

func (m *Module) EngineeringJobConfigDetails(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	jobID := chi.URLParam(r, "id")

	if jobID == "new" {
		return jobConfigDetailsPanel(JobConfig{Active: true}).Render(ctx, w)
	}

	client, err := getDatastoreClient(ctx, "df-fs-insights")
	if err != nil {
		return err
	}

	key, err := datastore.DecodeKey(jobID)
	if err != nil {
		return err
	}

	var props datastore.PropertyList
	err = client.Get(ctx, key, &props)
	if err != nil {
		return err
	}

	properties := make(map[string]interface{})
	for _, prop := range props {
		if prop.Name == "cron_override" {
			if nestedEntity, ok := prop.Value.(*datastore.Entity); ok {
				simpleMap := make(map[string]string)
				for _, nestedProp := range nestedEntity.Properties {
					if val, ok := nestedProp.Value.(string); ok {
						simpleMap[nestedProp.Name] = val
					}
				}
				properties[prop.Name] = simpleMap
				continue
			}
		}

		if prop.Name == "update_source_tables_config" {
			if ustcEntity, ok := prop.Value.(*datastore.Entity); ok {
				ustcMap := make(map[string]interface{})
				for _, ustcProp := range ustcEntity.Properties {
					if ustcProp.Name == "override" {
						simpleMap := make(map[string]string)
						if iList, ok := ustcProp.Value.([]interface{}); ok {
							for _, i := range iList {
								if nestedEntity, ok := i.(*datastore.Entity); ok && nestedEntity != nil {
									for _, nestedProp := range nestedEntity.Properties {
										if val, ok := nestedProp.Value.(string); ok {
											simpleMap[nestedProp.Name] = val
										}
									}
								}
							}
						} else if nestedEntity, ok := ustcProp.Value.(*datastore.Entity); ok && nestedEntity != nil {
							for _, nestedProp := range nestedEntity.Properties {
								if val, ok := nestedProp.Value.(string); ok {
									simpleMap[nestedProp.Name] = val
								}
							}
						}
						ustcMap[ustcProp.Name] = simpleMap
					} else if ustcProp.Name == "exclusions" {
						var values []string
						if exclusionsList, ok := ustcProp.Value.([]interface{}); ok {
							for _, item := range exclusionsList {
								if str, ok := item.(string); ok {
									values = append(values, str)
								}
							}
						}
						ustcMap[ustcProp.Name] = values
					} else {
						ustcMap[ustcProp.Name] = ustcProp.Value
					}
				}
				properties[prop.Name] = ustcMap
				continue
			}
		}

		if prop.Name == "mailing_list" {
			if mlEntity, ok := prop.Value.(*datastore.Entity); ok {
				mlMap := make(map[string]interface{})
				for _, mlProp := range mlEntity.Properties {
					if mlProp.Name == "receipt_emails" {
						if reEntity, ok := mlProp.Value.(*datastore.Entity); ok {
							reMap := make(map[string][]string)
							for _, reProp := range reEntity.Properties {
								if emails, ok := reProp.Value.([]interface{}); ok {
									emailStrs := make([]string, len(emails))
									for i, e := range emails {
										if emailStr, ok := e.(string); ok {
											emailStrs[i] = emailStr
										}
									}
									reMap[reProp.Name] = emailStrs
								}
							}
							mlMap[mlProp.Name] = reMap
						}
					} else {
						mlMap[mlProp.Name] = mlProp.Value
					}
				}
				properties[prop.Name] = mlMap
				continue
			}
		}

		properties[prop.Name] = prop.Value
	}

	strVal := func(k string) string { val, _ := properties[k].(string); return val }
	boolVal := func(k string) bool { val, _ := properties[k].(bool); return val }
	jsonVal := func(k string) string {
		if val, ok := properties[k]; ok {
			b, _ := json.Marshal(val)
			return string(b)
		}
		return "{}"
	}
	jsonArrVal := func(k string) string {
		if val, ok := properties[k]; ok {
			b, _ := json.Marshal(val)
			return string(b)
		}
		return "[]"
	}

	schedule := ""
	if val, ok := properties["schedule_tags"].(string); ok {
		schedule = val
	} else if arr, ok := properties["schedule_tags"].([]interface{}); ok && len(arr) > 0 {
		if strVal, ok := arr[0].(string); ok {
			schedule = strVal
		}
	}

	job := JobConfig{
		EncodedKey:               jobID,
		Name:                     strVal("name"),
		Type:                     strVal("type"),
		Active:                   boolVal("active"),
		Link:                     strVal("link"),
		Repository:               strVal("repository"),
		UniqueTag:                strVal("unique_tag"),
		Schedule:                 schedule,
		CountryAgnostic:          boolVal("country_agnostic"),
		CronDefault:              strVal("cron_default"),
		CountriesOverride:        jsonArrVal("countries_override"),
		CronOverride:             jsonVal("cron_override"),
		UpdateSourceTablesConfig: jsonVal("update_source_tables_config"),
		MailingList:              jsonVal("mailing_list"),
	}

	return jobConfigDetailsPanel(job).Render(ctx, w)
}

func (m *Module) EngineeringJobConfigSave(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	r.ParseForm()

	encodedKey := r.FormValue("id")
	m.l.Printf("Saving Job Config: %s (Name: %s)", encodedKey, r.FormValue("name"))

	client, err := getDatastoreClient(ctx, "df-fs-insights")
	if err != nil {
		return err
	}

	var key *datastore.Key
	if encodedKey == "" {
		key = datastore.IncompleteKey("AUTOMATION_V2", nil)
	} else {
		key, err = datastore.DecodeKey(encodedKey)
		if err != nil {
			http.Error(w, "Invalid entity ID.", http.StatusBadRequest)
			return nil
		}
	}

	var props datastore.PropertyList
	props = append(props, datastore.Property{Name: "name", Value: r.FormValue("name")})
	props = append(props, datastore.Property{Name: "type", Value: r.FormValue("type")})
	props = append(props, datastore.Property{Name: "active", Value: r.FormValue("active") == "true"})
	props = append(props, datastore.Property{Name: "country_agnostic", Value: r.FormValue("country_agnostic") == "true"})

	if r.FormValue("schedule_tags") == "cyclical" {
		props = append(props, datastore.Property{Name: "schedule_tags", Value: []interface{}{"cyclical"}})
	} else {
		props = append(props, datastore.Property{Name: "schedule_tags", Value: ""})
	}

	if r.FormValue("type") == "cloud-function" {
		props = append(props, datastore.Property{Name: "link", Value: r.FormValue("link")})
	} else {
		props = append(props, datastore.Property{Name: "repository", Value: r.FormValue("repository")})
		props = append(props, datastore.Property{Name: "unique_tag", Value: r.FormValue("unique_tag")})
	}

	if r.FormValue("schedule_tags") != "cyclical" {
		props = append(props, datastore.Property{Name: "cron_default", Value: r.FormValue("cron_default")})
	}

	var countriesOverride []interface{}
	if err := json.Unmarshal([]byte(r.FormValue("countries_override_json")), &countriesOverride); err == nil {
		props = append(props, datastore.Property{Name: "countries_override", Value: countriesOverride})
	}

	if r.FormValue("schedule_tags") != "cyclical" {
		var cronMap map[string]string
		if err := json.Unmarshal([]byte(r.FormValue("cron_override_json")), &cronMap); err == nil {
			var nestedProps datastore.PropertyList
			for k, v := range cronMap {
				nestedProps = append(nestedProps, datastore.Property{Name: k, Value: v})
			}
			props = append(props, datastore.Property{Name: "cron_override", Value: &datastore.Entity{Properties: nestedProps}, NoIndex: true})
		}
	}

	if r.FormValue("type") == "dataform" {
		var ustcMap map[string]interface{}
		if err := json.Unmarshal([]byte(r.FormValue("update_source_tables_config_json")), &ustcMap); err == nil {
			var ustcProps datastore.PropertyList
			if isEnabled, ok := ustcMap["is_enabled"].(bool); ok {
				ustcProps = append(ustcProps, datastore.Property{Name: "is_enabled", Value: isEnabled, NoIndex: false})
			}
			var maxStalenessValue int64 = 30
			if maxStaleness, ok := ustcMap["max_staleness"]; ok && maxStaleness != nil {
				if v, ok := maxStaleness.(float64); ok {
					maxStalenessValue = int64(v)
				}
			}
			ustcProps = append(ustcProps, datastore.Property{Name: "max_staleness", Value: maxStalenessValue, NoIndex: false})

			var overrideEntities []interface{}
			if overrideMap, ok := ustcMap["override"].(map[string]interface{}); ok {
				var nestedProps datastore.PropertyList
				for k, v := range overrideMap {
					if vStr, ok := v.(string); ok {
						nestedProps = append(nestedProps, datastore.Property{Name: k, Value: vStr, NoIndex: false})
					}
				}
				if len(nestedProps) > 0 {
					overrideEntities = append(overrideEntities, &datastore.Entity{Properties: nestedProps})
				}
			}
			ustcProps = append(ustcProps, datastore.Property{Name: "override", Value: overrideEntities, NoIndex: true})

			var exclusionValues []interface{}
			if exclusionsList, ok := ustcMap["exclusions"].([]interface{}); ok {
				for _, v := range exclusionsList {
					if vStr, ok := v.(string); ok && vStr != "" {
						exclusionValues = append(exclusionValues, vStr)
					}
				}
			}
			ustcProps = append(ustcProps, datastore.Property{Name: "exclusions", Value: exclusionValues, NoIndex: false})

			props = append(props, datastore.Property{Name: "update_source_tables_config", Value: &datastore.Entity{Properties: ustcProps}, NoIndex: true})
		}
	}

	var mlMap map[string]interface{}
	if err := json.Unmarshal([]byte(r.FormValue("mailing_list_json")), &mlMap); err == nil {
		var mlProps datastore.PropertyList
		if sendEmail, ok := mlMap["send_email"].(bool); ok {
			mlProps = append(mlProps, datastore.Property{Name: "send_email", Value: sendEmail, NoIndex: false})
		}
		if receiptEmails, ok := mlMap["receipt_emails"].(map[string]interface{}); ok {
			var emailProps datastore.PropertyList
			for country, emails := range receiptEmails {
				if emailList, ok := emails.([]interface{}); ok {
					emailProps = append(emailProps, datastore.Property{Name: country, Value: emailList, NoIndex: false})
				}
			}
			if len(emailProps) > 0 {
				mlProps = append(mlProps, datastore.Property{Name: "receipt_emails", Value: &datastore.Entity{Properties: emailProps}, NoIndex: true})
			}
		}
		props = append(props, datastore.Property{Name: "mailing_list", Value: &datastore.Entity{Properties: mlProps}, NoIndex: true})
	}

	if encodedKey == "" {
		_, err = client.Put(ctx, key, &props)
	} else {
		_, err = client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
			var placeholder datastore.PropertyList
			if err := tx.Get(key, &placeholder); err != nil {
				return err
			}
			_, err := tx.Put(key, &props)
			return err
		})
	}

	if err != nil {
		m.l.Printf("Failed to update entity in Datastore: %v", err)
		w.Write([]byte(`<div class="p-4 bg-red-50 text-red-700 rounded font-bold text-sm">Failed to save changes.</div>`))
		return nil
	}

	w.Header().Set("HX-Trigger", "job-saved")
	w.Write([]byte(`<div class="p-4 bg-green-50 text-green-700 rounded font-bold text-sm">Job Configuration Saved!</div>`))
	return nil
}
