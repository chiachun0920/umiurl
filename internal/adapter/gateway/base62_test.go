package gateway

import "testing"

func TestBase62CodeGeneratorGenerate(t *testing.T) {
	generator := NewBase62CodeGenerator(7)

	code, err := generator.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(code) != 7 {
		t.Fatalf("len(code) = %d, want 7", len(code))
	}

	allowed := map[rune]bool{}
	for _, char := range Base62Alphabet {
		allowed[char] = true
	}
	for _, char := range code {
		if !allowed[char] {
			t.Fatalf("code contains non-base62 character %q", char)
		}
	}
}
