package models

type TokenEntity struct {
	Base   string
	Lemma  string
	Pos    string
	Prob   float64
	Class  int
	Role   int
	Weight float64
	Sense  int
}

func NewTokenEntity(base string, lemma string, pos string, prob float64) *TokenEntity {
	return &TokenEntity{
		Base:  base,
		Lemma: lemma,
		Pos:   pos,
		Prob:  prob,
	}
}
