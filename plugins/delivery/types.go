package delivery

import (
	"errors"
	"fmt"
	"gothstack/app/db"
	"gothstack/plugins/auth"
	"time"

	"gorm.io/gorm"
)

// Event name constants
const (
	MealPlanCreatedEvent     = "delivery.mealPlanCreated"
	OrderCreatedEvent        = "delivery.orderCreated"
	OrderUpdatedEvent        = "delivery.orderUpdated"
	OrderCanceledEvent       = "delivery.orderCanceled"
	DeliveryAssignedEvent    = "delivery.deliveryAssigned"
	DeliveryCompletedEvent   = "delivery.deliveryCompleted"
	WeeklyMenuPublishedEvent = "delivery.weeklyMenuPublished"
)

// MealCenter represents the central kitchen/facility that prepares all meals
type MealCenter struct {
	gorm.Model
	Name        string
	Address     string
	Latitude    float64
	Longitude   float64
	PhoneNumber string
	IsActive    bool
	DaysMeals   []DaysMeals `gorm:"foreignKey:MealCenterID"`
}

// MealPlan represents a weekly or daily meal plan with multiple meal options
type DaysMeals struct {
	gorm.Model
	MealCenterID uint
	MealCenter   MealCenter `gorm:"foreignKey:MealCenterID"`
	Name         string
	Description  string
	MealDate     time.Time
	IsActive     bool
	MealOptions  []MealOption `gorm:"foreignKey:DaysMealsID"`
}

// MealOption represents a specific meal that can be ordered
type MealOption struct {
	gorm.Model
	DaysMealsID          uint
	Name                 string
	Description          string
	Price                float64
	Image                string
	NutritionalInfo      string
	IsAvailable          bool
	MaxDailyQuantity     int                   // Maximum that can be prepared per day
	CurrentDailyQuantity int                   // How many have been ordered for the next day
	DietaryRestrictions  []*DietaryRestriction `gorm:"many2many:meal_dietary_restrictions;"`
}

// DietaryRestriction represents different dietary needs
type DietaryRestriction struct {
	gorm.Model
	Name        string // e.g., "Diabetic", "Low Sodium", "Vegetarian"
	Description string
}

