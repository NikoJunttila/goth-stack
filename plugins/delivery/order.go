package delivery

import (
	"errors"
	"fmt"
	"gothstack/app/db"
	"gothstack/plugins/auth"
	"math"
	"time"

	"gorm.io/gorm"
)

// Order status constants
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusPreparing = "preparing"
	OrderStatusDelivery  = "out_for_delivery"
	OrderStatusDelivered = "delivered"
	OrderStatusCanceled  = "canceled"
)

// Order represents a meal order placed by a user
type Order struct {
	gorm.Model
	UserID        uint
	User          auth.User `gorm:"foreignKey:UserID;references:id"`
	UserProfileID uint
	UserProfile   UserProfile `gorm:"foreignKey:UserProfileID"`
	Status        string      `gorm:"default:pending"`
	DeliveryDate  time.Time
	Note          string
	TotalPrice    float64
	OrderItems    []OrderItem   `gorm:"foreignKey:OrderID"`
	Delivery      *DeliveryInfo `gorm:"foreignKey:OrderID"`
}

// OrderItem represents an individual meal option in an order
type OrderItem struct {
	gorm.Model
	OrderID      uint
	MealOptionID uint
	MealOption   MealOption `gorm:"foreignKey:MealOptionID"`
	Quantity     int
	Price        float64 // Price at time of order
}

// DeliveryInfo stores information about the delivery of an order
type DeliveryInfo struct {
	gorm.Model
	OrderID         uint
	Order           Order     `gorm:"foreignKey:OrderID"` // Add this line
	DriverID        *uint     // Optional, if assigned to a specific driver
	ScheduledTime   time.Time // When delivery is scheduled
	ActualTime      *time.Time
	DeliveryStatus  string
	DeliveryNotes   string
	DeliveryAddress string
	Latitude        float64 // Delivery location coordinates
	Longitude       float64
	// Set to true if using the delivery address instead of the profile address
	CustomAddress bool
}

// PurchaseMealOption allows a user to purchase a meal option
func PurchaseMealOption(userID, mealOptionID uint) (*Order, error) {
	// Use a transaction to ensure data consistency
	var order *Order
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		// 1. Fetch the meal option
		var mealOption MealOption
		if err := tx.First(&mealOption, mealOptionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("meal option not found")
			}
			return err
		}

		// 2. Check if the meal option is available
		if !mealOption.IsAvailable {
			return errors.New("meal option is not available")
		}

		// 3. Check if the requested quantity is available
		remainingQuantity := mealOption.MaxDailyQuantity - mealOption.CurrentDailyQuantity
		if 1 > remainingQuantity {
			return errors.New("insufficient quantity available")
		}

		// 4. Get the user profile
		var userProfile UserProfile
		if err := tx.Where("user_id = ?", userID).First(&userProfile).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user profile not found, please complete your profile before ordering")
			}
			return err
		}

		// 5. Get the days meals to retrieve the meal date
		var daysMeals DaysMeals
		if err := tx.First(&daysMeals, mealOption.DaysMealsID).Error; err != nil {
			return fmt.Errorf("error finding meal plan: %w", err)
		}

		// Use the meal date from DaysMeals as delivery date
		deliveryDate := daysMeals.MealDate

		// Make sure delivery date is not in the past
		if deliveryDate.Before(time.Now()) {
			return errors.New("delivery date cannot be in the past")
		}

		// 6. Create the order
		order = &Order{
			UserID:        userID,
			UserProfileID: userProfile.ID,
			Status:        OrderStatusPending,
			DeliveryDate:  deliveryDate, // Set from DaysMeals.MealDate
			Note:          "",
			TotalPrice:    mealOption.Price,
		}

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 7. Create the order item
		orderItem := OrderItem{
			OrderID:      order.ID,
			MealOptionID: mealOptionID,
			Quantity:     1,
			Price:        mealOption.Price,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			return err
		}

		// 8. Create delivery info
		deliveryInfo := DeliveryInfo{
			OrderID:         order.ID,
			ScheduledTime:   deliveryDate, // Set from DaysMeals.MealDate
			DeliveryStatus:  "scheduled",
			DeliveryNotes:   userProfile.DeliveryNotes,
			DeliveryAddress: userProfile.Address,
			Latitude:        userProfile.Latitude,
			Longitude:       userProfile.Longitude,
		}

		if err := tx.Create(&deliveryInfo).Error; err != nil {
			return err
		}

		// 9. Update the meal option's current daily quantity
		if err := tx.Model(&mealOption).Update("current_daily_quantity", mealOption.CurrentDailyQuantity+1).Error; err != nil {
			return err
		}

		// 10. Emit order created event (placeholder for your event system)
		// EmitEvent(OrderCreatedEvent, order.ID)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load the order with relationships
	if err := db.Get().Preload("OrderItems.MealOption").Preload("UserProfile").Preload("Delivery").First(&order, order.ID).Error; err != nil {
		return nil, err
	}

	return order, nil
}

