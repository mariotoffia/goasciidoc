package goparser

import (
	"strings"
	"testing"
)

func TestHasJSONTag(t *testing.T) {
	tests := []struct {
		name     string
		struct_  *GoStruct
		expected bool
	}{
		{
			name: "struct with json tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name: "Field1",
						Tag:  &GoTag{Value: "`json:\"field1\"`"},
					},
				},
			},
			expected: true,
		},
		{
			name: "struct without json tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name: "Field1",
						Tag:  nil,
					},
				},
			},
			expected: false,
		},
		{
			name: "struct with yaml tags but no json tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name: "Field1",
						Tag:  &GoTag{Value: "`yaml:\"field1\"`"},
					},
				},
			},
			expected: false,
		},
		{
			name: "struct with nested anonymous struct with json tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name: "Field1",
						Tag:  nil,
						AnonymousStruct: &GoStruct{
							Fields: []*GoField{
								{
									Name: "NestedField",
									Tag:  &GoTag{Value: "`json:\"nested\"`"},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name:     "empty struct",
			struct_:  &GoStruct{Fields: []*GoField{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.struct_.HasJSONTag()
			if result != tt.expected {
				t.Errorf("HasJSONTag() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasYAMLTag(t *testing.T) {
	tests := []struct {
		name     string
		struct_  *GoStruct
		expected bool
	}{
		{
			name: "struct with yaml tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name: "Field1",
						Tag:  &GoTag{Value: "`yaml:\"field1\"`"},
					},
				},
			},
			expected: true,
		},
		{
			name: "struct without yaml tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name: "Field1",
						Tag:  nil,
					},
				},
			},
			expected: false,
		},
		{
			name: "struct with json tags but no yaml tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name: "Field1",
						Tag:  &GoTag{Value: "`json:\"field1\"`"},
					},
				},
			},
			expected: false,
		},
		{
			name: "struct with nested anonymous struct with yaml tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name: "Field1",
						Tag:  nil,
						AnonymousStruct: &GoStruct{
							Fields: []*GoField{
								{
									Name: "NestedField",
									Tag:  &GoTag{Value: "`yaml:\"nested\"`"},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.struct_.HasYAMLTag()
			if result != tt.expected {
				t.Errorf("HasYAMLTag() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name     string
		struct_  *GoStruct
		expected string
	}{
		{
			name: "simple struct with json tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Name",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"name\"`"},
					},
					{
						Name:     "Age",
						Type:     "int",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"age\"`"},
					},
				},
			},
			expected: `{
  "name": "example",
  "age": 0
}`,
		},
		{
			name: "struct with omitempty tag",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Email",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"email,omitempty\"`"},
					},
				},
			},
			expected: `{
  "email": "example"
}`,
		},
		{
			name: "struct with ignored field",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Name",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"name\"`"},
					},
					{
						Name:     "Secret",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"-\"`"},
					},
				},
			},
			expected: `{
  "name": "example"
}`,
		},
		{
			name: "struct with slice",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Tags",
						Type:     "[]string",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"tags\"`"},
					},
				},
			},
			expected: `{
  "tags": ["example"]
}`,
		},
		{
			name: "struct with bool and float",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Active",
						Type:     "bool",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"active\"`"},
					},
					{
						Name:     "Score",
						Type:     "float64",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"score\"`"},
					},
				},
			},
			expected: `{
  "active": false,
  "score": 0.0
}`,
		},
		{
			name: "struct with map",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Metadata",
						Type:     "map[string]string",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"metadata\"`"},
					},
				},
			},
			expected: `{
  "metadata": {}
}`,
		},
		{
			name: "struct with pointer",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Name",
						Type:     "*string",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"name\"`"},
					},
				},
			},
			expected: `{
  "name": "example"
}`,
		},
		{
			name: "struct with anonymous nested struct",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Profile",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"profile\"`"},
						AnonymousStruct: &GoStruct{
							Fields: []*GoField{
								{
									Name:     "Bio",
									Type:     "string",
									Exported: true,
									Tag:      &GoTag{Value: "`json:\"bio\"`"},
								},
							},
						},
					},
				},
			},
			expected: `{
  "profile": {
    "bio": "example"
  }
}`,
		},
		{
			name: "struct with unexported field",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Name",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`json:\"name\"`"},
					},
					{
						Name:     "secret",
						Type:     "string",
						Exported: false,
						Tag:      &GoTag{Value: "`json:\"secret\"`"},
					},
				},
			},
			expected: `{
  "name": "example"
}`,
		},
		{
			name:     "empty struct",
			struct_:  &GoStruct{Fields: []*GoField{}},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.struct_.ToJSON()
			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("ToJSON() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestToYAML(t *testing.T) {
	tests := []struct {
		name     string
		struct_  *GoStruct
		expected string
	}{
		{
			name: "simple struct with yaml tags",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Name",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"name\"`"},
					},
					{
						Name:     "Age",
						Type:     "int",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"age\"`"},
					},
				},
			},
			expected: `name: "example"
age: 0`,
		},
		{
			name: "struct with omitempty tag",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Email",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"email,omitempty\"`"},
					},
				},
			},
			expected: `email: "example"`,
		},
		{
			name: "struct with ignored field",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Name",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"name\"`"},
					},
					{
						Name:     "Secret",
						Type:     "string",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"-\"`"},
					},
				},
			},
			expected: `name: "example"`,
		},
		{
			name: "struct with slice",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Tags",
						Type:     "[]string",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"tags\"`"},
					},
				},
			},
			expected: `tags:
  - "example"`,
		},
		{
			name: "struct with bool and float",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Active",
						Type:     "bool",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"active\"`"},
					},
					{
						Name:     "Score",
						Type:     "float64",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"score\"`"},
					},
				},
			},
			expected: `active: false
score: 0.0`,
		},
		{
			name: "struct with anonymous nested struct",
			struct_: &GoStruct{
				Fields: []*GoField{
					{
						Name:     "Profile",
						Exported: true,
						Tag:      &GoTag{Value: "`yaml:\"profile\"`"},
						AnonymousStruct: &GoStruct{
							Fields: []*GoField{
								{
									Name:     "Bio",
									Type:     "string",
									Exported: true,
									Tag:      &GoTag{Value: "`yaml:\"bio\"`"},
								},
							},
						},
					},
				},
			},
			expected: `profile:
  bio: "example"`,
		},
		{
			name:     "empty struct",
			struct_:  &GoStruct{Fields: []*GoField{}},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.struct_.ToYAML()
			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("ToYAML() = %q, want %q", result, tt.expected)
			}
		})
	}
}
