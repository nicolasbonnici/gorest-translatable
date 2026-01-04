package mocks

import (
	"context"
	"errors"

	"github.com/nicolasbonnici/gorest/database"
)

type MockDatabase struct {
	ExecFunc     func(ctx context.Context, query string, args ...interface{}) (database.Result, error)
	QueryFunc    func(ctx context.Context, query string, args ...interface{}) (database.Rows, error)
	QueryRowFunc func(ctx context.Context, query string, args ...interface{}) database.Row
}

func (m *MockDatabase) Exec(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, query, args...)
	}
	return &MockResult{rowsAffected: 1}, nil
}

func (m *MockDatabase) Query(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, query, args...)
	}
	return &MockRows{}, nil
}

func (m *MockDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) database.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, query, args...)
	}
	return &MockRow{}
}

func (m *MockDatabase) Connect(ctx context.Context, dsn string) error { return nil }
func (m *MockDatabase) Close() error                                  { return nil }
func (m *MockDatabase) Ping(ctx context.Context) error                { return nil }
func (m *MockDatabase) Begin(ctx context.Context) (database.Tx, error) {
	return nil, errors.New("not implemented")
}
func (m *MockDatabase) Dialect() database.Dialect                 { return nil }
func (m *MockDatabase) DriverName() string                        { return "mock" }
func (m *MockDatabase) Introspector() database.SchemaIntrospector { return nil }

type MockResult struct {
	rowsAffected int64
	lastInsertId int64
	err          error
}

func NewMockResult(rowsAffected int64) *MockResult {
	return &MockResult{rowsAffected: rowsAffected}
}

func (m *MockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, m.err
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.lastInsertId, m.err
}

type MockRow struct {
	ScanFunc func(dest ...interface{}) error
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return errors.New("no rows")
}

type MockRows struct {
	closed    bool
	closeErr  error
	scanErr   error
	nextCount int
	maxNext   int
}

func NewMockRows(maxNext int) *MockRows {
	return &MockRows{maxNext: maxNext}
}

func (m *MockRows) Next() bool {
	if m.nextCount < m.maxNext {
		m.nextCount++
		return true
	}
	return false
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	return nil
}

func (m *MockRows) Close() error {
	m.closed = true
	return m.closeErr
}

func (m *MockRows) Err() error {
	return nil
}
