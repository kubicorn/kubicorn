package try

import "time"

type Trier struct {
	maxAttempts int
	hangSeconds int
}

func NewTrier(attempts, hangSeconds int) *Trier {
	return &Trier{}
}

type tryFunc func(args ...interface{}) (bool, error)

func (t *Trier) Func(f tryFunc, args ...interface{}) (bool, error) {
	var err error
	for i := 0; i < t.maxAttempts; i++ {
		ok, err := f(args...)
		if err != nil {
			time.Sleep(time.Duration(t.hangSeconds) * time.Second)
			continue
		}
		if !ok {
			time.Sleep(time.Duration(t.hangSeconds) * time.Second)
			continue
		}
		return true, nil
	}
	return false, err
}

