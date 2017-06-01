package main

import (
	"testing"

	"gx/ipfs/QmeGapzEYCQkoEYN5x5MCPdj1zMGMHRjcPbA26sveo2XV4/go-ipfs-cmdkit"
)

func TestIsCientErr(t *testing.T) {
	t.Log("Catch both pointers and values")
	if !isClientError(cmdkit.Error{Code: cmdkit.ErrClient}) {
		t.Errorf("misidentified value")
	}
	if !isClientError(&cmdkit.Error{Code: cmdkit.ErrClient}) {
		t.Errorf("misidentified pointer")
	}
}
