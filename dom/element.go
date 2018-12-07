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
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace,omitempty"`
	Attributes map[string]interface{} `json:"attributes"`
	Text       string                 `json:"text,omitempty"`
}
