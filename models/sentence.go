package models

type SentenceEntity struct {
	Body     string
	Tokens   []*TokenEntity
	Weight   float64
	Sentence interface{}
}

func NewSentenceEntity() *SentenceEntity {
	return &SentenceEntity{}
}

func (e *SentenceEntity) AddTokenEntity(te *TokenEntity) {
	e.Tokens = append(e.Tokens, te)
}

func (e *SentenceEntity) SetBody(body string)              { e.Body = body }
func (e *SentenceEntity) SetSentence(sentence interface{}) { e.Sentence = sentence }

func (e *SentenceEntity) GetSentence() interface{} { return e.Sentence }
