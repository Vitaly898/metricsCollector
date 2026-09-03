package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsRetriableDBError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "connection exception",
			err:  &pgconn.PgError{Code: pgerrcode.ConnectionException},
			want: true,
		},
		{
			name: "connection failure",
			err:  &pgconn.PgError{Code: pgerrcode.ConnectionFailure},
			want: true,
		},
		{
			name: "serialization failure",
			err:  &pgconn.PgError{Code: pgerrcode.SerializationFailure},
			want: true,
		},
		{
			name: "deadlock detected",
			err:  &pgconn.PgError{Code: pgerrcode.DeadlockDetected},
			want: true,
		},
		{
			name: "unique violation is not retriable",
			err:  &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			want: false,
		},
		{
			name: "non-pg error is not retriable",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "wrapped pg error",
			err:  fmt.Errorf("query failed: %w", &pgconn.PgError{Code: pgerrcode.SerializationFailure}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetriableDBError(tt.err); got != tt.want {
				t.Errorf("isRetriableDBError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
