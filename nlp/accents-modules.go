package nlp

import (
	"regexp"

	set "gopkg.in/fatih/set.v0"
)

type AccentsModule interface {
	FixAccentuation(*set.Set, *sufrule)
}

type AccentsDefault struct {
}

func NewAccetsDefault() *AccentsDefault {
	return &AccentsDefault{}
}

func (a *AccentsDefault) FixAccentuation(candidates *set.Set, suf *sufrule) {
	//TODO
}

type AccentsES struct {
	llanaAcc        *regexp.Regexp
	agudaMal        *regexp.Regexp
	monosil         *regexp.Regexp
	lastVowelPutAcc *regexp.Regexp
	lastVowelNotAcc *regexp.Regexp
	anyVowelAcc     *regexp.Regexp

	withAcc    map[string]string
	withoutAcc map[string]string
}

func NewAccentsES() *AccentsES {
	return &AccentsES{
		withAcc:    make(map[string]string),
		withoutAcc: make(map[string]string),
	}
}

func (a *AccentsES) FixAccentuation(candidates *set.Set, suf *sufrule) {
	//TODO
}
