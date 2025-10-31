package asciidoc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProducerGenerateGolden(t *testing.T) {
	tdir := t.TempDir()

	goMod := "module example.com/docs\n\ngo 1.21"
	require.NoError(t, os.WriteFile(filepath.Join(tdir, "go.mod"), []byte(goMod), 0o644))

	pkgDir := filepath.Join(tdir, "sample")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))

	source := `package sample

import "fmt"

// Service handles greetings.
type Service struct {
	// Name is the service name.
	Name string
}

// ServiceOption configures a Service instance.
type ServiceOption func(*Service)

// Formatter formats names for presentation.
type Formatter interface {
	Format(string) string
}

// Config stores formatter configuration for services.
type Config struct {
	// Formatter applies formatting.
	Formatter Formatter
}

// NewService creates a new service with defaults.
func NewService(opts ...ServiceOption) *Service {
	svc := &Service{Name: DefaultGreeting}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// WithName sets the service name.
func WithName(name string) ServiceOption {
	return func(s *Service) {
		s.Name = name
	}
}

// FormatName uses the formatter when available.
func (c *Config) FormatName(name string) string {
	if c.Formatter == nil {
		return fmt.Sprintf("**%s**", name)
	}
	return c.Formatter.Format(name)
}

// Hook represents a service hook.
type Hook func(*Service) error

// ID is a typed identifier for services.
type ID string

// ServiceMap groups services by identifier.
type ServiceMap map[ID]*Service

const (
	// DefaultGreeting is the fallback message.
	DefaultGreeting = "hello"
)

// GlobalService is a reusable default service.
var GlobalService = &Service{Name: DefaultGreeting}
`

	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "service.go"), []byte(source), 0o644))

	var buff bytes.Buffer
	producer := NewProducer().
		Writer(&buff).
		Module(tdir).
		Include(pkgDir).
		IndexConfig(`{"title":"Integration Docs","version":"0.1.0","author":"Doc Bot","email":"doc@example.com","highlight":"none","images":"./img","web":"https://example.com","doctype":"article"}`)

	overrideAllDefaults(t, producer)

	producer.Generate()

	got := buff.String()
	goldenPath := filepath.Join("testdata", "producer_basic.golden")
	if update := os.Getenv("UPDATE_GOLDEN"); update != "" {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0o755))
		require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o644))
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err)
	assert.Equal(t, string(want), got)
}

func overrideAllDefaults(t *testing.T, p *Producer) {
	t.Helper()
	defaultsDir := filepath.Join("..", "defaults")
	entries, err := os.ReadDir(defaultsDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		p.OverrideFilePath(name, filepath.Join(defaultsDir, entry.Name()))
	}
}
