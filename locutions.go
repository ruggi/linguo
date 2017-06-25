package linguo

import (
	"container/list"
	"io/ioutil"
	"strconv"
	"strings"

	set "gopkg.in/fatih/set.v0"
)

const (
	LOCUTIONS_ST_P = 1 + iota
	LOCUTIONS_ST_M
	LOCUTIONS_ST_STOP
)

const (
	LOCUTIONS_TK_pref = 1 + iota
	LOCUTIONS_TK_mw
	LOCUTIONS_TK_prefL
	LOCUTIONS_TK_mwL
	LOCUTIONS_TK_prefP
	LOCUTIONS_TK_mwP
	LOCUTIONS_TK_other
)

const (
	LOCUTIONS_TAGSET = 1 + iota
	LOCUTIONS_MULTIWORDS
	LOCUTIONS_ONLYSELECTED
)

type LocutionStatus struct {
	AutomatStatus
	accMW, longestMW *set.Set
	components       []*Word
	overLongest      int
	mwAnalysis       *list.List
	mwAmbiguous      bool
}

func NewLocutionStatus() *LocutionStatus {
	return &LocutionStatus{
		accMW:      set.New(),
		longestMW:  set.New(),
		mwAnalysis: list.New(),
		components: make([]*Word, 0),
	}
}

type Locutions struct {
	Automat
	locut        map[string]string
	prefixes     *set.Set
	Tags         *TagSet
	onlySelected bool
}

func NewLocutions(locFile string) *Locutions {
	locutions := Locutions{
		locut:    make(map[string]string),
		prefixes: set.New(),
	}

	/*
		cfg := NewConfigFile(false, "##")
		cfg.AddSection("TagSetFile", LOCUTIONS_TAGSET)
		cfg.AddSection("Multiwords", LOCUTIONS_MULTIWORDS)
		cfg.AddSection("OnlySelected", LOCUTIONS_ONLYSELECTED)
	*/
	filestr, err := ioutil.ReadFile(locFile)
	if err != nil {
		panic("Error opening file " + locFile)
	}
	lines := strings.Split(string(filestr), "\n")

	for _, line := range lines {
		locutions.addLocution(line)
	}

	locutions.initialState = LOCUTIONS_ST_P
	locutions.stopState = LOCUTIONS_ST_STOP
	if locutions.final == nil {
		locutions.final = set.New()
	}
	locutions.final.Add(LOCUTIONS_ST_M)
	var s, t int
	for s = 0; s < AUTOMAT_MAX_STATES; s++ {
		for t = 0; t < AUTOMAT_MAX_TOKENS; t++ {
			locutions.trans[s][t] = LOCUTIONS_ST_STOP
		}
	}

	locutions.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_pref] = LOCUTIONS_ST_P
	locutions.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_prefL] = LOCUTIONS_ST_P
	locutions.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_prefP] = LOCUTIONS_ST_P
	locutions.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_mw] = LOCUTIONS_ST_M
	locutions.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_mwL] = LOCUTIONS_ST_M
	locutions.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_mwP] = LOCUTIONS_ST_M

	locutions.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_pref] = LOCUTIONS_ST_P
	locutions.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_prefL] = LOCUTIONS_ST_P
	locutions.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_prefP] = LOCUTIONS_ST_P
	locutions.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_mw] = LOCUTIONS_ST_M
	locutions.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_mwL] = LOCUTIONS_ST_M
	locutions.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_mwP] = LOCUTIONS_ST_M

	return &locutions
}

func (l *Locutions) BuildMultiword(se *Sentence, start *list.Element, end *list.Element, fs int, built *bool, st *LocutionStatus) *list.Element {
	mw := list.New()
	var form string
	for i := 0; i < st.shiftBegin && start != nil; i++ {
		start = start.Next()
	}
	var i *list.Element
	for i = start; i != end; i = i.Next() {
		mw.PushBack(i.Value.(*Word))
		form += i.Value.(*Word).getForm() + "_"
	}

	mw.PushBack(i.Value.(*Word))
	form += i.Value.(*Word).getForm()

	w := NewMultiword(form, mw)
	if l.ValidMultiWord(w, st) {
		end = end.Next()
		se.InsertBefore(w, start)
		for i = start; i != end; i = i.Next() {
			i.Value.(*Word).expired = true
		}
		i = end
		l.SetMultiwordAnalysis(w, fs, st)
		*built = true
	} else {
		l.ResetActions(st)
		i = start
		*built = false
	}

	return i
}

