package csv

// Options configures encoding/csv.Reader for a single read operation.
// Zero value is not ready-to-use; callers should start from DefaultOptions()
// and override fields as needed.
type Options struct {
	// FieldsPerRecord is passed to csv.Reader.FieldsPerRecord.
	// -1 = variable field count (default).
	FieldsPerRecord int

	// LazyQuotes allows bare quotes in unquoted fields (default true).
	// Needed for many real-world race/CSV dumps.
	LazyQuotes bool

	// TrimLeadingSpace is passed to csv.Reader.TrimLeadingSpace.
	TrimLeadingSpace bool

	// Comma is the field delimiter; 0 means ',' (csv package default).
	Comma rune

	// Comment, if non-zero, is the comment character for csv.Reader.
	Comment rune
}

// DefaultOptions matches the historical FLAG CSV behavior used by FRS:
// variable-width records and LazyQuotes enabled.
func DefaultOptions() Options {
	return Options{
		FieldsPerRecord:  -1,
		LazyQuotes:       true,
		TrimLeadingSpace: false,
		Comma:            0,
		Comment:          0,
	}
}
