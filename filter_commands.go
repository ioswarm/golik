package golik

type Filter struct {
	Filter string `json:"filter,omitempty"`
	From int `json:"from,omitempty"`
	Size int `json:"size,omitempty"`
	Meta map[string]string `json:"meta,omitempty"`
}

func (f *Filter) Condition() (Condition, error) {
	return Parse(f.Filter)
}


type Result struct {
	From int `json:"from"`
	Size int `json:"size,omitempty"`
	Count int `json:"count,omitempty"`
	Result []interface{} `json:"result"`
}

