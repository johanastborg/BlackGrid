package bigtable

import (
	"bytes"
	"testing"
	"time"

	"github.com/johanastborg/BlackGrid/pkg/mapping"
)

func TestPlanner_2DGeo(t *testing.T) {
	schema, err := mapping.NewGeo2DSchema("geo_events", "geo#")
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	planner := NewPlanner()

	sqlQuery := `SELECT id, latitude, longitude, event_type 
		FROM geo_events 
		WHERE latitude BETWEEN 37.75 AND 37.78 
		  AND longitude BETWEEN -122.45 AND -122.40 
		  AND event_type = 'CLICK'`

	plan, err := pPlan(planner, schema, sqlQuery, 16, HexFormat)
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if plan.TableName != "geo_events" {
		t.Errorf("expected table geo_events, got %s", plan.TableName)
	}

	if len(plan.ScanRanges) == 0 {
		t.Fatal("expected at least 1 scan range")
	}

	if len(plan.ScanRanges) > 16 {
		t.Errorf("expected at most 16 scan ranges, got %d", len(plan.ScanRanges))
	}

	// Verify lexicographical order of scan ranges
	for i := 0; i < len(plan.ScanRanges)-1; i++ {
		cur := plan.ScanRanges[i]
		next := plan.ScanRanges[i+1]
		if bytes.Compare(cur.StartKey, cur.EndKey) > 0 {
			t.Errorf("range %d: start key > end key", i)
		}
		if bytes.Compare(cur.EndKey, next.StartKey) > 0 {
			t.Errorf("range %d end key > range %d start key", i, i+1)
		}
	}

	// Verify residual filter extracted
	if len(plan.ResidualPredicates) != 1 {
		t.Errorf("expected 1 residual predicate, got %d: %+v", len(plan.ResidualPredicates), plan.ResidualPredicates)
	}
}

func TestPlanner_3DSpatioTemporal(t *testing.T) {
	start := time.Unix(1600000000, 0)
	end := time.Unix(1700000000, 0)
	schema, err := mapping.NewSpatioTemporalSchema("telemetry", "tele#", start, end)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	planner := NewPlanner()
	sqlQuery := `SELECT * FROM telemetry 
		WHERE latitude BETWEEN 10.0 AND 20.0 
		  AND longitude BETWEEN 50.0 AND 60.0 
		  AND timestamp BETWEEN 1610000000 AND 1620000000`

	plan, err := planner.Plan(schema, sqlQuery, 32, BinaryFormat)
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if len(plan.ScanRanges) == 0 {
		t.Fatal("expected at least 1 scan range")
	}

	// Validate Bigtable row key binary format prefix
	for _, r := range plan.ScanRanges {
		if !bytes.HasPrefix(r.StartKey, []byte("tele#")) {
			t.Errorf("start key missing prefix tele#: %q", r.StartKey)
		}
	}
}

func pPlan(p *Planner, schema *mapping.TableSchema, sqlQuery string, maxRanges int, format KeyFormat) (*QueryPlan, error) {
	return p.Plan(schema, sqlQuery, maxRanges, format)
}

