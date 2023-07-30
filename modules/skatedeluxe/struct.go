package skatedeluxe

import tls_client "github.com/bogdanfinn/tls-client"

type skateDeluxeTask struct {
	httpClient     tls_client.HttpClient
	Email          string
	Password       string
	Product_id     string
	PaymentType    string
	Delay          int
	SuccessWebhook string
	FailedWebhook  string
	TaskId         string
	Mode           string
	Proxy          string
	AuthToken      string
	Region         string
	Sizes          []Size
	Product_Price  float64
	Product_Image  string
	Product_Name   string
	ProxyArr       []string
}

// Responses

type EmailResponse struct {
	Status     string   `json:"status"`
	StatusCode int      `json:"status_code"`
	Customer   Customer `json:"customer"`
}

type Customer struct {
	Found    bool `json:"found,omitempty"`
	Status   int  `json:"status,omitempty"`
	Blocked  bool `json:"blocked,omitempty"`
	Closed   bool `json:"closed,omitempty"`
	Attempts int  `json:"attempts,omitempty"`
	Captcha  bool `json:"captcha,omitempty"`
}

type LoginPayLoad struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type LoginResponse struct {
	Status       string `json:"status,omitempty"`
	StatusCode   int    `json:"status_code,omitempty"`
	CountryID    string `json:"country_id,omitempty"`
	LanguageID   string `json:"language_id,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

type ATCPayload struct {
	ProductID         string        `json:"product_id,omitempty"`
	OptionID          string        `json:"option_id,omitempty"`
	OptionTransformID string        `json:"option_transform_id,omitempty"`
	Quantity          int           `json:"quantity,omitempty"`
	Children          []interface{} `json:"children"`
}

type Size struct {
	Size  string `json:"size"`
	Value string `json:"value"`
}

type AtcResponse struct {
	Status       string   `json:"status"`
	StatusCode   int      `json:"status_code"`
	Item         Item     `json:"item"`
	InvalidItems []string `json:"invalid_items"`
}

type Item struct {
	ItemID            *int     `json:"item_id,omitempty"`
	ParentItemID      *int     `json:"parent_item_id,omitempty"`
	PreviousItemID    int      `json:"previous_item_id"`
	CustomerID        *int     `json:"customer_id,omitempty"`
	SessionID         *int     `json:"session_id,omitempty"`
	ProductID         int      `json:"product_id"`
	OptionID          int      `json:"option_id"`
	OptionTransformID *int     `json:"option_transform_id,omitempty"`
	SetupID           *int     `json:"setup_id,omitempty"`
	Quantity          int      `json:"quantity"`
	TimeAdded         *string  `json:"time_added,omitempty"`
	TimeModified      *string  `json:"time_modified,omitempty"`
	Children          []string `json:"children"`
	ProductGroupID    *int     `json:"product_group_id,omitempty"`
}

type GetProductInfo struct {
	Items []struct {
		OptionTransformName string `json:"option_transform_name"`
		Product             struct {
			ProductID    int     `json:"product_id"`
			ProductPrice float64 `json:"product_price"`
			ProductImage string  `json:"product_image"`
			ProductName  string  `json:"product_name"`
		} `json:"product"`
	} `json:"items"`
}

type queryParam struct {
	key   string
	value string
}
