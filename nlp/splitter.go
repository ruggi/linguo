package nlp

import (
	"container/list"
	"strconv"
	"strings"

	set "gopkg.in/fatih/set.v0"
)

const SAME = 100
const VERY_LONG = 1000

const (
	SPLITTER_GENERAL = 1 + iota
	SPLITTER_MARKERS
	SPLITTER_SENT_END
	SPLITTER_SENT_START
)

type Splitter struct {
	SPLIT_AllowBetweenMarkers bool
	SPLIT_MaxWords            int64
	starters                  *set.Set
	enders                    map[string]bool
	markers                   map[string]int
}

func NewSplitter(splitterFile string) *Splitter {
	s := Splitter{
		starters: set.New(),
		enders:   make(map[string]bool),
		markers:  make(map[string]int),
	}

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("General", SPLITTER_GENERAL)
	cfg.AddSection("Markers", SPLITTER_MARKERS)
	cfg.AddSection("SentenceEnd", SPLITTER_SENT_END)
	cfg.AddSection("SentenceStart", SPLITTER_SENT_START)

	if !cfg.Open(splitterFile) {
		CRASH("Error opening file "+splitterFile, MOD_SPLITTER)
	}

	s.SPLIT_AllowBetweenMarkers = true
	s.SPLIT_MaxWords = 0

	nmk := 1
	line := ""

	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case SPLITTER_GENERAL:
			{
				name := items[0]
				if name == "AllowBetweenMarkers" {
					s.SPLIT_AllowBetweenMarkers, _ = strconv.ParseBool(items[1])
				} else if name == "MaxWords" {
					s.SPLIT_MaxWords, _ = strconv.ParseInt(items[1], 10, 64)
				} else {
					panic("Unexpected splitter option " + name)
				}
				break
			}
		case SPLITTER_MARKERS:
			{
				open := items[0]
				close := items[1]
				if open != close {
					s.markers[open] = nmk
					s.markers[close] = -nmk
				} else {
					s.markers[open] = SAME + nmk
					s.markers[close] = SAME + nmk
				}
				nmk++
				break
			}
		case SPLITTER_SENT_END:
			{
				name := items[0]
				value, _ := strconv.ParseBool(items[1])
				s.enders[name] = !value
				break
			}
		case SPLITTER_SENT_START:
			{
				s.starters.Add(line)
				break
			}
		default:
			break
		}
	}

	return &s
}

type SplitterStatus struct {
	BetweenMark  bool
	NoSplitCount int
	MarkType     *list.List
	MarkForm     *list.List
	buffer       *Sentence
	nsentence    int
}

func (s *Splitter) OpenSession() *SplitterStatus {
	return &SplitterStatus{
		BetweenMark:  false,
		NoSplitCount: 0,
		MarkType:     list.New(),
		MarkForm:     list.New(),
		buffer:       NewSentence(),
		nsentence:    0,
	}
}

func (s *Splitter) CloseSession(ses *SplitterStatus) {
	ses.MarkType = ses.MarkType.Init()
	ses.MarkForm = ses.MarkForm.Init()
	ses = nil
}

func (s *Splitter) Split(st *SplitterStatus, words []*Word, flush bool) []*Sentence {
	var sentences []*Sentence

	for i, w := range words {
		m := s.markers[w.getForm()]
		checkSplit := true

		if st.BetweenMark && !s.SPLIT_AllowBetweenMarkers && m != 0 && m == If(m > SAME, 1, -1).(int)*st.MarkType.Front().Value.(int) {
			st.MarkType.Remove(st.MarkType.Front())
			st.MarkForm.Remove(st.MarkForm.Front())
			if st.MarkForm.Len() == 0 {
				st.BetweenMark = false
				st.NoSplitCount = 0
			} else {
				st.NoSplitCount++
			}

			st.buffer.PushBack(w)
			checkSplit = false
		} else if m > 0 && !s.SPLIT_AllowBetweenMarkers {
			st.MarkForm.PushFront(w.getForm())
			st.MarkType.PushFront(m)
			st.BetweenMark = true
			st.NoSplitCount++
			st.buffer.PushBack(w)
			checkSplit = false
		} else if st.BetweenMark {
			st.NoSplitCount++
			if s.SPLIT_MaxWords == 0 || st.NoSplitCount <= int(s.SPLIT_MaxWords) {
				checkSplit = false
				st.buffer.PushBack(w)
			}

			if st.NoSplitCount == VERY_LONG {
			}
		}

		if checkSplit {
			e := s.enders[w.getForm()]
			if e {
				if e || s.endOfSentence(i, words) {
					st.buffer.PushBack(w)
					st.nsentence++
					st.buffer.sentID = strconv.Itoa(st.nsentence)
					sentences = append(sentences, st.buffer)
					nsentence := st.nsentence
					s.CloseSession(st)
					st = s.OpenSession()
					st.nsentence = nsentence
				} else {
					st.buffer.PushBack(w)
				}
			} else {
				st.buffer.PushBack(w)
			}
		}
	}

	if flush && st.buffer.Len() > 0 {
		st.nsentence++
		st.buffer.sentID = strconv.Itoa(st.nsentence)
		sentences = append(sentences, st.buffer)
		nsentence := st.nsentence
		s.CloseSession(st)
		st = s.OpenSession()
		st.nsentence = nsentence
	}

	return sentences
}

func (s *Splitter) endOfSentence(i int, words []*Word) bool {
	if i == len(words)-1 {
		return true
	} else {
		r := words[i+1]
		f := r.getForm()

		return strings.Title(f) == f || s.starters.Has(f)
	}
}
