package main

import (
	"fmt"
	"log"

	"github.com/epsierra/phinex-blog-api/src/utils" // Make sure the utils path is correct
)

func main() {
	// Sample data to encode into the token excluding business and locations
	decodedData := map[string]interface{}{
		"userId":          "6cf8e684-6eb8-4b69-b576-8de846caef33", // Example userId
		"role":            "user",                                 // Example role
		"isAuthenticated": true,                                   // Example authentication status
		"email":           "john.doe@example.com",                 // Example email
		"deviceId":        "device_001",                           // Example deviceId
	}

	// Generate the token
	token, err := utils.JwtEncode(decodedData)
	if err != nil {
		log.Fatalf("Error generating token: %v", err)
	}

	// Output the generated token
	fmt.Println("Generated Token:", token)
}
