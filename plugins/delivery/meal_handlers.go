package delivery

import (
	"fmt"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
)

// Validation schema for meal center
var mealCenterSchema = v.Schema{
	"Name":    v.Rules(v.Required, v.Max(255)),
	"Address": v.Rules(v.Required, v.Max(255)),
	"Phone":   v.Rules(v.Required, v.Max(20)),
}

// MealCenterFormValues struct for form handling
type MealCenterFormValues struct {
	Name    string  `form:"name"`
	Address string  `form:"address"`
	Phone   string  `form:"phone"`
	Lat     float64 `form:"lat"`
	Long    float64 `form:"long"`
	Success string
}

// GET handler to display the meal center form
func handleMealCenterForm(kit *kit.Kit) error {
	// Prepare empty form values
	values := MealCenterFormValues{}

	// Render the form
	return kit.Render(MealCenterShow(values))
}

// POST handler to process the meal center form submission
func handlePostMealCenter(kit *kit.Kit) error {
	// Parse and validate form values
	var values MealCenterFormValues
	errors, ok := v.Request(kit.Request, &values, mealCenterSchema)

	if !ok {
		fmt.Println(errors)
		return kit.Render(MealCenterForm(values, errors))
	}
	long, lat, err := fetchLongLat(values.Address)
	if err != nil {
		errors.Add("general", "Failed to fetch long lat")
		return kit.Render(MealCenterForm(values, errors))
	}
	// Create the meal center
	center, err := CreateMealCenter(
		values.Name,
		values.Address,
		values.Phone,
		lat,
		long,
	)

	if err != nil {
		fmt.Println(err)
		// Add general error
		errors.Add("general", "Failed to create meal center")
		return kit.Render(MealCenterForm(values, errors))
	}

	// Success: set success message and render form
	values.Success = fmt.Sprintf("Meal center '%s' created successfully!", center.Name)
	return kit.Render(MealCenterForm(values, v.Errors{}))
}
