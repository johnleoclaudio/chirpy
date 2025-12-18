package metrics

import (
	"fmt"
	"sync/atomic"
)

type API struct {
	FileserverHits atomic.Int32
}

func NewAPIMetrics() *API {
	return &API{}
}

func (c *API) IncMetric() {
	c.FileserverHits.Add(1)
}

func (c *API) GetMetrics() string {
	hits := c.FileserverHits.Load()
	return fmt.Sprintf("%v", hits)
}

func (c *API) ResetMetrics() bool {
	hits := c.FileserverHits.Load()
	success := c.FileserverHits.CompareAndSwap(hits, 0)
	return success
}
