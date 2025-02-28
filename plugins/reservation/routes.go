package reservation

import (
	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(router chi.Router, authConfig kit.AuthenticationConfig) {

	// Public routes
	router.Get("/reservations", kit.Handler(HandleListTimeSlots))

	// Protected routes - require authentication
	router.Group(func(auth chi.Router) {
		// Apply authentication middleware
		auth.Use(kit.WithAuthentication(authConfig, true))

		// User reservation routes
		auth.Get("/reservations/create", kit.Handler(HandleReservationForm))
		auth.Post("/reservations/create", kit.Handler(HandleCreateReservation))
		auth.Get("/reservations/my", kit.Handler(HandleUserReservations))
		auth.Post("/reservations/cancel/{id}", kit.Handler(HandleCancelReservation))

		// Admin routes - could add another middleware to check admin status
		auth.Group(func(admin chi.Router) {
			// Add admin middleware here if needed
			admin.Get("/admin/timeslots/create", kit.Handler(HandleCreateTimeSlotForm))
			admin.Post("/admin/timeslots/create", kit.Handler(HandleCreateTimeSlot))
		})
	})
}