// UserProfile extends the basic user with senior-specific information
type UserProfile struct {
	gorm.Model
	UserID              uint      `gorm:"not null;uniqueIndex"`
	User                auth.User `gorm:"foreignKey:UserID;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Address             string
	Latitude            float64
	Longitude           float64
	PhoneNumber         string
	DeliveryNotes       string                // Special instructions for the delivery person
	DietaryNotes        string                // Any dietary preferences or restrictions
	DietaryRestrictions []*DietaryRestriction `gorm:"many2many:user_dietary_restrictions;"`
}

// CreateMealCenter creates the central meal preparation facility
func CreateMealCenter(name, address, phone string, lat, long float64) (MealCenter, error) {
	center := MealCenter{
		Name:        name,
		Address:     address,
		PhoneNumber: phone,
		Latitude:    lat,
		Longitude:   long,
		IsActive:    true,
	}

	result := db.Get().Create(&center)
	return center, result.Error
}

// CreateMealPlan creates a new meal plan for a specific time period
func CreateMealPlan(mealCenterID uint, name, description string, mealDate time.Time) (DaysMeals, error) {
	// Check if meal center exists
	var center MealCenter
	if err := db.Get().First(&center, mealCenterID).Error; err != nil {
		return DaysMeals{}, errors.New("meal center not found")
	}

	plan := DaysMeals{
		MealCenterID: mealCenterID,
		Name:         name,
		Description:  description,
		MealDate:     mealDate,
		IsActive:     true,
	}

	result := db.Get().Create(&plan)
	return plan, result.Error
}

// CreateDietaryRestriction adds a new dietary restriction type
func CreateDietaryRestriction(name, description string) (DietaryRestriction, error) {
	restriction := DietaryRestriction{
		Name:        name,
		Description: description,
	}

	result := db.Get().Create(&restriction)
	return restriction, result.Error
}

// CreateMealOption adds a meal option to a meal plan
func CreateMealOption(
	DaysMealsID uint,
	name,
	description string,
	price float64,
	nutritionalInfo string,
	maxDaily int,
	dietaryRestrictionIDs []uint,
) (MealOption, error) {
	// Check if meal plan exists
	var plan DaysMeals
	if err := db.Get().First(&plan, DaysMealsID).Error; err != nil {
		return MealOption{}, errors.New("meal plan not found")
	}

	// Create meal option
	mealOption := MealOption{
		DaysMealsID:      DaysMealsID,
		Name:             name,
		Description:      description,
		Price:            price,
		NutritionalInfo:  nutritionalInfo,
		IsAvailable:      true,
		MaxDailyQuantity: maxDaily,
	}

	// Use transaction to handle the meal option and dietary restrictions
	err := db.Get().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&mealOption).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return MealOption{}, err
	}

	return mealOption, nil
}
func fetchLongLat(address string) (float64, float64, error) {
	// Use a geocoding service to get the latitude and longitude
	fmt.Println(address)
	return 0, 0, nil
}

// CreateUserProfile creates or updates a user profile with delivery information
func CreateUserProfile(
	userID uint,
	address string,
	phone, deliveryNotes, dietaryNotes string,
	dietaryRestrictionIDs []uint,
) (UserProfile, error) {
	// Check if profile already exists
	var profile UserProfile
	result := db.Get().Where("user_id = ?", userID).First(&profile)
	long, lat, _ := fetchLongLat(address)
	// Create new profile or update
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new profile
			profile = UserProfile{
				UserID:        userID,
				Address:       address,
				Latitude:      long,
				Longitude:     lat,
				PhoneNumber:   phone,
				DeliveryNotes: deliveryNotes,
				DietaryNotes:  dietaryNotes,
			}

			// Use transaction to handle the profile and dietary restrictions
			err := db.Get().Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&profile).Error; err != nil {
					return err
				}

				// Add dietary restrictions if any
				if len(dietaryRestrictionIDs) > 0 {
					for _, restrictionID := range dietaryRestrictionIDs {
						var restriction DietaryRestriction
						if err := tx.First(&restriction, restrictionID).Error; err != nil {
							return errors.New("dietary restriction not found: " + string(rune(restrictionID)))
						}

						if err := tx.Exec("INSERT INTO user_dietary_restrictions (user_profile_id, dietary_restriction_id) VALUES (?, ?)",
							profile.ID, restrictionID).Error; err != nil {
							return err
						}
					}
				}

				return nil
			})

			if err != nil {
				return UserProfile{}, err
			}
		} else {
			return UserProfile{}, result.Error
		}
	} else {
		// Update existing profile
		updates := map[string]interface{}{
			"address":        address,
			"latitude":       0,
			"longitude":      0,
			"phone_number":   phone,
			"delivery_notes": deliveryNotes,
			"dietary_notes":  dietaryNotes,
		}

		// Update profile
		if err := db.Get().Model(&profile).Updates(updates).Error; err != nil {
			return UserProfile{}, err
		}

		// Update dietary restrictions
		if len(dietaryRestrictionIDs) > 0 {
			// Clear existing restrictions
			if err := db.Get().Exec("DELETE FROM user_dietary_restrictions WHERE user_profile_id = ?", profile.ID).Error; err != nil {
				return UserProfile{}, err
			}

			// Add new restrictions
			for _, restrictionID := range dietaryRestrictionIDs {
				if err := db.Get().Exec("INSERT INTO user_dietary_restrictions (user_profile_id, dietary_restriction_id) VALUES (?, ?)",
					profile.ID, restrictionID).Error; err != nil {
					return UserProfile{}, err
				}
			}
		}
	}

	// Load the dietary restrictions
	db.Get().Preload("DietaryRestrictions").First(&profile, profile.ID)

	return profile, nil
}

// ResetDailyMealQuantities resets the current quantity counters for meal options
// This should be run each night after the day's delivery is completed
func ResetDailyMealQuantities() error {
	return db.Get().Model(&MealOption{}).Where("is_available = ?", true).Update("current_daily_quantity", 0).Error
}
func GetAllDietaryRestrictions() ([]DietaryRestriction, error) {
	var restrictions []DietaryRestriction
	result := db.Get().Find(&restrictions)
	return restrictions, result.Error
}
