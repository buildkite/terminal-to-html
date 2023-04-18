package terminal

import (
	"net/url"
)

const unsafeURLSubstitution = "#"

func sanitizeURL(s string) string {
	url, err := url.Parse(s)
	if err != nil {
		return unsafeURLSubstitution
	}

	// relative URLs (no scheme) are permitted
	if url.Scheme == "" {
		return url.String()
	}

	// allow-list schemes
	allowedSchemes := []string{"https", "http", "artifact"}
	for _, as := range allowedSchemes {
		if url.Scheme == as {
			return url.String()
		}
	}

	// default deny, catches e.g. "javascript:…"
	return unsafeURLSubstitution
}
