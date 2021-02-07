package main

// AdviceGetter interface
type AdviceGetter interface {
	GetAdvicesLimitedFor(topic string, amount int) ([]string, error)
	GetAdvicesFor(topic string) ([]string, error)
}

// CachedAdviceGetter struct just wraps the Advice query into a trivial cache manager
type CachedAdviceGetter struct {
	cache     CacheForAdvices
	retriever AdviceRetriever
	limiter   AdviceLimiter
}

// GetAdvicesLimitedFor function
func (g *CachedAdviceGetter) GetAdvicesLimitedFor(topic string, amount int) ([]string, error) {
	advices, err := g.GetAdvicesFor(topic)

	if err != nil {
		return nil, err
	}

	return g.limiter.LimitSliceTo(advices, amount), nil
}

// GetAdvicesFor function
func (g *CachedAdviceGetter) GetAdvicesFor(topic string) ([]string, error) {
	cached, _ := g.cache.Get(topic)
	if cached != nil {
		return cached, nil
	}

	advices, err := g.retriever.RetrieveForTopic(topic)
	if err != nil {
		return nil, err
	}

	stored, err := g.cache.Put(topic, advices)
	if err != nil {
		return nil, err
	}

	return stored, err
}
