package commands

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	cmdsutil "gx/ipfs/QmYiqbfRCkryYvJsxBopy77YEhxNZXTmq5Y2qiKyenc59C/go-ipfs-cmdkit"
)

type TestOutput struct {
	Foo, Bar string
	Baz      int
}

func TestMarshalling(t *testing.T) {
	cmd := &Command{}
	opts, _ := cmd.GetOptions(nil)

	req, _ := NewRequest(nil, nil, nil, nil, nil, opts)

	res := NewResponse(req)
	res.SetOutput(TestOutput{"beep", "boop", 1337})

	_, err := res.Marshal()
	if err == nil {
		t.Error("Should have failed (no encoding type specified in request)")
	}

	req.SetOption(cmdsutil.EncShort, JSON)

	reader, err := res.Marshal()
	if err != nil {
		t.Error(err, "Should have passed")
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	output := buf.String()
	if removeWhitespace(output) != "{\"Foo\":\"beep\",\"Bar\":\"boop\",\"Baz\":1337}" {
		t.Error("Incorrect JSON output")
	}

	res.SetError(fmt.Errorf("Oops!"), cmdsutil.ErrClient)
	reader, err = res.Marshal()
	if err != nil {
		t.Error("Should have passed")
	}
	buf.Reset()
	buf.ReadFrom(reader)
	output = buf.String()
	fmt.Println(removeWhitespace(output))
	if removeWhitespace(output) != "{\"Message\":\"Oops!\",\"Code\":1}" {
		t.Error("Incorrect JSON output")
	}
}

func TestErrTypeOrder(t *testing.T) {
	if cmdsutil.ErrNormal != 0 || cmdsutil.ErrClient != 1 || cmdsutil.ErrImplementation != 2 || cmdsutil.ErrNotFound != 3 {
		t.Fatal("ErrType order is wrong")
	}
}

func removeWhitespace(input string) string {
	input = strings.Replace(input, " ", "", -1)
	input = strings.Replace(input, "\t", "", -1)
	input = strings.Replace(input, "\n", "", -1)
	return strings.Replace(input, "\r", "", -1)
}
