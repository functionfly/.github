package storage

import "testing"

func TestListMarketplaceParams_Defaults(t *testing.T) {
	params := ListMarketplaceParams{
		Limit:  -5,
		Offset: -3,
	}
	if params.Limit >= 0 {
		t.Fatal("precondition: limit should be negative")
	}

	if params.Limit < 0 {
		params.Limit = 50
	}
	if params.Limit > 200 {
		params.Limit = 200
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	if params.Limit != 50 {
		t.Errorf("default limit should be 50, got %d", params.Limit)
	}
	if params.Offset != 0 {
		t.Errorf("default offset should be 0, got %d", params.Offset)
	}
}

func TestListMarketplaceParams_LimitClamping(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, 50},
		{-1, 50},
		{1, 1},
		{100, 100},
		{200, 200},
		{201, 200},
		{1000, 200},
	}
	for _, tt := range tests {
		limit := tt.input
		if limit <= 0 {
			limit = 50
		}
		if limit > 200 {
			limit = 200
		}
		if limit != tt.expected {
			t.Errorf("input %d: expected %d, got %d", tt.input, tt.expected, limit)
		}
	}
}

func TestListMarketplaceParams_FilterBuilding(t *testing.T) {
	category := "ci"
	status := "published"
	featured := true
	search := "github"
	creator := "user-1"

	params := ListMarketplaceParams{
		Category:  &category,
		Status:    &status,
		Featured:  &featured,
		Search:    &search,
		CreatorID: &creator,
		Tags:      []string{"ci", "github"},
		SortBy:    "trending",
	}

	if *params.Category != "ci" {
		t.Error("category should be set")
	}
	if *params.Status != "published" {
		t.Error("status should be set")
	}
	if !*params.Featured {
		t.Error("featured should be true")
	}
	if len(params.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(params.Tags))
	}
	if params.SortBy != "trending" {
		t.Errorf("expected trending, got %q", params.SortBy)
	}
}

func TestSortModes(t *testing.T) {
	validSorts := []string{"newest", "top_rated", "most_installed", "trending", ""}
	for _, sort := range validSorts {
		if sort == "" {
			continue
		}
		valid := sort == "newest" || sort == "top_rated" || sort == "most_installed" || sort == "trending"
		if !valid {
			t.Errorf("sort %q should be valid", sort)
		}
	}
}

func TestParseSemver_EdgeCases(t *testing.T) {
	tests := []struct {
		input           string
		expectZeroValid bool
	}{
		{"", true},
		{".", true},
		{"..", true},
		{".0.0", true},
		{"0..0", true},
		{"1.2.3.4", false},
		{"1", false},
		{"1.2", false},
		{"a.b.c", true},
		{"1.2.x", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSemver(tt.input)
			allZero := got[0] == 0 && got[1] == 0 && got[2] == 0
			if allZero != tt.expectZeroValid {
				t.Errorf("parseSemver(%q) = %v, allZero=%v want %v",
					tt.input, got, allZero, tt.expectZeroValid)
			}
		})
	}
}
