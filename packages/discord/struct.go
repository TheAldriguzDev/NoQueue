package discord

type DiscordParams struct {
	URL    string  `json:"url"`
	Embeds []Embed `json:"embeds"`
}

type Embed struct {
	Author    Author    `json:"author,omitempty"`
	Title     string    `json:"title,omitempty"`
	Url       string    `json:"url,omitempty"`
	Color     int       `json:"color,omitempty"`
	Timestamp string    `json:"timestamp,omitempty"`
	Footer    Footer    `json:"footer,omitempty"`
	Fields    []Field   `json:"fields,omitempty"`
	Thumbnail Thumbnail `json:"thumbnail,omitempty"`
}

type Author struct {
	Name    string `json:"name,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
	URL     string `json:"url,omitempty"`
}

type Footer struct {
	Text string `json:"text,omitempty"`
}

type Field struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

type Thumbnail struct {
	URL string `json:"url,omitempty"`
}
