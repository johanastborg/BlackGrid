package sql

import (
	"testing"
)

func TestParseQuery_Basic(t *testing.T) {
	query := "SELECT id, lat, lon FROM locations WHERE lat BETWEEN 37.0 AND 38.0 AND lon >= -122.5 AND lon <= -122.0 LIMIT 50"
	stmt, err := ParseQuery(query)
	if err != nil {
		t.Fatalf("unexpected error parsing query: %v", err)
	}

	if stmt.Table != "locations" {
		t.Errorf("expected table 'locations', got %q", stmt.Table)
	}

	if len(stmt.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(stmt.Fields))
	}

	if stmt.Limit != 50 {
		t.Errorf("expected limit 50, got %d", stmt.Limit)
	}
}

func TestParseQuery_SelectStar(t *testing.T) {
	query := "SELECT * FROM `my_project.dataset.sensors` WHERE x >= 10.5"
	stmt, err := ParseQuery(query)
	if err != nil {
		t.Fatalf("unexpected error parsing query: %v", err)
	}

	if len(stmt.Fields) != 1 || stmt.Fields[0] != "*" {
		t.Errorf("expected [*], got %+v", stmt.Fields)
	}

	if stmt.Table != "my_project.dataset.sensors" {
		t.Errorf("expected table 'my_project.dataset.sensors', got %q", stmt.Table)
	}
}

func TestExtractConstraints(t *testing.T) {
	query := `SELECT * FROM geo_data 
		WHERE latitude BETWEEN 37.70 AND 37.85 
		  AND longitude >= -122.50 
		  AND -122.35 >= longitude
		  AND status = 'ACTIVE'`

	stmt, err := ParseQuery(query)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	extracted, err := ExtractConstraints(stmt.Where)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}

	lat, ok := extracted.Constraints["latitude"]
	if !ok || lat.Min != 37.70 || lat.Max != 37.85 {
		t.Errorf("unexpected latitude constraint: %+v", lat)
	}

	lon, ok := extracted.Constraints["longitude"]
	if !ok || lon.Min != -122.50 || lon.Max != -122.35 {
		t.Errorf("unexpected longitude constraint: %+v", lon)
	}

	// status = 'ACTIVE' should be in residuals!
	if len(extracted.Residuals) != 1 {
		t.Errorf("expected 1 residual filter, got %d: %+v", len(extracted.Residuals), extracted.Residuals)
	}

	// Test BuildHypercube
	mins, maxs, err := extracted.BuildHypercube([]string{"latitude", "longitude"})
	if err != nil {
		t.Fatalf("BuildHypercube error: %v", err)
	}

	if mins[0] != 37.70 || maxs[0] != 37.85 || mins[1] != -122.50 || maxs[1] != -122.35 {
		t.Errorf("hypercube values incorrect: mins=%v, maxs=%v", mins, maxs)
	}
}

