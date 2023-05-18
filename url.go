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

	// deny-list known-XSS-dangerous URL schemes for <a href=""> etc.
	// An allow-list would be preferable, but we don't know what URL schemes are being legitimately
	// used in the wild, so that would be a breaking change, and likely require configurability.
	disallowedSchemes := []string{"javascript"}
	for _, ds := range disallowedSchemes {
		if url.Scheme == ds {
			return unsafeURLSubstitution
		}
	}

	// default allow
	return url.String()
}
