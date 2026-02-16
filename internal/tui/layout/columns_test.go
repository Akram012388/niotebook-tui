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
// ComputeColumns — percentage-based 20/60/20 split
// ---------------------------------------------------------------------------

func TestComputeColumnsThreeColumn(t *testing.T) {
	cols := ComputeColumns(120)
	if cols.Mode != ThreeColumn {
		t.Fatalf("expected ThreeColumn mode, got %d", cols.Mode)
	}
	// 120 * 20 / 100 = 24 per sidebar
	wantSidebar := 120 * SidebarPercent / 100
	if cols.Left != wantSidebar {
		t.Errorf("Left = %d, want %d", cols.Left, wantSidebar)
	}
	if cols.Right != wantSidebar {
		t.Errorf("Right = %d, want %d", cols.Right, wantSidebar)
	}
	// Center = 120 - 2*24 - 2 dividers = 70
	wantCenter := 120 - 2*wantSidebar - 2
	if cols.Center != wantCenter {
		t.Errorf("Center = %d, want %d", cols.Center, wantCenter)
	}
}

func TestComputeColumnsSymmetricSidebars(t *testing.T) {
	cols := ComputeColumns(140)
	if cols.Left != cols.Right {
		t.Errorf("sidebars should be symmetric: Left=%d, Right=%d", cols.Left, cols.Right)
	}
}

func TestComputeColumnsTwoColumn(t *testing.T) {
	cols := ComputeColumns(90)
	if cols.Mode != TwoColumn {
		t.Fatalf("expected TwoColumn mode, got %d", cols.Mode)
	}
	wantSidebar := 90 * SidebarPercent / 100
	if cols.Left != wantSidebar {
		t.Errorf("Left = %d, want %d", cols.Left, wantSidebar)
	}
	// Center = 90 - 18 - 1 divider = 71
	wantCenter := 90 - wantSidebar - 1
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

func TestComputeColumnsWidthSumsCorrectly(t *testing.T) {
	for _, w := range []int{100, 120, 140, 160, 200} {
		cols := ComputeColumns(w)
		dividers := 2 // left|center|right
		total := cols.Left + cols.Center + cols.Right + dividers
		if total != w {
			t.Errorf("width %d: Left(%d)+Center(%d)+Right(%d)+dividers(%d)=%d, want %d",
				w, cols.Left, cols.Center, cols.Right, dividers, total, w)
		}
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

// ---------------------------------------------------------------------------
// FocusState
// ---------------------------------------------------------------------------

func TestFocusStateDefault(t *testing.T) {
	fs := NewFocusState()
	if fs.Active() != FocusCenter {
		t.Errorf("default focus = %d, want FocusCenter", fs.Active())
	}
}

func TestFocusStateNextCycles(t *testing.T) {
	fs := NewFocusState()
	fs.Next() // Center(1) -> Right(2)
	if fs.Active() != FocusRight {
		t.Errorf("after Next: focus = %d, want FocusRight", fs.Active())
	}
	fs.Next() // Right(2) -> Left(0)
	if fs.Active() != FocusLeft {
		t.Errorf("after Next×2: focus = %d, want FocusLeft", fs.Active())
	}
	fs.Next() // Left(0) -> Center(1)
	if fs.Active() != FocusCenter {
		t.Errorf("after Next×3: focus = %d, want FocusCenter", fs.Active())
	}
}

func TestFocusStatePrevCycles(t *testing.T) {
	fs := NewFocusState()
	fs.Prev() // Center(1) -> Left(0)
	if fs.Active() != FocusLeft {
		t.Errorf("after Prev: focus = %d, want FocusLeft", fs.Active())
	}
	fs.Prev() // Left(0) -> Right(2)
	if fs.Active() != FocusRight {
		t.Errorf("after Prev×2: focus = %d, want FocusRight", fs.Active())
	}
	fs.Prev() // Right(2) -> Center(1)
	if fs.Active() != FocusCenter {
		t.Errorf("after Prev×3: focus = %d, want FocusCenter", fs.Active())
	}
}

func TestFocusStateReset(t *testing.T) {
	fs := NewFocusState()
	fs.Next()
	fs.Reset()
	if fs.Active() != FocusCenter {
		t.Errorf("after Reset: focus = %d, want FocusCenter", fs.Active())
	}
}
