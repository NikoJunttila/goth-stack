package delivery

import (
	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(router chi.Router, authConfig kit.AuthenticationConfig) {
	// Public routes - no authentication required
	router.Group(func(public chi.Router) {
		public.Get("/deliveries", kit.Handler(handleListDeliveries))

		public.Get("/orders-for-day/{id}", kit.Handler(handleGetMealsForDay))

		public.Get("/daily/meals/{id}", kit.Handler(handleShowMeals))

		public.Get("/meal-plans/{id}", kit.Handler(handleGetMealPlan))
		public.Get("/meal-plans", kit.Handler(handleListMealPlans))
		public.Get("/meal-plans/new", kit.Handler(handleMealPlanForm))
		public.Post("/meal-plans/new", kit.Handler(handlePostMealPlan))

		public.Get("/create-meal-option/{id}", kit.Handler(handleMealOptionForm))
		public.Post("/create-meal-option", kit.Handler(handlePostMealOption))
		public.Post("/create-meal-center", kit.Handler(handlePostMealCenter))
		public.Get("/create-meal-center", kit.Handler(handleMealCenterForm))

	})

	// Protected routes - authentication required
	router.Group(func(auth chi.Router) {
		// Apply authentication middleware
		auth.Use(kit.WithAuthentication(authConfig, true))
		auth.Get("/create-profile", kit.Handler(handleUserProfileForm))
		auth.Post("/create-profile", kit.Handler(handlePostUserProfile))
		auth.Post("/meals/{id}/buy", kit.Handler(handleMealPurchase))
		// auth.Post("/meal", kit.Handler(handlePostMeal))

		// Meal center management (admin only)
		/* 	auth.Group(func(admin chi.Router) {
			admin.Use(kit.WithRole("admin"))
			admin.Post("/meal-centers", kit.Handler(handleCreateMealCenter))
			admin.Post("/meal-plans", kit.Handler(handleCreateMealPlan))
			admin.Post("/dietary-restrictions", kit.Handler(handleCreateDietaryRestriction))
			admin.Post("/meal-options", kit.Handler(handleCreateMealOption))
		}) */

		// Driver routes
		/* 	auth.Group(func(driver chi.Router) {
			driver.Use(kit.WithRole("driver"))
			auth.Post("/drivers", kit.Handler(handleCreateDriver))
			auth.Post("/delivery-routes", kit.Handler(handleCreateDeliveryRoute))
			auth.Post("/delivery-routes/{id}/optimize", kit.Handler(handleOptimizeDeliveryRoute))
			auth.Post("/delivery-routes/{id}/start", kit.Handler(handleStartDeliveryRoute))
			auth.Post("/orders/{id}/delivered", kit.Handler(handleMarkOrderDelivered))
		}) */

		// Admin routes for order and delivery management
		/* 	auth.Group(func(admin chi.Router) {
			admin.Use(kit.WithRole("admin"))
			admin.Get("/orders/tomorrow", kit.Handler(handleGetTomorrowsOrders))
			admin.Post("/reset-meal-quantities", kit.Handler(handleResetDailyMealQuantities))
		}) */
	})
}
