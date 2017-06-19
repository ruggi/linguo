package models

import "fmt"

type Entity struct {
	Model string
	Score float64
	Value string
}

func NewEntity(model string, score float64, value string) *Entity {
	return &Entity{
		Model: model,
		Score: score,
		Value: value,
	}
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s:%0.3f:%s", e.Model, e.Score, e.Value)
}

type UnknownEntity struct {
	Name      string
	Frequency int64
}

func NewUnknownEntity(name string, frequency int64) *UnknownEntity {
	return &UnknownEntity{
		Name:      name,
		Frequency: frequency,
	}
}
