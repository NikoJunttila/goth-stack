package reservation

import (
	"fmt"
	"strconv"
	"time"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"
)

// Form validation schemas
var timeSlotSchema = v.Schema{
	"title":     v.Rules(v.Min(1), v.Max(100)),
	"startTime": v.Rules(v.Min(1)), // These will be validated further in handler
	"endTime":   v.Rules(v.Min(1)),
	"capacity":  v.Rules(v.Min(1)),
}

var reservationSchema = v.Schema{
	"timeSlotID": v.Rules(v.Min(1)),
	"notes":      v.Rules(v.Max(500)),
}

// Form data structures
type TimeSlotFormValues struct {
	Title          string
	StartTime      string
	EndTime        string
	Capacity       string
	SuccessMessage string
}

type ReservationFormValues struct {
	TimeSlotID     string
	Notes          string
	SuccessMessage string
}

// Page data structures
type TimeSlotPageData struct {
	TimeSlots []TimeSlot
}

type ReservationPageData struct {
	Reservations []Reservation
	TimeSlots    []TimeSlot
}

// Handler functions
func HandleListTimeSlots(kit *kit.Kit) error {
	slots, err := GetAvailableTimeSlots()
	if err != nil {
		return err
	}
	return kit.Render(TimeSlotsList(TimeSlotPageData{TimeSlots: slots}))
}

func HandleCreateTimeSlotForm(kit *kit.Kit) error {
	return kit.Render(CreateTimeSlotForm(TimeSlotFormValues{}, v.Errors{}))
}

func HandleCreateTimeSlot(kit *kit.Kit) error {
	var values TimeSlotFormValues
	errors, ok := v.Request(kit.Request, &values, timeSlotSchema)
	if !ok {
		return kit.Render(CreateTimeSlotForm(values, errors))
	}

	// Parse time values
	startTime, err := time.Parse("2006-01-02T15:04", values.StartTime)
	if err != nil {
		errors["startTime"] = []string{"Invalid start time format"}
		return kit.Render(CreateTimeSlotForm(values, errors))
	}

	endTime, err := time.Parse("2006-01-02T15:04", values.EndTime)
	if err != nil {
		errors["endTime"] = []string{"Invalid end time format"}
		return kit.Render(CreateTimeSlotForm(values, errors))
	}

	if startTime.After(endTime) {
		errors["startTime"] = []string{"Start time must be before end time"}
		return kit.Render(CreateTimeSlotForm(values, errors))
	}

	if startTime.Before(time.Now()) {
		errors["startTime"] = []string{"Start time cannot be in the past"}
		return kit.Render(CreateTimeSlotForm(values, errors))
	}

	capacity, err := strconv.Atoi(values.Capacity)
	if err != nil || capacity < 1 {
		errors["capacity"] = []string{"Capacity must be a positive number"}
		return kit.Render(CreateTimeSlotForm(values, errors))
	}

	slot, err := CreateTimeSlot(startTime, endTime, values.Title, capacity)
	if err != nil {
		return kit.Render(CreateTimeSlotForm(values, errors))
	}
	fmt.Println(slot)

	values.SuccessMessage = fmt.Sprintf("New time slot created: %s", values.Title)
	return kit.Render(CreateTimeSlotForm(values, errors))
}

func HandleReservationForm(kit *kit.Kit) error {
	slots, err := GetAvailableTimeSlots()
	if err != nil {
		return err
	}
	return kit.Render(CreateReservationForm(ReservationFormValues{}, slots))
}

func HandleCreateReservation(kit *kit.Kit) error {
	// Ensure user is authenticated
	userID := GetUserID(kit)
	if userID == 0 {
		return kit.Redirect(303, "/login")
	}

	var values ReservationFormValues
	errors, ok := v.Request(kit.Request, &values, reservationSchema)

	slots, _ := GetAvailableTimeSlots() // Get slots for re-rendering the form if needed

	if !ok {
		return kit.Render(CreateReservationForm(values, slots))
	}

	timeSlotID, err := strconv.ParseUint(values.TimeSlotID, 10, 32)
	if err != nil {
		errors["timeSlotID"] = []string{"Invalid time slot ID"}
		return kit.Render(CreateReservationForm(values, slots))
	}

	reservation, err := ReserveTimeSlot(uint(timeSlotID), userID, values.Notes)
	if err != nil {
		errors["timeSlotID"] = []string{err.Error()}
		return kit.Render(CreateReservationForm(values, slots))
	}
	fmt.Println(reservation)

	// Get updated slots list after reservation
	updatedSlots, _ := GetAvailableTimeSlots()

	values.SuccessMessage = "Reservation confirmed!"
	return kit.Render(CreateReservationForm(values, updatedSlots))
}

func HandleUserReservations(kit *kit.Kit) error {
	userID := GetUserID(kit)
	if userID == 0 {
		return kit.Redirect(303, "/login")
	}

	reservations, err := GetUserReservations(userID)
	if err != nil {
		return err
	}

	timeSlots, _ := GetAvailableTimeSlots()

	return kit.Render(UserReservations(ReservationPageData{
		Reservations: reservations,
		TimeSlots:    timeSlots,
	}))
}

func HandleCancelReservation(kit *kit.Kit) error {
	userID := GetUserID(kit)
	if userID == 0 {
		return kit.Redirect(303, "/login")
	}
	reservationIDStr := chi.URLParam(kit.Request, "id")
	// reservationIDStr := kit.Request.URL.Query().Get("id")
	reservationID, err := strconv.ParseUint(reservationIDStr, 10, 32)
	if err != nil {
		return kit.Text(400, "Invalid reservation ID")
	}

	// Verify that this reservation belongs to the user
	reservations, _ := GetUserReservations(userID)

	var userOwnsReservation bool
	for _, r := range reservations {
		if r.ID == uint(reservationID) {
			userOwnsReservation = true
			break
		}
	}

	if !userOwnsReservation {
		return kit.Text(403, "You don't have permission to cancel this reservation")
	}

	err = CancelReservation(uint(reservationID))
	if err != nil {
		return kit.Text(500, "Failed to cancel reservation")
	}

	// Get updated reservations
	updatedReservations, _ := GetUserReservations(userID)
	timeSlots, _ := GetAvailableTimeSlots()

	return kit.Render(UserReservations(ReservationPageData{
		Reservations: updatedReservations,
		TimeSlots:    timeSlots,
	}))
}

// Helper function to get user ID from kit context
func GetUserID(kit *kit.Kit) uint {
	session := kit.GetSession("user")
	userIDInterface := session.Values["userID"]
	if userIDInterface == nil {
		return 0
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		return 0
	}

	return userID
}
