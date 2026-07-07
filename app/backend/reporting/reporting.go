package reporting

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

type FunctionArea struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Hidden      bool
}

type Report struct {
	ID             string
	Type           Type
	Link           string
	Title          string
	Description    string
	DateAdded      int64
	CC             string
	FunctionalArea string
	URLCountryCode string
	BranchCode     string
	ConsultantCode string
	Architect      string
	Builders       string
	BusinessOwners map[string]string
	Guide          string
	CycleCode      string
	DateCode       string
	PeriodCode     string
	PrevPeriodCode string
	PrevCycleCode  string
	PageTarget1    string
	PageTarget2    string
	PageTarget3    string
}

func (r Report) HumanDate() string {
	return humanize.Time(time.Unix(r.DateAdded, 0))
}

type Type string

const (
	TypeJasper Type = "JASPER"
	TypeLooker Type = "LOOKER"
	TypeHex    Type = "HEX"
	TypeDoc    Type = "GOOGLE_DOC"
	TypeTool   Type = "TOOL"
)

var FunctionalAreas = []FunctionArea{
	{ID: "collections", Name: "Collections", Icon: "folder.svg"},
	{ID: "credit", Name: "Credit", Icon: "credit_card.svg"},
	{ID: "finance", Name: "Finance", Icon: "calculator.svg"},
	{ID: "ops", Name: "Operations", Icon: "delivery_truck.svg"},
	{ID: "hr", Name: "Human Resources", Icon: "suitcase.svg"},
	{ID: "it", Name: "Information Technology", Icon: "laptop.svg"},
	{ID: "marketing", Name: "Marketing", Icon: "target.svg"},
	{ID: "product", Name: "Product", Icon: "light_bulb.svg"},
	{ID: "sales", Name: "Sales", Icon: "money_bag.svg"},
	{ID: "ujuzi", Name: "Ujuzi", Icon: "abacus.svg"},
	{ID: "other", Name: "Other", Icon: "folder.svg", Hidden: true},
}

var FunctionalAreasMap = make(map[string]FunctionArea)

func init() {
	for _, fa := range FunctionalAreas {
		FunctionalAreasMap[fa.ID] = fa
	}
}

func (r Report) GetIcon() string {
	for _, fa := range FunctionalAreas {
		if fa.ID == r.FunctionalArea {
			return fmt.Sprintf("/assets/img/%s", fa.Icon)
		}
	}
	return "/assets/img/folder.svg"
}
