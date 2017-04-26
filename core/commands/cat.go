package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	core "github.com/ipfs/go-ipfs/core"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"

	cmds "gx/ipfs/QmUZBejTzVRuN8ubr2LC8FG7YexRMsNnzM2s2Pi4JxJd5P/go-ipfs-cmds"
	"gx/ipfs/Qmf7G7FikwUsm48Jm4Yw4VBGNZuyRaAMzpWDJcW8V71uV2/go-ipfs-cmdkit"
)

const progressBarMinSize = 1024 * 1024 * 8 // show progress bar for outputs > 8MiB

var CatCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline:          "Show IPFS object data.",
		ShortDescription: "Displays the data contained by an IPFS or IPNS object(s) at the given path.",
	},

	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("ipfs-path", true, true, "The path to the IPFS object(s) to be outputted.").EnableStdin(),
	},
	Run: func(req cmds.Request, re cmds.ResponseEmitter) {
		node, err := req.InvocContext().GetNode()
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		if !node.OnlineMode() {
			if err := node.SetupOfflineRouting(); err != nil {
				err2 := re.SetError(err, cmdsutil.ErrNormal)
				if err2 != nil {
					log.Error(err)
				}
				return
			}
		}

		readers, length, err := cat(req.Context(), node, req.Arguments())

		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		/*
			if err := corerepo.ConditionalGC(req.Context(), node, length); err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}

				return
			}
		*/

		re.SetLength(length)

		reader := io.MultiReader(readers...)
		// Since the reader returns the error that a block is missing, we need to take
		// Emit errors and send them to the client. Usually we don't do that because
		// it means the connection is broken or we supplied an illegal argument etc.
		err = re.Emit(reader)
		if err != nil {
			err = re.SetError(err, cmdsutil.ErrNormal)
			if err != nil {
				log.Error(err)
			}
		}
		re.Close()
	},
	PostRun: map[cmds.EncodingType]func(cmds.Request, cmds.ResponseEmitter) cmds.ResponseEmitter{
		cmds.CLI: func(req cmds.Request, re cmds.ResponseEmitter) cmds.ResponseEmitter {
			reNext, res := cmds.NewChanResponsePair(req)

			go func() {
				if res.Length() > 0 && res.Length() < progressBarMinSize {
					if err := cmds.Copy(re, res); err != nil {
						err2 := re.SetError(err, cmdsutil.ErrNormal)
						if err2 != nil {
							log.Error(err)
						}
					}

					return
				}

				// Copy closes by itself, so we must not do this before
				defer re.Close()

				v, err := res.Next()
				if err != nil {
					if err == cmds.ErrRcvdError {
						err2 := re.SetError(res.Error().Message, res.Error().Code)
						if err2 != nil {
							log.Error(err)
						}
					} else {
						err2 := re.SetError(res.Error(), cmdsutil.ErrNormal)
						if err2 != nil {
							log.Error(err)
						}
					}

					return
				}

				r, ok := v.(io.Reader)
				if !ok {
					err2 := re.SetError(fmt.Sprintf("expected io.Reader, not %T", v), cmdsutil.ErrNormal)
					if err2 != nil {
						log.Error(err)
					}
					return
				}

				bar, reader := progressBarForReader(os.Stderr, r, int64(res.Length()))
				bar.Start()

				err = re.Emit(reader)
				if err != nil {
					log.Error(err)
				}
			}()

			return reNext
		},
	},
}

func cat(ctx context.Context, node *core.IpfsNode, paths []string) ([]io.Reader, uint64, error) {
	readers := make([]io.Reader, 0, len(paths))
	length := uint64(0)
	for _, fpath := range paths {
		read, err := coreunix.Cat(ctx, node, fpath)
		if err != nil {
			return nil, 0, err
		}
		readers = append(readers, read)
		length += uint64(read.Size())
	}
	return readers, length, nil
}
