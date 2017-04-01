package commands

import (
	"fmt"

	"gx/ipfs/QmYiqbfRCkryYvJsxBopy77YEhxNZXTmq5Y2qiKyenc59C/go-ipfs-cmdkit"

	cmds "github.com/ipfs/go-ipfs/commands"
)

var daemonShutdownCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Shut down the ipfs daemon",
	},
	Run: func(req cmds.Request, res cmds.Response) {
		nd, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		if nd.LocalMode() {
			res.SetError(fmt.Errorf("daemon not running"), cmdsutil.ErrClient)
			return
		}

		if err := nd.Process().Close(); err != nil {
			log.Error("error while shutting down ipfs daemon:", err)
		}
	},
}
