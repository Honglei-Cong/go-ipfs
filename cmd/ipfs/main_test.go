package main

import (
	"testing"

	"gx/ipfs/QmYiqbfRCkryYvJsxBopy77YEhxNZXTmq5Y2qiKyenc59C/go-ipfs-cmdkit"
)

func TestIsCientErr(t *testing.T) {
	t.Log("Catch both pointers and values")
	if !isClientError(cmdsutil.Error{Code: cmdsutil.ErrClient}) {
		t.Errorf("misidentified value")
	}
	if !isClientError(&cmdsutil.Error{Code: cmdsutil.ErrClient}) {
		t.Errorf("misidentified pointer")
	}
}
