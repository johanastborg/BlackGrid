package sql

import (
	"fmt"
	"math"
)

// RangeConstraint holds the extracted min and max bounds for a single dimension column.
type RangeConstraint struct {
	ColName string
	Min     float64
	Max     float64
	HasMin  bool
	HasMax  bool
}

// ExtractedQuery holds the extracted dimensional constraints and any residual predicates.
type ExtractedQuery struct {
	Constraints map[string]*RangeConstraint
	Residuals   []Expr
}

// ExtractConstraints analyzes a WHERE clause expression and extracts dimensional range constraints.
func ExtractConstraints(where Expr) (*ExtractedQuery, error) {
	eq := &ExtractedQuery{
		Constraints: make(map[string]*RangeConstraint),
		Residuals:   make([]Expr, 0),
	}

	if where == nil {
		return eq, nil
	}

	if err := eq.traverse(where); err != nil {
		return nil, err
	}

	return eq, nil
}

func (eq *ExtractedQuery) getOrCreate(colName string) *RangeConstraint {
	c, ok := eq.Constraints[colName]
	if !ok {
		c = &RangeConstraint{
			ColName: colName,
			Min:     math.Inf(-1),
			Max:     math.Inf(1),
		}
		eq.Constraints[colName] = c
	}
	return c
}

func (eq *ExtractedQuery) traverse(expr Expr) error {
	switch e := expr.(type) {
	case *ParenExpr:
		return eq.traverse(e.Expr)

	case *BinaryExpr:
		if e.Op == "AND" {
			if err := eq.traverse(e.Left); err != nil {
				return err
			}
			return eq.traverse(e.Right)
		}

		// Comparison: e.g. col >= 10, 10 <= col, col = 5
		handled, err := eq.extractComparison(e)
		if err != nil {
			return err
		}
		if !handled {
			eq.Residuals = append(eq.Residuals, e)
		}
		return nil

	case *BetweenExpr:
		if e.Not {
			// NOT BETWEEN is a non-contiguous residual filter
			eq.Residuals = append(eq.Residuals, e)
			return nil
		}
		col, ok := e.Expr.(*ColRef)
		if !ok {
			eq.Residuals = append(eq.Residuals, e)
			return nil
		}
		low, lowOk := e.Low.(*NumberLit)
		high, highOk := e.High.(*NumberLit)
		if !lowOk || !highOk {
			eq.Residuals = append(eq.Residuals, e)
			return nil
		}

		c := eq.getOrCreate(col.Name)
		if low.Value > c.Min {
			c.Min = low.Value
			c.HasMin = true
		}
		if high.Value < c.Max {
			c.Max = high.Value
			c.HasMax = true
		}
		return nil

	default:
		eq.Residuals = append(eq.Residuals, expr)
		return nil
	}
}

func (eq *ExtractedQuery) extractComparison(b *BinaryExpr) (bool, error) {
	// Case 1: ColRef <op> NumberLit
	if col, ok := b.Left.(*ColRef); ok {
		if num, ok := b.Right.(*NumberLit); ok {
			c := eq.getOrCreate(col.Name)
			val := num.Value
			switch b.Op {
			case ">=":
				if val > c.Min {
					c.Min = val
					c.HasMin = true
				}
			case ">":
				// Closed interval approximation
				if val > c.Min {
					c.Min = val
					c.HasMin = true
				}
			case "<=":
				if val < c.Max {
					c.Max = val
					c.HasMax = true
				}
			case "<":
				if val < c.Max {
					c.Max = val
					c.HasMax = true
				}
			case "=":
				c.Min = val
				c.Max = val
				c.HasMin = true
				c.HasMax = true
			default:
				return false, nil
			}
			return true, nil
		}
	}

	// Case 2: NumberLit <op> ColRef (e.g. 10 <= col)
	if num, ok := b.Left.(*NumberLit); ok {
		if col, ok := b.Right.(*ColRef); ok {
			c := eq.getOrCreate(col.Name)
			val := num.Value
			switch b.Op {
			case "<=": // num <= col  =>  col >= num
				if val > c.Min {
					c.Min = val
					c.HasMin = true
				}
			case "<": // num < col   =>  col > num
				if val > c.Min {
					c.Min = val
					c.HasMin = true
				}
			case ">=": // num >= col  =>  col <= num
				if val < c.Max {
					c.Max = val
					c.HasMax = true
				}
			case ">": // num > col   =>  col < num
				if val < c.Max {
					c.Max = val
					c.HasMax = true
				}
			case "=":
				c.Min = val
				c.Max = val
				c.HasMin = true
				c.HasMax = true
			default:
				return false, nil
			}
			return true, nil
		}
	}

	return false, nil
}

// BuildHypercube constructs min/max arrays for the given dimensions according to the extracted constraints.
func (eq *ExtractedQuery) BuildHypercube(dimNames []string) ([]float64, []float64, error) {
	mins := make([]float64, len(dimNames))
	maxs := make([]float64, len(dimNames))

	for i, name := range dimNames {
		c, ok := eq.Constraints[name]
		if !ok || !c.HasMin || !c.HasMax {
			return nil, nil, fmt.Errorf("missing bounded range constraint for dimension %q", name)
		}
		if c.Min > c.Max {
			return nil, nil, fmt.Errorf("contradictory constraint for %q: min %v > max %v", name, c.Min, c.Max)
		}
		mins[i] = c.Min
		maxs[i] = c.Max
	}

	return mins, maxs, nil
}

