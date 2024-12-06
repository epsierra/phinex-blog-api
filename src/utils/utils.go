package utils

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// JWT Secret Key - Assuming it's stored in an environment variable
var appSecretKey = os.Getenv("APP_SECRET_KEY")

// JwtEncode generates a JWT token with a payload similar to the TypeScript example
func JwtEncode(data map[string]interface{}) (string, error) {
	// Creating a new JWT token with the claims
	claims := jwt.MapClaims{
		"userId":          data["userId"],
		"isAuthenticated": data["isAuthenticated"],
		"email":           data["email"],
		"deviceId":        data["deviceId"],
	}

	// Create token using secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encodedToken, err := token.SignedString([]byte(appSecretKey))
	if err != nil {
		return "", err
	}

	return encodedToken, nil
}

// JwtDecode decodes the JWT token and returns the claims
func JwtDecode(tokenString string) (map[string]interface{}, error) {
	// Parse and validate the token using the secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure that the method used for signing is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Return the secret key to validate the token
		return []byte(appSecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	// Extract the claims (decoded data) from the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		data := make(map[string]interface{})
		// Extract fields from the token's claims
		if userId, exists := claims["userId"]; exists {
			data["userId"] = userId
		}
		if role, exists := claims["role"]; exists {
			data["role"] = role
		}
		if isAuthenticated, exists := claims["isAuthenticated"]; exists {
			data["isAuthenticated"] = isAuthenticated
		}
		if email, exists := claims["email"]; exists {
			data["email"] = email
		}
		if businessId, exists := claims["businessId"]; exists {
			data["businessId"] = businessId
		}
		if businessLocations, exists := claims["businessLocations"]; exists {
			data["businessLocations"] = businessLocations
		}
		if deviceId, exists := claims["deviceId"]; exists {
			data["deviceId"] = deviceId
		}

		return data, nil
	}

	return nil, errors.New("invalid token")
}

// HashData hashes sensitive data (like passwords)
func HashData(data interface{}) (string, error) {
	strData := String(data)
	// Generate salt for hashing
	salt, err := bcrypt.GenerateFromPassword([]byte(strData), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	// Return the hashed data
	return string(salt), nil
}

// MatchWithHashedData compares input data with hashed data
func MatchWithHashedData(data interface{}, hashedData string) (bool, error) {
	strData := String(data)
	// Compare the provided data with the stored hashed data
	err := bcrypt.CompareHashAndPassword([]byte(hashedData), []byte(strData))
	if err != nil {
		return false, err
	}
	// If no error, the data matches
	return true, nil
}

// Helper function to convert any data to a string
func String(data interface{}) string {
	// Ensure we can safely convert data into string
	switch v := data.(type) {
	case string:
		return v
	default:
		// Convert non-string types to string
		bytes, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(bytes)
	}
}