func (l *Locutions) addLocution(line string) {
	if line == "" {
		return
	}
	var prefix, key, lemma, tag string
	var p int
	items := Split(line, " ")
	key = items[0]

	lemma = items[1]
	tag = items[2]

	data := lemma + " " + tag
	var t [2]string
	i := 0

	for k := 3; k < len(items); k++ {
		t[i] = items[k]
		if i == 1 {
			data += "#" + t[0] + " " + t[1]
			t[0] = ""
			t[1] = ""
		}
		i = 1 - i
	}

	if t[0] == "" {
		t[0] = "I"
	}
	data += "|" + t[0]

	l.locut[key] = data

	prefix = ""
	p = strings.Index(key, "_")
	for p > -1 {
		prefix += key[0 : p+1]
		l.prefixes.Add(prefix)
		key = key[p+1:]
		p = strings.Index(key, "_")
	}
}

func (l *Locutions) setOnlySelected(b bool) {
	l.onlySelected = b
}

func (l *Locutions) check(s string, acc *set.Set, mw *bool, pref *bool, st *LocutionStatus) bool {
	if l.locut[s] != "" {
		acc.Add(s)
		st.longestMW = acc
		st.overLongest = 0
		*mw = true
	} else if l.prefixes.Has(s + "_") {
		acc.Add(s)
		*pref = true
	}

	return *mw || *pref
}

func (l *Locutions) ComputeToken(state int, j *list.Element, se *Sentence) int {
	st := se.getProcessingStatus().(*LocutionStatus)
	if st.components == nil {
		st.components = make([]*Word, 0)
	}
	st.components = append(st.components, j.Value.(*Word))
	var form, lem, tag string
	form = j.Value.(*Word).getLCForm()

	token := LOCUTIONS_TK_other

	acc := set.New()
	mw := false
	pref := false

	if j.Value.(*Word).Len() == 0 {
		if st.accMW.Size() == 0 {
			l.check(form, acc, &mw, &pref, st)
		} else {
			for _, i := range st.accMW.List() {
				l.check(i.(string)+"_"+form, acc, &mw, &pref, st)
			}
		}
	} else {
		first := j.Value.(*Word).Front()

		if l.onlySelected {
			first = j.Value.(*Word).selectedBegin(0).Element
		}
		for a := first; a != nil; a = a.Next() {
			bm := false
			bp := false
			lem = "<" + a.Value.(*Analysis).getLemma() + ">"
			tag = a.Value.(*Analysis).getTag()
			if l.Tags != nil {
				tag = l.Tags.GetShortTag(tag)
			}
			if st.accMW.Size() == 0 {
				l.check(form, acc, &bm, &bp, st)
				l.check(lem, acc, &bm, &bp, st)
				if l.check(tag, acc, &bm, &bp, st) {
					j.Value.(*Word).unselectAllAnalysis(0)
					a.Value.(*Analysis).markSelected(0)
				}

				mw = mw || bm
				pref = pref || bp
			} else {
				for _, i := range st.accMW.List() {
					l.check(i.(string)+"_"+form, acc, &bm, &bp, st)
					l.check(i.(string)+"_"+lem, acc, &bm, &bp, st)
					if l.check(i.(string)+"_"+tag, acc, &bm, &bp, st) {
						j.Value.(*Word).unselectAllAnalysis(0)
						a.Value.(*Analysis).markSelected(0)
					}
					mw = mw || bm
					pref = pref || bp
				}
			}
		}
	}

	if mw {
		token = LOCUTIONS_TK_mw
	} else if pref {
		token = LOCUTIONS_TK_pref
	}

	st.overLongest++
	st.accMW = acc

	return token
}

