package terminal

import (
	"fmt"
	"strconv"
	"strings"
)

const bkNamespace = "bk"

// Parse an Application Program Command sequence, which may or may not be a
// Buildkite APC, e.g. bk;t=123123234234234;llamas=blah
func (p *parser) parseBuildkiteAPC(sequence string) (map[string]string, error) {
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

		key, val := tokenParts[0], tokenParts[1]
		switch key {
		case "t":
			t, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("t key has non-integer value %q: %w", val, err)
			}
			p.lastTimestamp = t
			data[key] = val

		case "dt":
			dt, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("dt key has non-integer value %q: %w", val, err)
			}
			// Convert dt into t
			p.lastTimestamp += dt
			data["t"] = strconv.FormatInt(p.lastTimestamp, 10)

		default:
			data[key] = val
		}
	}

	return data, nil
}
