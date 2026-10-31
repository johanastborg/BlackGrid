package mapping

import (
	"errors"
	"fmt"
	"time"
)

// DimensionType identifies the real-world data type of a dimension.
type DimensionType string

const (
	TypeFloat64   DimensionType = "FLOAT64"
	TypeInt64     DimensionType = "INT64"
	TypeTimestamp DimensionType = "TIMESTAMP"
)

// Dimension defines a single dimension mapped into the Morton space.
type Dimension struct {
	Name    string
	Type    DimensionType
	MinVal  float64 // Minimum domain value
	MaxVal  float64 // Maximum domain value
	Bits    int     // Bits allocated for this dimension
}

// Validate checks dimension configuration validity.
func (d Dimension) Validate() error {
	if d.Name == "" {
		return errors.New("dimension name cannot be empty")
	}
	if d.Bits <= 0 || d.Bits > 64 {
		return fmt.Errorf("dimension %q has invalid bit allocation %d (must be 1..64)", d.Name, d.Bits)
	}
	if d.MinVal >= d.MaxVal {
		return fmt.Errorf("dimension %q has invalid range: min %v >= max %v", d.Name, d.MinVal, d.MaxVal)
	}
	return nil
}

// TableSchema defines how a Bigtable table's row keys are structured.
type TableSchema struct {
	TableName      string
	RowKeyPrefix   string      // e.g. "geo#" or "sensor#"
	Dimensions     []Dimension // Ordered list of dimensions interleaved in the Morton code
	TotalBits      int
}

// NewTableSchema creates and validates a TableSchema.
func NewTableSchema(tableName, prefix string, dims []Dimension) (*TableSchema, error) {
	if tableName == "" {
		return nil, errors.New("table name cannot be empty")
	}
	if len(dims) == 0 {
		return nil, errors.New("at least one dimension required")
	}

	totalBits := 0
	dimNames := make(map[string]bool)
	for _, d := range dims {
		if err := d.Validate(); err != nil {
			return nil, err
		}
		if dimNames[d.Name] {
			return nil, fmt.Errorf("duplicate dimension name: %q", d.Name)
		}
		dimNames[d.Name] = true
		totalBits += d.Bits
	}

	if totalBits > 64 {
		return nil, fmt.Errorf("total dimension bits %d exceeds 64", totalBits)
	}

	return &TableSchema{
		TableName:    tableName,
		RowKeyPrefix: prefix,
		Dimensions:   dims,
		TotalBits:    totalBits,
	}, nil
}

// FindDimension returns the dimension with the specified name.
func (s *TableSchema) FindDimension(name string) (*Dimension, int, bool) {
	for i := range s.Dimensions {
		if s.Dimensions[i].Name == name {
			return &s.Dimensions[i], i, true
		}
	}
	return nil, -1, false
}

// Helper constructors for common spatial / temporal schemas

// NewGeo2DSchema creates a standard 2D Geographic (Latitude, Longitude) table schema.
// Bits: 32 bits for lat, 32 bits for lon -> 64-bit Morton code (~sub-meter precision).
func NewGeo2DSchema(tableName, prefix string) (*TableSchema, error) {
	return NewTableSchema(tableName, prefix, []Dimension{
		{Name: "latitude", Type: TypeFloat64, MinVal: -90.0, MaxVal: 90.0, Bits: 32},
		{Name: "longitude", Type: TypeFloat64, MinVal: -180.0, MaxVal: 180.0, Bits: 32},
	})
}

// NewSpatioTemporalSchema creates a 3D (lat, lon, time) schema with 21 bits each (63 bits total).
func NewSpatioTemporalSchema(tableName, prefix string, startTime, endTime time.Time) (*TableSchema, error) {
	return NewTableSchema(tableName, prefix, []Dimension{
		{Name: "latitude", Type: TypeFloat64, MinVal: -90.0, MaxVal: 90.0, Bits: 21},
		{Name: "longitude", Type: TypeFloat64, MinVal: -180.0, MaxVal: 180.0, Bits: 21},
		{Name: "timestamp", Type: TypeTimestamp, MinVal: float64(startTime.Unix()), MaxVal: float64(endTime.Unix()), Bits: 21},
	})
}

