package commands

import (
	"io"
	"strings"

	cmds "github.com/ipfs/go-ipfs/commands"
	core "github.com/ipfs/go-ipfs/core"
	e "github.com/ipfs/go-ipfs/core/commands/e"
	"github.com/ipfs/go-ipfs/core/coreunix"
	dag "github.com/ipfs/go-ipfs/merkledag"
	path "github.com/ipfs/go-ipfs/path"
	tar "github.com/ipfs/go-ipfs/tar"

	"gx/ipfs/QmadYQbq2fJpaRE3XhpMLH68NNxmWMwfMQy1ntr1cKf7eo/go-ipfs-cmdkit"
)

var TarCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Utility functions for tar files in ipfs.",
	},

	Subcommands: map[string]*cmds.Command{
		"add": tarAddCmd,
		"cat": tarCatCmd,
	},
}

var tarAddCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Import a tar file into ipfs.",
		ShortDescription: `
'ipfs tar add' will parse a tar file and create a merkledag structure to
represent it.
`,
	},

	Arguments: []cmdsutil.Argument{
		cmdsutil.FileArg("file", true, false, "Tar file to add.").EnableStdin(),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		nd, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		fi, err := req.Files().NextFile()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		node, err := tar.ImportTar(fi, nd.DAG)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		c := node.Cid()

		fi.FileName()
		res.SetOutput(&coreunix.AddedObject{
			Name: fi.FileName(),
			Hash: c.String(),
		})
	},
	Type: coreunix.AddedObject{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			o, ok := v.(*coreunix.AddedObject)
			if !ok {
				return nil, e.TypeErr(o, v)
			}
			return strings.NewReader(o.Hash + "\n"), nil
		},
	},
}

var tarCatCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Export a tar file from IPFS.",
		ShortDescription: `
'ipfs tar cat' will export a tar file from a previously imported one in IPFS.
`,
	},

	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("path", true, false, "ipfs path of archive to export.").EnableStdin(),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		nd, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		p, err := path.ParsePath(req.Arguments()[0])
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		root, err := core.Resolve(req.Context(), nd.Namesys, nd.Resolver, p)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		rootpb, ok := root.(*dag.ProtoNode)
		if !ok {
			res.SetError(dag.ErrNotProtobuf, cmdsutil.ErrNormal)
			return
		}

		r, err := tar.ExportTar(req.Context(), rootpb, nd.DAG)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		res.SetOutput(r)
	},
}
