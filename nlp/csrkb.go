package nlp

import (
	"container/list"
	"io/ioutil"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const VERTEX_NOT_FOUND = -1
const RE_WNP = "^[NARV]"

const (
	UKB_RELATION_FILE = 1 + iota
	UKB_REX_WNPOS
	UKB_PR_PARAMS
)

type CSRKB struct {
	maxIterations int
	threshold     float64
	damping       float64
	vertexIndex   map[string]int
	outCoef       []float64
	firstEdge     []int
	numEdges      []int
	edges         []int
	numVertices   int
}

type IntPair struct {
	first  int
	second int
}

type IntPairsArray []IntPair

func (a IntPairsArray) Less(i, j int) bool {
	return a[i].first < a[j].first || (a[i].first == a[j].first && a[i].second < a[j].second)
}

func (a IntPairsArray) Len() int { return len(a) }

func (a IntPairsArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func List2IntPairsArray(ls *list.List) []IntPair {
	out := make([]IntPair, ls.Len())
	for i, l := 0, ls.Front(); i < ls.Len() && l != nil; i, l = i+1, l.Next() {
		out[i] = l.Value.(IntPair)
	}
	return out
}

func IntPairsArray2List(a IntPairsArray) *list.List {
	out := list.New()
	for _, i := range a {
		out.PushBack(i)
	}
	return out
}

func NewCSRKB(kbFile string, nit int, thr float64, damp float64) *CSRKB {
	ukb := CSRKB{
		vertexIndex:   make(map[string]int),
		maxIterations: nit,
		threshold:     thr,
		damping:       damp,
	}

	var syn1, syn2 string
	var pos1, pos2 int
	rels := list.New()
	ukb.numVertices = 0

	fileString, err := ioutil.ReadFile(kbFile)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(fileString), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		items := Split(line, " ")

		syn1 = items[0]
		syn2 = items[1]

		pos1 = ukb.addVertex(syn1)
		if syn2 != "-" {
			pos2 = ukb.addVertex(syn2)

			rels.PushBack(IntPair{pos1, pos2})
			rels.PushBack(IntPair{pos2, pos1})
		}

	}

	ukb.fillCSRTables(ukb.numVertices, rels)
	return &ukb
}

func (ukb *CSRKB) fillCSRTables(nv int, rels *list.List) {
	tmpA := List2IntPairsArray(rels)
	sort.Sort(IntPairsArray(tmpA))
	rels = IntPairsArray2List(tmpA)

	ukb.edges = make([]int, rels.Len())
	ukb.firstEdge = make([]int, nv)
	ukb.numEdges = make([]int, nv)
	ukb.outCoef = make([]float64, nv)

	n := 0
	r := 0

	p := rels.Front()
	for p != nil && n < nv {
		ukb.firstEdge[n] = r
		for p != nil && p.Value.(IntPair).first == n {
			ukb.edges[r] = p.Value.(IntPair).second
			r++
			p = p.Next()
		}
		ukb.numEdges[n] = r - ukb.firstEdge[n]
		ukb.outCoef[n] = 1 / float64(ukb.numEdges[n])
		n++
	}
}

func (ukb *CSRKB) addVertex(s string) int {
	ukb.vertexIndex[s] = ukb.numVertices
	ukb.numVertices++
	return ukb.vertexIndex[s]
}

func (ukb *CSRKB) size() int { return ukb.numVertices }

func (ukb *CSRKB) getVertex(s string) int {
	out := ukb.vertexIndex[s]
	if out > 0 {
		return out
	} else {
		return VERTEX_NOT_FOUND
	}
}

func (ukb *CSRKB) pageRank(pv []float64) {
	var ranks [2][]float64
	CURRENT := 0
	NEXT := 1
	initVal := 1.0 / float64(ukb.numVertices)

	ranks[CURRENT] = ArrayFloatInit(ukb.numVertices, initVal)
	ranks[NEXT] = ArrayFloatInit(ukb.numVertices, 0.0)

	nit := 0
	change := ukb.threshold
	for nit < ukb.maxIterations && change >= ukb.threshold {
		change = 0

		for v := 0; v < ukb.numVertices; v++ {
			rank := 0.0
			for e := ukb.firstEdge[v]; e < ukb.firstEdge[v]+ukb.numEdges[v]; e++ {
				u := ukb.edges[e]
				rank += ranks[CURRENT][u] * ukb.outCoef[u]
			}

			ranks[NEXT][v] = rank*ukb.damping + pv[v]*(1-ukb.damping)
			change += math.Abs(ranks[NEXT][v] - ranks[CURRENT][v])
		}

		tmp := NEXT
		NEXT = CURRENT
		CURRENT = tmp
		nit++
	}

	ArrayFloatSwap(pv, ranks[CURRENT])
}

type UKB struct {
	wn       *CSRKB
	RE_wnpos *regexp.Regexp
}

func NewUKB(wsdFile string) *UKB {
	ukb := UKB{
		RE_wnpos: regexp.MustCompile(RE_WNP),
	}

	path := wsdFile[0:strings.LastIndex(wsdFile, "/")]
	var relFile string

	var thr float64 = 0.000001
	var nit int = 30
	var damp float64 = 0.85

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("RelationFile", UKB_RELATION_FILE)
	cfg.AddSection("RE_Wordnet_PoS", UKB_REX_WNPOS)
	cfg.AddSection("PageRankParameters", UKB_PR_PARAMS)

	if !cfg.Open(wsdFile) {
		panic("Error loading file " + wsdFile)
	}

	line := ""

	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case UKB_RELATION_FILE:
			{
				fname := items[0]
				if strings.HasPrefix(fname, "../") {
					wsdFile = strings.Replace(wsdFile, "./", "", -1)
					path = wsdFile[0:strings.Index(wsdFile, "/")]
					relFile = path + "/" + strings.Replace(fname, "../", "", -1)
				} else {
					relFile = path + "/" + strings.Replace(fname, "./", "", -1)
				}
				break
			}
		case UKB_REX_WNPOS:
			{
				ukb.RE_wnpos = regexp.MustCompile(line)
				break
			}
		case UKB_PR_PARAMS:
			{
				key := items[0]
				if key == "Threshold" {
					thr, _ = strconv.ParseFloat(items[1], 64)
				} else if key == "MaxIterations" {
					nit, _ = strconv.Atoi(items[1])
				} else if key == "Damping" {
					damp, _ = strconv.ParseFloat(items[1], 64)
				}
				break
			}
		default:
			break
		}
	}

	if relFile == "" {
		panic("No relation file provided in UKB configuration file " + wsdFile)
	}

	ukb.wn = NewCSRKB(relFile, nit, thr, damp)

	return &ukb
}

