package main

import (
	"testing"

	"gx/ipfs/QmT7xnHPBQcMbgpcDJ81opQZzU4LfLCFv5U1B6YERMRsDj/go-ipfs-cmdkit"
)

func TestIsCientErr(t *testing.T) {
	t.Log("Only catch pointers")
	if isClientError(cmdkit.Error{Code: cmdkit.ErrClient}) {
		t.Errorf("misidentified value")
	}
	if !isClientError(&cmdkit.Error{Code: cmdkit.ErrClient}) {
		t.Errorf("misidentified pointer")
	}
}
