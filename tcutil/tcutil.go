package tcutil

import (
	"fmt"
	"strconv"
)

// DefaultAsIEC toggles whether the IEC format is used by default for byte
// counts/rates/limits' string conversion.
var DefaultAsIEC bool = true

// ByteFormatter is the shared interface for byte counts, rates, and limits.
type ByteFormatter interface {
	// Int64 returns the actual value in a int64.
	Int64() int64

	// Format formats the value in a human readable string, with the passed
	// precision.
	Format(bool, int) string

	// String satisfies the fmt.Stringer interface, returning a human readable
	// string.
	String() string

	// Add adds the passed value to itself, returning the sum of the two values.
	Add(interface{}) interface{}
}

// Format formats size i using the supplied precision, as a human readable
// string (1 B, 2 kB, 3 MB, 5 GB, ...). When asIEC is true, will format the
// amount as a IEC size (1 B, 2 KiB, 4 MiB, 5 GiB, ...)
func Format(i int64, asIEC bool, precision int) string {
	c, sizes, end := int64(1000), "kMGTPEZY", "B"
	if asIEC {
		c, sizes, end = 1024, "KMGTPEZY", "iB"
	}
	if i < c {
		return fmt.Sprintf("%d B", i)
	}
	exp, div := 0, c
	for n := i / c; n >= c; n /= c {
		div *= c
		exp++
	}
	return fmt.Sprintf("%."+strconv.Itoa(precision)+"f %c%s", float64(i)/float64(div), sizes[exp], end)
}

// ByteCount wraps a byte count as int64.
type ByteCount int64

// Int64 returns the byte count as an int64.
func (bc ByteCount) Int64() int64 {
	return int64(bc)
}

// Format formats the byte count.
func (bc ByteCount) Format(asIEC bool, prec int) string {
	return Format(int64(bc), asIEC, prec)
}

// String satisfies the fmt.Stringer interface.
func (bc ByteCount) String() string {
	return bc.Format(DefaultAsIEC, 2)
}

// Add adds i to the byte count.
func (bc ByteCount) Add(i interface{}) interface{} {
	return bc + i.(ByteCount)
}

// Rate is a bytes per second rate.
type Rate int64

// Int64 returns the rate as an int64.
func (r Rate) Int64() int64 {
	return int64(r)
}

// Format formats the rate.
func (r Rate) Format(asIEC bool, prec int) string {
	return Format(int64(r), asIEC, prec) + "/s"
}

// String satisfies the fmt.Stringer interface.
func (r Rate) String() string {
	return r.Format(DefaultAsIEC, 2)
}

// Add adds i to the byte count.
func (r Rate) Add(i interface{}) interface{} {
	return r + i.(Rate)
}

// Limit is a K bytes per second limit.
type Limit int64

// Int64 returns the rate as an int64.
func (l Limit) Int64() int64 {
	return int64(l)
}

// Format formats the rate.
func (l Limit) Format(asIEC bool, prec int) string {
	return Format(int64(l*1000), asIEC, prec) + "/s"
}

// String satisfies the fmt.Stringer interface.
func (l Limit) String() string {
	return l.Format(DefaultAsIEC, 2)
}

// Add adds i to the byte count.
func (l Limit) Add(i interface{}) interface{} {
	return l + i.(Limit)
}

// Percent wraps a float64.
type Percent float64

// String satisfies the fmt.Stringer interface.
func (p Percent) String() string {
	return fmt.Sprintf("%.f%%", float64(p)*100)
}
