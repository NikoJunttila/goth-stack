package reservation

import (
	"errors"
	"gothstack/app/db"
	"time"

	"gorm.io/gorm"
)

// Event name constants
const (
	TimeSlotCreatedEvent     = "reservation.timeSlotCreated"
	TimeSlotReservedEvent    = "reservation.timeSlotReserved"
	ReservationCanceledEvent = "reservation.reservationCanceled"
)

// TimeSlot represents a time period that can be reserved
type TimeSlot struct {
	gorm.Model
	StartTime    time.Time
	EndTime      time.Time
	Available    bool
	Title        string
	Capacity     int           // For multiple users per slot, if needed
	Reservations []Reservation `gorm:"foreignKey:TimeSlotID"`
}

// Reservation represents a user's booking of a time slot
type Reservation struct {
	gorm.Model
	TimeSlotID uint
	TimeSlot   TimeSlot `gorm:"foreignKey:TimeSlotID"`
	UserID     uint
	Notes      string
	Status     string // "confirmed", "canceled", etc.
}

// CreateTimeSlot creates a new time slot
func CreateTimeSlot(startTime, endTime time.Time, title string, capacity int) (TimeSlot, error) {
	if startTime.After(endTime) {
		return TimeSlot{}, errors.New("start time must be before end time")
	}

	slot := TimeSlot{
		StartTime: startTime,
		EndTime:   endTime,
		Available: true,
		Title:     title,
		Capacity:  capacity,
	}

	result := db.Get().Create(&slot)
	return slot, result.Error
}

// GetTimeSlot retrieves a time slot by ID
func GetTimeSlot(id uint) (TimeSlot, error) {
	var slot TimeSlot
	result := db.Get().Preload("Reservations").First(&slot, id)
	return slot, result.Error
}

// GetAvailableTimeSlots retrieves all available time slots
func GetAvailableTimeSlots() ([]TimeSlot, error) {
	var slots []TimeSlot
	result := db.Get().Where("available = ?", true).
		Where("end_time > ?", time.Now()).
		Order("start_time asc").
		Find(&slots)
	return slots, result.Error
}

// GetTimeSlotsByDateRange retrieves time slots within a date range
func GetTimeSlotsByDateRange(start, end time.Time) ([]TimeSlot, error) {
	var slots []TimeSlot
	result := db.Get().
		Where("start_time >= ? AND start_time <= ?", start, end).
		Order("start_time asc").
		Find(&slots)
	return slots, result.Error
}

// ReserveTimeSlot creates a reservation for a time slot
func ReserveTimeSlot(timeSlotID, userID uint, notes string) (Reservation, error) {
	// Use a transaction to ensure atomicity
	var reservation Reservation
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		// Get the time slot with locking
		var slot TimeSlot
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&slot, timeSlotID).Error; err != nil {
			return err
		}

		// Check if the slot is available
		if !slot.Available {
			return errors.New("time slot is not available")
		}

		// Check if it's in the past
		if slot.EndTime.Before(time.Now()) {
			return errors.New("cannot reserve a time slot in the past")
		}

		// Count existing reservations
		var count int64
		tx.Model(&Reservation{}).Where("time_slot_id = ? AND status != ?", timeSlotID, "canceled").Count(&count)

		// Check capacity
		if int(count) >= slot.Capacity {
			return errors.New("time slot is at full capacity")
		}

		// Create the reservation
		reservation = Reservation{
			TimeSlotID: timeSlotID,
			UserID:     userID,
			Notes:      notes,
			Status:     "confirmed",
		}

		if err := tx.Create(&reservation).Error; err != nil {
			return err
		}

		// If this is a single-use time slot or it's now at capacity, mark as unavailable
		if slot.Capacity == 1 || int(count+1) >= slot.Capacity {
			if err := tx.Model(&slot).Update("available", false).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return reservation, err
}

// CancelReservation cancels a reservation and updates the time slot availability
func CancelReservation(reservationID uint) error {
	return db.Get().Transaction(func(tx *gorm.DB) error {
		// Get the reservation
		var reservation Reservation
		if err := tx.First(&reservation, reservationID).Error; err != nil {
			return err
		}

		// Update the reservation status
		if err := tx.Model(&reservation).Update("status", "canceled").Error; err != nil {
			return err
		}

		// Get the time slot
		var slot TimeSlot
		if err := tx.First(&slot, reservation.TimeSlotID).Error; err != nil {
			return err
		}

		// Make the time slot available again if it's not in the past
		if slot.EndTime.After(time.Now()) {
			if err := tx.Model(&slot).Update("available", true).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetUserReservations retrieves all reservations for a specific user
func GetUserReservations(userID uint) ([]Reservation, error) {
	var reservations []Reservation
	result := db.Get().
		Preload("TimeSlot").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&reservations)
	return reservations, result.Error
}

// using goose
/* func initialize() {
	db.Get().AutoMigrate(&TimeSlot{}, &Reservation{})
} */
