package ini_test

import (
	"fmt"
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

func ExampleNewParser() {
	const example = `# This is a comment.
[Section]
setting=val ; This is also a comment.
other=some\n\
\#thing`

	p := ini.NewParser(strings.NewReader(example))
	for {
		t, err := p.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			panic(err)
		}

		switch t := t.(type) {
		case *ini.SectionToken:
			fmt.Printf("%v:\n", t.Name)
		case *ini.SettingToken:
			fmt.Printf("\t%v: %v\n", t.Left, strings.TrimSpace(t.Right))
		case *ini.CommentToken:
			fmt.Println(t)
		}
	}
	// Output: # This is a comment.
	// Section:
	//	setting: val
	//; This is also a comment.
	//	other: some
	//#thing
}
