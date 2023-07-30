package auth

import "time"

type WhopResponse struct {
	ID                    string       `json:"id"`
	Product               string       `json:"product"`
	User                  string       `json:"user"`
	Plan                  string       `json:"plan"`
	Email                 string       `json:"email"`
	Status                string       `json:"status"`
	Valid                 bool         `json:"valid"`
	CancelAtPeriodEnd     bool         `json:"cancel_at_period_end"`
	PaymentProcessor      string       `json:"payment_processor"`
	LicenseKey            string       `json:"license_key"`
	UserMetadata          UserMetadata `json:"metadata"`
	Quantity              int          `json:"quantity"`
	WalletAddress         *string      `json:"wallet_address"`
	CustomFieldsResponses struct{}     `json:"custom_fields_responses"`
	Discord               Discord      `json:"discord"`
	NFTTokens             []string     `json:"nft_tokens"`
	ExpiresAt             *time.Time   `json:"expires_at"`
	RenewalPeriodStart    *time.Time   `json:"renewal_period_start"`
	RenewalPeriodEnd      *time.Time   `json:"renewal_period_end"`
	CreatedAt             int64        `json:"created_at"`
	ManageURL             string       `json:"manage_url"`
	AffiliatePageURL      string       `json:"affiliate_page_url"`
	CheckoutSession       *string      `json:"checkout_session"`
	AccessPass            string       `json:"access_pass"`
	Deliveries            struct{}     `json:"deliveries"`
}

type UserMetadata struct {
	Key string `json:"key"`
}

type Discord struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	ImageURL string `json:"image_url"`
}
