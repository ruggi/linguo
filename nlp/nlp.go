package nlp

import (
	"strings"

	set "gopkg.in/fatih/set.v0"

	"github.com/federicoruggi/linguo/models"
)

type NLPEngine struct {
	options       *NLPOptions
	tokenizer     *Tokenizer
	splitter      *Splitter
	morfo         *Maco
	tagger        *HMMTagger
	grammar       *Grammar
	shallowParser *ChartParser
	sense         *Senses
	dsb           *UKB
	disambiguator *Disambiguator
	filter        *set.Set
	mitie         *MITIE
}

func NewNLPEngine(options *NLPOptions) *NLPEngine {
	e := &NLPEngine{
		options: options,
	}

	if options.TokenizerFile != "" {
		e.tokenizer = NewTokenizer(options.DataPath + "/" + options.Lang + "/" + options.TokenizerFile)
	}

	if options.SplitterFile != "" {
		e.splitter = NewSplitter(options.DataPath + "/" + options.Lang + "/" + options.SplitterFile)
	}

	if options.MorfoOptions != nil {
		e.morfo = NewMaco(options.MorfoOptions)
	}

	if options.SenseFile != "" {
		e.sense = NewSenses(options.DataPath + "/" + options.Lang + "/" + options.SenseFile)
	}

	if options.TaggerFile != "" {
		e.tagger = NewHMMTagger(options.DataPath+"/"+options.Lang+"/"+options.TaggerFile, true, FORCE_TAGGER, 1)
	}

	if options.ShallowParserFile != "" {
		e.grammar = NewGrammar(options.DataPath + "/" + options.Lang + "/" + options.ShallowParserFile)
		e.shallowParser = NewChartParser(e.grammar)
	}

	if options.UKBFile != "" {
		e.dsb = NewUKB(options.DataPath + "/" + options.Lang + "/" + options.UKBFile)
	}

	if options.DisambiguatorFile != "" {
		e.disambiguator = NewDisambiguator(options.DataPath + "/" + options.DisambiguatorFile)
	}

	e.mitie = NewMITIE(options.DataPath + "/" + options.Lang + "/mitie/ner_model.dat")
	return e
}

type Result struct {
	Sentences       []*models.SentenceEntity
	Entities        []*models.Entity
	UnknownEntities []*models.UnknownEntity
}

func (e *NLPEngine) Workflow(input string) Result {
	tokens := e.tokenizer.Tokenize(input, 0)

	sid := e.splitter.OpenSession()
	sentences := e.splitter.Split(sid, tokens, true)
	e.splitter.CloseSession(sid)

	for _, sentence := range sentences {
		if e.morfo != nil {
			e.morfo.Analyze(sentence)
		}
		if e.sense != nil {
			e.sense.Analyze(sentence)
		}
		if e.tagger != nil {
			e.tagger.Analyze(sentence)
		}
		if e.shallowParser != nil {
			e.shallowParser.Analyze(sentence)
		}
	}

	if e.dsb != nil {
		e.dsb.Analyze(sentences)
	}

	var sentenceEntities []*models.SentenceEntity
	entitiesFrequency := make(map[string]int64)

	for _, s := range sentences {
		se := models.NewSentenceEntity()
		body := ""
		for ww := s.Front(); ww != nil; ww = ww.Next() {
			w := ww.Value.(*Word)
			a := w.Front().Value.(*Analysis)
			te := models.NewTokenEntity(w.getForm(), a.getLemma(), a.getTag(), a.getProb())
			if a.getTag() == "NP" {
				entitiesFrequency[w.getForm()]++
			}
			body += w.getForm() + " "
			se.AddTokenEntity(te)
		}
		body = strings.Trim(body, " ")
		se.SetBody(body)
		se.SetSentence(s)

		sentenceEntities = append(sentenceEntities, se)
	}

	entities := e.mitie.Process(input)
	var unknownEntities []*models.UnknownEntity

	for name, frequency := range entitiesFrequency {
		name = strings.Replace(name, "_", " ", -1)

		found := false
		for _, e := range entities {
			if e.Value == name {
				found = true
				break
			}
		}
		if found {
			continue
		}

		ue := models.NewUnknownEntity(name, frequency)
		unknownEntities = append(unknownEntities, ue)
	}

	return Result{
		Sentences:       sentenceEntities,
		Entities:        entities,
		UnknownEntities: unknownEntities,
	}
}
