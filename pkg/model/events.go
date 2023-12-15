package model

type EventsRequest struct {
	Name   string `json:"name"`
	Fights []Fight
}

type EventResponse struct {
	EventId int32   `json:"event_id"`
	Name    string  `json:"name"`
	Fights  []Fight `json:"fights"`
	IsDone  bool    `json:"is_done"`
}

type FullEventResponse struct {
	EventId int32           `json:"event_id"`
	Name    string          `json:"name"`
	Fights  []FightResponse `json:"fights"`
	IsDone  bool            `json:"is_done"`
}

type FightResultRequest struct {
	FightId    int32 `json:"fight_id"`
	WinnerId   int32 `json:"winner_id"`
	NotContest bool  `json:"not_contest"`
}