package feed

import (
	"context"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

// NotifyingStore wraps a store.URLStore and calls notify after any
// mutating operation that succeeds. notify is called in a goroutine
// so it does not block the caller.
type NotifyingStore struct {
	store.URLStore
	notify func()
}

// NewNotifyingStore wraps s and calls notify (in a new goroutine) after
// a successful UpdateStatus, Update, or Delete.
func NewNotifyingStore(s store.URLStore, notify func()) *NotifyingStore {
	return &NotifyingStore{URLStore: s, notify: notify}
}

func (n *NotifyingStore) UpdateStatus(ctx context.Context, id int64, status string, audioPath *string) (model.URL, error) {
	u, err := n.URLStore.UpdateStatus(ctx, id, status, audioPath)
	if err == nil {
		go n.notify()
	}
	return u, err
}

func (n *NotifyingStore) Update(ctx context.Context, id int64, title, description *string) (model.URL, error) {
	u, err := n.URLStore.Update(ctx, id, title, description)
	if err == nil {
		go n.notify()
	}
	return u, err
}

func (n *NotifyingStore) Delete(ctx context.Context, id int64) error {
	err := n.URLStore.Delete(ctx, id)
	if err == nil {
		go n.notify()
	}
	return err
}