func (ukb *UKB) initSynsetVector(sentences []*Sentence, pv []float64) {
	nw := 0
	uniq := make(map[string]*Word)
	for _, s := range sentences {
		for w := s.Front(); w != nil; w = w.Next() {
			if ukb.RE_wnpos.MatchString(w.Value.(*Word).getTag(0)) {
				key := w.Value.(*Word).getLCForm() + "#" + strings.ToLower(w.Value.(*Word).getTag(0))[0:1]
				if uniq[key] == nil {
					nw++
					uniq[key] = w.Value.(*Word)
				}
			}
		}
	}

	for _, u := range uniq {
		lsen := u.getSenses(0)
		nsyn := lsen.Len()
		for s := lsen.Front(); s != nil; s = s.Next() {
			syn := ukb.wn.getVertex(s.Value.(FloatPair).first)
			if syn != VERTEX_NOT_FOUND {
				pv[syn] += (1.0 / float64(nw)) * (1.0 / float64(nsyn))
			}
		}
	}
}

type FloatPair struct {
	first  string
	second float64
}

type FloatPairsArray []FloatPair

func (a FloatPairsArray) Less(i, j int) bool {
	return a[i].first < a[j].first || (a[i].first == a[j].first && a[i].second < a[j].second)
}

func (a FloatPairsArray) Len() int { return len(a) }

func (a FloatPairsArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func List2FloatPairsArray(ls *list.List) []FloatPair {
	out := make([]FloatPair, ls.Len())
	for i, l := 0, ls.Front(); i < ls.Len() && l != nil; i, l = i+1, l.Next() {
		out[i] = l.Value.(FloatPair)
	}
	return out
}

func FloatPairsArray2List(a FloatPairsArray) *list.List {
	out := list.New()
	for _, i := range a {
		out.PushBack(i)
	}
	return out
}

func (ukb *UKB) extractRanksToSentences(sentences []*Sentence, pv []float64) {
	for _, s := range sentences {
		for w := s.Front(); w != nil; w = w.Next() {
			lsen := w.Value.(*Word).getSenses(0)
			for p := lsen.Front(); p != nil; p = p.Next() {
				syn := ukb.wn.getVertex(p.Value.(FloatPair).first)
				if syn != VERTEX_NOT_FOUND {
					lsen.InsertAfter(FloatPair{p.Value.(FloatPair).first, pv[syn]}, p)
					lsen.Remove(p)
				}
			}

			a := List2FloatPairsArray(lsen)
			sort.Sort(FloatPairsArray(a))
			lsen = FloatPairsArray2List(a)
			w.Value.(*Word).setSenses(lsen, 0)
		}
	}
}

func (ukb *UKB) Analyze(sentences []*Sentence) {
	pv := make([]float64, ukb.wn.size())
	ukb.initSynsetVector(sentences, pv)
	ukb.wn.pageRank(pv)
	ukb.extractRanksToSentences(sentences, pv)
}
