package persistance

type Create struct {
	Entity interface{} `json:"entity,omitempty"`
}

type Get struct {
	Id interface{} `json:"id,omitempty"`
}

type Update struct {
	Entity interface{} `json:"entity,omitempty"`
}

type Delete struct {
	Id interface{} `json:"id,omitempty"`
}
