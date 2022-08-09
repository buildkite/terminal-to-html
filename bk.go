package terminal

import (
	"fmt"
	"strings"
)

const bkNamespace = "bk"

// Parse an Application Program Command sequence, which may or may not be a
// Buildkite APC, e.g. bk;t=123123234234234;llamas=blah
func parseApcBk(sequence string) (map[string]string, error) {
	if !strings.HasPrefix(sequence, bkNamespace+";") {
		return nil, nil
	}

	tokens, err := tokenizeString(sequence[3:], ';', '\\')
	if err != nil {
		return nil, err
	}

	data := map[string]string{}

	for _, token := range tokens {
		tokenParts := strings.SplitN(token, "=", 2)
		if len(tokenParts) != 2 {
			return nil, fmt.Errorf("Failed to read key=value from token %q", token)
		}
		data[tokenParts[0]] = tokenParts[1]
	}

	return data, nil
}
