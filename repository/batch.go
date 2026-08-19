package repository

import "context"

type BatchItem struct {
	Key   string
	Apply func(context.Context) error
}

type BatchResult struct {
	Key   string
	Error error
	OK    bool
}

func RunBatch(ctx context.Context, items []BatchItem) []BatchResult {
	results := make([]BatchResult, 0, len(items))
	for _, item := range items {
		select {
		case <-ctx.Done():
			results = append(results, BatchResult{Key: item.Key, Error: ctx.Err(), OK: false})
			return results
		default:
		}
		err := item.Apply(ctx)
		results = append(results, BatchResult{
			Key:   item.Key,
			Error: err,
			OK:    err == nil,
		})
	}
	return results
}
