package types

import (
	"encoding/json"

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

// MarshalJSON pass through
func (c *ChainedOrderedMap) MarshalJSON() ([]byte, error) {
	return c.m.MarshalJSON()
}

func (c *ChainedOrderedMap) UnmarshalJSON(b []byte) error {
	c.m = orderedmap.New()
	return c.m.UnmarshalJSON(b)
}

// MarshalYAML where orderedmap.OrderedMap cannot handle yaml marshaling
func (c *ChainedOrderedMap) MarshalYAML() (interface{}, error) {
	data, err := c.m.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var in map[string]map[string]interface{}
	if err := json.Unmarshal(data, &in); err != nil {
		return nil, err
	}

	return in, nil
}

func (c *ChainedOrderedMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var in map[string]map[string]interface{}
	if err := unmarshal(&in); err != nil {
		return err
	}

	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return c.UnmarshalJSON(data)
}
