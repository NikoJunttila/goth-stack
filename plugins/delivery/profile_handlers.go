package delivery

import (
	"errors"
	"fmt"
	"gothstack/app/db"
	"gothstack/plugins/auth"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"gorm.io/gorm"
)

var profileSchema = v.Schema{
	"DietaryNotes": v.Rules(v.Required, v.Max(255)),
	"Address":      v.Rules(v.Required, v.Max(255)),
	"Phone":        v.Rules(v.Required, v.Max(20)),
}

type UserProfileFormValues struct {
	Address       string `form:"address"`
	Phone         string `form:"phone"`
	DeliveryNotes string `form:"delivery_notes"`
	DietaryNotes  string `form:"dietary_notes"`
	//DietaryRestrictionIDs []string `form:"dietaryRestrictionIDs"`
	Success string
}

// GET handler to display the user profile form
func handleUserProfileForm(kit *kit.Kit) error {
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	// Fetch existing profile, if any
	var profile UserProfile
	err := db.Get().Where("user_id = ?", userID).Preload("DietaryRestrictions").First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Fetch all dietary restrictions for the form
	var restrictions []DietaryRestriction
	if err := db.Get().Find(&restrictions).Error; err != nil {
		return err
	}

	// Prepare form values and selected restrictions
	values := UserProfileFormValues{}
	selected := make(map[uint]bool)
	if profile.ID != 0 {
		// Pre-fill form with existing profile data
		values.Address = profile.Address
		values.Phone = profile.PhoneNumber
		values.DeliveryNotes = profile.DeliveryNotes
		values.DietaryNotes = profile.DietaryNotes
		for _, r := range profile.DietaryRestrictions {
			selected[r.ID] = true
		}
	}

	// Prepare data for the template
	data := UserProfileFormValues{
		Address:       values.Address,
		Phone:         values.Phone,
		DeliveryNotes: values.DeliveryNotes,
		DietaryNotes:  values.DietaryNotes,
		//DietaryRestrictionIDs: []string{},
	}

	return kit.Render(ProfileShow(data))
}

// POST handler to process the form submission
func handlePostUserProfile(kit *kit.Kit) error {
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	// Parse and validate form values
	var values UserProfileFormValues
	errors, ok := v.Request(kit.Request, &values, profileSchema)
	fmt.Println(values.DietaryNotes)
	if !ok {
		fmt.Println(errors)
		return kit.Render(UserProfileForm(values, errors))
	}
	_, err := CreateUserProfile(
		userID,
		values.Address,
		values.Phone,
		values.DeliveryNotes,
		values.DietaryNotes,
		[]uint{},
	)
	if err != nil {
		fmt.Println(err)
		return kit.Render(UserProfileForm(values, errors))
	}

	// Success: redirect to profile view page
	return kit.Render(UserProfileForm(values, v.Errors{}))
}
