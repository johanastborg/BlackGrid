package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/johanastborg/BlackGrid/pkg/bigtable"
	"github.com/johanastborg/BlackGrid/pkg/mapping"
)

func main() {
	queryFlag := flag.String("query", "", "GoogleSQL query to plan")
	tableFlag := flag.String("table", "geo_events", "Target Bigtable table name")
	prefixFlag := flag.String("prefix", "geo#", "Bigtable row key prefix")
	maxRangesFlag := flag.Int("max-ranges", 16, "Maximum scan ranges for Bigtable")
	formatFlag := flag.String("format", "hex", "Row key format ('hex' or 'binary')")
	flag.Parse()

	query := *queryFlag
	if query == "" {
		if flag.NArg() > 0 {
			query = strings.Join(flag.Args(), " ")
		} else {
			// Default demo query
			query = `SELECT id, latitude, longitude, event_name 
FROM geo_events 
WHERE latitude BETWEEN 37.75 AND 37.80 
  AND longitude BETWEEN -122.45 AND -122.40 
  AND status = 'ACTIVE' 
LIMIT 100`
		}
	}

	fmt.Printf("=== BlackGrid GoogleSQL to Bigtable Morton Scan Planner ===\n\n")
	fmt.Printf("Input Query:\n  %s\n\n", query)

	// Build default 2D geographic schema
	schema, err := mapping.NewGeo2DSchema(*tableFlag, *prefixFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating table schema: %v\n", err)
		os.Exit(1)
	}

	keyFormat := bigtable.HexFormat
	if strings.ToLower(*formatFlag) == "binary" {
		keyFormat = bigtable.BinaryFormat
	}

	planner := bigtable.NewPlanner()
	plan, err := planner.Plan(schema, query, *maxRangesFlag, keyFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Planning error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Target Table: %s\n", plan.TableName)
	fmt.Printf("Dimensions:   ")
	for i, d := range schema.Dimensions {
		if i > 0 {
			fmt.Printf(" x ")
		}
		fmt.Printf("%s [%.4f, %.4f]", d.Name, plan.HypercubeMin[i], plan.HypercubeMax[i])
	}
	fmt.Printf("\n\n")

	if len(plan.ResidualPredicates) > 0 {
		fmt.Printf("Residual Predicates (Post-Filters / Bigtable Column Filters):\n")
		for _, r := range plan.ResidualPredicates {
			fmt.Printf("  • %s\n", r)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("Generated Bigtable Scan Ranges (%d total, max budget %d):\n", plan.TotalScanRanges, *maxRangesFlag)
	for i, r := range plan.ScanRanges {
		mortonInterval := plan.MortonIntervals[i]
		fmt.Printf("  [%2d] Morton Range: [%016x, %016x] (size: %d)\n",
			i+1, mortonInterval.Start, mortonInterval.End, mortonInterval.Size())
		fmt.Printf("       Start Key:    %s\n", r.StartHex)
		fmt.Printf("       End Key:      %s\n", r.EndHex)
	}
	fmt.Println()
}

