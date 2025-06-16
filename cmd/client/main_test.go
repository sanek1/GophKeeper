package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainPackage(t *testing.T) {
	t.Run("PackageLoads", func(t *testing.T) {
		// Test that main package can be loaded and imported
		assert.True(t, true)
	})

	t.Run("VersionInfo", func(t *testing.T) {
		// Test version variables exist
		assert.NotNil(t, version)
		assert.NotNil(t, commitHash)
		assert.NotNil(t, buildDate)
		assert.NotNil(t, buildNum)
	})
}
