package commands

import (
	cmds "github.com/ipfs/go-ipfs/commands"

	"gx/ipfs/QmadYQbq2fJpaRE3XhpMLH68NNxmWMwfMQy1ntr1cKf7eo/go-ipfs-cmdkit"
)

var DiagCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Generate diagnostic reports.",
	},

	Subcommands: map[string]*cmds.Command{
		"sys":  sysDiagCmd,
		"cmds": ActiveReqsCmd,
	},
}
