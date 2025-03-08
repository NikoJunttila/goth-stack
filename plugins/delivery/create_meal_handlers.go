package delivery

import (
	"fmt"
	"gothstack/app/db"
	"strconv"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
)

// Validation schema for meal option
var mealOptionSchema = v.Schema{
	"Name":             v.Rules(v.Required, v.Max(255)),
	"Description":      v.Rules(v.Required, v.Max(1000)),
	"Price":            v.Rules(v.Required, v.Min(0)),
	"NutritionalInfo":  v.Rules(v.Required),
	"MaxDailyQuantity": v.Rules(v.Required, v.Min(0)),
	"MealPlanID":       v.Rules(v.Required),
}

// MealOptionFormValues struct for form handling
type MealOptionFormValues struct {
	MealPlanID          string   `form:"meal_plan_id"`
	Name                string   `form:"name"`
	Description         string   `form:"description"`
	Price               string   `form:"price"`
	NutritionalInfo     string   `form:"nutritional_info"`
	MaxDailyQuantity    string   `form:"max_daily_quantity"`
	DietaryRestrictions []string `form:"dietary_restrictions"`
	Success             string
}

// GET handler to display the meal option form
func handleMealOptionForm(kit *kit.Kit) error {
	// Get mealPlanID from URL parameters or query string
	mealPlanID := "1"

	// Prepare empty form values
	values := MealOptionFormValues{
		MealPlanID: mealPlanID,
	}

	// Fetch available dietary restrictions for the form
	restrictions, err := GetAllDietaryRestrictions()
	if err != nil {
		// Add error handling as needed
		return err
	}

	// Render the form
	return kit.Render(MealOptionShow(values, restrictions))
}

// POST handler to process the meal option form submission
func handlePostMealOption(kit *kit.Kit) error {
	// Parse and validate form values
	var values MealOptionFormValues
	errors, ok := v.Request(kit.Request, &values, mealOptionSchema)
	if !ok {
		// Fetch dietary restrictions for re-rendering the form
		restrictions, _ := GetAllDietaryRestrictions()
		return kit.Render(MealOptionForm(values, restrictions, errors))
	}

	// Convert form values to appropriate types
	mealPlanID, err := strconv.ParseUint(values.MealPlanID, 10, 64)
	if err != nil {
		errors.Add("MealPlanID", "Invalid meal plan ID")
		restrictions, _ := GetAllDietaryRestrictions()
		return kit.Render(MealOptionForm(values, restrictions, errors))
	}

	price, err := strconv.ParseFloat(values.Price, 64)
	if err != nil {
		errors.Add("Price", "Invalid price value")
		restrictions, _ := GetAllDietaryRestrictions()
		return kit.Render(MealOptionForm(values, restrictions, errors))
	}

	maxDaily, err := strconv.Atoi(values.MaxDailyQuantity)
	if err != nil {
		errors.Add("MaxDailyQuantity", "Invalid quantity value")
		restrictions, _ := GetAllDietaryRestrictions()
		return kit.Render(MealOptionForm(values, restrictions, errors))
	}

	// Convert dietary restriction IDs
	var dietaryRestrictionIDs []uint
	for _, id := range values.DietaryRestrictions {
		idVal, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			continue // Skip invalid IDs
		}
		dietaryRestrictionIDs = append(dietaryRestrictionIDs, uint(idVal))
	}

	// Create the meal option
	mealOption, err := CreateMealOption(
		uint(mealPlanID),
		values.Name,
		values.Description,
		price,
		values.NutritionalInfo,
		maxDaily,
		dietaryRestrictionIDs,
	)

	if err != nil {
		// Add general error
		errors.Add("general", "Failed to create meal option: "+err.Error())
		restrictions, _ := GetAllDietaryRestrictions()
		return kit.Render(MealOptionForm(values, restrictions, errors))
	}

	// Success: set success message and render form
	values.Success = fmt.Sprintf("Meal option '%s' created successfully!", mealOption.Name)
	restrictions, _ := GetAllDietaryRestrictions()
	return kit.Render(MealOptionForm(values, restrictions, v.Errors{}))
}
func handleShowMeals(kit *kit.Kit) error {
	// Get mealPlanID from URL parameters
	mealPlanIDStr := "1"
	mealPlanID, err := strconv.ParseUint(mealPlanIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid meal plan ID: %w", err)
	}

	// Fetch all meal options for the given meal plan ID
	var options []MealOption
	if err := db.Get().Where("meal_plan_id = ?", mealPlanID).Find(&options).Error; err != nil {
		return fmt.Errorf("error fetching meal options: %w", err)
	}

	return kit.Render(MealOptionList(options, mealPlanIDStr))
}
