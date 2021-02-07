package main

// AdviceLimiter interface is used to trim the string slice by a defined amount
type AdviceLimiter interface {
	LimitSliceTo(advices []string, amount int) []string
}

// SimpleAdviceLimiter struct is the simplest implementation of the interface above
type SimpleAdviceLimiter struct{}

// LimitSliceTo function just returns a new copy of the given slice with the first i elements
func (limiter *SimpleAdviceLimiter) LimitSliceTo(advices []string, i int) []string {
	if i < 0 {
		return make([]string, 0)
	}

	l := len(advices)

	var amount int
	if l > i {
		amount = i
	} else {
		amount = l
	}

	ret := make([]string, amount)
	for cur := 0; cur < amount; cur++ {
		ret[cur] = advices[cur]
	}
	return ret
}
