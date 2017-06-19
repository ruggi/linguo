package nlp

type NLPOptions struct {
	DataPath          string
	Lang              string
	TokenizerFile     string
	SplitterFile      string
	TaggerFile        string
	ShallowParserFile string
	SenseFile         string
	UKBFile           string
	DisambiguatorFile string
	MorfoOptions      *MacoOptions
}

func NewNLPOptions(dataPath string, lang string) *NLPOptions {
	return &NLPOptions{
		DataPath: dataPath,
		Lang:     lang,
	}
}

func (o *NLPOptions) TokenizerFilePath(path string) *NLPOptions {
	o.TokenizerFile = path
	return o
}

func (o *NLPOptions) SplitterFilePath(path string) *NLPOptions {
	o.SplitterFile = path
	return o
}

func (o *NLPOptions) TaggerFilePath(path string) *NLPOptions {
	o.TaggerFile = path
	return o
}

func (o *NLPOptions) ShallowParserFilePath(path string) *NLPOptions {
	o.ShallowParserFile = path
	return o
}

func (o *NLPOptions) SenseFilePath(path string) *NLPOptions {
	o.SenseFile = path
	return o
}

func (o *NLPOptions) UKBFilePath(path string) *NLPOptions {
	o.UKBFile = path
	return o
}

func (o *NLPOptions) DisambiguatorFilePath(path string) *NLPOptions {
	o.DisambiguatorFile = path
	return o
}

func (o *NLPOptions) WithMorfoOptions(options *MacoOptions) *NLPOptions {
	o.MorfoOptions = options
	return o
}
