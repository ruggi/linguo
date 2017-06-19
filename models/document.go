package models

import (
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

type DocumentEntity struct {
	id          string
	timestamp   int64
	Status      string
	URL         string `param:"url"`
	Title       string `param:"title"`
	Description string `param:"description"`
	Keywords    string `param:"keywords"`
	Content     string `param:"content"`
	TopImage    string
	Language    string `param:"lang"`
	Sentences   []*SentenceEntity
	Unknown     map[string]int64
	Entities    []*Entity
}

func NewDocumentEntity() *DocumentEntity {
	return &DocumentEntity{}
}

func (e *DocumentEntity) Init() {
	u4, _ := uuid.NewV4()
	e.id = u4.String()
	e.timestamp = time.Now().UnixNano()
	e.Unknown = make(map[string]int64)
	e.Status = ""
}

func (e *DocumentEntity) AddSentenceEntity(se *SentenceEntity) {
	e.Sentences = append(e.Sentences, se)
}

func (e *DocumentEntity) AddUnknownEntity(name string, frequency int64) {
	e.Unknown[name] = frequency
}

func (e *DocumentEntity) String() string {
	return e.URL
}
