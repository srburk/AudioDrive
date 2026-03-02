package feed

import (
	"context"
	"log"

	"audiodrive/internal/model"
)

// Item is the enriched unit flowing through the pipeline.
type Item struct {
	URL            model.URL
	AudioSizeBytes int64 // populated by SizeProcessor; 0 if unknown
}

// Processor is a step in the feed build pipeline.
type Processor func(ctx context.Context, item *Item) error

// Pipeline runs a sequence of processors against a filtered list of URLs.
type Pipeline struct {
	steps []Processor
}

// New returns an empty Pipeline.
func New() *Pipeline { return &Pipeline{} }

// Add appends a processor step and returns the pipeline for chaining.
func (p *Pipeline) Add(proc Processor) *Pipeline {
	p.steps = append(p.steps, proc)
	return p
}

// Run filters urls to those with status==done and a non-nil AudioPath,
// constructs Items, then runs each processor step in order.
// A step error skips that item (logged); other items are still processed.
// The returned error is always nil; item-level errors are handled internally.
func (p *Pipeline) Run(ctx context.Context, urls []model.URL) ([]Item, error) {
	var result []Item
	for _, u := range urls {
		if u.Status != model.StatusDone || u.AudioPath == nil {
			continue
		}
		item := Item{URL: u}
		skip := false
		for _, step := range p.steps {
			if err := step(ctx, &item); err != nil {
				log.Printf("feed: pipeline step error for item %d: %v", u.ID, err)
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, item)
		}
	}
	return result, nil
}
