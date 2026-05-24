package domain

import "strings"

func IsHiddenTransactionID(hiddenIDs []string, id string) bool {
	normalizedID := strings.TrimSpace(id)
	if normalizedID == "" {
		return false
	}

	for _, hiddenID := range hiddenIDs {
		if strings.TrimSpace(hiddenID) == normalizedID {
			return true
		}
	}

	return false
}
