package commands

import (
	cmds "github.com/ipfs/go-ipfs/commands"

	"gx/ipfs/QmeGapzEYCQkoEYN5x5MCPdj1zMGMHRjcPbA26sveo2XV4/go-ipfs-cmdkit"
)

var DiagCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Generate diagnostic reports.",
	},

	Subcommands: map[string]*cmds.Command{
		"sys":  sysDiagCmd,
		"cmds": ActiveReqsCmd,
	},
}