package gemini

import (
	"context"
	"testing"
)

func TestStatusReport(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	tests := []struct {
		name     string
		metrics  string
		wantErr  bool
		validate func(t *testing.T, report string)
	}{
		{
			name: "healthy system",
			metrics: `
Cache utilization: low
Hit rate: improving
Latency: stable
Memory usage: normal
Active connections: steady`,
			validate: func(t *testing.T, report string) {
				assertContains(t, report, "performance", "Should discuss performance")
				assertContains(t, report, "stable", "Should mention stability")
				assertContains(t, report, "memory", "Should mention resource usage")
			},
		},
		{
			name: "system under load",
			metrics: `
Cache utilization: high
Hit rate: decreasing
Latency: increased
Memory usage: elevated
Active connections: increasing`,
			validate: func(t *testing.T, report string) {
				assertContains(t, report, "load", "Should mention system load")
				assertContains(t, report, "performance", "Should discuss performance")
				assertContains(t, report, "recommend", "Should include recommendations")
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := client.GenerateStatusReport(ctx, tt.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateStatusReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, report)
			}
		})
	}
}
