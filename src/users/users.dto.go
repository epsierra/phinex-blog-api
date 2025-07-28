package users

// CreateUserDto defines the input for creating a user
type CreateUserDto struct {
	FullName string `json:"fullName" validate:"required" example:"John Doe"`
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"password123"`
}

// UpdateUserDto defines the input for updating a user
type UpdateUserDto struct {
	FullName  string `json:"fullName" example:"John Doe"`
	FirstName string `json:"firstName" example:"John"`
	LastName  string `json:"lastName" example:"Doe"`
	Gender    string `json:"gender,omitempty" example:"Male"`
	Dob       string `json:"dob,omitempty" example:"1990-01-01"`
	Email     string `json:"email" validate:"email" example:"john.doe@example.com"`
	Password  string `json:"password" validate:"min=6" example:"newpassword123"`
}

// UserResponse represents the response for user operations
type UserResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// FollowUnfollowDto defines the input for follow/unfollow
type FollowUnfollowDto struct {
	FollowerId  string `json:"followerId" validate:"required" example:"user-id-1"`
	FollowingId string `json:"followingId" validate:"required" example:"user-id-2"`
}

// FollowResponse represents the response for follow/unfollow actions
type FollowResponse struct {
	Followed bool `json:"followed"`
}
