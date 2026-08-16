package zenmanage

// DefaultsCollection stores default values by flag key.
type DefaultsCollection struct {
	values map[string]any
}

// NewDefaultsCollection creates an empty defaults collection.
func NewDefaultsCollection() *DefaultsCollection {
	return &DefaultsCollection{values: map[string]any{}}
}

// DefaultsFromMap creates defaults from a map.
func DefaultsFromMap(values map[string]any) *DefaultsCollection {
	c := NewDefaultsCollection()
	for k, v := range values {
		c.values[k] = v
	}
	return c
}

// Set sets a default value.
func (c *DefaultsCollection) Set(key string, value any) {
	if c.values == nil {
		c.values = map[string]any{}
	}
	c.values[key] = value
}

// Get gets a default value.
func (c *DefaultsCollection) Get(key string) (any, bool) {
	if c == nil || c.values == nil {
		return nil, false
	}
	v, ok := c.values[key]
	return v, ok
}

// Has checks if key exists.
func (c *DefaultsCollection) Has(key string) bool {
	_, ok := c.Get(key)
	return ok
}

// Delete removes a key.
func (c *DefaultsCollection) Delete(key string) {
	if c == nil || c.values == nil {
		return
	}
	delete(c.values, key)
}

// Clear removes all entries.
func (c *DefaultsCollection) Clear() {
	if c == nil {
		return
	}
	c.values = map[string]any{}
}

// Keys returns all keys.
func (c *DefaultsCollection) Keys() []string {
	if c == nil || c.values == nil {
		return nil
	}
	keys := make([]string, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}
	return keys
}

// Size returns item count.
func (c *DefaultsCollection) Size() int {
	if c == nil || c.values == nil {
		return 0
	}
	return len(c.values)
}

// All returns a shallow copy.
func (c *DefaultsCollection) All() map[string]any {
	if c == nil || c.values == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(c.values))
	for k, v := range c.values {
		out[k] = v
	}
	return out
}
