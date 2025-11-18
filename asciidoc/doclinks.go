package asciidoc

import (
	"fmt"
	"go/types"
	"regexp"
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser"
	"golang.org/x/tools/go/packages"
)

// DocReference represents a parsed reference from documentation
type DocReference struct {
	Original     string  // Original text from backticks (e.g., "pkg.MyType")
	PackagePath  string  // Full package import path (e.g., "github.com/user/pkg")
	PackageAlias string  // Package alias or name used in reference
	Receiver     string  // Struct/interface name for methods
	Identifier   string  // Type/method/function/package name
	Kind         RefKind // What this reference points to
	IsExternal   bool    // Whether reference is to external package
}

type RefKind int

const (
	RefUnknown RefKind = iota
	RefType        // Struct, interface, type alias, etc.
	RefMethod      // Method on a type
	RefFunction    // Package-level function
	RefPackage     // Package reference
	RefField       // Struct field
	RefConstant    // Package-level constant
	RefVariable    // Package-level variable
)

// backtickPattern matches content within backticks
var backtickPattern = regexp.MustCompile("`([^`]+)`")

// processDocumentation processes documentation strings and replaces backtick references with links
func (t *TemplateContext) processDocumentation(doc string) string {
	if doc == "" || t.Config == nil || t.Config.TypeLinks == TypeLinksDisabled {
		return doc
	}

	return backtickPattern.ReplaceAllStringFunc(doc, func(match string) string {
		// Extract content without backticks
		content := strings.Trim(match, "`")
		if content == "" {
			return match
		}

		// Try to parse and resolve as a reference
		ref := t.parseReference(content)
		if ref == nil {
			// Not a valid reference, return as-is
			return match
		}

		// Try to resolve the reference
		if !t.resolveReference(ref) {
			// Could not resolve, return as-is
			return match
		}

		// Generate link for the resolved reference
		link := t.generateDocLink(ref)
		if link == "" {
			return match
		}

		return link
	})
}

// parseReference parses a backtick reference into components
func (t *TemplateContext) parseReference(ref string) *DocReference {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}

	// Check if it's a builtin type - don't link these
	if _, ok := builtinTypes[ref]; ok {
		return nil
	}

	dr := &DocReference{
		Original: ref,
		Kind:     RefUnknown,
	}

	// Split by dots
	parts := strings.Split(ref, ".")

	switch len(parts) {
	case 1:
		// Unqualified identifier: MyType, MyFunc, or package name
		// We'll determine if it's a package during resolution
		dr.Identifier = parts[0]

	case 2:
		// Could be:
		// - pkg.Type (qualified type)
		// - Receiver.Method (method on receiver)
		// We'll determine which during resolution
		// For now, set both and let resolution logic decide
		dr.PackageAlias = parts[0]
		dr.Receiver = parts[0] // Also set as potential receiver
		dr.Identifier = parts[1]

	case 3:
		// pkg.Receiver.Method or nested package path
		// We'll try both interpretations during resolution
		dr.PackageAlias = parts[0]
		dr.Receiver = parts[1]
		dr.Identifier = parts[2]

	default:
		// Fully qualified path like github.com/user/pkg.Type
		// or github.com/user/pkg.Receiver.Method
		// Find where package path ends
		dr.PackagePath, dr.Receiver, dr.Identifier = t.splitFullyQualifiedRef(ref, parts)
		if dr.PackagePath != "" {
			// Extract package alias from path
			dr.PackageAlias = parts[len(parts)-2]
			if dr.Receiver != "" {
				dr.PackageAlias = parts[len(parts)-3]
			}
		}
	}

	return dr
}

// splitFullyQualifiedRef splits a fully qualified reference like github.com/user/pkg.Type
func (t *TemplateContext) splitFullyQualifiedRef(ref string, parts []string) (pkgPath, receiver, identifier string) {
	// Look for package path containing '/'
	lastSlashIdx := strings.LastIndex(ref, "/")
	if lastSlashIdx == -1 {
		// No slash, not a full path
		return "", "", ""
	}

	// Everything up to the last package component is the path
	// Find the first '.' after the last '/'
	dotAfterSlashIdx := strings.Index(ref[lastSlashIdx:], ".")
	if dotAfterSlashIdx == -1 {
		// No dot after slash, could be just a package reference
		return ref, "", ""
	}

	dotAfterSlashIdx += lastSlashIdx
	pkgPath = ref[:dotAfterSlashIdx]
	rest := ref[dotAfterSlashIdx+1:]

	// rest is either "Type" or "Receiver.Method"
	dotIdx := strings.Index(rest, ".")
	if dotIdx == -1 {
		identifier = rest
	} else {
		receiver = rest[:dotIdx]
		identifier = rest[dotIdx+1:]
	}

	return pkgPath, receiver, identifier
}

