package commands

import (
	"gx/ipfs/QmYiqbfRCkryYvJsxBopy77YEhxNZXTmq5Y2qiKyenc59C/go-ipfs-cmdkit"

	cmds "github.com/ipfs/go-ipfs/commands"
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
