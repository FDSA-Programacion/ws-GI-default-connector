package application

import (
	"fmt"
	"ws-int-httr/internal/domain"
)

type BookingService struct {
	bookingProvider domain.BookingProvider
}

func NewBookingService(bookingProvider domain.BookingProvider) *BookingService {
	return &BookingService{bookingProvider: bookingProvider}
}

// Implements: Availability
func (s *BookingService) Availability(req *domain.AvailRequest) (*domain.BaseJsonRS[*domain.AvailResponse], error) {
	response, err := s.bookingProvider.SendAvail(req)
	if err != nil {
		return nil, fmt.Errorf("error al obtener disponibilidad de OT: %w", err)
	}

	return response, nil
}

// Implements: PreBook
func (s *BookingService) PreBook(req *domain.PreBookRequest) (*domain.BaseJsonRS[*domain.PreBookResponse], error) {
	response, err := s.bookingProvider.SendPreBook(req)
	if err != nil {
		return nil, fmt.Errorf("error al obtener pre-reserva de OT: %w", err)
	}

	return response, nil
}

// Implements: Book
func (s *BookingService) Book(req *domain.BookRequest) (*domain.BaseJsonRS[*domain.BookResponse], error) {
	response, err := s.bookingProvider.SendBook(req)
	if err != nil {
		return nil, fmt.Errorf("error al confirmar la reserva de OT: %w", err)
	}

	return response, nil
}

// Implements: Cancel
func (s *BookingService) Cancel(req *domain.CancelRequest) (*domain.BaseJsonRS[*domain.CancelResponse], error) {
	response, err := s.bookingProvider.SendCancel(req)
	if err != nil {
		return nil, fmt.Errorf("error al confirmar la reserva de OT: %w", err)
	}

	return response, nil
}
