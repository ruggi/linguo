package linguo

import (
	"container/list"
	"strconv"
	"strings"
)

type ChartParser struct {
	gram *Grammar
}

func NewChartParser(gram *Grammar) *ChartParser {
	return &ChartParser{
		gram: gram,
	}
}

func (c *ChartParser) getStartSymbol() string {
	return c.gram.getStartSymbol()
}

func (c *ChartParser) Analyze(s *Sentence) {
	for k := 0; k < s.numKBest(); k++ {
		ch := NewChart(c.gram)
		ch.loadSentence(s, k)
		ch.parse()
		tr := ch.getTree(ch.getSize()-1, 0, "")
		for w, n := s.Front(), tr.begin(); w != nil && n.pnode != tr.end().pnode; n = n.PlusPlus() {
			if n.pnode.numChildren() == 0 {
				n.pnode.info.(*Node).setWord(w.Value.(*Word))
				n.pnode.info.(*Node).setLabel(w.Value.(*Word).getTag(k))
				w = w.Next()
			}
		}

		tr.buildNodeIndex(s.sentID)
		s.setParseTree(tr, k)
	}
}

type Edge struct {
	*Rule
	matched  *list.List
	backpath *list.List
}

func NewEdge() *Edge {
	c := Edge{}
	c.Rule = NewRule()
	return &c
}

func NewEdgeFromString(s string, ls *list.List, pgov int) *Edge {
	c := Edge{
		matched:  list.New(),
		backpath: list.New(),
	}
	c.Rule = NewRuleFromString(s, ls, pgov)
	return &c
}

func NewEdgeFromEdge(edge *Edge) *Edge {
	matched := list.New()
	matched.PushBackList(edge.matched)
	backpath := list.New()
	backpath.PushBackList(edge.backpath)
	c := Edge{
		matched:  matched,
		backpath: backpath,
		Rule:     NewRuleFromRule(edge.Rule),
	}
	return &c
}

func (c *Edge) String() string {
	output := ""
	output += "Head:" + c.getHead() + "\n"
	output += "Governor:" + strconv.Itoa(c.getGovernor()) + "\n"
	output += "Right:\n"
	right := c.getRight()
	for r := right.Front(); r != nil; r = r.Next() {
		output += "    - " + r.Value.(string) + "\n"
	}

	output += "Matched:\n"
	matched := c.getMatched()
	for m := matched.Front(); m != nil; m = m.Next() {
		output += "    - " + m.Value.(string) + "\n"
	}

	output += "Backpath:\n"
	backpath := c.getBackpath()
	for b := backpath.Front(); b != nil; b = b.Next() {
		output += "    - (" + strconv.Itoa(b.Value.(Pair).first.(int)) + "," + strconv.Itoa(b.Value.(Pair).second.(int)) + ")\n"
	}

	return output
}

func (c *Edge) getMatched() *list.List {
	return c.matched
}

func (c *Edge) getBackpath() *list.List {
	return c.backpath
}

func (c *Edge) active() bool {
	return c.right.Len() > 0
}

func (c *Edge) shift(a, b int) {
	c.matched.PushBack(c.right.Front().Value.(string))
	c.right.Remove(c.right.Front())
	c.backpath.PushBack(Pair{a, b})
}

type Chart struct {
	table []*list.List
	size  int
	gram  *Grammar
}

func NewChart(gram *Grammar) *Chart {
	return &Chart{
		gram: gram,
	}
}

func (c *Chart) getSize() int { return c.size }

func (c *Chart) loadSentence(s *Sentence, k int) {
	var j, n int
	var w *list.Element

	n = s.Len()
	c.table = make([]*list.List, (1+n)*n/2)

	j = 0
	l := list.New()
	for w = s.Front(); w != nil; w = w.Next() {
		ce := list.New()
		for a := w.Value.(*Word).selectedBegin(k).Element; a != w.Value.(*Word).selectedBegin(k).Next(); a = a.Next() {
			e := NewEdgeFromString(a.Value.(*Analysis).getTag(), l, 0)
			ce.PushBack(e)
			c.findAllRules(e, ce, 0, j)

			e1 := NewEdgeFromString(a.Value.(*Analysis).getTag()+"("+w.Value.(*Word).getLCForm()+")", l, 0)
			ce.PushBack(e1)
			c.findAllRules(e1, ce, 0, j)

			e2 := NewEdgeFromString(a.Value.(*Analysis).getTag()+"<"+a.Value.(*Analysis).getLemma()+">", l, 0)
			ce.PushBack(e2)
			c.findAllRules(e2, ce, 0, j)
		}

		c.table[c.index(0, j)] = ce
		j++
	}
	c.size = j

	for k := 0; k < j; k++ {
		ce := c.table[c.index(0, k)]
		for e := ce.Front(); e != nil; e = e.Next() {
			//TODO
		}
	}
}

