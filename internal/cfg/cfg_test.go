package cfg

import (
	"flag"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid defaults",
			cfg:     Config{DrainSeconds: 60, ShutdownBudgetSeconds: 90, HTTPPort: 8080},
			wantErr: false,
		},
		{
			name:    "valid boundary low",
			cfg:     Config{DrainSeconds: 1, ShutdownBudgetSeconds: 2, HTTPPort: 1},
			wantErr: false,
		},
		{
			name:    "valid boundary high",
			cfg:     Config{DrainSeconds: 299, ShutdownBudgetSeconds: 300, HTTPPort: 65535},
			wantErr: false,
		},
		{
			name:    "drain zero",
			cfg:     Config{DrainSeconds: 0, ShutdownBudgetSeconds: 90, HTTPPort: 8080},
			wantErr: true,
		},
		{
			name:    "drain negative",
			cfg:     Config{DrainSeconds: -1, ShutdownBudgetSeconds: 90, HTTPPort: 8080},
			wantErr: true,
		},
		{
			name:    "drain exceeds max",
			cfg:     Config{DrainSeconds: 301, ShutdownBudgetSeconds: 400, HTTPPort: 8080},
			wantErr: true,
		},
		{
			name:    "shutdown budget zero",
			cfg:     Config{DrainSeconds: 60, ShutdownBudgetSeconds: 0, HTTPPort: 8080},
			wantErr: true,
		},
		{
			name:    "shutdown budget exceeds max",
			cfg:     Config{DrainSeconds: 60, ShutdownBudgetSeconds: 301, HTTPPort: 8080},
			wantErr: true,
		},
		{
			name:    "shutdown budget must exceed drain",
			cfg:     Config{DrainSeconds: 90, ShutdownBudgetSeconds: 90, HTTPPort: 8080},
			wantErr: true,
		},
		{
			name:    "shutdown budget less than drain",
			cfg:     Config{DrainSeconds: 90, ShutdownBudgetSeconds: 60, HTTPPort: 8080},
			wantErr: true,
		},
		{
			name:    "http port zero",
			cfg:     Config{DrainSeconds: 60, ShutdownBudgetSeconds: 90, HTTPPort: 0},
			wantErr: true,
		},
		{
			name:    "http port negative",
			cfg:     Config{DrainSeconds: 60, ShutdownBudgetSeconds: 90, HTTPPort: -1},
			wantErr: true,
		},
		{
			name:    "http port exceeds max",
			cfg:     Config{DrainSeconds: 60, ShutdownBudgetSeconds: 90, HTTPPort: 65536},
			wantErr: true,
		},
		{
			name:    "multiple errors",
			cfg:     Config{DrainSeconds: 0, ShutdownBudgetSeconds: 0, HTTPPort: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegisterFlags(t *testing.T) {
	var c Config
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	c.RegisterFlags(fs)

	// Verify defaults by parsing with no args
	if err := fs.Parse(nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if c.DrainSeconds != 60 {
		t.Errorf("DrainSeconds = %d, want 60", c.DrainSeconds)
	}
	if c.ShutdownBudgetSeconds != 90 {
		t.Errorf("ShutdownBudgetSeconds = %d, want 90", c.ShutdownBudgetSeconds)
	}
	if c.HTTPPort != 8080 {
		t.Errorf("HTTPPort = %d, want 8080", c.HTTPPort)
	}
}

func TestRegisterFlags_Override(t *testing.T) {
	var c Config
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	c.RegisterFlags(fs)

	if err := fs.Parse([]string{"-drain-seconds=10", "-shutdown-budget-seconds=20", "-http-port=9090"}); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if c.DrainSeconds != 10 {
		t.Errorf("DrainSeconds = %d, want 10", c.DrainSeconds)
	}
	if c.ShutdownBudgetSeconds != 20 {
		t.Errorf("ShutdownBudgetSeconds = %d, want 20", c.ShutdownBudgetSeconds)
	}
	if c.HTTPPort != 9090 {
		t.Errorf("HTTPPort = %d, want 9090", c.HTTPPort)
	}
}

func FuzzValidate(f *testing.F) {
	f.Add(60, 90, 8080)
	f.Add(0, 0, 0)
	f.Add(1, 2, 1)
	f.Add(300, 300, 65535)
	f.Add(-1, -1, -1)
	f.Add(301, 301, 65536)

	f.Fuzz(func(t *testing.T, drain, budget, port int) {
		c := Config{
			DrainSeconds:          drain,
			ShutdownBudgetSeconds: budget,
			HTTPPort:              port,
		}
		// Validate must not panic regardless of input
		err := c.Validate()

		// Cross-check: if all fields are in valid ranges and budget > drain, there should be no error
		drainOK := drain >= 1 && drain <= 300
		budgetOK := budget >= 1 && budget <= 300
		portOK := port >= 1 && port <= 65535
		budgetGtDrain := budget > drain

		if drainOK && budgetOK && portOK && budgetGtDrain && err != nil {
			t.Errorf("valid config returned error: %v", err)
		}
		if (!drainOK || !budgetOK || !portOK || !budgetGtDrain) && err == nil {
			t.Errorf("invalid config (drain=%d, budget=%d, port=%d) returned nil error", drain, budget, port)
		}
	})
}
