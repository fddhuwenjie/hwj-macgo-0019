package domain

import "time"

type Category struct {
	BaseAggregate
	name        string
	description string
}

func NewCategory(id, name, description string, now time.Time) (*Category, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if name == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewCategory", ErrInvalidInvariant)
	}
	c := &Category{name: name, description: description}
	c.setID(id)
	c.setVersion(1)
	c.setCreatedAt(now)
	c.setUpdatedAt(now)
	return c, nil
}

func (c *Category) Name() string {
	if c == nil {
		return ""
	}
	return c.name
}

func (c *Category) Description() string {
	if c == nil {
		return ""
	}
	return c.description
}

func (c *Category) Rename(newName string, now time.Time) error {
	if c == nil {
		return ErrNilArgument
	}
	if newName == "" {
		return NewError(ErrorKindInvalidArgument, "Category.Rename", ErrInvalidInvariant)
	}
	c.name = newName
	c.bumpVersion(now)
	return nil
}

func (c *Category) Describe(description string, now time.Time) error {
	if c == nil {
		return ErrNilArgument
	}
	c.description = description
	c.bumpVersion(now)
	return nil
}

func (c *Category) Clone() *Category {
	if c == nil {
		return nil
	}
	cp := *c
	return &cp
}
