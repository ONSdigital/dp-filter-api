package observation

import (
	"database/sql/driver"
	"time"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//go:generate moq -out observationtest/bolt_connectiongit a.go -pkg observationtest . Conn

type Conn interface {
	// PrepareNeo prepares a neo4j specific statement
	PrepareNeo(query string) (bolt.Stmt, error)
	// PreparePipeline prepares a neo4j specific pipeline statement
	// Useful for running multiple queries at the same time
	PreparePipeline(query ...string) (bolt.PipelineStmt, error)
	// QueryNeo queries using the neo4j-specific interface
	QueryNeo(query string, params map[string]interface{}) (bolt.Rows, error)
	// QueryNeoAll queries using the neo4j-specific interface and returns all row data and output metadata
	QueryNeoAll(query string, params map[string]interface{}) ([][]interface{}, map[string]interface{}, map[string]interface{}, error)
	// QueryPipeline queries using the neo4j-specific interface
	// pipelining multiple statements
	QueryPipeline(query []string, params ...map[string]interface{}) (bolt.PipelineRows, error)
	// ExecNeo executes a query using the neo4j-specific interface
	ExecNeo(query string, params map[string]interface{}) (bolt.Result, error)
	// ExecPipeline executes a query using the neo4j-specific interface
	// pipelining multiple statements
	ExecPipeline(query []string, params ...map[string]interface{}) ([]bolt.Result, error)
	// Close closes the connection
	Close() error
	// Begin starts a new transaction
	Begin() (driver.Tx, error)
	// SetChunkSize is used to set the max chunk size of the
	// bytes to send to Neo4j at once
	SetChunkSize(uint16)
	// SetTimeout sets the read/write timeouts for the
	// connection to Neo4j
	SetTimeout(time.Duration)
}
