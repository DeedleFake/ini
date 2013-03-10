package ini_test

import (
	"github.com/DeedleFake/ini"
	"io"
	"strings"
	"testing"
)

const test = `[Test 1]
This=is
# Comment
a=test.`

func TestNext(t *testing.T) {
	p := ini.NewParser(strings.NewReader(test))

	for {
		tok, err := p.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			t.Fatalf("p.Next(): %v", err)
		}

		switch tok := tok.(type) {
		case *ini.SectionToken:
			t.Logf("Section: %v", tok.Name)
		case *ini.SettingToken:
			t.Logf("Setting: %v: '%v'", tok.Left, tok.Right)
		case *ini.CommentToken:
			t.Logf("Comment: '%v'", tok.Comment)
		default:
			panic(tok)
		}
	}
}
