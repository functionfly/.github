package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	// pgx driver is registered via the main postgres driver import
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type PostgresConnector struct {
	logger *logrus.Logger
}

func NewPostgresConnector(logger *logrus.Logger) *PostgresConnector {
	if logger == nil {
		logger = logrus.New()
	}
	return &PostgresConnector{logger: logger}
}

func (c *PostgresConnector) Name() string { return "PostgreSQL" }
func (c *PostgresConnector) Icon() string { return "database" }
func (c *PostgresConnector) IsConfigured() bool { return true }

func (c *PostgresConnector) Authenticate(ctx context.Context, creds map[string]string) error {
	dsn := creds["connection_string"]
	if dsn == "" {
		return fmt.Errorf("connection string is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.PingContext(ctx)
}

func (c *PostgresConnector) FetchData(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	dsn := config["connection_string"]
	query := config["query"]

	if dsn == nil || query == nil {
		return nil, fmt.Errorf("connection_string and query are required")
	}

	db, err := sql.Open("postgres", dsn.(string))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query.(string))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return json.Marshal(results)
}

func (c *PostgresConnector) Search(ctx context.Context, query string, config map[string]interface{}) ([]SearchResult, error) {
	dsn := config["connection_string"]
	if dsn == nil {
		return nil, fmt.Errorf("connection_string is required")
	}

	db, err := sql.Open("postgres", dsn.(string))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Use a parameterized query to prevent SQL injection.
	rows, err := db.QueryContext(ctx,
		"SELECT table_schema, table_name FROM information_schema.tables WHERE table_name LIKE $1 LIMIT 10",
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var tableSchema, tableName string
		if err := rows.Scan(&tableSchema, &tableName); err == nil {
			results = append(results, SearchResult{
				Title:   tableName,
				Content: fmt.Sprintf("Schema: %s", tableSchema),
				URL:     fmt.Sprintf("postgres://%s/%s", tableSchema, tableName),
			})
		}
	}
	return results, nil
}
