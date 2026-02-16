package layout

import (
	"testing"
)

// ---------------------------------------------------------------------------
// ModeForWidth
// ---------------------------------------------------------------------------

func TestModeForWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  LayoutMode
	}{
		{"wide terminal returns ThreeColumn", 120, ThreeColumn},
		{"medium terminal returns TwoColumn", 90, TwoColumn},
		{"narrow terminal returns SingleColumn", 60, SingleColumn},
		{"exact ThreeColumnMin returns ThreeColumn", ThreeColumnMin, ThreeColumn},
		{"exact TwoColumnMin returns TwoColumn", TwoColumnMin, TwoColumn},
		{"one below ThreeColumnMin returns TwoColumn", ThreeColumnMin - 1, TwoColumn},
		{"one below TwoColumnMin returns SingleColumn", TwoColumnMin - 1, SingleColumn},
		{"zero width returns SingleColumn", 0, SingleColumn},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ModeForWidth(tc.width)
			if got != tc.want {
				t.Errorf("ModeForWidth(%d) = %d, want %d", tc.width, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeColumns
// ---------------------------------------------------------------------------

func TestComputeColumnsThreeColumn(t *testing.T) {
	cols := ComputeColumns(120)
	if cols.Mode != ThreeColumn {
		t.Fatalf("expected ThreeColumn mode, got %d", cols.Mode)
	}
	if cols.Left != LeftWidth {
		t.Errorf("Left = %d, want %d", cols.Left, LeftWidth)
	}
	// Center = 120 - 20 - 18 - 2 dividers = 80
	wantCenter := 120 - LeftWidth - RightWidth - 2
	if cols.Center != wantCenter {
		t.Errorf("Center = %d, want %d", cols.Center, wantCenter)
	}
	if cols.Right != RightWidth {
		t.Errorf("Right = %d, want %d", cols.Right, RightWidth)
	}
}

func TestComputeColumnsTwoColumn(t *testing.T) {
	cols := ComputeColumns(90)
	if cols.Mode != TwoColumn {
		t.Fatalf("expected TwoColumn mode, got %d", cols.Mode)
	}
	if cols.Left != LeftWidth {
		t.Errorf("Left = %d, want %d", cols.Left, LeftWidth)
	}
	// Center = 90 - 20 - 1 divider = 69
	wantCenter := 90 - LeftWidth - 1
	if cols.Center != wantCenter {
		t.Errorf("Center = %d, want %d", cols.Center, wantCenter)
	}
	if cols.Right != 0 {
		t.Errorf("Right = %d, want 0", cols.Right)
	}
}

func TestComputeColumnsSingleColumn(t *testing.T) {
	cols := ComputeColumns(60)
	if cols.Mode != SingleColumn {
		t.Fatalf("expected SingleColumn mode, got %d", cols.Mode)
	}
	if cols.Left != 0 {
		t.Errorf("Left = %d, want 0", cols.Left)
	}
	if cols.Center != 60 {
		t.Errorf("Center = %d, want 60", cols.Center)
	}
	if cols.Right != 0 {
		t.Errorf("Right = %d, want 0", cols.Right)
	}
}

func TestComputeColumnsCenterMinimum(t *testing.T) {
	// Edge case: width so small that center would be negative
	cols := ComputeColumns(0)
	if cols.Center < 0 {
		t.Errorf("Center should never be negative, got %d", cols.Center)
	}
}

// ---------------------------------------------------------------------------
// RenderColumns
// ---------------------------------------------------------------------------

func TestRenderColumnsThreeColumn(t *testing.T) {
	result := RenderColumns(120, 10, "left", "center", "right")
	if result == "" {
		t.Error("RenderColumns(120, 10, ...) returned empty string")
	}
}

func TestRenderColumnsTwoColumn(t *testing.T) {
	result := RenderColumns(90, 10, "left", "center", "right")
	if result == "" {
		t.Error("RenderColumns(90, 10, ...) returned empty string")
	}
}

func TestRenderColumnsSingleColumn(t *testing.T) {
	result := RenderColumns(60, 10, "left", "center", "right")
	if result == "" {
		t.Error("RenderColumns(60, 10, ...) returned empty string")
	}
}

func TestRenderColumnsZeroHeight(t *testing.T) {
	// Should not panic with zero height.
	result := RenderColumns(120, 0, "left", "center", "right")
	_ = result // just ensure no panic
}

func TestRenderColumnsEmptyContent(t *testing.T) {
	// Should not panic with empty content strings.
	result := RenderColumns(120, 5, "", "", "")
	if result == "" {
		t.Error("RenderColumns with empty content returned empty string")
	}
}
