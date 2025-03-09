package delivery

import (
	"fmt"
	"gothstack/app/db"
	"gothstack/plugins/auth"
	"net/http"
	"strconv"
	"strings"

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

// Handler to process the purchase
func handleMealPurchase(kit *kit.Kit) error {
	// Parse form values
	// Get user ID from the session (adjust according to your auth system)
	mealOptionIDStr := chi.URLParam(kit.Request, "id")
	mealOptionID, err := strconv.ParseUint(mealOptionIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid meal option ID: %w", err)
	}
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID

	// Call the business logic function to purchase the meal
	order, err := PurchaseMealOption(userID, uint(mealOptionID))
	if err != nil {
		// Handle errors (e.g., insufficient quantity, meal not available)
		return fmt.Errorf("failed to purchase meal: %w", err)
	}
	fmt.Println(order)
	// Redirect to order confirmation page
	return kit.Redirect(http.StatusSeeOther, "/meal-plans")
}

func handleGetMealsForDay(kit *kit.Kit) error {
	// Parse form values
	// Get user ID from the session (adjust according to your auth system)
	mealOptionIDStr := chi.URLParam(kit.Request, "id")
	mealOptionID, err := strconv.ParseUint(mealOptionIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid meal option ID: %w", err)
	}

	orders, err := FindOrdersByDaysMealsID(uint(mealOptionID))
	if err != nil {
		// Handle errors (e.g., insufficient quantity, meal not available)
		return fmt.Errorf("failed to purchase meal: %w", err)
	}

	// Print orders in a more readable format
	fmt.Printf("Found %d orders for DaysMeals ID %d:\n", len(orders), mealOptionID)

	for i, order := range orders {
		fmt.Printf("\n--- Order #%d ---\n", i+1)
		fmt.Printf("Order ID: %d\n", order.ID)
		fmt.Printf("Status: %s\n", order.Status)
		fmt.Printf("Delivery Date: %s\n", order.DeliveryDate.Format("Mon, Jan 2, 2006"))
		fmt.Printf("Total Price: $%.2f\n", order.TotalPrice)

		// User information
		fmt.Printf("User: ID %d", order.UserID)
		if order.UserProfile.ID > 0 {
			fmt.Printf(" - %s\n", order.UserProfile.PhoneNumber)
			fmt.Printf("Delivery Address: %s\n", order.UserProfile.Address)
		} else {
			fmt.Println(" (profile not loaded)")
		}

		// Order items
		fmt.Println("Items:")
		for j, item := range order.OrderItems {
			fmt.Printf("  %d. %s (Qty: %d, Price: $%.2f)\n",
				j+1,
				item.MealOption.Name,
				item.Quantity,
				item.Price)
		}

		// Delivery info
		if order.Delivery != nil {
			fmt.Printf("Delivery Status: %s\n", order.Delivery.DeliveryStatus)
			fmt.Printf("Scheduled Time: %s\n", order.Delivery.ScheduledTime.Format("3:04 PM"))
			if order.Delivery.DeliveryNotes != "" {
				fmt.Printf("Delivery Notes: %s\n", order.Delivery.DeliveryNotes)
			}
		}

		fmt.Println(strings.Repeat("-", 30))
	}

	// Redirect to order confirmation page
	return kit.Redirect(http.StatusSeeOther, "/meal-plans")
}

func handleListDeliveries(kit *kit.Kit) error {
	// First, fetch the DaysMeals record using the provided ID
	var id string = "1"
	var daysMeal DaysMeals
	if result := db.Get().First(&daysMeal, id); result.Error != nil {
		return fmt.Errorf("failed to find day's meals with ID %s: %w", id, result.Error)
	}

	// Extract the meal date from the fetched DaysMeal
	mealDate := daysMeal.MealDate

	// Get the meal center ID from the daysMeal
	mealCenterID := daysMeal.MealCenterID

	// Now call GetDeliveriesForDriver with the correct parameters
	driverID := uint(0) // 0 means "unassigned drivers"
	deliveries, err := GetDeliveriesForDriver(driverID, mealDate, mealCenterID)
	if err != nil {
		return fmt.Errorf("failed to get deliveries: %w", err)
	}

	// Pretty print the deliveries
	if len(deliveries) == 0 {
		fmt.Println("No deliveries found for the specified date and criteria")
	} else {
		fmt.Println("\n===== DELIVERIES FOR", mealDate.Format("2006-01-02"), "=====")
		fmt.Println("Total Deliveries:", len(deliveries))
		fmt.Println("--------------------------------------------")

		for i, delivery := range deliveries {
			fmt.Printf("Delivery #%d (ID: %d)\n", i+1, delivery.ID)
			fmt.Printf("  Order ID: %d\n", delivery.OrderID)
			fmt.Printf("  Status: %s\n", delivery.DeliveryStatus)
			fmt.Printf("  Scheduled Time: %s\n", delivery.ScheduledTime.Format("2006-01-02 15:04:05"))

			if delivery.ActualTime != nil {
				fmt.Printf("  Actual Time: %s\n", delivery.ActualTime.Format("2006-01-02 15:04:05"))
			} else {
				fmt.Printf("  Actual Time: Not delivered yet\n")
			}

			if delivery.DriverID != nil {
				fmt.Printf("  Driver ID: %d\n", *delivery.DriverID)
			} else {
				fmt.Printf("  Driver: Unassigned\n")
			}

			fmt.Printf("  Address: %s\n", delivery.DeliveryAddress)
			fmt.Printf("  Coordinates: (%.6f, %.6f)\n", delivery.Latitude, delivery.Longitude)

			if delivery.DeliveryNotes != "" {
				fmt.Printf("  Notes: %s\n", delivery.DeliveryNotes)
			}

			if delivery.CustomAddress {
				fmt.Println("  Custom address is being used")
			}

			// Print customer info if available
			if delivery.Order.UserProfile.ID > 0 {
				fmt.Printf("  Customer: %s\n", delivery.Order.UserProfile.PhoneNumber)
			}

			fmt.Println("--------------------------------------------")
		}
	}

	return kit.Render(DeliveryList(deliveries))
}
