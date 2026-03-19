package application

import (
	"errors"
	"testing"

	"ws-int-httr/internal/domain"

	"github.com/stretchr/testify/require"
)

type fakeBookingProvider struct {
	availFn   func(*domain.AvailRequest) (*domain.BaseJsonRS[*domain.AvailResponse], error)
	preBookFn func(*domain.PreBookRequest) (*domain.BaseJsonRS[*domain.PreBookResponse], error)
	bookFn    func(*domain.BookRequest) (*domain.BaseJsonRS[*domain.BookResponse], error)
	cancelFn  func(*domain.CancelRequest) (*domain.BaseJsonRS[*domain.CancelResponse], error)
}

func (f *fakeBookingProvider) SendAvail(req *domain.AvailRequest) (*domain.BaseJsonRS[*domain.AvailResponse], error) {
	return f.availFn(req)
}

func (f *fakeBookingProvider) SendPreBook(req *domain.PreBookRequest) (*domain.BaseJsonRS[*domain.PreBookResponse], error) {
	return f.preBookFn(req)
}

func (f *fakeBookingProvider) SendBook(req *domain.BookRequest) (*domain.BaseJsonRS[*domain.BookResponse], error) {
	return f.bookFn(req)
}

func (f *fakeBookingProvider) SendCancel(req *domain.CancelRequest) (*domain.BaseJsonRS[*domain.CancelResponse], error) {
	return f.cancelFn(req)
}

func TestNewBookingService(t *testing.T) {
	t.Parallel()

	svc := NewBookingService(newProviderStub())
	require.NotNil(t, svc)
}

func TestBookingService_Availability(t *testing.T) {
	t.Parallel()

	req := &domain.AvailRequest{}
	expected := &domain.BaseJsonRS[*domain.AvailResponse]{Success: "OK"}

	t.Run("success", func(t *testing.T) {
		svc := NewBookingService(&fakeBookingProvider{
			availFn: func(got *domain.AvailRequest) (*domain.BaseJsonRS[*domain.AvailResponse], error) {
				require.Same(t, req, got)
				return expected, nil
			},
			preBookFn: newProviderStub().preBookFn,
			bookFn:    newProviderStub().bookFn,
			cancelFn:  newProviderStub().cancelFn,
		})

		got, err := svc.Availability(req)
		require.NoError(t, err)
		require.Same(t, expected, got)
	})

	t.Run("error wraps provider error", func(t *testing.T) {
		providerErr := errors.New("provider failed")
		p := newProviderStub()
		p.availFn = func(*domain.AvailRequest) (*domain.BaseJsonRS[*domain.AvailResponse], error) {
			return nil, providerErr
		}
		svc := NewBookingService(p)

		got, err := svc.Availability(req)
		require.Nil(t, got)
		require.Error(t, err)
		require.ErrorIs(t, err, providerErr)
		require.Contains(t, err.Error(), "error al obtener disponibilidad de OT")
	})
}

func TestBookingService_PreBook(t *testing.T) {
	t.Parallel()

	req := &domain.PreBookRequest{}
	expected := &domain.BaseJsonRS[*domain.PreBookResponse]{Success: "OK"}

	t.Run("success", func(t *testing.T) {
		p := newProviderStub()
		p.preBookFn = func(got *domain.PreBookRequest) (*domain.BaseJsonRS[*domain.PreBookResponse], error) {
			require.Same(t, req, got)
			return expected, nil
		}
		svc := NewBookingService(p)

		got, err := svc.PreBook(req)
		require.NoError(t, err)
		require.Same(t, expected, got)
	})

	t.Run("error wraps provider error", func(t *testing.T) {
		providerErr := errors.New("provider failed")
		p := newProviderStub()
		p.preBookFn = func(*domain.PreBookRequest) (*domain.BaseJsonRS[*domain.PreBookResponse], error) {
			return nil, providerErr
		}
		svc := NewBookingService(p)

		got, err := svc.PreBook(req)
		require.Nil(t, got)
		require.Error(t, err)
		require.ErrorIs(t, err, providerErr)
		require.Contains(t, err.Error(), "error al obtener pre-reserva de OT")
	})
}

func TestBookingService_Book(t *testing.T) {
	t.Parallel()

	req := &domain.BookRequest{}
	expected := &domain.BaseJsonRS[*domain.BookResponse]{Success: "OK"}

	t.Run("success", func(t *testing.T) {
		p := newProviderStub()
		p.bookFn = func(got *domain.BookRequest) (*domain.BaseJsonRS[*domain.BookResponse], error) {
			require.Same(t, req, got)
			return expected, nil
		}
		svc := NewBookingService(p)

		got, err := svc.Book(req)
		require.NoError(t, err)
		require.Same(t, expected, got)
	})

	t.Run("error wraps provider error", func(t *testing.T) {
		providerErr := errors.New("provider failed")
		p := newProviderStub()
		p.bookFn = func(*domain.BookRequest) (*domain.BaseJsonRS[*domain.BookResponse], error) {
			return nil, providerErr
		}
		svc := NewBookingService(p)

		got, err := svc.Book(req)
		require.Nil(t, got)
		require.Error(t, err)
		require.ErrorIs(t, err, providerErr)
		require.Contains(t, err.Error(), "error al confirmar la reserva de OT")
	})
}

func TestBookingService_Cancel(t *testing.T) {
	t.Parallel()

	req := &domain.CancelRequest{}
	expected := &domain.BaseJsonRS[*domain.CancelResponse]{Success: "OK"}

	t.Run("success", func(t *testing.T) {
		p := newProviderStub()
		p.cancelFn = func(got *domain.CancelRequest) (*domain.BaseJsonRS[*domain.CancelResponse], error) {
			require.Same(t, req, got)
			return expected, nil
		}
		svc := NewBookingService(p)

		got, err := svc.Cancel(req)
		require.NoError(t, err)
		require.Same(t, expected, got)
	})

	t.Run("error wraps provider error", func(t *testing.T) {
		providerErr := errors.New("provider failed")
		p := newProviderStub()
		p.cancelFn = func(*domain.CancelRequest) (*domain.BaseJsonRS[*domain.CancelResponse], error) {
			return nil, providerErr
		}
		svc := NewBookingService(p)

		got, err := svc.Cancel(req)
		require.Nil(t, got)
		require.Error(t, err)
		require.ErrorIs(t, err, providerErr)
		require.Contains(t, err.Error(), "error al confirmar la reserva de OT")
	})
}

func newProviderStub() *fakeBookingProvider {
	return &fakeBookingProvider{
		availFn: func(*domain.AvailRequest) (*domain.BaseJsonRS[*domain.AvailResponse], error) {
			return nil, nil
		},
		preBookFn: func(*domain.PreBookRequest) (*domain.BaseJsonRS[*domain.PreBookResponse], error) {
			return nil, nil
		},
		bookFn: func(*domain.BookRequest) (*domain.BaseJsonRS[*domain.BookResponse], error) {
			return nil, nil
		},
		cancelFn: func(*domain.CancelRequest) (*domain.BaseJsonRS[*domain.CancelResponse], error) {
			return nil, nil
		},
	}
}
