package application

import (
	"context"
	"fmt"
)

type BatchUseCase struct {
	attempts *AttemptUseCase
}

func NewBatchUseCase(attempts *AttemptUseCase) *BatchUseCase {
	return &BatchUseCase{attempts: attempts}
}

func (uc *BatchUseCase) ApplyBatch(ctx context.Context, commands []AttemptCommand) BatchResult {
	result := BatchResult{
		Items: make([]ItemResult, 0, len(commands)),
	}
	for i, cmd := range commands {
		_, err := uc.attempts.Apply(ctx, cmd)
		item := ItemResult{Index: i, Key: cmd.IdempotencyKey, Success: err == nil}
		if err != nil {
			item.Error = err.Error()
			result.Failed++
		} else {
			result.Succeeded++
		}
		result.Items = append(result.Items, item)
	}
	return result
}

func (uc *BatchUseCase) RecordBatch(ctx context.Context, outcomes []AttemptOutcome) BatchResult {
	result := BatchResult{Items: make([]ItemResult, 0, len(outcomes))}
	for i, out := range outcomes {
		var err error
		if out.Terminated || out.Success {
			err = uc.attempts.RecordSuccess(ctx, out.AttemptID)
		} else {
			err = uc.attempts.RecordFailure(ctx, out.AttemptID, out.Reason)
		}
		item := ItemResult{Index: i, Key: out.AttemptID, Success: err == nil}
		if err != nil {
			item.Error = fmt.Sprintf("outcome %d: %v", i, err)
			result.Failed++
		} else {
			result.Succeeded++
		}
		result.Items = append(result.Items, item)
	}
	return result
}
