package gemini

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDocGenerator(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	generator := NewDocGenerator(client)

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package test

func Add(a, b int) int {
	return a + b
}

type Calculator struct{}

func (c *Calculator) Multiply(x, y float64) (float64, error) {
	return x * y, nil
}
`
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("generate doc for function", func(t *testing.T) {
		code := `func Add(a, b int) int`
		doc, err := generator.GenerateFuncDoc(context.Background(), code)
		if err != nil {
			t.Fatalf("Failed to generate function doc: %v", err)
		}

		assertContains(t, doc, "Add", "Should contain function name")
		assertContains(t, doc, "Parameters", "Should contain parameters section")
		assertContains(t, doc, "Returns", "Should contain returns section")
	})

	t.Run("generate doc for method", func(t *testing.T) {
		code := `func (c *Calculator) Multiply(x, y float64) (float64, error)`
		doc, err := generator.GenerateFuncDoc(context.Background(), code)
		if err != nil {
			t.Fatalf("Failed to generate method doc: %v", err)
		}

		assertContains(t, doc, "Multiply", "Should contain method name")
		assertContains(t, doc, "Parameters", "Should contain parameters section")
		assertContains(t, doc, "Returns", "Should contain returns section")
		assertContains(t, doc, "error", "Should mention error handling")
	})

	t.Run("generate doc for file", func(t *testing.T) {
		docs, err := generator.GenerateFileDoc(context.Background(), testFile)
		if err != nil {
			t.Fatalf("Failed to generate file doc: %v", err)
		}

		if len(docs) != 2 {
			t.Errorf("Expected 2 function docs, got %d", len(docs))
		}

		// Check Add function doc
		if doc, ok := docs["Add"]; ok {
			assertContains(t, doc, "Add", "Should contain function name")
			assertContains(t, doc, "Parameters", "Should contain parameters section")
		} else {
			t.Error("Add function doc not found")
		}

		// Check Multiply method doc
		if doc, ok := docs["Multiply"]; ok {
			assertContains(t, doc, "Multiply", "Should contain method name")
			assertContains(t, doc, "error", "Should mention error handling")
		} else {
			t.Error("Multiply method doc not found")
		}
	})

	t.Run("auto doc for new function", func(t *testing.T) {
		doc, err := generator.AutoDoc(context.Background(), testFile, "Add")
		if err != nil {
			t.Fatalf("Failed to auto-generate doc: %v", err)
		}

		assertContains(t, doc, "Add", "Should contain function name")
		assertContains(t, doc, "Parameters", "Should contain parameters section")
		assertContains(t, doc, "Returns", "Should contain returns section")
	})
}
