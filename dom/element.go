package dom

type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Top    int `json:"top"`
	Left   int `json:"left"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
}

type Element struct {
	ID         string         `json:"id,omitempty"`
	Name       string         `json:"name,omitempty"`
	Namespace  string         `json:"namespace,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Text       string         `json:"text,omitempty"`
	HTML       string         `json:"html,omitempty"`
	Children   []*Element     `json:"children,omitempty"`
}
