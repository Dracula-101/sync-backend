package model

type CommunitySearchResult struct {
	Community
	Score          float64  `json:"score,omitempty"`
	RelevanceScore float64  `json:"relevanceScore,omitempty"`
	Matched        []string `json:"matched,omitempty"` // Fields that matched the search query
}
