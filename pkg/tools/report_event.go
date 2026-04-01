package tools

import (
	"context"
	"time"
)

// ReportSavedEvent is emitted when a save_report tool writes a report to disk.
type ReportSavedEvent struct {
	Title     string    `json:"title"`
	FilePath  string    `json:"file_path"`
	Summary   string    `json:"summary"`
	Timestamp time.Time `json:"timestamp"`
}

// ReportSavedCallback is called when a report artifact is saved.
type ReportSavedCallback func(event ReportSavedEvent) error

type reportSavedCallbackKey struct{}

// ContextWithReportSavedCallback embeds a ReportSavedCallback into the context.
func ContextWithReportSavedCallback(ctx context.Context, callback ReportSavedCallback) context.Context {
	return context.WithValue(ctx, reportSavedCallbackKey{}, callback)
}

// ReportSavedCallbackFromCtx returns the ReportSavedCallback embedded in the context, or nil.
func ReportSavedCallbackFromCtx(ctx context.Context) ReportSavedCallback {
	if v, ok := ctx.Value(reportSavedCallbackKey{}).(ReportSavedCallback); ok {
		return v
	}
	return nil
}
