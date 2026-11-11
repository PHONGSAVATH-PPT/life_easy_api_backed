package config

import "testing"

func TestValidateDatabaseURL(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want bool
	}{
		{"postgres scheme", "postgres://user:password@postgres:5432/life_easy?sslmode=disable", true},
		{"postgresql scheme", "postgresql://user:password@localhost:5432/life_easy", true},
		{"wrong scheme", "mysql://user:password@localhost:3306/life_easy", false},
		{"missing database", "postgres://user:password@localhost:5432", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validateDatabaseURL(test.dsn) == nil
			if got != test.want {
				t.Fatalf("validateDatabaseURL(%q) valid = %v, want %v", test.dsn, got, test.want)
			}
		})
	}
}