func (l *Locutions) ResetActions(st *LocutionStatus) {
	st.longestMW.Clear()
	st.accMW.Clear()
	st.components = make([]*Word, 0)
	st.mwAnalysis = st.mwAnalysis.Init()
}

func (l *Locutions) ValidMultiWord(w *Word, st *LocutionStatus) bool {
	var lemma, tag, check, par string
	var nc int
	la := list.New()
	valid := false
	ambiguous := false

	for _, m := range st.longestMW.List() {
		form := m.(string)
		if l.locut[form] != "" {

			mwData := l.locut[form]
			p := strings.Index(mwData, "|")
			tags := mwData[0:p]
			ldata := Split(tags, "#")
			amb := mwData[p+1:]
			ambiguous = ambiguous || (amb == "A")

			for _, k := range ldata {
				items := Split(k, " ")
				lemma = items[0]
				tag = items[1]

				p = If(string(lemma[0]) == "$", 0, -1).(int)

				for p > -1 {
					lf := lemma[p+1 : p+2]
					var repl string
					if lf == "F" {
						repl = st.components[p-1].getLCForm()
					} else if lf == "L" {
						repl = st.components[p-1].getLemma(0)
					} else {
						panic("Invalid lemma in locution entry " + form + " " + lemma + " " + tag)
					}

					lemma = strings.Replace(lemma, lemma[p:p+3], repl, p)
					p = strings.Index(lemma, "$")
				}

				if string(tag[0]) != "$" {
					la.PushBack(NewAnalysis(lemma, tag))
					valid = true
				} else {
					p = If(string(tag[0]) == ":", 0, -1).(int)
					if p > -1 {
						panic("Invalid tag in locution entry: " + form + " " + lemma + " " + tag)
					}

					check = tag[p+1:]
					nc, _ = strconv.Atoi(tag[1 : p-1])

					found := false
					for a := st.components[nc-1].Front(); a != nil; a = a.Next() {

						par = a.Value.(*Analysis).getTag()
						if strings.Index(par, check) == 0 {
							found = true
							la.PushBack(NewAnalysis(lemma, par))
						}
					}

					if !found {

					}
					valid = found
				}
			}
		}
	}

	st.mwAnalysis = la
	st.mwAmbiguous = ambiguous
	return valid
}

func (l *Locutions) SetMultiwordAnalysis(i *Word, fstate int, st *LocutionStatus) {
	i.setAnalysis(List2Array(st.mwAnalysis)...)
	i.setAmbiguousMw(st.mwAmbiguous)

}

func (l *Locutions) matching(se *Sentence, i *list.Element) bool {
	var j, sMatch, eMatch *list.Element
	var newstate, state, token, fstate int
	found := false

	pst := NewLocutionStatus()
	se.setProcessingStatus(pst)

	state = l.initialState
	fstate = 0
	l.ResetActions(pst)

	pst.shiftBegin = 0

	sMatch = i
	eMatch = nil
	for j = i; state != l.stopState && j != nil; j = j.Next() {
		token = l.ComputeToken(state, j, se)
		newstate = l.trans[state][token]

		state = newstate
		if l.final.Has(state) {
			eMatch = j
			fstate = state
		}
	}

	if eMatch != nil {
		i = l.BuildMultiword(se, sMatch, eMatch, fstate, &found, pst)
	}
	se.clearProcessingStatus()

	return found
}

func (l *Locutions) analyze(se *Sentence) {
	if l.analyzeMatching(se) {
		se.rebuildWordIndex()
	}
}

func (l *Locutions) analyzeMatching(se *Sentence) bool {
	for i := se.Front(); i != nil; i = i.Next() {
		if !i.Value.(*Word).isLocked() {
			if l.matching(se, i) {
				for {
					if i == nil || !i.Value.(*Word).expired {
						return true
					}
					i = i.Next()
				}
			}
		}
	}
	return false
}
