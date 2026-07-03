// package uj deals with application level controls
package uj

import (
	"net/http"
	"net/url"
	"time"

	"github.com/Rockup-Consulting/std/core/web"
	"github.com/a-h/templ"
)

func IsEmbedded(r *http.Request) bool {
	v := r.FormValue("embedded")
	return v == "true"
}

// Redirect does an http.Redirect with certain Ujuzi specific params included
func Redirect(w http.ResponseWriter, r *http.Request, link string) error {
	u, err := url.Parse(link)
	if err != nil {
		return err
	}

	q := u.Query()

	embedded := r.FormValue("embedded")
	cc := r.FormValue("cc")

	if embedded != "" {
		q.Set("embedded", embedded)
	}

	if cc != "" {
		q.Set("cc", cc)
	}

	u.RawQuery = q.Encode()
	link = u.String()

	http.Redirect(w, r, link, http.StatusSeeOther)
	return nil
}

func DeleteCookieAndRedirect(w http.ResponseWriter, r *http.Request, link string, cookie string, now time.Time) error {
	u, err := url.Parse(link)
	if err != nil {
		return err
	}

	q := u.Query()

	embedded := r.FormValue("embedded")
	cc := r.FormValue("cc")

	if embedded != "" {
		q.Set("embedded", embedded)
	}

	if cc != "" {
		q.Set("cc", cc)
	}

	u.RawQuery = q.Encode()
	link = u.String()

	web.DeleteCookieAndRedirect(w, r, link, cookie, http.StatusTemporaryRedirect, now)

	return nil
}

func Href(r *http.Request, link string) string {
	u, err := url.Parse(link)
	if err != nil {
		panic(err)
	}

	q := u.Query()

	embedded := r.FormValue("embedded")
	cc := r.FormValue("cc")

	if embedded != "" {
		q.Set("embedded", embedded)
	}

	if cc != "" {
		q.Set("cc", cc)
	}

	u.RawQuery = q.Encode()
	link = u.String()

	return link
}

func SafeHref(r *http.Request, href string) templ.SafeURL {
	return templ.SafeURL(Href(r, href))
}