// resolveReference attempts to resolve a reference using imports and type information
func (t *TemplateContext) resolveReference(ref *DocReference) bool {
	if ref == nil {
		return false
	}

	// Special case: single identifier that might be a package name
	// Check if it's in the imports
	if ref.PackageAlias == "" && ref.Receiver == "" && ref.Identifier != "" {
		if pkgPath := t.importPathForAlias(ref.Identifier, t.File); pkgPath != "" {
			// It's a package reference
			ref.PackagePath = pkgPath
			ref.PackageAlias = ref.Identifier
			ref.Identifier = "" // Clear identifier to mark as package-only
		}
	}

	// For two-part references, determine if first part is package or receiver
	if ref.PackageAlias != "" && ref.Receiver != "" && ref.PackageAlias == ref.Receiver {
		// Both are set to the same value, need to disambiguate
		if pkgPath := t.importPathForAlias(ref.PackageAlias, t.File); pkgPath != "" {
			// It's a package reference (pkg.Type or pkg.Function)
			ref.PackagePath = pkgPath
			ref.Receiver = "" // Clear receiver
		} else {
			// It's likely a receiver reference (Type.Method)
			ref.PackageAlias = "" // Clear package alias
			// Keep receiver and use current package
		}
	}

	// Now resolve package path if not already set
	if ref.PackagePath == "" && ref.PackageAlias != "" {
		ref.PackagePath = t.resolvePackagePath(ref.PackageAlias)
	}

	// If still no package path and we have an identifier (not package-only ref),
	// use current file's package
	if ref.PackagePath == "" && ref.Identifier != "" && t.File != nil {
		ref.PackagePath = t.packagePathForFile(t.File)
	}

	// Determine if this is internal or external
	if ref.PackagePath != "" {
		// Standard library packages (no dots/slashes) are always external
		if !strings.Contains(ref.PackagePath, ".") && !strings.Contains(ref.PackagePath, "/") {
			ref.IsExternal = true
		} else {
			ref.IsExternal = !t.isInternalImport(ref.PackagePath)
		}
	}

	// Try to determine the kind of reference
	// For external references, we'll make best-effort guesses
	if ref.IsExternal {
		// For external refs, guess based on naming conventions
		if ref.Receiver != "" {
			ref.Kind = RefMethod
		} else if ref.PackageAlias == "" && ref.PackagePath != "" && ref.Identifier == "" {
			ref.Kind = RefPackage
		} else if ref.Identifier != "" {
			// Could be type or function - default to type
			ref.Kind = RefType
		}
		return true
	}

	// For internal references, try to actually resolve using type information
	// This is best-effort - if we can't load package info, we'll still generate links
	return t.resolveInternalReference(ref)
}

// resolvePackagePath resolves a package alias to a full import path
func (t *TemplateContext) resolvePackagePath(alias string) string {
	// Try to resolve from imports
	if path := t.importPathForAlias(alias, t.File); path != "" {
		return path
	}

	// Check if it's a standard library package (no dots in name)
	if !strings.Contains(alias, ".") && !strings.Contains(alias, "/") {
		// Likely a standard library package like "fmt", "os", etc.
		return alias
	}

	// Could be a package name without import (same package)
	// or a reference that doesn't resolve
	return ""
}

