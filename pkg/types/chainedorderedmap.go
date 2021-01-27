package types

import (
	"github.com/iancoleman/orderedmap"
)

type ChainedOrderedMap struct {
	m *orderedmap.OrderedMap
}

func NewOrderedMap() *ChainedOrderedMap {
	return &ChainedOrderedMap{
		m: orderedmap.New(),
	}
}

func (c *ChainedOrderedMap) Set(key string, value interface{}) *ChainedOrderedMap {
	c.m.Set(key, value)
	return c
}

func (c *ChainedOrderedMap) GetMap() *orderedmap.OrderedMap {
	return c.m
}
