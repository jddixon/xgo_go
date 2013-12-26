package md

// xgo/md/old_parser.go

import (
	"fmt"
	gl "github.com/jddixon/xgo/lex"
	"io"
)

var _ = fmt.Print

type State int

const (
	START State = iota
	NONSEP_COLL
	MAYBE_COLL
	SEP_COLL
)

var (
	SEP_CHAR    = []rune{CR, LF}
	FOUR_SPACES = []rune("    ")

	OPEN_EM      = []rune("<em>")
	CLOSE_EM     = []rune("</em>")
	H_RULE       = []rune("<hr />")
	OPEN_STRONG  = []rune("<strong>")
	CLOSE_STRONG = []rune("</strong>")
)

type OldParser struct {
	lexer *gl.LexInput
	state State

	lineSep          *LineSep
	crCount, lfCount int
	maybes           []rune
	seps             []rune

	curText []rune
	curPara *Para
	downers []MarkdownI

	// for handling emphasis
	emphChar    rune
	emphDoubled bool

	// our little dictionary
	dict map[string]*Definition
}

func NewOldParser(reader io.Reader) (p *OldParser, err error) {
	lx, err := gl.NewNewLexInput(reader)
	if err == nil {
		p = &OldParser{
			lexer: lx,
			state: START,
		}
	}
	return
}

func (p *OldParser) Parse() ([]MarkdownI, error) {
	var (
		ch           rune
		err          error
		leadingSpace bool

		// header handling
		collected bool
		hashCount int
	)
	lx := p.lexer
	ch, err = lx.NextCh()
	for err == nil {
		if err != nil {
			break
		}
		if p.state == START {
			if len(p.curText) == 0 {
				if ch == SPACE { // leading tab?
					leadingSpace = true
					goto NEXT
				} else if ch == '#' && !leadingSpace {
					collected, hashCount, err = p.collectHeader()
					if collected {
						p.state = START // XXX
					}
					_ = hashCount
					goto NEXT
				}
			}
			if ch == BACKSLASH {
				var nextChar rune
				p.curText = append(p.curText, BACKSLASH)
				nextChar, err = lx.PeekCh()
				if escaped(nextChar) {
					ch, err = lx.NextCh()
					p.curText = append(p.curText, ch)
				}
				p.state = NONSEP_COLL
			} else if ch == '_' || ch == '*' {
				// scan ahead for matching ch; if we find it, we
				// wrap it properly and append it to p.curText and
				// return true.  Otherwise we push whatever has been
				// scanned back on input, append ch to p.curText,
				// and return false.
				p.oldParseEmph(ch)
			} else if ch == CR || ch == LF {
				if ch == CR {
					p.crCount++
				} else {
					p.lfCount++
				}
				p.seps = append(p.seps, ch)
				p.state = SEP_COLL
			} else {
				p.curText = append(p.curText, ch)
				p.state = NONSEP_COLL
			}
		} else if p.state == SEP_COLL {
			if ch == CR || ch == LF {
				if ch == CR {
					p.crCount++
					if p.crCount < 3 {
						p.seps = append(p.seps, ch)
					}
				} else {
					p.lfCount++
					if p.lfCount < 3 {
						p.seps = append(p.seps, ch)
					}
				}
				// p.state unchanged
			} else {
				p.lineSep, err = NewLineSep(p.seps)
				p.seps = p.seps[:0]
				p.crCount = 0
				p.lfCount = 0
				p.downers = append(p.downers, p.lineSep)
				p.lineSep = nil
				p.curText = append(p.curText, ch)
				p.state = NONSEP_COLL
			}
		} else if p.state == NONSEP_COLL {
			if ch == BACKSLASH {
				var nextChar rune
				p.curText = append(p.curText, BACKSLASH)
				nextChar, err = lx.PeekCh()
				if escaped(nextChar) {
					ch, err = lx.NextCh()
					p.curText = append(p.curText, ch)
				}
			} else if len(p.curText) == 0 && ch == SPACE { // leading tab?
				// ignore
			} else if ch == '_' || ch == '*' {
				// scan ahead for matching ch; if we find it, we
				// wrap it properly and append it to p.curText and
				// return true.  Otherwise we push whatever has been
				// scanned back on input, append ch to p.curText,
				// and return false.
				p.oldParseEmph(ch)
			} else if ch == CR || ch == LF {
				if ch == CR {
					p.crCount = 1
				} else {
					p.lfCount = 1
				}
				p.maybes = append(p.maybes, ch)
				p.state = MAYBE_COLL
			} else {
				p.curText = append(p.curText, ch)
				// p.state unchanged
			}
		} else if p.state == MAYBE_COLL {
			if ch == SPACE || ch == TAB {
				// ignore it
			} else if ch == CR || ch == LF {
				p.maybes = append(p.maybes, ch)
				if ch == CR {
					p.crCount++
				} else {
					p.lfCount++
				}
				if p.crCount > 1 || p.lfCount > 1 {
					t := NewText(p.curText)
					para := NewPara(t)
					p.downers = append(p.downers, para)
					p.curText = p.curText[:0]
					p.seps = make([]rune, len(p.maybes))
					copy(p.seps, p.maybes)
					p.maybes = p.maybes[:0]
					p.state = SEP_COLL
				}
			} else {
				// If the last nonSep is a space (or tab?) we
				// make the nonSep a para, insert a p.lineSep,
				// and start a new para.
				lastChar := p.curText[len(p.curText)-1]
				if lastChar == SPACE || lastChar == TAB {
					if lastChar == TAB {
						p.curText = p.curText[:len(p.curText)-1]
						p.curText = append(p.curText, FOUR_SPACES...)
					}
					t := NewText(p.curText)
					para := NewPara(t)
					p.downers = append(p.downers, para)
					p.curText = p.curText[:0]
					p.lineSep, _ = NewLineSep(p.maybes)
					p.downers = append(p.downers, p.lineSep)
					p.maybes = p.maybes[:0]
				} else {
					p.curText = append(p.curText, p.maybes...)
					p.maybes = p.maybes[:0]
				}
				lx.PushBack(ch)
				p.state = NONSEP_COLL
			}
		}
	NEXT:
		ch, err = lx.NextCh()
	}
	if err == io.EOF {
		if p.state == SEP_COLL {
			p.seps = p.seps[:0] // just discard
		} else if p.state == NONSEP_COLL || p.state == MAYBE_COLL {
			lastChar := p.curText[len(p.curText)-1]
			if lastChar == TAB {
				p.curText = p.curText[:len(p.curText)-1]
				p.curText = append(p.curText, FOUR_SPACES...)
			}
			t := NewText(p.curText)
			para := NewPara(t)
			p.downers = append(p.downers, para)
			p.curText = p.curText[:0]
		}
		err = nil
	}
	if err == nil && len(p.curText) > 0 {
		t := NewText(p.curText)
		para := NewPara(t)
		p.downers = append(p.downers, para)
	}
	return p.downers, err
}