// resolveInternalReference attempts to resolve an internal reference using type information
func (t *TemplateContext) resolveInternalReference(ref *DocReference) bool {
	// For now, we'll do basic resolution based on naming
	// Full type checking would require loading package information

	if ref.Receiver != "" {
		// Has receiver, likely a method
		ref.Kind = RefMethod
		return true
	}

	if ref.Identifier == "" {
		// No identifier, likely a package reference
		ref.Kind = RefPackage
		return true
	}

	// Check if identifier starts with uppercase (exported)
	if len(ref.Identifier) > 0 {
		first := rune(ref.Identifier[0])
		if first >= 'A' && first <= 'Z' {
			// Could be Type, Function, Constant, or Variable
			// Default to Type as it's most common in docs
			ref.Kind = RefType
			return true
		}
	}

	// Lowercase, less common in docs but could be unexported type/function
	ref.Kind = RefType
	return true
}

// generateDocLink generates an AsciiDoc link for a resolved reference
func (t *TemplateContext) generateDocLink(ref *DocReference) string {
	if ref == nil || ref.PackagePath == "" {
		return ""
	}

	linkText := ref.Original

	// Handle package-only references
	if ref.Kind == RefPackage {
		if ref.IsExternal {
			if t.Config.TypeLinks == TypeLinksInternalExternal {
				return fmt.Sprintf("link:https://pkg.go.dev/%s[%s]", ref.PackagePath, linkText)
			}
			return "`" + linkText + "`"
		}
		// Internal package reference - link to package anchor
		anchor := anchorID(ref.PackagePath, "")
		return fmt.Sprintf("<<%s,%s>>", strings.TrimSuffix(anchor, "."), linkText)
	}

	// Build anchor ID
	var anchor string
	if ref.Receiver != "" {
		// Method reference: pkg.Receiver.Method
		anchor = anchorID(ref.PackagePath, ref.Receiver+"."+ref.Identifier)
	} else {
		// Type/Function reference: pkg.Identifier
		anchor = anchorID(ref.PackagePath, ref.Identifier)
	}

	// Generate link based on internal/external and mode
	if ref.IsExternal {
		if t.Config.TypeLinks == TypeLinksInternalExternal {
			return t.generateExternalLink(ref, linkText)
		}
		// External but not linking external refs
		return "`" + linkText + "`"
	}

	// Internal reference
	return t.generateInternalLink(ref, anchor, linkText)
}

// generateExternalLink creates a pkg.go.dev link
func (t *TemplateContext) generateExternalLink(ref *DocReference, linkText string) string {
	switch ref.Kind {
	case RefPackage:
		return fmt.Sprintf("link:https://pkg.go.dev/%s[%s]", ref.PackagePath, linkText)

	case RefMethod:
		if ref.Receiver != "" {
			return fmt.Sprintf("link:https://pkg.go.dev/%s#%s.%s[%s]",
				ref.PackagePath, ref.Receiver, ref.Identifier, linkText)
		}
		return fmt.Sprintf("link:https://pkg.go.dev/%s#%s[%s]",
			ref.PackagePath, ref.Identifier, linkText)

	case RefType, RefFunction, RefConstant, RefVariable:
		return fmt.Sprintf("link:https://pkg.go.dev/%s#%s[%s]",
			ref.PackagePath, ref.Identifier, linkText)

	default:
		return fmt.Sprintf("link:https://pkg.go.dev/%s#%s[%s]",
			ref.PackagePath, ref.Identifier, linkText)
	}
}

// generateInternalLink creates an internal AsciiDoc link
func (t *TemplateContext) generateInternalLink(ref *DocReference, anchor, linkText string) string {
	// Check if this is a cross-module reference in separate mode
	if t.Config.SubModuleMode == SubModuleSeparate && t.isWorkspaceImport(ref.PackagePath) {
		// Cross-module reference - use file link
		targetModule := t.Workspace.ModuleForPath(ref.PackagePath)
		if targetModule != nil {
			shortName := goparser.ModuleShortName(targetModule)
			return fmt.Sprintf("link:%s.adoc#%s[%s]", shortName, anchor, linkText)
		}
	}

	// Same module or single/merged mode - use anchor link
	return fmt.Sprintf("<<%s,%s>>", anchor, linkText)
}

// loadPackageTypes loads type information for a package
func (t *TemplateContext) loadPackageTypes(pkgPath string) (*types.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedName | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package not found: %s", pkgPath)
	}

	if pkgs[0].Types == nil {
		return nil, fmt.Errorf("no type information for package: %s", pkgPath)
	}

	return pkgs[0].Types, nil
}
