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
