package observability

import "regexp"

var idPattern = regexp.MustCompile(`/[0-9a-fA-F-]{6,}`)

func NormalizeRoute(path string) string {
	// Replace IDs, UUIDs, numeric segments
	return idPattern.ReplaceAllString(path, "/:id")
}
