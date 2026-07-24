package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDomainPackagesRemainTransportIndependent(t *testing.T) {
	t.Parallel()
	internalRoot := filepath.Clean("..")
	banned := []string{
		"net/http",
		"database/sql",
		"github.com/go-chi/",
		"websocket",
		"sqlite",
		"sqlc",
		"log/slog",
		"/transport",
		"/config",
		"/contracts/generated",
		"svelte",
	}
	allowedExternal := []string{
		"github.com/google/uuid",
		"github.com/rivo/uniseg",
		"golang.org/x/text/",
	}

	err := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
			!isDomainPath(path) {
			return nil
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, imported := range file.Imports {
			importPath, unquoteErr := strconv.Unquote(imported.Path.Value)
			if unquoteErr != nil {
				return unquoteErr
			}
			for _, fragment := range banned {
				if strings.Contains(importPath, fragment) {
					t.Errorf("%s imports forbidden package %s", path, importPath)
				}
			}
			if strings.Contains(importPath, ".") && !strings.Contains(importPath, "/internal/domain") &&
				!containsPrefix(allowedExternal, importPath) {
				t.Errorf("%s imports unapproved external package %s", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func isDomainPath(path string) bool {
	clean := filepath.ToSlash(path)
	return strings.Contains(clean, "/domain/") || strings.HasPrefix(clean, "../domain/")
}

func containsPrefix(prefixes []string, value string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
