package md

// xgo/md/header.go

import (
	"fmt"
	"io"
	"strings"
)

type Header struct {
	runes []rune
}

func NewHeader(n int, title []rune) (h *Header, err error) {
	if n < 1 || 6 < n {
		err = HeaderNOutOfRange
	} else if len(title) == 0 {
		err = EmptyHeaderTitle
	} else {
		text := fmt.Sprintf("<h%d>%s</h%d>", n, string(title), n)
		h = &Header{runes: []rune(text)}
	}
	return
}

func (h *Header) Get() []rune {
	return h.runes
}

// Collect headers preceded by 1-6 hash signs ('#') and optionally
// terminated by one hash sign.
//
func (p *Parser) collectHeader() (collected bool, hashCount int, err error) {

	fmt.Printf("Entering collectHeader()\n")

	lx := p.lexer

	var (
		atEOF bool
		runes []rune
	)

	// count leading hashes -----------------------------------------
	hashCount = 1 // we enter having seen one '#'
	ch, err := lx.NextCh()
	for err == nil && !atEOF {
		if ch != '#' {
			break
		}
		hashCount++
		ch, err = lx.NextCh()
		if err == io.EOF {
			atEOF = true
			err = nil
		}
	}

	// collect the title --------------------------------------------
	for err == nil {
		if ch == '\r' || ch == '\n' {
			break
		}
		runes = append(runes, ch)
		if atEOF {
			break
		}
		ch, err = lx.NextCh()
		if err == io.EOF {
			atEOF = true
			err = nil
		}
	}
	// if we have a title -------------------------------------------
	if err == nil && len(runes) > 0 {

		// XXX UNDER AS-YET-UNDERSTOOD CIRCUMSTANCES we get a trailing
		// null byte
		if runes[len(runes)-1] == rune(0) {
			runes = runes[:len(runes)-1]
		}
		// drop any trailing hash sign --------------------
		if runes[len(runes)-1] == '#' {
			runes = runes[:len(runes)-1]
		}
		title := strings.TrimSpace(string(runes))
		runes = []rune(title)

		// create the object ------------------------------
		var h *Header
		h, _ = NewHeader(hashCount, runes)
		p.bits = append(p.bits, h)
		collected = true
	}
	return
}
