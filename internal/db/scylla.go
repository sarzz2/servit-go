package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gocql/gocql"
)

var ScyllaSession *gocql.Session

// InitScylla initializes the ScyllaDB connection
func InitScylla(hosts []string) error {
	// First connection without a keyspace to create it.
	cluster := gocql.NewCluster(hosts...)
	cluster.ProtoVersion = 4
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to connect to ScyllaDB: %w", err)
	}
	defer session.Close()

	// Create keyspace if it doesn't exist.
	err = session.Query(`
		CREATE KEYSPACE IF NOT EXISTS messaging 
		WITH replication = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}`).Exec()
	if err != nil {
		return fmt.Errorf("failed to create keyspace: %w", err)
	}

	// Now connect again with the keyspace specified.
	cluster.Keyspace = "messaging"
	scyllaSession, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session with keyspace: %w", err)
	}

	ScyllaSession = scyllaSession
	return nil
}

func RunMigrations(migrationsDir string) error {
	// Read the migration directory entries.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter for .cql files and sort them to ensure they run in order.
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Execute each migration.
	for _, fileName := range migrationFiles {
		filePath := filepath.Join(migrationsDir, fileName)
		fmt.Printf("Applying migration: %s\n", fileName)

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		// Assume each migration file contains a single CQL statement.
		cqlQuery := string(data)
		if err := ScyllaSession.Query(cqlQuery).Exec(); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}
		fmt.Printf("Migration %s applied successfully.\n", fileName)
	}
	return nil
}
