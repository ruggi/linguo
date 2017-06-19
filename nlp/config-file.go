package nlp

import (
	"io/ioutil"
	"strings"
)

type ConfigFile struct {
	lines                       []string
	sectionsOpen, sectionsClose map[string]int
	section                     int
	SECTION_NONE                int
	SECTION_UNKNOWN             int
	filename                    string
	sectionStart                bool
	commentPrefix               string
	lineNum                     int
	skipUnknownSections         bool
	unkName                     string
}

func NewConfigFile(skip bool, comment string) ConfigFile {
	if comment == "" {
		comment = "##"
	}
	return ConfigFile{
		SECTION_NONE:        -1,
		SECTION_UNKNOWN:     -2,
		skipUnknownSections: skip,
		commentPrefix:       comment,
		sectionsOpen:        make(map[string]int),
		sectionsClose:       make(map[string]int),
	}
}

func (c *ConfigFile) IsOpenSection(s string) bool {
	return len(s) > 2 && s[0] == '<' && s[1] != '/' && s[len(s)-1] == '>'
}

func (c *ConfigFile) IsCloseSection(s string) bool {
	return len(s) > 3 && s[0] == '<' && s[1] == '/' && s[len(s)-1] == '>'
}

func (c *ConfigFile) IsComment(s string) bool {
	return s == "" || strings.Contains(s, c.commentPrefix)
}

func (c *ConfigFile) AddSection(key string, section int) {
	c.sectionsOpen["<"+key+">"] = section
	c.sectionsClose["</"+key+">"] = section
}

func (c *ConfigFile) PrintSections() {
	for k, v := range c.sectionsOpen {
		println(k, v)
	}
}

func (c *ConfigFile) Open(filename string) bool {
	c.filename = filename
	c.section = c.SECTION_NONE
	fileString, err := ioutil.ReadFile(filename)
	if err != nil {
		return false
	}
	lines := strings.Split(string(fileString), "\n")
	c.lines = make([]string, len(lines))
	copy(c.lines, lines)
	c.lineNum = -1
	return true
}

func (c *ConfigFile) GetSection() int {
	return c.section
}

func (c *ConfigFile) GetLineNum() int {
	return c.lineNum
}

func (c *ConfigFile) AtSectionStart() bool {
	return c.sectionStart
}

func (c *ConfigFile) GetContentLine(line *string) bool {
	c.lineNum++
	c.sectionStart = false
	for k := c.lineNum; k < len(c.lines); k++ {
		*line = string(c.lines[k])
		if c.section == c.SECTION_NONE {
			if c.IsOpenSection(*line) {
				section := c.sectionsOpen[*line]
				if section == 0 {
					if c.skipUnknownSections {
						c.section = c.SECTION_UNKNOWN
						c.unkName = (*line)[1 : len(*line)-1]
					}
				} else {
					c.section = section
					c.sectionStart = true
				}
			}
		} else if c.section != c.SECTION_NONE {
			if c.IsCloseSection(*line) {
				s := c.sectionsClose[*line]
				if s == 0 {
					if c.skipUnknownSections && c.section == c.SECTION_UNKNOWN {
						if c.unkName == (*line)[2:len(*line)-1] {
							c.section = c.SECTION_NONE
						}
					}
				} else if s == c.section {
					c.section = c.SECTION_NONE
				}
			} else if !c.IsOpenSection(*line) && c.section != c.SECTION_UNKNOWN && !c.IsComment(*line) {
				return true
			}
		}
		c.lineNum++
	}

	return c.lineNum < len(c.lines)
}
