package hang

import "time"

type Hanger struct {
	Ratio float64
}

func (h *Hanger) Hang() {
	time.Sleep(time.Duration(h.Ratio) * time.Second)
	h.Ratio = h.Ratio * h.Ratio
}
