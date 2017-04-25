package main

import (
	"testing"

	"gx/ipfs/QmadYQbq2fJpaRE3XhpMLH68NNxmWMwfMQy1ntr1cKf7eo/go-ipfs-cmdkit"
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
