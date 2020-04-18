package types

import "fmt"

type TxMarker struct {
	Height int64  `json:"height"`
	Memo   string `json:"memo"`
}

type TxMarkers []TxMarker

func NewTxMarker(height int64, memo string) TxMarker {
	return TxMarker{
		Height: height,
		Memo:   memo,
	}
}

func (m TxMarker) IsEmpty() bool {
	if m.Height == 0 {
		return true
	}
	if len(m.Memo) == 0 {
		return true
	}
	return false
}

// String implement of fmt.Stringer
func (m TxMarker) String() string {
	return fmt.Sprintf("Height: %d | Memo: %s", m.Height, m.Memo)
}

// Pops a memo out of the marker
func (mrks TxMarkers) Pop() (TxMarker, TxMarkers) {
	if len(mrks) == 0 {
		return TxMarker{}, nil
	}
	pop := mrks[0]
	markers := mrks[1:len(mrks)]

	return pop, markers
}

func (mrks TxMarkers) FilterByMinHeight(minHeight int64) TxMarkers {
	result := make(TxMarkers, 0)
	for _, mark := range mrks {
		if mark.Height >= minHeight {
			result = append(result, mark)
		}
	}
	return result
}
