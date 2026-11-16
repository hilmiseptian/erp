package utils

import (
	"math/big"
	"strings"
)

// ParseDecimal validates a base-10 decimal string and returns its exact value.
// maxInteger and maxFraction limit the number of digits before and after the
// decimal point. A leading minus sign is accepted; callers enforce domain rules.
func ParseDecimal(value string, maxInteger, maxFraction int) (*big.Rat, bool) {
	if value == "" || maxInteger < 1 || maxFraction < 0 {
		return nil, false
	}

	unsigned := value
	if strings.HasPrefix(unsigned, "-") || strings.HasPrefix(unsigned, "+") {
		unsigned = unsigned[1:]
	}
	if unsigned == "" {
		return nil, false
	}

	parts := strings.Split(unsigned, ".")
	if len(parts) > 2 || len(parts[0]) == 0 || len(parts[0]) > maxInteger {
		return nil, false
	}
	if len(parts) == 2 && (len(parts[1]) == 0 || len(parts[1]) > maxFraction) {
		return nil, false
	}
	for _, part := range parts {
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return nil, false
			}
		}
	}

	valueRat, ok := new(big.Rat).SetString(value)
	return valueRat, ok
}
