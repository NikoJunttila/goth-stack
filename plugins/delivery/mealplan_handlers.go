package delivery

import (
	"errors"
	"fmt"
	"gothstack/app/db"
	"time"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"gorm.io/gorm"
)

// Validation schema for meal plan
var mealPlanSchema = v.Schema{
	"Name":        v.Rules(v.Required, v.Max(255)),
	"Description": v.Rules(v.Max(1000)),
	"MealDate":    v.Rules(v.Required),
}

// MealPlanFormValues struct for form handling
type MealPlanFormValues struct {
	MealCenterID uint   `form:"meal_center_id"`
	Name         string `form:"name"`
	Description  string `form:"description"`
	MealDate     string `form:"meal_date"` // Format: YYYY-MM-DD
	Success      string
}

// GET handler to display the meal plan form
func handleMealPlanForm(kit *kit.Kit) error {

	// Fetch all meal centers for dropdown
	var centers []MealCenter
	if err := db.Get().Find(&centers).Error; err != nil {
		return err
	}
	// Get pre-selected meal center if provided in query
	var centerID uint = 1

	// Prepare form values
	values := MealPlanFormValues{
		MealCenterID: centerID,
		MealDate:     time.Now().Format("2006-01-02"),
	}

	// Render the form
	return kit.Render(MealPlanShow(values, centers))
}

// POST handler to process the meal plan form submission
func handlePostMealPlan(kit *kit.Kit) error {
	fmt.Println("handlePostMealPlan")
	// Parse and validate form values
	var values MealPlanFormValues
	errors, ok := v.Request(kit.Request, &values, mealPlanSchema)
	fmt.Println(values)
	// Fetch all meal centers for dropdown (needed for re-rendering form with errors)
	var centers []MealCenter
	if err := db.Get().Find(&centers).Error; err != nil {
		return err
	}

	if !ok {
		fmt.Println(ok, errors)
		return kit.Render(MealPlanForm(values, errors, centers))
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", values.MealDate)
	if err != nil {
		errors.Add("StartDate", "Invalid date format")
		return kit.Render(MealPlanForm(values, errors, centers))
	}

	// Create the meal plan
	plan, err := CreateMealPlan(
		values.MealCenterID,
		values.Name,
		values.Description,
		startDate,
	)

	if err != nil {
		fmt.Println(err)
		// Add general error
		errors.Add("general", "Failed to create meal plan: "+err.Error())
		return kit.Render(MealPlanForm(values, errors, centers))
	}

	// Success: set success message and render form
	values.Success = fmt.Sprintf("Meal plan '%s' created successfully!", plan.Name)
	return kit.Render(MealPlanForm(values, v.Errors{}, centers))
}

// Function to get a meal plan by ID
func handleGetMealPlan(kit *kit.Kit) error {
	// Get ID from URL parameters
	id := 1
	fmt.Println("fetching meal plans")
	var plan DaysMeals
	// Make sure to use Preload properly
	err := db.Get().Preload("MealCenter").First(&plan, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println(err)
			return err
		}
		fmt.Println(err)
		return err
	}
	var mealOptions []MealOption
	if err := db.Get().Where("days_meals_id = ?", id).Find(&mealOptions).Error; err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(mealOptions)
	return kit.Render(ShowAllMealsInDay(mealOptions, "friday"))
}

// Function to list all meal plans
func handleListMealPlans(kit *kit.Kit) error {
	var plans []DaysMeals
	if err := db.Get().Preload("MealCenter").Find(&plans).Error; err != nil {
		return err
	}

	// Filter by meal center if specified
	var filteredPlans []DaysMeals
	var centerID uint = 1

	for _, plan := range plans {
		if plan.MealCenterID == centerID {
			filteredPlans = append(filteredPlans, plan)
		}
	}
	plans = filteredPlans

	// Get meal centers for filtering dropdown
	var centers []MealCenter
	if err := db.Get().Find(&centers).Error; err != nil {
		return err
	}

	return kit.Render(MealPlanList(plans, centers))
}
