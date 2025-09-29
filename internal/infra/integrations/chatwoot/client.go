package chatwoot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// Client implements the ChatwootClient interface
type Client struct {
	baseURL    string
	token      string
	accountID  string
	httpClient *http.Client
	logger     *logger.Logger
}

// NewClient creates a new Chatwoot API client
func NewClient(baseURL, token, accountID string, logger *logger.Logger) *Client {
	return &Client{
		baseURL:   baseURL,
		token:     token,
		accountID: accountID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CreateInbox creates a new inbox in Chatwoot
func (c *Client) CreateInbox(name, webhookURL string) (*ports.ChatwootInbox, error) {
	c.logger.InfoWithFields("Creating Chatwoot inbox", map[string]interface{}{
		"name":        name,
		"webhook_url": webhookURL,
	})

	payload := map[string]interface{}{
		"name": name,
		"channel": map[string]interface{}{
			"type":        "api",
			"webhook_url": webhookURL,
		},
	}

	var inbox ports.ChatwootInbox
	err := c.makeRequest("POST", "/inboxes", payload, &inbox)
	if err != nil {
		return nil, fmt.Errorf("failed to create inbox: %w", err)
	}

	return &inbox, nil
}

// ListInboxes lists all inboxes
func (c *Client) ListInboxes() ([]ports.ChatwootInbox, error) {
	c.logger.Info("Listing Chatwoot inboxes")

	var response struct {
		Payload []ports.ChatwootInbox `json:"payload"`
	}

	err := c.makeRequest("GET", "/inboxes", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list inboxes: %w", err)
	}

	return response.Payload, nil
}

// GetInbox gets an inbox by ID
func (c *Client) GetInbox(inboxID int) (*ports.ChatwootInbox, error) {
	c.logger.InfoWithFields("Getting Chatwoot inbox", map[string]interface{}{
		"inbox_id": inboxID,
	})

	var inbox ports.ChatwootInbox
	err := c.makeRequest("GET", fmt.Sprintf("/inboxes/%d", inboxID), nil, &inbox)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbox: %w", err)
	}

	return &inbox, nil
}

// UpdateInbox updates an inbox
func (c *Client) UpdateInbox(inboxID int, updates map[string]interface{}) error {
	c.logger.InfoWithFields("Updating Chatwoot inbox", map[string]interface{}{
		"inbox_id": inboxID,
		"updates":  updates,
	})

	err := c.makeRequest("PATCH", fmt.Sprintf("/inboxes/%d", inboxID), updates, nil)
	if err != nil {
		return fmt.Errorf("failed to update inbox: %w", err)
	}

	return nil
}

// DeleteInbox deletes an inbox
func (c *Client) DeleteInbox(inboxID int) error {
	c.logger.InfoWithFields("Deleting Chatwoot inbox", map[string]interface{}{
		"inbox_id": inboxID,
	})

	err := c.makeRequest("DELETE", fmt.Sprintf("/inboxes/%d", inboxID), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete inbox: %w", err)
	}

	return nil
}

// CreateContact creates a new contact
func (c *Client) CreateContact(phone, name string, inboxID int) (*ports.ChatwootContact, error) {
	c.logger.InfoWithFields("Creating Chatwoot contact", map[string]interface{}{
		"phone":    phone,
		"name":     name,
		"inbox_id": inboxID,
	})

	payload := map[string]interface{}{
		"name":         name,
		"phone_number": phone,
		"inbox_id":     inboxID,
	}

	var contact ports.ChatwootContact
	err := c.makeRequest("POST", "/contacts", payload, &contact)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return &contact, nil
}

// FindContact finds a contact by phone number
func (c *Client) FindContact(phone string, inboxID int) (*ports.ChatwootContact, error) {
	c.logger.InfoWithFields("Finding Chatwoot contact", map[string]interface{}{
		"phone":    phone,
		"inbox_id": inboxID,
	})

	var response struct {
		Payload []ports.ChatwootContact `json:"payload"`
	}

	// URL encode the phone number to handle + and other special characters
	encodedPhone := url.QueryEscape(phone)
	err := c.makeRequest("GET", fmt.Sprintf("/contacts/search?q=%s", encodedPhone), nil, &response)
	if err != nil {
		c.logger.ErrorWithFields("Failed to search contact", map[string]interface{}{
			"phone":         phone,
			"encoded_phone": encodedPhone,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to find contact: %w", err)
	}

	c.logger.InfoWithFields("Contact search response", map[string]interface{}{
		"phone":         phone,
		"encoded_phone": encodedPhone,
		"payload_count": len(response.Payload),
		"response":      response,
	})

	if len(response.Payload) == 0 {
		return nil, fmt.Errorf("contact not found")
	}

	contact := &response.Payload[0]
	c.logger.InfoWithFields("Contact found", map[string]interface{}{
		"contact_id": contact.ID,
		"phone":      contact.PhoneNumber,
		"name":       contact.Name,
	})

	return contact, nil
}

// ListContactConversations lists all conversations for a contact (following Evolution API logic)
func (c *Client) ListContactConversations(contactID int) ([]ports.ChatwootConversation, error) {
	c.logger.InfoWithFields("Listing contact conversations", map[string]interface{}{
		"contact_id": contactID,
	})

	var response struct {
		Payload []ports.ChatwootConversation `json:"payload"`
	}

	err := c.makeRequest("GET", fmt.Sprintf("/contacts/%d/conversations", contactID), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list contact conversations: %w", err)
	}

	c.logger.InfoWithFields("Contact conversations listed", map[string]interface{}{
		"contact_id":          contactID,
		"conversations_count": len(response.Payload),
	})

	return response.Payload, nil
}

// GetContact gets a contact by ID
func (c *Client) GetContact(contactID int) (*ports.ChatwootContact, error) {
	c.logger.InfoWithFields("Getting Chatwoot contact", map[string]interface{}{
		"contact_id": contactID,
	})

	var contact ports.ChatwootContact
	err := c.makeRequest("GET", fmt.Sprintf("/contacts/%d", contactID), nil, &contact)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return &contact, nil
}

// UpdateContactAttributes updates contact attributes
func (c *Client) UpdateContactAttributes(contactID int, attributes map[string]interface{}) error {
	c.logger.InfoWithFields("Updating Chatwoot contact attributes", map[string]interface{}{
		"contact_id": contactID,
		"attributes": attributes,
	})

	payload := map[string]interface{}{
		"custom_attributes": attributes,
	}

	err := c.makeRequest("PUT", fmt.Sprintf("/contacts/%d", contactID), payload, nil)
	if err != nil {
		return fmt.Errorf("failed to update contact attributes: %w", err)
	}

	return nil
}

// CreateConversation creates a new conversation
func (c *Client) CreateConversation(contactID, inboxID int) (*ports.ChatwootConversation, error) {
	c.logger.InfoWithFields("Creating Chatwoot conversation", map[string]interface{}{
		"contact_id": contactID,
		"inbox_id":   inboxID,
	})

	payload := map[string]interface{}{
		"contact_id": contactID,
		"inbox_id":   inboxID,
	}

	var conversation ports.ChatwootConversation
	err := c.makeRequest("POST", "/conversations", payload, &conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return &conversation, nil
}

// GetConversation gets a conversation by contact and inbox
func (c *Client) GetConversation(contactID, inboxID int) (*ports.ChatwootConversation, error) {
	c.logger.InfoWithFields("Getting Chatwoot conversation", map[string]interface{}{
		"contact_id": contactID,
		"inbox_id":   inboxID,
	})

	var response struct {
		Payload []ports.ChatwootConversation `json:"payload"`
	}

	err := c.makeRequest("GET", fmt.Sprintf("/conversations?contact_id=%d&inbox_id=%d", contactID, inboxID), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if len(response.Payload) == 0 {
		return nil, fmt.Errorf("conversation not found")
	}

	return &response.Payload[0], nil
}

// GetConversationByID gets a conversation by ID
func (c *Client) GetConversationByID(conversationID int) (*ports.ChatwootConversation, error) {
	c.logger.InfoWithFields("Getting Chatwoot conversation by ID", map[string]interface{}{
		"conversation_id": conversationID,
	})

	var conversation ports.ChatwootConversation
	err := c.makeRequest("GET", fmt.Sprintf("/conversations/%d", conversationID), nil, &conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return &conversation, nil
}

// UpdateConversationStatus updates conversation status
func (c *Client) UpdateConversationStatus(conversationID int, status string) error {
	c.logger.InfoWithFields("Updating Chatwoot conversation status", map[string]interface{}{
		"conversation_id": conversationID,
		"status":          status,
	})

	payload := map[string]interface{}{
		"status": status,
	}

	err := c.makeRequest("POST", fmt.Sprintf("/conversations/%d/toggle_status", conversationID), payload, nil)
	if err != nil {
		return fmt.Errorf("failed to update conversation status: %w", err)
	}

	return nil
}

// SendMessage sends a message to a conversation
func (c *Client) SendMessage(conversationID int, content string) (*ports.ChatwootMessage, error) {
	c.logger.InfoWithFields("Sending message to Chatwoot", map[string]interface{}{
		"conversation_id": conversationID,
		"content":         content,
	})

	payload := map[string]interface{}{
		"content":      content,
		"message_type": "incoming", // Messages from WhatsApp are incoming to Chatwoot
	}

	var message ports.ChatwootMessage
	err := c.makeRequest("POST", fmt.Sprintf("/conversations/%d/messages", conversationID), payload, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &message, nil
}

// SendMediaMessage sends a media message to a conversation
func (c *Client) SendMediaMessage(conversationID int, content string, attachment io.Reader, filename string) (*ports.ChatwootMessage, error) {
	c.logger.InfoWithFields("Sending media message to Chatwoot", map[string]interface{}{
		"conversation_id": conversationID,
		"filename":        filename,
	})

	// TODO: Implement multipart form data upload for media
	// For now, just send as text message
	return c.SendMessage(conversationID, content)
}

// GetMessages gets messages from a conversation
func (c *Client) GetMessages(conversationID int, before int) ([]ports.ChatwootMessage, error) {
	c.logger.InfoWithFields("Getting messages from Chatwoot", map[string]interface{}{
		"conversation_id": conversationID,
		"before":          before,
	})

	var response struct {
		Payload []ports.ChatwootMessage `json:"payload"`
	}

	url := fmt.Sprintf("/conversations/%d/messages", conversationID)
	if before > 0 {
		url += fmt.Sprintf("?before=%d", before)
	}

	err := c.makeRequest("GET", url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return response.Payload, nil
}

// GetAccount gets account information
func (c *Client) GetAccount() (*ports.ChatwootAccount, error) {
	c.logger.Info("Getting Chatwoot account")

	var account ports.ChatwootAccount
	err := c.makeRequest("GET", "", nil, &account)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// UpdateAccount updates account information
func (c *Client) UpdateAccount(updates map[string]interface{}) error {
	c.logger.InfoWithFields("Updating Chatwoot account", map[string]interface{}{
		"updates": updates,
	})

	err := c.makeRequest("PATCH", "", updates, nil)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// makeRequest makes an HTTP request to the Chatwoot API
func (c *Client) makeRequest(method, endpoint string, payload interface{}, result interface{}) error {
	url := fmt.Sprintf("%s/api/v1/accounts/%s%s", c.baseURL, c.accountID, endpoint)

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_access_token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
