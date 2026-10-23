package common

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

// ErrAdminIPNotAllowed is returned when a request from an admin-role user
// originates from an IP address outside the configured admin allowlist.
var ErrAdminIPNotAllowed = errors.New("administrator access is not allowed from this IP address")

// The admin IP allowlist is configured exclusively through out-of-band sources
// (environment variable and/or a standalone file) and is deliberately not
// editable from the web console: a hijacked admin session must not be able to
// lift the restriction, and a misconfigured allowlist must not lock operators
// out of the only interface that could fix it. Recovery is always server-side.
var adminAllowedIPEntries []string

// InitAdminIPAllowlist loads the admin IP allowlist from ADMIN_ALLOWED_IPS
// (comma or newline separated) and the standalone file named by
// ADMIN_ALLOWED_IPS_FILE (one IP or CIDR per line, '#' starts a comment).
// Entries from both sources are merged. An empty result disables the
// restriction, preserving the previous behavior for existing deployments.
// Any invalid entry is an error, so a typo cannot silently narrow or widen
// the allowlist.
func InitAdminIPAllowlist() error {
	var rawEntries []string
	if envValue := strings.TrimSpace(os.Getenv("ADMIN_ALLOWED_IPS")); envValue != "" {
		envValue = strings.ReplaceAll(envValue, ",", "\n")
		rawEntries = append(rawEntries, strings.Split(envValue, "\n")...)
	}
	if path := strings.TrimSpace(os.Getenv("ADMIN_ALLOWED_IPS_FILE")); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read ADMIN_ALLOWED_IPS_FILE %q: %w", path, err)
		}
		rawEntries = append(rawEntries, strings.Split(string(data), "\n")...)
	}
	entries := make([]string, 0, len(rawEntries))
	for _, raw := range rawEntries {
		entry := strings.TrimSpace(raw)
		if entry == "" || strings.HasPrefix(entry, "#") {
			continue
		}
		if net.ParseIP(entry) == nil {
			if _, _, err := net.ParseCIDR(entry); err != nil {
				return fmt.Errorf("invalid admin IP allowlist entry %q: must be an IP address or a CIDR", entry)
			}
		}
		entries = append(entries, entry)
	}
	adminAllowedIPEntries = entries
	return nil
}

func AdminIPAllowlistEnabled() bool {
	return len(adminAllowedIPEntries) > 0
}

func AdminIPAllowlistCount() int {
	return len(adminAllowedIPEntries)
}

// CheckAdminClientIP reports whether a client IP may be used by admin-role
// users. When no allowlist is configured every IP is allowed, and ordinary
// users are never routed through this check. An unparseable IP is rejected
// (fail-closed) once the restriction is active.
func CheckAdminClientIP(clientIP string) error {
	if len(adminAllowedIPEntries) == 0 {
		return nil
	}
	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil {
		return ErrAdminIPNotAllowed
	}
	if !IsIpInCIDRList(ip, adminAllowedIPEntries) {
		return ErrAdminIPNotAllowed
	}
	return nil
}
