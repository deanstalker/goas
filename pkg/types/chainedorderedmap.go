package types

import (
	"encoding/json"
	"log"

	"github.com/iancoleman/orderedmap"
)

// ChainedOrderMap to enable chaining on orderedmap.OrderedMap
type ChainedOrderedMap struct {
	m *orderedmap.OrderedMap
}

// NewOrderedMap starts a new chained instance
func NewOrderedMap() *ChainedOrderedMap {
	return &ChainedOrderedMap{
		m: orderedmap.New(),
	}
}

// Set key/value on orderedmap.OrderedMap instance
func (c *ChainedOrderedMap) Set(key string, value interface{}) *ChainedOrderedMap {
	c.m.Set(key, value)
	return c
}

// Get key from orderedmap.OrderedMap
func (c *ChainedOrderedMap) Get(key string) (interface{}, bool) {
	return c.m.Get(key)
}

// GetMap fetches the underlying orderedmap.OrderedMap
func (c *ChainedOrderedMap) GetMap() *orderedmap.OrderedMap {
	return c.m
}

// MarshalJSON pass through
func (c *ChainedOrderedMap) MarshalJSON() ([]byte, error) {
	return c.m.MarshalJSON()
}

// MarshalYAML where orderedmap.OrderedMap cannot handle yaml marshalling
func (c *ChainedOrderedMap) MarshalYAML() (interface{}, error) {
	data, err := c.m.MarshalJSON()
	log.Printf("marshal json: %s", data)
	if err != nil {
		return nil, err
	}

	var in map[string]map[string]interface{}
	if err := json.Unmarshal(data, &in); err != nil {
		return nil, err
	}
	log.Printf("unmarshal json: %s", in)

	return in, nil
}
