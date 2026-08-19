package repository

import (
	"context"

	"retryengine/domain"
)

type LocalWindowRepository struct {
	localRepoBase
}

func NewLocalWindowRepository(store *LocalStore) *LocalWindowRepository {
	return &LocalWindowRepository{localRepoBase{store: store, prefix: PrefixWindow}}
}

func (r *LocalWindowRepository) Save(ctx context.Context, window *domain.Window) error {
	return r.save(ctx, window)
}

func (r *LocalWindowRepository) FindByID(ctx context.Context, id string) (*domain.Window, error) {
	var v domain.Window
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalWindowRepository) FindAll(ctx context.Context) ([]*domain.Window, error) {
	var out []*domain.Window
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalWindowRepository) Create(ctx context.Context, v *domain.Window) error {
	return r.Save(ctx, v)
}
func (r *LocalWindowRepository) Update(ctx context.Context, v *domain.Window) error {
	return r.Save(ctx, v)
}
func (r *LocalWindowRepository) Get(ctx context.Context, id string) (*domain.Window, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalWindowRepository) List(ctx context.Context) ([]*domain.Window, error) {
	return r.FindAll(ctx)
}
func (r *LocalWindowRepository) Upsert(ctx context.Context, v *domain.Window) error {
	return r.Save(ctx, v)
}
func (r *LocalWindowRepository) Find(ctx context.Context, id string) (*domain.Window, error) {
	return r.FindByID(ctx, id)
}

type LocalReservationRepository struct {
	localRepoBase
}

func NewLocalReservationRepository(store *LocalStore) *LocalReservationRepository {
	return &LocalReservationRepository{localRepoBase{store: store, prefix: PrefixReservation}}
}

func (r *LocalReservationRepository) Save(ctx context.Context, reservation *domain.Reservation) error {
	return r.save(ctx, reservation)
}

func (r *LocalReservationRepository) FindByID(ctx context.Context, id string) (*domain.Reservation, error) {
	var v domain.Reservation
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalReservationRepository) FindAll(ctx context.Context) ([]*domain.Reservation, error) {
	var out []*domain.Reservation
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalReservationRepository) Create(ctx context.Context, v *domain.Reservation) error {
	return r.Save(ctx, v)
}
func (r *LocalReservationRepository) Update(ctx context.Context, v *domain.Reservation) error {
	return r.Save(ctx, v)
}
func (r *LocalReservationRepository) Get(ctx context.Context, id string) (*domain.Reservation, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalReservationRepository) List(ctx context.Context) ([]*domain.Reservation, error) {
	return r.FindAll(ctx)
}
func (r *LocalReservationRepository) Upsert(ctx context.Context, v *domain.Reservation) error {
	return r.Save(ctx, v)
}
func (r *LocalReservationRepository) Find(ctx context.Context, id string) (*domain.Reservation, error) {
	return r.FindByID(ctx, id)
}

type LocalFailureStreakRepository struct {
	localRepoBase
}

func NewLocalFailureStreakRepository(store *LocalStore) *LocalFailureStreakRepository {
	return &LocalFailureStreakRepository{localRepoBase{store: store, prefix: PrefixFailureStreak}}
}

func (r *LocalFailureStreakRepository) Save(ctx context.Context, streak *domain.FailureStreak) error {
	return r.save(ctx, streak)
}

func (r *LocalFailureStreakRepository) FindByID(ctx context.Context, id string) (*domain.FailureStreak, error) {
	var v domain.FailureStreak
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalFailureStreakRepository) FindAll(ctx context.Context) ([]*domain.FailureStreak, error) {
	var out []*domain.FailureStreak
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalFailureStreakRepository) Create(ctx context.Context, v *domain.FailureStreak) error {
	return r.Save(ctx, v)
}
func (r *LocalFailureStreakRepository) Update(ctx context.Context, v *domain.FailureStreak) error {
	return r.Save(ctx, v)
}
func (r *LocalFailureStreakRepository) Get(ctx context.Context, id string) (*domain.FailureStreak, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalFailureStreakRepository) List(ctx context.Context) ([]*domain.FailureStreak, error) {
	return r.FindAll(ctx)
}
func (r *LocalFailureStreakRepository) Upsert(ctx context.Context, v *domain.FailureStreak) error {
	return r.Save(ctx, v)
}
func (r *LocalFailureStreakRepository) Find(ctx context.Context, id string) (*domain.FailureStreak, error) {
	return r.FindByID(ctx, id)
}

