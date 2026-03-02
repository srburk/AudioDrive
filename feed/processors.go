package feed

import (
	"context"
	"os"
)

// SizeProcessor stats the audio file and populates AudioSizeBytes.
// It is best-effort: a nil AudioPath or missing/inaccessible file
// leaves AudioSizeBytes at 0 and returns no error.
func SizeProcessor(_ context.Context, item *Item) error {
	if item.URL.AudioPath == nil {
		return nil
	}
	info, err := os.Stat(*item.URL.AudioPath)
	if err != nil {
		return nil
	}
	item.AudioSizeBytes = info.Size()
	return nil
}
