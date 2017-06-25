package linguo

import (
	"sync"
)

type Engine struct {
	semaphore *sync.Mutex
	NLP       *NLPEngine
	Ready     bool
}

func NewEngine() *Engine {
	return &Engine{
		semaphore: new(sync.Mutex),
		Ready:     false,
	}
}

func (e *Engine) InitNLP(path, lang string) {
	e.semaphore.Lock()
	defer e.semaphore.Unlock()

	if e.Ready {
		return
	}

	options := e.makeOptions(path, lang)
	nlpEngine := NewNLPEngine(options)

	e.NLP = nlpEngine
	e.Ready = true
}

func (e *Engine) makeOptions(path, lang string) *NLPOptions {
	macoOptions := NewMacoOptions(path, lang).
		PunctuationFilePath("/common/punct.dat").
		DictionaryFilePath("/" + lang + "/dicc.src").
		LocutionsFilePath("/" + lang + "/locucions-extended.dat").
		NPdataFilePath("/" + lang + "/np.dat").
		ProbabilityFilePath("/" + lang + "/probabilitats.dat")

	return NewNLPOptions(path, lang).
		TokenizerFilePath("/tokenizer.dat").
		SplitterFilePath("/splitter.dat").
		TaggerFilePath("/tagger.dat").
		ShallowParserFilePath("/chunker/grammar-chunk.dat").
		SenseFilePath("/senses.dat").
		UKBFilePath("/ukb.dat").
		DisambiguatorFilePath("/common/knowledge.dat").
		WithMorfoOptions(macoOptions)
}