func (c *Chart) parse() {
	var k, i, a int
	for k = 1; k < c.size; k++ {
		for i = 0; i < c.size-k; i++ {
			ce := list.New()
			for a = 0; a < k; a++ {
				for ed := c.table[c.index(a, i)].Front(); ed != nil; ed = ed.Next() {
					e := NewEdgeFromEdge(ed.Value.(*Edge))
					if e.active() {
						ls := e.getRight()
						if c.canExtend(ls.Front().Value.(string), k-a-1, i+a+1) {
							e.shift(k-a-1, i+a+1)
							ce.PushBack(e)
							if !e.active() {
								c.findAllRules(e, ce, k, i)
							}
						}
					}
				}
			}
			c.table[c.index(k, i)] = ce
		}
	}

	c.ndump()

	best := NewEdge()
	gotroot := false

	for ed := c.table[c.index(c.size-1, 0)].Front(); ed != nil; ed = ed.Next() {
		if !ed.Value.(*Edge).active() && !c.gram.isNoTop(ed.Value.(*Edge).getHead()) && c.betterEdge(ed.Value.(*Edge), best) {
			gotroot = true
			best = ed.Value.(*Edge)
		}
	}

	if !gotroot {
		lp := c.cover(c.size-1, 0)
		ls := list.New()
		for p := lp.Front(); p != nil; p = p.Next() {
			best = NewEdge()
			for ed := c.table[c.index(p.Value.(Pair).first.(int), p.Value.(Pair).second.(int))].Front(); ed != nil; ed = ed.Next() {
				if !ed.Value.(*Edge).active() && c.betterEdge(ed.Value.(*Edge), best) {
					best = ed.Value.(*Edge)
				}
			}

			ls.PushBack(best.getHead())
		}

		e1 := NewEdgeFromString(c.gram.getStartSymbol(), ls, GRAMMAR_NOGOV)

		for p := lp.Front(); p != nil; p = p.Next() {
			e1.shift(p.Value.(Pair).first.(int), p.Value.(Pair).second.(int))
		}

		c.table[c.index(c.size-1, 0)].PushBack(e1)
	}

}

func (c *Chart) cover(a, b int) *list.List {
	x := 0
	y := 0
	var i, j int
	var f bool
	var ed *list.Element
	var lp, lr *list.List

	if a < 0 || b < 0 || a+b >= c.size {
		return list.New()
	}

	f = false

	best := NewEdge()

	for i = a; !f && i >= 0; i-- {
		for j = b; j < b+(a-i)+1; j++ {
			for ed = c.table[c.index(i, j)].Front(); ed != nil; ed = ed.Next() {
				if !ed.Value.(*Edge).active() && c.betterEdge(ed.Value.(*Edge), best) {
					x = i
					y = j
					best = ed.Value.(*Edge)
					f = true
				}
			}
		}
	}

	lp = c.cover(y-b-1, b)
	lr = c.cover((a+b)-(x+y+1), x+y+1)

	lp.PushBack(Pair{x, y})
	for tlr := lr.Front(); tlr != nil; tlr = tlr.Next() {
		lp.PushBack(tlr.Value.(Pair))
	}

	return lp
}

func (c *Chart) betterEdge(e1 *Edge, e2 *Edge) bool {
	h1 := e1.getHead()
	h2 := e2.getHead()
	start := c.gram.getStartSymbol()

	if h1 == start && h2 != start {
		return true
	}
	if h1 != start && h2 == start {
		return false
	}

	if c.gram.isTerminal(h1) && c.gram.isTerminal(h2) {
		return c.gram.getSpecificity(h1) < c.gram.getSpecificity(h2)
	}

	if !c.gram.isTerminal(h1) && !c.gram.isTerminal(h2) {
		if c.gram.getPriority(h1) < c.gram.getPriority(h2) {
			return true
		}
		if c.gram.getPriority(h1) > c.gram.getPriority(h2) {
			return false
		}

		return e1.getMatched().Len() > e2.getMatched().Len()
	}

	return !c.gram.isTerminal(h1) && c.gram.isTerminal(h2)
}

func (c *Chart) index(i, j int) int {
	return j + i*(c.size+1) - (i+1)*i/2
}

func (c *Chart) canExtend(hd string, i int, j int) bool {
	b := false
	for ed := c.table[c.index(i, j)].Front(); !b && ed != nil; ed = ed.Next() {
		b = (!ed.Value.(*Edge).active() && c.checkMatch(hd, ed.Value.(*Edge).getHead()))
	}

	return b
}

func (c *Chart) checkMatch(searched string, found string) bool {
	var s, t, m string
	if searched == found {
		return true
	}

	n := strings.Index(searched, "*")
	if n == -1 {
		return false
	}

	if strings.Index(found, searched[0:n]) != 0 {
		return false
	}

	n = MultiIndex(found, "(<")
	if n == -1 {
		s = found
		t = ""
	} else {
		s = found[0:n]
		t = found[n:]
	}

	n = MultiIndex(searched, "(<")
	if n == -1 {
		m = ""
	} else {
		m = searched[n:]
	}

	file := strings.Index(m, "\"") > -1

	if !file {
		return (s+m == found)
	} else {
		return c.gram.inFileMap(t, m)
	}
}

