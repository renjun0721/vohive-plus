package api

import "strings"

// An existing menu belongs to its original network dialog. A transient IMS
// disconnect must not send its continuation/cancellation to cellular AT+CUSD.
func ussdSessionUsesVoWiFi(sessionID string, currentlyActive bool) bool {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "cs" {
		return false
	}
	if sessionID != "" {
		return true
	}
	// Preserve the optional session_id behavior of the cancellation API.
	return currentlyActive
}