// FindOrdersByDaysMealsID retrieves all orders associated with a specific DaysMeals ID
func FindOrdersByDaysMealsID(daysMealsID uint) ([]Order, error) {
	var orders []Order

	// Start the query
	query := db.Get().
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Joins("JOIN meal_options ON order_items.meal_option_id = meal_options.id").
		Where("meal_options.days_meals_id = ?", daysMealsID).
		// Avoid duplicates if an order has multiple items from the same DaysMeals
		Group("orders.id")

	// Execute the query with preloaded relationships for comprehensive order information
	err := query.
		Preload("OrderItems.MealOption").
		Preload("UserProfile").
		Preload("User").
		Preload("Delivery").
		Find(&orders).Error

	if err != nil {
		return nil, fmt.Errorf("error finding orders for DaysMeals ID %d: %w", daysMealsID, err)
	}

	return orders, nil
}

// Get a user's orders
func GetUserOrders(userID uint) ([]Order, error) {
	var orders []Order
	result := db.Get().Where("user_id = ?", userID).Preload("OrderItems.MealOption").Preload("Delivery").Find(&orders)
	return orders, result.Error
}

// Get a specific order
func GetOrder(orderID uint) (*Order, error) {
	var order Order
	result := db.Get().Preload("OrderItems.MealOption").Preload("UserProfile").Preload("Delivery").First(&order, orderID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, result.Error
	}
	return &order, nil
}

// CancelOrder cancels an existing order if it's in a cancelable state
func CancelOrder(orderID, userID uint) error {
	return db.Get().Transaction(func(tx *gorm.DB) error {
		// Get the order
		var order Order
		if err := tx.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("order not found")
			}
			return err
		}

		// Check if order can be canceled
		if order.Status != OrderStatusPending && order.Status != OrderStatusConfirmed {
			return errors.New("order cannot be canceled in its current state")
		}

		// Update order status
		if err := tx.Model(&order).Update("status", OrderStatusCanceled).Error; err != nil {
			return err
		}

		// Update delivery status
		if err := tx.Model(&DeliveryInfo{}).Where("order_id = ?", orderID).Update("delivery_status", "canceled").Error; err != nil {
			return err
		}

		// Update meal option availability by returning quantities
		var orderItems []OrderItem
		if err := tx.Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
			return err
		}

		for _, item := range orderItems {
			var mealOption MealOption
			if err := tx.First(&mealOption, item.MealOptionID).Error; err != nil {
				continue // Skip if not found
			}

			newQuantity := mealOption.CurrentDailyQuantity - item.Quantity
			if newQuantity < 0 {
				newQuantity = 0
			}

			if err := tx.Model(&mealOption).Update("current_daily_quantity", newQuantity).Error; err != nil {
				return err
			}
		}

		// Emit order canceled event (placeholder for your event system)
		// EmitEvent(OrderCanceledEvent, order.ID)

		return nil
	})
}

