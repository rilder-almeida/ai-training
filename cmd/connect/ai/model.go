package ai

// MetaData represents the metadata that is assoicated with a board.
type MetaData struct {
	Markers  int    `json:"markers" bson:"markers"`
	Moves    []int  `json:"moves" bson:"moves"`
	Feedback string `json:"feedback" bson:"feedback"`
}

// Board represents connect 4 board information.
type Board struct {
	ID        string    `bson:"board_id"`
	Board     string    `bson:"board"`
	MetaData  MetaData  `bson:"meta_data"`
	Embedding []float64 `bson:"embedding"`
}

// SimilarBoard represents connect 4 board found in the similarity search.
type SimilarBoard struct {
	ID        string    `bson:"board_id"`
	Board     string    `bson:"board"`
	MetaData  MetaData  `bson:"meta_data"`
	Embedding []float64 `bson:"embedding"`
	Score     float64   `bson:"score"`
}

// PickResponse provides the LLM's choice for the next move.
type PickResponse struct {
	Column   int
	Reason   string
	Attmepts int
}
