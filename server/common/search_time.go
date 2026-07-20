package common

import (
	"context"
	"time"
)

// SearchTimeRange holds optional inclusive modification-time bounds in Unix milliseconds.
type SearchTimeRange struct {
	From *int64
	To   *int64
}

type searchTimeRangeKeyType struct{}

var searchTimeRangeKey = searchTimeRangeKeyType{}

// WithSearchTimeRange attaches a time range to ctx for search engines to consume.
func WithSearchTimeRange(ctx context.Context, r SearchTimeRange) context.Context {
	return context.WithValue(ctx, searchTimeRangeKey, r)
}

// SearchTimeRangeFromContext returns the time range attached to ctx, if any.
func SearchTimeRangeFromContext(ctx context.Context) SearchTimeRange {
	if ctx == nil {
		return SearchTimeRange{}
	}
	if v, ok := ctx.Value(searchTimeRangeKey).(SearchTimeRange); ok {
		return v
	}
	return SearchTimeRange{}
}

// HasSearchTimeRange reports whether either bound is set.
func HasSearchTimeRange(r SearchTimeRange) bool {
	return r.From != nil || r.To != nil
}

// InTimeRange reports whether t falls within the inclusive [from, to] bounds
// expressed in Unix milliseconds. When a range is active, zero times never match.
// When both bounds are nil, every time matches (including zero).
func InTimeRange(t time.Time, from, to *int64) bool {
	if from == nil && to == nil {
		return true
	}
	if t.IsZero() {
		return false
	}
	ms := t.UnixNano() / int64(time.Millisecond)
	if from != nil && ms < *from {
		return false
	}
	if to != nil && ms > *to {
		return false
	}
	return true
}

// FileTimeMs returns the file's modification time as Unix milliseconds, or 0 if unknown.
func FileTimeMs(f IFile) int64 {
	if file, ok := f.(File); ok && file.FTime != 0 {
		return file.FTime
	}
	t := f.ModTime()
	if t.IsZero() {
		return 0
	}
	return t.UnixNano() / int64(time.Millisecond)
}
