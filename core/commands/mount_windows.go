package commands

import (
	"errors"

	"gx/ipfs/QmYiqbfRCkryYvJsxBopy77YEhxNZXTmq5Y2qiKyenc59C/go-ipfs-cmdkit"

	cmds "github.com/ipfs/go-ipfs/commands"
)

var MountCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline:          "Not yet implemented on Windows.",
		ShortDescription: "Not yet implemented on Windows. :(",
	},

	Run: func(req cmds.Request, res cmds.Response) {
		res.SetError(errors.New("Mount isn't compatible with Windows yet"), cmdsutil.ErrNormal)
	},
}
