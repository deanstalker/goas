package util

import "github.com/iancoleman/orderedmap"

type ChainedOrderedMap struct {
	m *orderedmap.OrderedMap
}

func (c ChainedOrderedMap) New() ChainedOrderedMap {
	c.m = orderedmap.New()
	return c
}

func (c ChainedOrderedMap) Set(key string, value interface{}) ChainedOrderedMap {
	c.m.Set(key, value)
	return c
}

func (c ChainedOrderedMap) GetMap() *orderedmap.OrderedMap {
	return c.m
}
