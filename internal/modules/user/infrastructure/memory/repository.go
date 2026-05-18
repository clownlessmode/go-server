package memory

import (
	"context"
	"sync"
	"time"

	"project/internal/modules/user/domain"
)

type Repository struct {
	mu     sync.RWMutex
	nextID int64
	data   map[int64]*domain.User
}

func NewRepository() *Repository {
	return &Repository{
		nextID: 1,
		data:   make(map[int64]*domain.User),
	}
}

func (r *Repository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	created := cloneUser(user)
	created.ID = r.nextID
	created.CreatedAt = now
	created.UpdatedAt = now

	r.data[created.ID] = created
	r.nextID++

	return cloneUser(created), nil
}

func (r *Repository) List(ctx context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*domain.User, 0, len(r.data))
	for _, user := range r.data {
		users = append(users, cloneUser(user))
	}

	return users, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.data[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	return cloneUser(user), nil
}

func (r *Repository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.data {
		if user.Login == login {
			return cloneUser(user), nil
		}
	}

	return nil, domain.ErrUserNotFound
}

func (r *Repository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.data[user.ID]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	updated := cloneUser(user)
	updated.CreatedAt = current.CreatedAt
	updated.UpdatedAt = time.Now().UTC()

	r.data[updated.ID] = updated

	return cloneUser(updated), nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[id]; !ok {
		return domain.ErrUserNotFound
	}

	delete(r.data, id)
	return nil
}

func cloneUser(user *domain.User) *domain.User {
	if user == nil {
		return nil
	}

	cloned := *user
	return &cloned
}
