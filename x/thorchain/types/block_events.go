package types

// BlockEvents is aggregation of all the events happened in a block
type BlockEvents struct {
	Height int64  `json:"height"`
	Events Events `json:"events"`
}

// NewBlockEvents create a new instance of BlockEvents
func NewBlockEvents(height int64) *BlockEvents {
	return &BlockEvents{
		Height: height,
		Events: make(Events, 0),
	}
}

// AddEvent - add the given event to block
func (b *BlockEvents) AddEvent(event Event) {
	b.Events = append(b.Events, event)
}
