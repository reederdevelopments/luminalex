package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func AddDistinct[T any](list []T, newItems []T, compare func(T, T) bool) []T {
	uniqueList := make([]T, len(list))
	copy(uniqueList, list)

	for _, newItem := range newItems {
		exists := false
		for _, existingItem := range uniqueList {
			if compare(existingItem, newItem) {
				exists = true
				break
			}
		}
		if !exists {
			uniqueList = append(uniqueList, newItem)
		}
	}
	return uniqueList
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func InList(item string, list []string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func MustMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("ERROR: MustMarshal failed: %v", err)
		return ""
	}
	return string(b)
}

func TimeAgo(t time.Time) string {
	const (
		day   = 24 * time.Hour
		month = 30 * day
		year  = 12 * month
	)

	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case diff < day:
		h := int(diff.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case diff < month:
		d := int(diff / day)
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	case diff < year:
		m := int(diff / month)
		if m == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", m)
	default:
		y := int(diff / year)
		if y == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", y)
	}
}
