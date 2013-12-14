package xmlpull

import (
	"fmt"
)

var _ = fmt.Print

// A simplistic parse of
// [28]  doctypedecl ::= '<!DOCTYPE' S Name (S ExternalID)? S? ('['
//                      (markupdecl | DeclSep)* ']' S?)? '>'
//
func (p *Parser) parseDocTypeDecl() (err error) {

	var runes []rune

	// Enter having seen <!D

	err = p.ExpectStr("OCTYPE")
	if err == nil {
		var ch rune
		runes = []rune("<!DOCTYPE")

		bracketDepth := 0
		ch, err = p.NextCh()
		for err == nil {
			runes = append(runes, ch)
			if ch == '[' {
				bracketDepth++
			} else if ch == ']' {
				bracketDepth--
			} else if ch == '>' && bracketDepth == 0 {
				break
			}
			ch, err = p.NextCh()
		}
	}
	if err == nil {
		p.docTypeDecl = string(runes)
	}
	return
}
