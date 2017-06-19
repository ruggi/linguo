package nlp

import set "gopkg.in/fatih/set.v0"

type Accent struct {
	who AccentsModule
}

//Create the appropriate accents module (according to received options), and create a wrapper to access it.
func NewAccent(lang string) *Accent {
	accent := Accent{}
	if lang == "es" {
		//Create spanish accent handler
		who := NewAccentsES()
		accent.who = who
	} else {
		//Create Default (null) accent handler. Ok for English.
		who := NewAccetsDefault()
		accent.who = who
	}
	return &accent
}

//Wrapper methods: just call the wrapped accents module.
func (a *Accent) FixAccentutation(candidates *set.Set, suf *sufrule) {
	a.who.FixAccentuation(candidates, suf)
}
