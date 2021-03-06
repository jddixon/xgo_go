package xmlpull

// xgo/xml/xmlpull/parseProlog.go

import (
	// e "errors"
	"fmt"
)

var _ = fmt.Print

// [1]  document ::= prolog element Misc*
// [22] prolog ::= XMLDecl? Misc* (doctypedecl Misc*)?
// [27] Misc ::= Comment | PI | S
// [28] doctypedecl ::= '<!DOCTYPE' S Name (S ExternalID)? S? ('['
//                      (markupdecl | DeclSep)* ']' S?)? '>'
// [39] element ::= EmptyElemTag | STag content ETag
// [40] STag ::= '<' Name (S Attribute)* S? '>'
//
func (p *Parser) parseProlog() (err error) {

	var ch rune

	// optional XMLDecl ---------------------------------------------
	// if XMLDecl is present, it begins with "<?xml" followed by an S
	var found bool
	found, err = p.AcceptStr("<?xml")
	if found {
		// DEBUG
		fmt.Println("parseProlog: '<?xml' SEEN")
		// END
		err = p.parseXmlDecl()
	}

	// zero or more Misc --------------------------------------------

	// XXX STUB

	// optional (doctypedecl followed by zero or more Misc ----------

	if err == nil {
		ch, err = p.NextCh()
		p.afterLT = false
		gotS := false
		for err == nil {
			// deal with Misc
			// deal with docdecl --> mark it!
			// else parseStartTag seen <[^/]
			if ch == '<' {
				if gotS && p.tokenizing {
					// posEnd = pos - 1
					p.afterLT = true
					p.curEvent = IGNORABLE_WHITESPACE
					break
				}
				ch, err = p.NextCh()
				if ch == '?' {
					// check if it is 'xml'
					// deal with XMLDecl
					var isPI bool
					isPI, err = p.parsePI() // skipping XMLDecl
					if err != nil {
						break
					}
					if isPI {
						if p.tokenizing {
							p.curEvent = PROCESSING_INSTRUCTION
							break
						}
					} else {
						// skip over - continue tokenizing
						//posStart = pos
						gotS = false
					}

				} else if ch == '!' {
					ch, err = p.NextCh()
					if ch == 'D' {
						if p.seenDocdecl {
							err = p.NewXmlPullError(
								"only one docdecl allowed in XML document")
							break
						}
						p.seenDocdecl = true
						p.parseDocTypeDecl()
						if p.tokenizing {
							p.curEvent = DOCDECL
							break
						}
					} else if ch == '-' {
						p.parseComment()
						if p.tokenizing {
							p.curEvent = COMMENT
							break
						}
					} else {
						err = p.NewXmlPullError(
							"unexpected markup <!" + printableChar(ch))
						break
					}
				} else if ch == '/' {
					err = p.NewXmlPullError(
						"start tag name cannot begin with '/'\n")
					break
				} else if isNameStartChar(ch) {
					p.rootElmSeen = true
					p.PushBack(ch)
					// DEBUG
					fmt.Printf("parseProlog: '%c' seen and pushed back, "+
						"invoking parseStartTag\n", ch)
					// END

					// XXX RETURNS PullEvent, error; these are lost!
					p.parseStartTag()
					break
				} else {
					msg := fmt.Sprintf(
						"expected start tag name, but cannot begin with '%c'\n",
						ch)
					err = p.NewXmlPullError(msg)
					break
				}
			} else if p.IsS(ch) {
				gotS = true
			} else {
				msg := fmt.Sprintf(
					"only whitespace allowed before start tag, not '%c'\n",
					ch)
				err = p.NewXmlPullError(msg)
				break
			}
			ch, err = p.NextCh()
		} // end for
	}

	return
}
