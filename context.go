package zenmanage

import (
	"encoding/json"
)

// Attribute is a high-level context attribute helper.
type Attribute struct {
	key    string
	values []string
}

// NewAttribute creates a context attribute.
func NewAttribute(key string, values []string) Attribute {
	return Attribute{key: key, values: append([]string{}, values...)}
}

// Key returns the attribute key.
func (a Attribute) Key() string {
	return a.key
}

// Values returns a copy of attribute values.
func (a Attribute) Values() []string {
	return append([]string{}, a.values...)
}

// AddValue appends a value to the attribute.
func (a *Attribute) AddValue(value string) {
	a.values = append(a.values, value)
}

// Context captures the evaluation context.
type Context struct {
	data ContextData
}

// NewContext creates a context.
func NewContext(typ, identifier, name string, attributes []Attribute) Context {
	ctxAttrs := make([]ContextAttribute, 0, len(attributes))
	for _, a := range attributes {
		vals := make([]ContextValue, 0, len(a.values))
		for _, v := range a.values {
			vals = append(vals, ContextValue{Value: v})
		}
		ctxAttrs = append(ctxAttrs, ContextAttribute{Key: a.key, Values: vals})
	}

	return Context{data: ContextData{
		Type:       typ,
		Name:       name,
		Identifier: identifier,
		Attributes: ctxAttrs,
	}}
}

// SingleContext creates a simple context with optional name.
func SingleContext(typ, identifier, name string) Context {
	return NewContext(typ, identifier, name, nil)
}

// ContextFromData creates context from already-structured data.
func ContextFromData(data ContextData) Context {
	return Context{data: data}
}

// Type returns the context type.
func (c Context) Type() string {
	return c.data.Type
}

// Name returns the context name.
func (c Context) Name() string {
	return c.data.Name
}

// Identifier returns the context identifier.
func (c Context) Identifier() string {
	return c.data.Identifier
}

// Attributes returns a copy of all attributes.
func (c Context) Attributes() []ContextAttribute {
	attrs := make([]ContextAttribute, len(c.data.Attributes))
	copy(attrs, c.data.Attributes)
	return attrs
}

// AddAttribute adds an attribute to the context.
func (c *Context) AddAttribute(attr Attribute) {
	vals := make([]ContextValue, 0, len(attr.values))
	for _, v := range attr.values {
		vals = append(vals, ContextValue{Value: v})
	}
	c.data.Attributes = append(c.data.Attributes, ContextAttribute{Key: attr.key, Values: vals})
}

// GetAttribute retrieves an attribute by key.
func (c Context) GetAttribute(key string) (ContextAttribute, bool) {
	for _, attr := range c.data.Attributes {
		if attr.Key == key {
			return attr, true
		}
	}
	return ContextAttribute{}, false
}

// IsEmpty determines whether context has meaningful fields.
func (c Context) IsEmpty() bool {
	return c.data.Type == "" && c.data.Name == "" && c.data.Identifier == "" && len(c.data.Attributes) == 0
}

// Data returns a copy of context data.
func (c Context) Data() ContextData {
	return c.data
}

// JSON serializes context into API representation.
func (c Context) JSON() ([]byte, error) {
	return json.Marshal(c.data)
}
