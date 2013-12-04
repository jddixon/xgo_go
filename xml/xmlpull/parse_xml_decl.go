package xmlpull

import (
	e "errors"
	"fmt"
)

// [25] Eq ::= S? '=' S?
func (p *Parser) expectEq() (err error) {
	lx := &p.LexInput
	lx.SkipS()
	ch, err := lx.NextCh()
	if err == nil && ch != '=' {
		msg := fmt.Sprintf(
			"parseXmlDecl.expectEq: expected = , found %c\n", ch)
		err = e.New(msg)
	}
	if err == nil {
		lx.SkipS() // closes Eq
	}
	return
}
func (p *Parser) expectQuoteCh() (quoteCh rune, err error) {
	lx := &p.LexInput
	quoteCh, err = lx.NextCh()
	if err == nil && quoteCh != '\'' && quoteCh != '"' {
		msg := fmt.Sprintf("expected quotation mark, found '%c'", quoteCh)
		err = e.New(msg)
	}
	return
}

// [81] EncName ::= [A-Za-z] ([A-Za-z0-9._] | '-')*
func (p *Parser) getEncodingStartCh() (ch rune, err error) {
	lx := &p.LexInput
	ch, err = lx.NextCh()
	if err == nil {
		if !('a' <= ch && ch <= 'z') && !('A' <= ch && ch <= 'Z') {
			msg := fmt.Sprintf("cannot start encoding name: '%c'\n", ch)
			err = e.New(msg)
		}
	}
	return
}

func (p *Parser) getEncodingNameCh(quoteCh rune) (ch rune, err error) {
	lx := &p.LexInput
	ch, err = lx.NextCh()
	if err == nil && ch != quoteCh {
		if !('a' <= ch && ch <= 'z') && !('A' <= ch && ch <= 'Z') &&
			!('0' <= ch && ch <= '9') && (ch != '.') && (ch != '_') &&
			(ch != '-') {
			msg := fmt.Sprintf("illegal character in encoding name: '%c'\n", ch)
			err = e.New(msg)
		}
	}
	return
}

// Function called after encountering <?xmlS at the beginning of the input,
// where S as usual represents a space.
//
func (p *Parser) parseXmlDecl() (xmlDeclVersion, xmlDeclEncoding string,
	xmlDeclStandalone bool, err error) {

	var (
		found       bool
		ch, quoteCh rune
	)
	lx := &p.LexInput

	// [23] XMLDecl ::= '<?xml' VersionInfo EncodingDecl? SDDecl? S? '?>'
	// [24] VersionInfo ::= S 'version' Eq ("'" VersionNum "'" | '"' VersionNum '"')
	// We are on first S past <?xml
	lx.SkipS()
	err = lx.ExpectStr("version")

	if err == nil {
		err = p.expectEq()
		if err == nil {
			quoteCh, err = p.expectQuoteCh()
		}
	}

	// [26] VersionNum ::= ([a-zA-Z0-9_.:] | '-')+
	if err == nil {
		var vRunes []rune
		ch, err = lx.NextCh()
		for err == nil && ch != quoteCh {
			if ('a' <= ch && ch <= 'z') ||
				('A' <= ch && ch <= 'Z') ||
				('0' <= ch && ch <= '9') ||
				(ch == '_') || (ch == '.') || (ch == ':') || (ch == '-') {
				vRunes = append(vRunes, ch)
			} else {
				msg := fmt.Sprintf(
					"Not an acceptable version character: '%c'", ch)
				err = e.New(msg)
				break
			}
			ch, err = lx.NextCh()
		}
		if err == nil {
			// ch is guaranteed to be quoteCh
			xmlDeclVersion = string(vRunes)
			if xmlDeclVersion != "1.0" {
				err = OnlyVersion1_0
			}
		}
	}

	// [80] EncodingDecl ::= S 'encoding' Eq ('"' EncName '"' | "'" EncName "'" )
	if err == nil {
		lx.SkipS()
		found, err = lx.AcceptStr("encoding")
		if err == nil {
			if found {
				var eRunes []rune
				err = p.expectEq()
				if err == nil {
					var eStartCh, eNameCh rune
					quoteCh, err = p.expectQuoteCh()
					if err == nil {
						eStartCh, err = p.getEncodingStartCh()
						if err == nil {
							eRunes = append(eRunes, eStartCh)
							for err == nil && eNameCh != quoteCh {
								eNameCh, err = p.getEncodingNameCh(quoteCh)
								if err == nil {
									if eNameCh == quoteCh {
										break
									}
									eRunes = append(eRunes, eNameCh)
								}
							}
						}
					}
				}
				if err == nil {
					xmlDeclEncoding = string(eRunes)
				}
			}
		}
	}
	// [32] SDDecl ::= S 'standalone' Eq (("'" ('yes' | 'no') "'") | ('"' ('yes' | 'no') '"'))
	if err == nil {
		lx.SkipS()
		found, err = lx.AcceptStr("standalone")
		if err == nil && found {
			var foundYes, foundNo bool
			err = p.expectEq()
			if err == nil {
				quoteCh, err = p.expectQuoteCh()
				if err == nil {
					foundYes, err = p.AcceptStr("yes")
				}
				if err == nil && !foundYes {
					foundNo, err = p.AcceptStr("no")
				}
				if err == nil {
					if foundYes {
						xmlDeclStandalone = true
					} else if foundNo {
						xmlDeclStandalone = false
					} else {
						err = MustBeYesOrNo
					}
				}
				ch, err = lx.NextCh()
				if err == nil && ch != quoteCh {
					err = MissingDeclClosingQuote
				}
			}
		}
	}
	// expecting ?>
	if err == nil {
		lx.SkipS()
		err = lx.ExpectStr("?>")
	}
	return
}