type LocalSuspensionRepository struct {
	localRepoBase
}

func NewLocalSuspensionRepository(store *LocalStore) *LocalSuspensionRepository {
	return &LocalSuspensionRepository{localRepoBase{store: store, prefix: PrefixSuspension}}
}

func (r *LocalSuspensionRepository) Save(ctx context.Context, suspension *domain.Suspension) error {
	return r.save(ctx, suspension)
}

func (r *LocalSuspensionRepository) FindByID(ctx context.Context, id string) (*domain.Suspension, error) {
	var v domain.Suspension
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalSuspensionRepository) FindAll(ctx context.Context) ([]*domain.Suspension, error) {
	var out []*domain.Suspension
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalSuspensionRepository) Create(ctx context.Context, v *domain.Suspension) error {
	return r.Save(ctx, v)
}
func (r *LocalSuspensionRepository) Update(ctx context.Context, v *domain.Suspension) error {
	return r.Save(ctx, v)
}
func (r *LocalSuspensionRepository) Get(ctx context.Context, id string) (*domain.Suspension, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalSuspensionRepository) List(ctx context.Context) ([]*domain.Suspension, error) {
	return r.FindAll(ctx)
}
func (r *LocalSuspensionRepository) Upsert(ctx context.Context, v *domain.Suspension) error {
	return r.Save(ctx, v)
}
func (r *LocalSuspensionRepository) Find(ctx context.Context, id string) (*domain.Suspension, error) {
	return r.FindByID(ctx, id)
}

type LocalRecoveryCredentialRepository struct {
	localRepoBase
}

func NewLocalRecoveryCredentialRepository(store *LocalStore) *LocalRecoveryCredentialRepository {
	return &LocalRecoveryCredentialRepository{localRepoBase{store: store, prefix: PrefixRecoveryCredential}}
}

func (r *LocalRecoveryCredentialRepository) Save(ctx context.Context, credential *domain.RecoveryCredential) error {
	return r.save(ctx, credential)
}

func (r *LocalRecoveryCredentialRepository) FindByID(ctx context.Context, id string) (*domain.RecoveryCredential, error) {
	var v domain.RecoveryCredential
	if err := r.findByID(ctx, id, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LocalRecoveryCredentialRepository) FindAll(ctx context.Context) ([]*domain.RecoveryCredential, error) {
	var out []*domain.RecoveryCredential
	if err := r.findAll(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocalRecoveryCredentialRepository) Create(ctx context.Context, v *domain.RecoveryCredential) error {
	return r.Save(ctx, v)
}
func (r *LocalRecoveryCredentialRepository) Update(ctx context.Context, v *domain.RecoveryCredential) error {
	return r.Save(ctx, v)
}
func (r *LocalRecoveryCredentialRepository) Get(ctx context.Context, id string) (*domain.RecoveryCredential, error) {
	return r.FindByID(ctx, id)
}
func (r *LocalRecoveryCredentialRepository) List(ctx context.Context) ([]*domain.RecoveryCredential, error) {
	return r.FindAll(ctx)
}
func (r *LocalRecoveryCredentialRepository) Upsert(ctx context.Context, v *domain.RecoveryCredential) error {
	return r.Save(ctx, v)
}
func (r *LocalRecoveryCredentialRepository) Find(ctx context.Context, id string) (*domain.RecoveryCredential, error) {
	return r.FindByID(ctx, id)
}
