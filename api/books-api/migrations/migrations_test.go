package migrations

import (
	"io"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedMigrationSourceUsesCanonicalSQL(t *testing.T) {
	source, err := iofs.New(Files, ".")
	require.NoError(t, err)
	defer source.Close()

	version, err := source.First()
	require.NoError(t, err)
	require.Equal(t, uint(1), version)

	up, identifier, err := source.ReadUp(version)
	require.NoError(t, err)
	require.Equal(t, "initial_schema", identifier)
	actual, err := io.ReadAll(up)
	require.NoError(t, err)
	expected, err := os.ReadFile("000001_initial_schema.up.sql")
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
