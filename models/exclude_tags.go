package models

type ExcludeTags struct {
	Tags []Tag `json:"tags"`
}

type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
