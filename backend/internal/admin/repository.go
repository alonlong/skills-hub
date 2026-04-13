package admin

import "context"

type Repository interface {
	GetTask(ctx context.Context, taskID string) (ReviewTask, error)
	Decide(ctx context.Context, taskID string, reviewerID string, comment string, decision string, nextVersionStatus string) error
}
