package repository

import (
	"context"

	"retryengine/domain"
)

type LocalAttemptRepository struct {
	localRepoBase
}

func NewLocalAttemptRepository(store *LocalStore) *LocalAttemptRepository {
	return &LocalAttemptRepository{localRepoBase{store: store, prefix: PrefixAttempt}}
}

func (r *LocalAttemptRepository) Save(ctx context.Context, attempt *domain.Attempt) error {
	return r.save(ctx, attempt)
}

func (r *LocalAttemptRepository) FindByID(ctx context.Context, id string) (*domain.Attempt, error) {
	var v domain.Attempt
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalAttemptRepository) FindAll(ctx context.Context) ([]*domain.Attempt, error) {
	var out []*domain.Attempt
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalAttemptRepository) Create(ctx context.Context, v *domain.Attempt) error {
	return r.Save(ctx, v)
}
func (r *LocalAttemptRepository) Update(ctx context.Context, v *domain.Attempt) error {
	return r.Save(ctx, v)
}
func (r *LocalAttemptRepository) Get(ctx context.Context, id string) (*domain.Attempt, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalAttemptRepository) List(ctx context.Context) ([]*domain.Attempt, error) {
	return r.FindAll(ctx)
}
func (r *LocalAttemptRepository) Upsert(ctx context.Context, v *domain.Attempt) error {
	return r.Save(ctx, v)
}
func (r *LocalAttemptRepository) Find(ctx context.Context, id string) (*domain.Attempt, error) {
	return r.FindByID(ctx, id)
}

type LocalResultRepository struct {
	localRepoBase
}

func NewLocalResultRepository(store *LocalStore) *LocalResultRepository {
	return &LocalResultRepository{localRepoBase{store: store, prefix: PrefixResult}}
}

func (r *LocalResultRepository) Save(ctx context.Context, result *domain.Result) error {
	return r.save(ctx, result)
}

func (r *LocalResultRepository) FindByID(ctx context.Context, id string) (*domain.Result, error) {
	var v domain.Result
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalResultRepository) FindAll(ctx context.Context) ([]*domain.Result, error) {
	var out []*domain.Result
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalResultRepository) Create(ctx context.Context, v *domain.Result) error {
	return r.Save(ctx, v)
}
func (r *LocalResultRepository) Update(ctx context.Context, v *domain.Result) error {
	return r.Save(ctx, v)
}
func (r *LocalResultRepository) Get(ctx context.Context, id string) (*domain.Result, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalResultRepository) List(ctx context.Context) ([]*domain.Result, error) {
	return r.FindAll(ctx)
}
func (r *LocalResultRepository) Upsert(ctx context.Context, v *domain.Result) error {
	return r.Save(ctx, v)
}
func (r *LocalResultRepository) Find(ctx context.Context, id string) (*domain.Result, error) {
	return r.FindByID(ctx, id)
}

type LocalBackoffPlanRepository struct {
	localRepoBase
}

func NewLocalBackoffPlanRepository(store *LocalStore) *LocalBackoffPlanRepository {
	return &LocalBackoffPlanRepository{localRepoBase{store: store, prefix: PrefixBackoffPlan}}
}

func (r *LocalBackoffPlanRepository) Save(ctx context.Context, plan *domain.BackoffPlan) error {
	return r.save(ctx, plan)
}

func (r *LocalBackoffPlanRepository) FindByID(ctx context.Context, id string) (*domain.BackoffPlan, error) {
	var v domain.BackoffPlan
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalBackoffPlanRepository) FindAll(ctx context.Context) ([]*domain.BackoffPlan, error) {
	var out []*domain.BackoffPlan
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalBackoffPlanRepository) Create(ctx context.Context, v *domain.BackoffPlan) error {
	return r.Save(ctx, v)
}
func (r *LocalBackoffPlanRepository) Update(ctx context.Context, v *domain.BackoffPlan) error {
	return r.Save(ctx, v)
}
func (r *LocalBackoffPlanRepository) Get(ctx context.Context, id string) (*domain.BackoffPlan, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalBackoffPlanRepository) List(ctx context.Context) ([]*domain.BackoffPlan, error) {
	return r.FindAll(ctx)
}
func (r *LocalBackoffPlanRepository) Upsert(ctx context.Context, v *domain.BackoffPlan) error {
	return r.Save(ctx, v)
}
func (r *LocalBackoffPlanRepository) Find(ctx context.Context, id string) (*domain.BackoffPlan, error) {
	return r.FindByID(ctx, id)
}
