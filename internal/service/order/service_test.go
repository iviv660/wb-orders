package order

import (
	"context"
	"errors"
	"testing"

	"app/internal/mocks"
	"app/internal/model"

	"github.com/stretchr/testify/require"
)

func newTestService() (
	ctx context.Context,
	svc *Service,
	repo *mocks.MockRepository,
	cache *mocks.MockCache,
) {
	ctx = context.Background()
	repo = new(mocks.MockRepository)
	cache = new(mocks.MockCache)
	svc = New(repo, cache)
	return
}

func Test_ProcessOrder_OK(t *testing.T) {
	ctx, svc, repo, cache := newTestService()

	order := model.Order{OrderUUID: "uid-1"}

	repo.On("SetOrder", ctx, order).Return(nil).Once()
	cache.On("Set", "order:"+order.OrderUUID, order).Return(nil).Once()

	err := svc.ProcessOrder(ctx, order)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_ProcessOrder_RepoError(t *testing.T) {
	ctx, svc, repo, cache := newTestService()

	order := model.Order{OrderUUID: "uid-1"}
	errRepo := errors.New("repo error")

	repo.On("SetOrder", ctx, order).Return(errRepo).Once()

	err := svc.ProcessOrder(ctx, order)
	require.ErrorIs(t, err, errRepo)

	cache.AssertNotCalled(t, "Set")
	repo.AssertExpectations(t)
}

func Test_Get_CacheHit(t *testing.T) {
	ctx, svc, repo, cache := newTestService()

	key := "order:uid-1"
	want := model.Order{OrderUUID: "uid-1"}

	cache.On("Get", key).Return(want, nil).Once()

	got, err := svc.Get(ctx, "uid-1")
	require.NoError(t, err)
	require.Equal(t, want, got)

	repo.AssertNotCalled(t, "GetOrder")
	cache.AssertExpectations(t)
}

func Test_Get_CacheMiss_RepoOK_CacheSetOK(t *testing.T) {
	ctx, svc, repo, cache := newTestService()

	key := "order:uid-1"
	want := model.Order{OrderUUID: "uid-1"}

	cache.On("Get", key).Return(model.Order{}, errors.New("cache miss")).Once()
	repo.On("GetOrder", ctx, "uid-1").Return(want, nil).Once()
	cache.On("Set", key, want).Return(nil).Once()

	got, err := svc.Get(ctx, "uid-1")
	require.NoError(t, err)
	require.Equal(t, want, got)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_Get_CacheMiss_RepoError(t *testing.T) {
	ctx, svc, repo, cache := newTestService()

	key := "order:uid-1"
	errRepo := errors.New("repo error")

	cache.On("Get", key).Return(model.Order{}, errors.New("cache miss")).Once()
	repo.On("GetOrder", ctx, "uid-1").Return(model.Order{}, errRepo).Once()

	_, err := svc.Get(ctx, "uid-1")
	require.ErrorIs(t, err, errRepo)

	cache.AssertNotCalled(t, "Set")
	repo.AssertExpectations(t)
}
