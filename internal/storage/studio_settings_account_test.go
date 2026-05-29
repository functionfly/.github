package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAccountPreferences(t *testing.T) {
	prefs := DefaultAccountPreferences()
	assert.False(t, prefs.LaunchAtLogin)
	assert.True(t, prefs.MinimizeToTrayOnClose)
	assert.True(t, prefs.RestoreLastWorkspace)
	assert.True(t, prefs.OpenLinksExternally)
	assert.Nil(t, prefs.LastWorkspace)
}

func TestApplyAccountPreferencesPatch(t *testing.T) {
	current := DefaultAccountPreferences()
	falseVal := false
	trueVal := true

	patched := ApplyAccountPreferencesPatch(current, AccountPreferencesPatch{
		LaunchAtLogin: &trueVal,
		OpenLinksExternally: &falseVal,
	})

	assert.True(t, patched.LaunchAtLogin)
	assert.True(t, patched.MinimizeToTrayOnClose)
	assert.False(t, patched.OpenLinksExternally)
}

func TestMergeAccountPreferencesPreservesLastWorkspace(t *testing.T) {
	existing := DefaultAccountPreferences()
	existing.LastWorkspace = &LastWorkspaceState{Route: "/studio/workflows"}

	incoming := DefaultAccountPreferences()
	incoming.LaunchAtLogin = true

	merged := MergeAccountPreferences(existing, incoming)
	assert.True(t, merged.LaunchAtLogin)
	assert.NotNil(t, merged.LastWorkspace)
	assert.Equal(t, "/studio/workflows", merged.LastWorkspace.Route)
}

func TestDefaultSettingsIncludesAccountPreferences(t *testing.T) {
	settings := DefaultSettings()
	assert.NotNil(t, settings.AccountPreferences)
	assert.True(t, settings.AccountPreferences.MinimizeToTrayOnClose)
}
