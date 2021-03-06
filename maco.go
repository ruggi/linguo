package linguo

type MacoOptions struct {
	Path                                                                                                                              string
	Lang                                                                                                                              string
	LocutionsFile, QuantitiesFile, AffixFile, CompoundFile, DictionaryFile, ProbabilityFile, NPdataFile, PunctuationFile, UserMapFile string
	Decimal, Thousand                                                                                                                 string
	ProbabilityThreshold                                                                                                              float64
	InverseDict, RetokContractions                                                                                                    bool
}

func NewMacoOptions(path, lang string) *MacoOptions {
	return &MacoOptions{
		Path:                 path,
		Lang:                 lang,
		UserMapFile:          "",
		LocutionsFile:        "",
		QuantitiesFile:       "",
		AffixFile:            "",
		ProbabilityFile:      "",
		DictionaryFile:       "",
		NPdataFile:           "",
		PunctuationFile:      "",
		CompoundFile:         "",
		Decimal:              "",
		Thousand:             "",
		ProbabilityThreshold: 0.001,
		InverseDict:          false,
		RetokContractions:    true,
	}
}

func (m *MacoOptions) UserMapFilePath(path string) *MacoOptions {
	m.UserMapFile = m.Path + path
	return m
}

func (m *MacoOptions) LocutionsFilePath(path string) *MacoOptions {
	m.LocutionsFile = m.Path + path
	return m
}

func (m *MacoOptions) QuantitiesFilePath(path string) *MacoOptions {
	m.QuantitiesFile = m.Path + path
	return m
}

func (m *MacoOptions) AffixFilePath(path string) *MacoOptions {
	m.AffixFile = m.Path + path
	return m
}

func (m *MacoOptions) ProbabilityFilePath(path string) *MacoOptions {
	m.ProbabilityFile = m.Path + path
	return m
}

func (m *MacoOptions) DictionaryFilePath(path string) *MacoOptions {
	m.DictionaryFile = m.Path + path
	return m
}

func (m *MacoOptions) NPdataFilePath(path string) *MacoOptions {
	m.NPdataFile = m.Path + path
	return m
}

func (m *MacoOptions) PunctuationFilePath(path string) *MacoOptions {
	m.PunctuationFile = m.Path + path
	return m
}

func (m *MacoOptions) CompoundFilePath(path string) *MacoOptions {
	m.CompoundFile = m.Path + path
	return m
}

func (this *MacoOptions) SetNumericalPoint(dec string, tho string) {
	this.Decimal = dec
	this.Thousand = tho
}

func (this *MacoOptions) SetThreshold(t float64) {
	this.ProbabilityThreshold = t
}

func (this *MacoOptions) SetInverseDict(b bool) {
	this.InverseDict = b
}

func (this *MacoOptions) SetRetokContractions(b bool) {
	this.RetokContractions = b
}

type Maco struct {
	MultiwordsDetection, NumbersDetection, PunctuationDetection, DatesDetection, QuantitiesDetection, DictionarySearch, ProbabilityAssignment, UserMap, NERecognition bool
	loc                                                                                                                                                               *Locutions
	dic                                                                                                                                                               *Dictionary
	prob                                                                                                                                                              *Probability
	punct                                                                                                                                                             *Punts
	npm                                                                                                                                                               *NER
	/*
		numb *Numbers
		dates *Dates
		quant *Quantities

		user *regexp.Regexp
	*/
}

func NewMaco(opts *MacoOptions) *Maco {
	this := Maco{
		MultiwordsDetection:   false,
		NumbersDetection:      false,
		PunctuationDetection:  false,
		DatesDetection:        false,
		QuantitiesDetection:   false,
		DictionarySearch:      false,
		ProbabilityAssignment: false,
		UserMap:               false,
		NERecognition:         false,
	}

	if opts.PunctuationFile != "" {
		this.punct = NewPunts(opts.PunctuationFile)
		this.PunctuationDetection = true
	}

	if opts.DictionaryFile != "" {
		this.dic = NewDictionary(opts.Lang, opts.DictionaryFile, opts.AffixFile, opts.CompoundFile, opts.InverseDict, opts.RetokContractions)
		this.DictionarySearch = true
	}

	if opts.LocutionsFile != "" {
		this.loc = NewLocutions(opts.LocutionsFile)
		this.MultiwordsDetection = true
	}

	if opts.NPdataFile != "" {
		this.npm = NewNER(opts.NPdataFile)
		this.NERecognition = true
	}

	if opts.ProbabilityFile != "" {
		this.prob = NewProbability(opts.ProbabilityFile, opts.ProbabilityThreshold)
		this.ProbabilityAssignment = true
	}

	return &this
}

func (this *Maco) Analyze(s *Sentence) {
	if this.PunctuationDetection && this.punct != nil {
		this.punct.analyze(s)
	}

	if this.DictionarySearch && this.dic != nil {
		this.dic.Analyze(s)
	}

	if this.MultiwordsDetection && this.loc != nil {
		this.loc.analyze(s)
	}

	if this.NERecognition && this.npm != nil {
		this.npm.who.analyze(s)
	}

	if this.ProbabilityAssignment && this.prob != nil {
		this.prob.Analyze(s)
	}
}
