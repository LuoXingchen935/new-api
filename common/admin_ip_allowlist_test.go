package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setAdminIPAllowlistEnv restores the environment after each case so the
// package-level allowlist never leaks across tests.
func setAdminIPAllowlistEnv(t *testing.T, ips, file string) {
	t.Helper()
	t.Setenv("ADMIN_ALLOWED_IPS", ips)
	t.Setenv("ADMIN_ALLOWED_IPS_FILE", file)
}

func resetAdminIPAllowlist(t *testing.T) {
	t.Helper()
	previous := adminAllowedIPEntries
	t.Cleanup(func() { adminAllowedIPEntries = previous })
	adminAllowedIPEntries = nil
}

func TestInitAdminIPAllowlistMergesEnvAndFile(t *testing.T) {
	resetAdminIPAllowlist(t)
	file := t.TempDir() + "/allow.txt"
	require.NoError(t, os.WriteFile(file, []byte("# ops network\n10.0.0.0/8\n203.0.113.5\n\n"), 0o600))
	setAdminIPAllowlistEnv(t, "198.51.100.7, 192.0.2.0/24", file)

	require.NoError(t, InitAdminIPAllowlist())
	assert.True(t, AdminIPAllowlistEnabled())
	assert.Equal(t, 4, AdminIPAllowlistCount())
	assert.NoError(t, CheckAdminClientIP("198.51.100.7"))
	assert.NoError(t, CheckAdminClientIP("192.0.2.99"))
	assert.NoError(t, CheckAdminClientIP("10.1.2.3"))
	assert.NoError(t, CheckAdminClientIP("203.0.113.5"))
	assert.ErrorIs(t, CheckAdminClientIP("8.8.8.8"), ErrAdminIPNotAllowed)
}

func TestInitAdminIPAllowlistEmptyDisablesRestriction(t *testing.T) {
	resetAdminIPAllowlist(t)
	setAdminIPAllowlistEnv(t, "  ", "")

	require.NoError(t, InitAdminIPAllowlist())
	assert.False(t, AdminIPAllowlistEnabled())
	assert.NoError(t, CheckAdminClientIP("8.8.8.8"))
	assert.NoError(t, CheckAdminClientIP(""))
}

func TestInitAdminIPAllowlistRejectsInvalidEntry(t *testing.T) {
	resetAdminIPAllowlist(t)
	setAdminIPAllowlistEnv(t, "198.51.100.7,not-an-ip", "")

	err := InitAdminIPAllowlist()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not-an-ip")
}

func TestCheckAdminClientIPRejectsUnparseableIP(t *testing.T) {
	resetAdminIPAllowlist(t)
	setAdminIPAllowlistEnv(t, "198.51.100.7", "")
	require.NoError(t, InitAdminIPAllowlist())

	// Fail-closed: once the restriction is active an unusable client IP is denied.
	assert.ErrorIs(t, CheckAdminClientIP("garbage"), ErrAdminIPNotAllowed)
	assert.ErrorIs(t, CheckAdminClientIP(""), ErrAdminIPNotAllowed)
}
