package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"zpwoot/internal/domain/contact"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/logger"
)

// MockWhatsAppClient implements the contact.WhatsAppClient interface for testing
type MockWhatsAppClient struct{}

func (m *MockWhatsAppClient) IsOnWhatsApp(ctx context.Context, phoneNumbers []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, phone := range phoneNumbers {
		result[phone] = map[string]interface{}{
			"phone_number":   phone,
			"is_on_whatsapp": true,
			"jid":            phone + "@s.whatsapp.net",
			"is_business":    false,
			"verified_name":  "",
		}
	}
	return result, nil
}

func (m *MockWhatsAppClient) GetProfilePictureInfo(ctx context.Context, jid string, preview bool) (map[string]interface{}, error) {
	return map[string]interface{}{
		"jid":         jid,
		"url":         "https://example.com/profile.jpg",
		"id":          "12345",
		"type":        "image",
		"direct_path": "/path/to/image",
		"has_picture": true,
	}, nil
}

func (m *MockWhatsAppClient) GetUserInfo(ctx context.Context, jids []string) ([]map[string]interface{}, error) {
	var users []map[string]interface{}
	for _, jid := range jids {
		users = append(users, map[string]interface{}{
			"jid":           jid,
			"phone_number":  jid,
			"name":          "Test User",
			"status":        "Hey there! I am using WhatsApp.",
			"picture_id":    "12345",
			"is_business":   false,
			"verified_name": "",
			"is_contact":    true,
			"last_seen":     nil,
			"is_online":     false,
		})
	}
	return users, nil
}

func (m *MockWhatsAppClient) GetBusinessProfile(ctx context.Context, jid string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"jid":         jid,
		"name":        "Test Business",
		"category":    "Technology",
		"description": "A test business",
		"website":     "https://example.com",
		"email":       "test@example.com",
		"address":     "123 Test St",
		"verified":    true,
	}, nil
}

func main() {
	// Create logger
	logger := logger.New()

	// Test with mock client
	fmt.Println("=== Testing with Mock Client ===")
	mockClient := &MockWhatsAppClient{}
	testContactService(mockClient, logger)

	// Test with real wameow client (will fail without proper setup)
	fmt.Println("\n=== Testing with WameowClient (will show interface compatibility) ===")
	wameowClient := &wameow.WameowClient{}
	testContactServiceInterface(wameowClient, logger)
}

func testContactService(client contact.WhatsAppClient, logger *logger.Logger) {
	// Create contact service
	service := contact.NewService(client, logger)

	ctx := context.Background()

	// Test CheckWhatsApp
	fmt.Println("Testing CheckWhatsApp...")
	checkReq := &contact.CheckWhatsAppRequest{
		SessionID:    "test-session",
		PhoneNumbers: []string{"+5511999999999", "+5511888888888"},
	}

	checkResp, err := service.CheckWhatsApp(ctx, checkReq)
	if err != nil {
		log.Printf("Error checking WhatsApp status: %v", err)
	} else {
		fmt.Printf("Check result: %d contacts checked\n", len(checkResp.Results))
		for _, result := range checkResp.Results {
			fmt.Printf("  %s: %t\n", result.PhoneNumber, result.IsOnWhatsApp)
		}
	}

	// Test GetProfilePicture
	fmt.Println("\nTesting GetProfilePicture...")
	profileReq := &contact.GetProfilePictureRequest{
		SessionID: "test-session",
		JID:       "5511999999999@s.whatsapp.net",
		Preview:   false,
	}
	
	profileResp, err := service.GetProfilePicture(ctx, profileReq)
	if err != nil {
		log.Printf("Error getting profile picture: %v", err)
	} else {
		fmt.Printf("Profile picture URL: %s\n", profileResp.URL)
	}

	// Test GetUserInfo
	fmt.Println("\nTesting GetUserInfo...")
	userReq := &contact.GetUserInfoRequest{
		SessionID: "test-session",
		JIDs:      []string{"5511999999999@s.whatsapp.net"},
	}
	
	userResp, err := service.GetUserInfo(ctx, userReq)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
	} else {
		fmt.Printf("User info: %d users found\n", len(userResp.Users))
		for _, user := range userResp.Users {
			fmt.Printf("  %s: %s\n", user.JID, user.Name)
		}
	}

	// Test GetBusinessProfile
	fmt.Println("\nTesting GetBusinessProfile...")
	businessReq := &contact.GetBusinessProfileRequest{
		SessionID: "test-session",
		JID:       "5511999999999@s.whatsapp.net",
	}
	
	businessResp, err := service.GetBusinessProfile(ctx, businessReq)
	if err != nil {
		log.Printf("Error getting business profile: %v", err)
	} else {
		fmt.Printf("Business profile: %s (%s)\n", businessResp.Profile.Name, businessResp.Profile.Category)
	}
}

func testContactServiceInterface(client contact.WhatsAppClient, logger *logger.Logger) {
	// This just tests that WameowClient implements the interface
	service := contact.NewService(client, logger)
	fmt.Printf("WameowClient successfully implements contact.WhatsAppClient interface: %T\n", service)
}