// OptimizeDeliveryRoute creates an optimized delivery route for a driver
// using coordinates to minimize travel distance with a nearest neighbor approach
func OptimizeDeliveryRoute(deliveries []DeliveryInfo, startLat, startLng float64) ([]DeliveryInfo, error) {
	if len(deliveries) == 0 {
		return deliveries, nil
	}
	fmt.Println("Optimizing route for", len(deliveries), "deliveries")
	fmt.Println("*******Starting coordinates:", startLat, startLng)
	// Create a copy of the deliveries to sort
	optimizedRoute := make([]DeliveryInfo, len(deliveries))
	copy(optimizedRoute, deliveries)

	// Track which deliveries have been added to the route
	visited := make([]bool, len(optimizedRoute))
	result := make([]DeliveryInfo, 0, len(optimizedRoute))

	// Start from the provided starting point (e.g., restaurant or depot)
	currentLat, currentLng := startLat, startLng

	// Nearest neighbor algorithm - repeatedly find the closest unvisited point
	for len(result) < len(optimizedRoute) {
		nextIdx := -1
		minDistance := math.MaxFloat64

		// Find the closest unvisited delivery
		for i, delivery := range optimizedRoute {
			if !visited[i] {
				// Calculate distance using the Haversine formula
				distance := calculateDistance(currentLat, currentLng, delivery.Latitude, delivery.Longitude)

				if distance < minDistance {
					minDistance = distance
					nextIdx = i
				}
			}
		}

		if nextIdx == -1 {
			break // Should never happen unless there's a data issue
		}

		// Mark as visited and add to result
		visited[nextIdx] = true
		result = append(result, optimizedRoute[nextIdx])

		// Update current position
		currentLat = optimizedRoute[nextIdx].Latitude
		currentLng = optimizedRoute[nextIdx].Longitude
	}

	return result, nil
}

// calculateDistance uses the Haversine formula to calculate the distance between two points on Earth
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // km

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Differences
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad

	// Haversine formula
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadius * c
	return distance
}

func GetDeliveriesForDriver(driverID uint, deliveryDate time.Time, mealCenterID uint) ([]DeliveryInfo, error) {
	var deliveries []DeliveryInfo

	// Get all deliveries for the specified day and driver (if assigned)
	query := db.Get().
		Joins("JOIN orders ON delivery_infos.order_id = orders.id").
		Where("DATE(delivery_infos.scheduled_time) = DATE(?)", deliveryDate)

	// Add driver condition if specified
	if driverID > 0 {
		query = query.Where("delivery_infos.driver_id = ?", driverID)
	} else {
		// Unassigned deliveries only
		query = query.Where("delivery_infos.driver_id IS NULL")
	}

	// First, get the deliveries without preloading
	if err := query.Find(&deliveries).Error; err != nil {
		return nil, fmt.Errorf("error finding deliveries: %w", err)
	}

	// Then load orders and user profiles separately
	if len(deliveries) > 0 {
		orderIDs := make([]uint, len(deliveries))
		for i, d := range deliveries {
			orderIDs[i] = d.OrderID
		}

		// Load orders with user profiles
		var orders []Order
		if err := db.Get().
			Preload("UserProfile").
			Where("id IN ?", orderIDs).
			Find(&orders).Error; err != nil {
			return nil, fmt.Errorf("error finding orders: %w", err)
		}

		// Map orders to deliveries
		orderMap := make(map[uint]Order)
		for _, o := range orders {
			orderMap[o.ID] = o
		}

		for i := range deliveries {
			if order, exists := orderMap[deliveries[i].OrderID]; exists {
				deliveries[i].Order = order
			}
		}
	}

	// Get the meal center coordinates
	var mealCenter MealCenter
	if err := db.Get().First(&mealCenter, mealCenterID).Error; err != nil {
		return nil, fmt.Errorf("error finding meal center: %w", err)
	}

	// Optimize the route
	optimizedDeliveries, err := OptimizeDeliveryRoute(deliveries, mealCenter.Latitude, mealCenter.Longitude)
	if err != nil {
		return nil, fmt.Errorf("error optimizing route: %w", err)
	}

	return optimizedDeliveries, nil
}
