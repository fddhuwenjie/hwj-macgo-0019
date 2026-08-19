package repository

import (
	"context"

	"retryengine/domain"
)

type localRepoBase struct {
	store  *LocalStore
	prefix string
}

func (b *localRepoBase) save(ctx context.Context, aggregate interface{}) error {
	return b.store.Save(ctx, b.prefix, aggregate)
}

func (b *localRepoBase) findByID(ctx context.Context, id string, dest interface{}) error {
	return b.store.FindByID(ctx, b.prefix, id, dest)
}

func (b *localRepoBase) findAll(ctx context.Context, dest interface{}) error {
	return b.store.FindAll(ctx, b.prefix, dest)
}

type LocalCategoryRepository struct {
	localRepoBase
}

func NewLocalCategoryRepository(store *LocalStore) *LocalCategoryRepository {
	return &LocalCategoryRepository{localRepoBase{store: store, prefix: PrefixCategory}}
}

func (r *LocalCategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	return r.save(ctx, category)
}

func (r *LocalCategoryRepository) FindByID(ctx context.Context, id string) (*domain.Category, error) {
	var v domain.Category
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalCategoryRepository) FindAll(ctx context.Context) ([]*domain.Category, error) {
	var out []*domain.Category
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalCategoryRepository) Create(ctx context.Context, v *domain.Category) error {
	return r.Save(ctx, v)
}
func (r *LocalCategoryRepository) Update(ctx context.Context, v *domain.Category) error {
	return r.Save(ctx, v)
}
func (r *LocalCategoryRepository) Get(ctx context.Context, id string) (*domain.Category, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalCategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	return r.FindAll(ctx)
}
func (r *LocalCategoryRepository) Upsert(ctx context.Context, v *domain.Category) error {
	return r.Save(ctx, v)
}
func (r *LocalCategoryRepository) Find(ctx context.Context, id string) (*domain.Category, error) {
	return r.FindByID(ctx, id)
}

type LocalBudgetPolicyRepository struct {
	localRepoBase
}

func NewLocalBudgetPolicyRepository(store *LocalStore) *LocalBudgetPolicyRepository {
	return &LocalBudgetPolicyRepository{localRepoBase{store: store, prefix: PrefixBudgetPolicy}}
}

func (r *LocalBudgetPolicyRepository) Save(ctx context.Context, policy *domain.BudgetPolicy) error {
	return r.save(ctx, policy)
}

func (r *LocalBudgetPolicyRepository) FindByID(ctx context.Context, id string) (*domain.BudgetPolicy, error) {
	var v domain.BudgetPolicy
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalBudgetPolicyRepository) FindAll(ctx context.Context) ([]*domain.BudgetPolicy, error) {
	var out []*domain.BudgetPolicy
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalBudgetPolicyRepository) Create(ctx context.Context, v *domain.BudgetPolicy) error {
	return r.Save(ctx, v)
}
func (r *LocalBudgetPolicyRepository) Update(ctx context.Context, v *domain.BudgetPolicy) error {
	return r.Save(ctx, v)
}
func (r *LocalBudgetPolicyRepository) Get(ctx context.Context, id string) (*domain.BudgetPolicy, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalBudgetPolicyRepository) List(ctx context.Context) ([]*domain.BudgetPolicy, error) {
	return r.FindAll(ctx)
}
func (r *LocalBudgetPolicyRepository) Upsert(ctx context.Context, v *domain.BudgetPolicy) error {
	return r.Save(ctx, v)
}
func (r *LocalBudgetPolicyRepository) Find(ctx context.Context, id string) (*domain.BudgetPolicy, error) {
	return r.FindByID(ctx, id)
}
