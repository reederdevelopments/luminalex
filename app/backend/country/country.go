package country

import (
	"context"
	"fmt"
)

type key int8

const (
	ctxKey key = 0

	ZA = "za"
	UG = "ug"
	KE = "ke"
	ZM = "zm"
	TZ = "tz"
)

type Country struct {
	Name string
	Code string
}

var (
	List = []Country{
		{
			Name: "South Africa",
			Code: "za",
		},
		{
			Name: "Kenya",
			Code: "ke",
		},
		{
			Name: "Uganda",
			Code: "ug",
		},
		{
			Name: "Zambia",
			Code: "zm",
		},
		{
			Name: "Tanzania",
			Code: "tz",
		},
	}
)

func Set(ctx context.Context, c Country) context.Context {
	return context.WithValue(ctx, ctxKey, c)
}

func FromCtx(ctx context.Context) Country {
	c, ok := ctx.Value(ctxKey).(Country)
	if !ok {
		panic("Country not on context")
	}

	return c
}

func CountryToCode(country string) string {
	if country == "" {
		return ""
	}

	for _, c := range List {
		if country == c.Name {
			return c.Code
		}
	}
	return ""
}

func CodeToCountry(code string) Country {
	for _, c := range List {
		if code == c.Code {
			return c
		}
	}

	panic(fmt.Sprintf("invalid country code %q", code))
}
