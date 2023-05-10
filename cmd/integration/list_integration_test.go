//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("list all resources", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		// Check that all the resources in the test policy are listed
		assertListsAllResources(t, stdOut, cli)
		// Should not contain any timestamps or other metadata
		assert.NotContains(t, stdOut, "\"created_at\":")
	})

	t.Run("list with inspect", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list", "--inspect")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)

		assertListsAllResources(t, stdOut, cli)
		// Should contain timestamps and other metadata
		assertContainsMetadata(t, stdOut, cli)
	})

	t.Run("list with kind", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list", "--kind", "variable")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)

		assert.Contains(t, stdOut, cli.account+":variable:meow")
		assert.Contains(t, stdOut, cli.account+":variable:woof")
		assert.NotContains(t, stdOut, cli.account+":policy:root")
		assert.NotContains(t, stdOut, cli.account+":user:alice")
		assert.NotContains(t, stdOut, cli.account+":host:bob")

		// Should not contain any timestamps or other metadata
		assertDoesNotContainMetadata(t, stdOut, cli)
	})

	t.Run("list with kind and inspect", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list", "--kind", "variable", "--inspect")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)

		assert.Contains(t, stdOut, cli.account+":variable:meow")
		assert.Contains(t, stdOut, cli.account+":variable:woof")
		assert.NotContains(t, stdOut, cli.account+":host:bob")

		// Should contain timestamps and other metadata
		assertContainsMetadata(t, stdOut, cli)
	})

	t.Run("list with search", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list", "--search", "meow")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)

		assert.Contains(t, stdOut, cli.account+":variable:meow")
		assert.NotContains(t, stdOut, cli.account+":variable:woof")
		assert.NotContains(t, stdOut, cli.account+":policy:root")
		assert.NotContains(t, stdOut, cli.account+":user:alice")
		assert.NotContains(t, stdOut, cli.account+":host:bob")

		// Should not contain any timestamps or other metadata
		assertDoesNotContainMetadata(t, stdOut, cli)
	})

	t.Run("list with role", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list", "--role", cli.account+":user:alice")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)

		assert.Contains(t, stdOut, cli.account+":variable:meow")
		assert.NotContains(t, stdOut, cli.account+":user:alice")
		assert.NotContains(t, stdOut, cli.account+":variable:woof")
		assert.NotContains(t, stdOut, cli.account+":policy:root")
		assert.NotContains(t, stdOut, cli.account+":host:bob")
	})

	t.Run("list with limit", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list", "--limit", "2")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)

		assert.Contains(t, stdOut, cli.account+":host:bob")
		assert.Contains(t, stdOut, cli.account+":policy:root")
		assert.NotContains(t, stdOut, cli.account+":variable:meow")
		assert.NotContains(t, stdOut, cli.account+":variable:woof")
		assert.NotContains(t, stdOut, cli.account+":user:alice")
	})

	t.Run("list with offset", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list", "--offset", "2")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)

		assert.Contains(t, stdOut, cli.account+":variable:meow")
		assert.Contains(t, stdOut, cli.account+":variable:woof")
		assert.Contains(t, stdOut, cli.account+":user:alice")
		assert.NotContains(t, stdOut, cli.account+":host:bob")
		assert.NotContains(t, stdOut, cli.account+":policy:root")
	})

	t.Run("list with limit and offset", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("list", "--limit", "2", "--offset", "1")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)

		assert.Contains(t, stdOut, cli.account+":policy:root")
		assert.Contains(t, stdOut, cli.account+":user:alice")
		assert.NotContains(t, stdOut, cli.account+":host:bob")
		assert.NotContains(t, stdOut, cli.account+":variable:woof")
		assert.NotContains(t, stdOut, cli.account+":variable:meow")
	})
}

func assertListsAllResources(t *testing.T, stdOut string, cli *testConjurCLI) {
	assert.Contains(t, stdOut, cli.account+":policy:root")
	assert.Contains(t, stdOut, cli.account+":variable:meow")
	assert.Contains(t, stdOut, cli.account+":variable:woof")
	assert.Contains(t, stdOut, cli.account+":user:alice")
	assert.Contains(t, stdOut, cli.account+":host:bob")
}

func assertContainsMetadata(t *testing.T, stdOut string, cli *testConjurCLI) {
	assert.Contains(t, stdOut, "\"created_at\":")
	assert.Contains(t, stdOut, "\"annotations\":")
	assert.Contains(t, stdOut, "\"owner\":")
	assert.Contains(t, stdOut, "\"permissions\":")
	assert.Contains(t, stdOut, "\"policy\":")
}

func assertDoesNotContainMetadata(t *testing.T, stdOut string, cli *testConjurCLI) {
	assert.NotContains(t, stdOut, "\"created_at\":")
	assert.NotContains(t, stdOut, "\"annotations\":")
	assert.NotContains(t, stdOut, "\"owner\":")
	assert.NotContains(t, stdOut, "\"permissions\":")
	assert.NotContains(t, stdOut, "\"policy\":")
}
