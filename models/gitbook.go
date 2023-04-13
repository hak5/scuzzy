package models

type AISearchRequest struct {
	Query           string   `json:"query"`
	PreviousQueries []string `json:"previousQueries"`
}

type AISearchResult struct {
	Answer struct {
		Text              string   `json:"text"`
		FollowupQuestions []string `json:"followupQuestions"`
		Pages             []struct {
			Page     string   `json:"page"`
			Revision string   `json:"revision"`
			Space    string   `json:"space"`
			Sections []string `json:"sections"`
		} `json:"pages"`
	} `json:"answer"`
}
