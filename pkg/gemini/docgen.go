package gemini

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// DocGenerator generates API documentation using Gemini
type DocGenerator struct {
	client *Client
}

// NewDocGenerator creates a new DocGenerator
func NewDocGenerator(client *Client) *DocGenerator {
	return &DocGenerator{
		client: client,
	}
}

// GenerateFuncDoc generates documentation for a specific function
func (g *DocGenerator) GenerateFuncDoc(ctx context.Context, code string) (string, error) {
	return g.client.GenerateAPIDoc(ctx, code)
}

// GenerateFileDoc generates documentation for all functions in a file
func (g *DocGenerator) GenerateFileDoc(ctx context.Context, filePath string) (map[string]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %v", err)
	}

	docs := make(map[string]string)
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			// Get the function's code
			// Note: We'll use source code reconstruction instead of position-based extraction
			code := fmt.Sprintf("func %s%s %s",
				g.receiverString(fn),
				fn.Name.Name,
				g.formatFuncType(fn.Type))

			// Generate documentation
			doc, err := g.GenerateFuncDoc(ctx, code)
			if err != nil {
				return nil, fmt.Errorf("failed to generate doc for function %s: %v", fn.Name.Name, err)
			}
			docs[fn.Name.Name] = doc
		}
	}

	return docs, nil
}

// Helper function to format function receiver
func (g *DocGenerator) receiverString(fn *ast.FuncDecl) string {
	if fn.Recv == nil {
		return ""
	}
	var recv string
	for _, field := range fn.Recv.List {
		if star, ok := field.Type.(*ast.StarExpr); ok {
			if ident, ok := star.X.(*ast.Ident); ok {
				recv = fmt.Sprintf("(%s *%s) ", field.Names[0], ident.Name)
			}
		} else if ident, ok := field.Type.(*ast.Ident); ok {
			recv = fmt.Sprintf("(%s %s) ", field.Names[0], ident.Name)
		}
	}
	return recv
}

// Helper function to format function type
func (g *DocGenerator) formatFuncType(t *ast.FuncType) string {
	var params []string
	if t.Params != nil {
		for _, p := range t.Params.List {
			param := fmt.Sprintf("%s %s",
				strings.Join(g.extractNames(p.Names), ", "),
				g.formatExpr(p.Type))
			params = append(params, param)
		}
	}

	var results []string
	if t.Results != nil {
		for _, r := range t.Results.List {
			result := g.formatExpr(r.Type)
			if len(r.Names) > 0 {
				result = fmt.Sprintf("%s %s",
					strings.Join(g.extractNames(r.Names), ", "),
					result)
			}
			results = append(results, result)
		}
	}

	return fmt.Sprintf("(%s) (%s)",
		strings.Join(params, ", "),
		strings.Join(results, ", "))
}

// Helper function to extract names from ast.Ident slice
func (g *DocGenerator) extractNames(idents []*ast.Ident) []string {
	var names []string
	for _, ident := range idents {
		names = append(names, ident.Name)
	}
	return names
}

// Helper function to format expressions
func (g *DocGenerator) formatExpr(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + g.formatExpr(t.X)
	case *ast.SelectorExpr:
		return g.formatExpr(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + g.formatExpr(t.Elt)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// AutoDoc automatically generates documentation for newly added or modified functions
func (g *DocGenerator) AutoDoc(ctx context.Context, filePath string, funcName string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %v", err)
	}

	// Find the specific function
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == funcName {
			code := fmt.Sprintf("func %s%s %s",
				g.receiverString(fn),
				fn.Name.Name,
				g.formatFuncType(fn.Type))

			return g.GenerateFuncDoc(ctx, code)
		}
	}

	return "", fmt.Errorf("function %s not found", funcName)
}
