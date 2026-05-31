package runner

import "testing"

func TestParseSize(t *testing.T) {
	tests := []struct {
		in   string
		want int64
	}{
		{"", 0},
		{"42", 42},
		{"1KB", 1024},
		{"1.5MB", 1572864},
		{"2G", 2 << 30},
	}
	for _, tt := range tests {
		got, err := ParseSize(tt.in)
		if err != nil {
			t.Fatalf("ParseSize(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("ParseSize(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestBudgetExhausted(t *testing.T) {
	b := NewBudget(100)
	if b.Exhausted() {
		t.Fatal("new budget should not be exhausted")
	}
	b.Add(40)
	if b.Exhausted() {
		t.Fatal("budget exhausted too early")
	}
	b.Add(60)
	if !b.Exhausted() {
		t.Fatal("budget should be exhausted at cap")
	}
}