func (c *Chart) findAllRules(e *Edge, ce *list.List, k int, i int) {
	d := list.New()
	if c.gram.isTerminal(e.getHead()) {
		lr := c.gram.getRulesRightWildcard(e.getHead()[0:1])
		for r := lr.Front(); r != nil; r = r.Next() {
			newR := NewRuleFromRule(r.Value.(*Rule))
			if c.checkMatch(newR.getRight().Front().Value.(string), e.getHead()) {
				ed := NewEdgeFromString(newR.getHead(), newR.getRight(), newR.getGovernor())
				ed.shift(k, i)
				ce.PushBack(ed)
				if !ed.active() {
					d.PushBack(ed.getHead())
				}
			}
		}
	}

	d.PushBack(e.getHead())
	for d.Len() > 0 {
		lr := c.gram.getRulesRight(d.Front().Value.(string))
		for r := lr.Front(); r != nil; r = r.Next() {
			newR := NewRuleFromRule(r.Value.(*Rule))
			ed := NewEdgeFromString(newR.getHead(), newR.getRight(), newR.getGovernor())
			ed.shift(k, i)
			ce.PushBack(ed)
			if !ed.active() {
				d.PushBack(ed.getHead())
			}
		}
		d.Remove(d.Front())
	}
}

func (c *Chart) ndump() {
	for a := 0; a < c.size; a++ {
		for i := 0; i < c.size-a; i++ {
			if c.table[c.index(a, i)].Len() > 0 {
				for ed := c.table[c.index(a, i)].Front(); ed != nil; ed = ed.Next() {
					// TODO
				}
			}
		}
	}
}

func (c *Chart) dump() {
	for a := 0; a < c.size; a++ {
		for i := 0; i < c.size-a; i++ {
			if c.table[c.index(a, i)].Len() > 0 {
				out := "Cell (" + strconv.Itoa(a) + "," + strconv.Itoa(i) + ")\n"
				for ed := c.table[c.index(a, i)].Front(); ed != nil; ed = ed.Next() {
					out += "    " + ed.Value.(*Edge).getHead() + " ==>"
					ls := ed.Value.(*Edge).getMatched()
					for s := ls.Front(); s != nil; s = s.Next() {
						out += " " + s.Value.(string)
					}
					out += " ."
					ls = ed.Value.(*Edge).getRight()
					for s := ls.Front(); s != nil; s = s.Next() {
						out += " " + s.Value.(string)
					}
					lp := ed.Value.(*Edge).getBackpath()
					out += "   Backpath:"
					for p := lp.Front(); p != nil; p = p.Next() {
						out += "(" + strconv.Itoa(p.Value.(Pair).first.(int)) + "," + strconv.Itoa(p.Value.(Pair).second.(int)) + ")"
					}
					out += "\n"
				}
			}
		}
	}
}

func (c *Chart) getTree(x int, y int, lab string) *ParseTree {
	label := lab
	if label == "" {
		best := NewEdge()
		for ed := c.table[c.index(x, y)].Front(); ed != nil; ed = ed.Next() {
			if !ed.Value.(*Edge).active() && !c.gram.isHidden(ed.Value.(*Edge).getHead()) && c.betterEdge(ed.Value.(*Edge), best) {
				label = ed.Value.(*Edge).getHead()
				best = ed.Value.(*Edge)
			}
		}
	}
	node := NewNodeFromLabel(label)
	tr := NewOneNodeParseTree(node)

	if label == c.gram.getStartSymbol() || !c.gram.isTerminal(label) {
		best := NewEdge()
		for ed := c.table[c.index(x, y)].Front(); ed != nil; ed = ed.Next() {
			if !ed.Value.(*Edge).active() && label == ed.Value.(*Edge).getHead() && c.betterEdge(ed.Value.(*Edge), best) {
				best = ed.Value.(*Edge)
			}
		}

		r := best.getMatched()
		bp := best.getBackpath()
		g := best.getGovernor()

		headset := false

		for ch, s, p := 0, r.Front(), bp.Front(); s != nil && p != nil; ch, s, p = ch+1, s.Next(), p.Next() {
			child := c.getTree(p.Value.(Pair).first.(int), p.Value.(Pair).second.(int), s.Value.(string))
			childLabel := child.begin().pnode.info.(*Node).getLabel()
			if c.gram.isHidden(childLabel) || c.gram.isOnlyTop(childLabel) || (c.gram.isFlat(childLabel) && label == childLabel) {
				for x := child.siblingBegin(); x.pnode != child.siblingEnd().pnode; x = x.siblingPlusPlus() {
					if ch == g {
						headset = true
					} else {
						x.pnode.info.(*Node).setHead(false)
					}
					tr.appendChild(x.pnode)
				}
			} else {
				if ch == g {
					child.begin().pnode.info.(*Node).setHead(true)
					headset = true
				}
				tr.appendChild(child)
			}
		}

		if !headset && label != c.gram.getStartSymbol() {
			//TODO
		}

	}

	return tr
}
