package bigtable

import (
	"encoding/hex"
	"fmt"

	"github.com/johanastborg/BlackGrid/pkg/mapping"
	"github.com/johanastborg/BlackGrid/pkg/sql"
	"github.com/johanastborg/BlackGrid/pkg/zcurve"
)

// RowScanRange represents a single Bigtable scan range.
type RowScanRange struct {
	StartKey []byte
	EndKey   []byte
	StartHex string
	EndHex   string
}

// QueryPlan encapsulates the complete execution plan for a Bigtable scan.
type QueryPlan struct {
	TableName          string
	ScanRanges         []RowScanRange
	MortonIntervals    []zcurve.Interval
	HypercubeMin       []float64
	HypercubeMax       []float64
	ResidualPredicates []string
	TotalScanRanges    int
	KeyFormat          KeyFormat
}

// Planner converts GoogleSQL queries into optimized Bigtable scan plans.
type Planner struct{}

// NewPlanner creates a new Bigtable query planner.
func NewPlanner() *Planner {
	return &Planner{}
}

// Plan converts a GoogleSQL string and a TableSchema into an execution plan.
func (p *Planner) Plan(schema *mapping.TableSchema, sqlQuery string, maxRanges int, format KeyFormat) (*QueryPlan, error) {
	stmt, err := sql.ParseQuery(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL query: %w", err)
	}
	return p.PlanAST(schema, stmt, maxRanges, format)
}

// PlanAST converts a parsed SelectStmt and a TableSchema into an execution plan.
func (p *Planner) PlanAST(schema *mapping.TableSchema, stmt *sql.SelectStmt, maxRanges int, format KeyFormat) (*QueryPlan, error) {
	if stmt.Where == nil {
		return nil, fmt.Errorf("query must have a WHERE clause to generate Morton scan ranges")
	}

	extracted, err := sql.ExtractConstraints(stmt.Where)
	if err != nil {
		return nil, fmt.Errorf("failed to extract constraints: %w", err)
	}

	dimCount := len(schema.Dimensions)
	dimNames := make([]string, dimCount)
	for i, d := range schema.Dimensions {
		dimNames[i] = d.Name
	}

	mins, maxs, err := extracted.BuildHypercube(dimNames)
	if err != nil {
		return nil, fmt.Errorf("failed to build dimensional hypercube: %w", err)
	}

	// Quantize continuous bounds to integer coordinate box
	queryBox := zcurve.BoxND{
		Min: make([]uint64, dimCount),
		Max: make([]uint64, dimCount),
	}

	bitsPerDim := schema.Dimensions[0].Bits
	for i, dim := range schema.Dimensions {
		qMin, qMax, err := mapping.QuantizeRange(&dim, mins[i], maxs[i])
		if err != nil {
			return nil, err
		}
		queryBox.Min[i] = qMin
		queryBox.Max[i] = qMax
	}

	decomposer, err := zcurve.NewDecomposer(dimCount, bitsPerDim, maxRanges)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize decomposer: %w", err)
	}

	intervals := decomposer.DecomposeND(queryBox)

	scanRanges := make([]RowScanRange, len(intervals))
	for i, in := range intervals {
		startKey, endKey := BuildRangeKeys(schema.RowKeyPrefix, in.Start, in.End, format)
		scanRanges[i] = RowScanRange{
			StartKey: startKey,
			EndKey:   endKey,
			StartHex: hex.EncodeToString(startKey),
			EndHex:   hex.EncodeToString(endKey),
		}
	}

	residuals := make([]string, len(extracted.Residuals))
	for i, r := range extracted.Residuals {
		residuals[i] = r.String()
	}

	return &QueryPlan{
		TableName:          schema.TableName,
		ScanRanges:         scanRanges,
		MortonIntervals:    intervals,
		HypercubeMin:       mins,
		HypercubeMax:       maxs,
		ResidualPredicates: residuals,
		TotalScanRanges:    len(scanRanges),
		KeyFormat:          format,
	}, nil
}

