package commands

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ipfs/go-ipfs/blocks"
	util "github.com/ipfs/go-ipfs/blocks/blockstore/util"
	e "github.com/ipfs/go-ipfs/core/commands/e"
	"gx/ipfs/QmRTwaSETX8m9rVAD9QacsoxFMURcuSoLDhf1jtABzCcLP/go-ipfs-cmds"
	"gx/ipfs/QmYiqbfRCkryYvJsxBopy77YEhxNZXTmq5Y2qiKyenc59C/go-ipfs-cmdkit"

	mh "gx/ipfs/QmVGtdTZdTFaLsaj2RwdVG8jcjNNcp1DE914DKZ2kHmXHw/go-multihash"
	cid "gx/ipfs/QmYhQaCYEcaPPjxJX7YcPcVKkQfRy6sJ7B3XmGFk82XYdQ/go-cid"
)

type BlockStat struct {
	Key  string
	Size int
}

func (bs BlockStat) String() string {
	return fmt.Sprintf("Key: %s\nSize: %d\n", bs.Key, bs.Size)
}

var BlockCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Interact with raw IPFS blocks.",
		ShortDescription: `
'ipfs block' is a plumbing command used to manipulate raw IPFS blocks.
Reads from stdin or writes to stdout, and <key> is a base58 encoded
multihash.
`,
	},

	Subcommands: map[string]*cmds.Command{
		"stat": blockStatCmd,
		"get":  blockGetCmd,
		"put":  blockPutCmd,
		"rm":   blockRmCmd,
	},
}

var blockStatCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Print information of a raw IPFS block.",
		ShortDescription: `
'ipfs block stat' is a plumbing command for retrieving information
on raw IPFS blocks. It outputs the following to stdout:

	Key  - the base58 encoded multihash
	Size - the size of the block in bytes

`,
	},

	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("key", true, false, "The base58 multihash of an existing block to stat.").EnableStdin(),
	},
	Run: func(req cmds.Request, re cmds.ResponseEmitter) {
		b, err := getBlockForKey(req, req.Arguments()[0])
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		err = re.Emit(&BlockStat{
			Key:  b.Cid().String(),
			Size: len(b.RawData()),
		})
		if err != nil {
			log.Error(err)
		}
	},
	Type: BlockStat{},
	Encoders: map[cmds.EncodingType]func(cmds.Request) func(io.Writer) cmds.Encoder{
		cmds.Text: cmds.MakeEncoder(func(w io.Writer, v interface{}) error {
			bs, ok := v.(*BlockStat)
			if !ok {
				return e.TypeErr(bs, v)
			}
			_, err := fmt.Fprintf(w, "%s", bs)
			return err
		}),
	},
}

var blockGetCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Get a raw IPFS block.",
		ShortDescription: `
'ipfs block get' is a plumbing command for retrieving raw IPFS blocks.
It outputs to stdout, and <key> is a base58 encoded multihash.
`,
	},

	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("key", true, false, "The base58 multihash of an existing block to get.").EnableStdin(),
	},
	Run: func(req cmds.Request, re cmds.ResponseEmitter) {
		b, err := getBlockForKey(req, req.Arguments()[0])
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		err = re.Emit(bytes.NewReader(b.RawData()))
		if err != nil {
			log.Error(err)
		}
	},
}

var blockPutCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Store input as an IPFS block.",
		ShortDescription: `
'ipfs block put' is a plumbing command for storing raw IPFS blocks.
It reads from stdin, and <key> is a base58 encoded multihash.
`,
	},

	Arguments: []cmdsutil.Argument{
		cmdsutil.FileArg("data", true, false, "The data to be stored as an IPFS block.").EnableStdin(),
	},
	Options: []cmdsutil.Option{
		cmdsutil.StringOption("format", "f", "cid format for blocks to be created with.").Default("v0"),
		cmdsutil.StringOption("mhtype", "multihash hash function").Default("sha2-256"),
		cmdsutil.IntOption("mhlen", "multihash hash length").Default(-1),
	},
	Run: func(req cmds.Request, re cmds.ResponseEmitter) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		file, err := req.Files().NextFile()
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		err = file.Close()
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		var pref cid.Prefix
		pref.Version = 1

		format, _, _ := req.Option("format").String()
		switch format {
		case "cbor":
			pref.Codec = cid.DagCBOR
		case "protobuf":
			pref.Codec = cid.DagProtobuf
		case "raw":
			pref.Codec = cid.Raw
		case "v0":
			pref.Version = 0
			pref.Codec = cid.DagProtobuf
		default:
			err := fmt.Errorf("unrecognized format: %s", format)
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}

			return
		}

		mhtype, _, _ := req.Option("mhtype").String()
		mhtval, ok := mh.Names[mhtype]
		if !ok {
			err := fmt.Errorf("unrecognized multihash function: %s", mhtype)
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}
		pref.MhType = mhtval

		mhlen, _, err := req.Option("mhlen").Int()
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}
		pref.MhLength = mhlen

		bcid, err := pref.Sum(data)
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		b, err := blocks.NewBlockWithCid(data, bcid)
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		k, err := n.Blocks.AddBlock(b)
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		err = re.Emit(&BlockStat{
			Key:  k.String(),
			Size: len(data),
		})
		if err != nil {
			log.Error(err)
		}
	},
	Encoders: map[cmds.EncodingType]func(cmds.Request) func(io.Writer) cmds.Encoder{
		cmds.Text: cmds.MakeEncoder(func(w io.Writer, v interface{}) error {
			bs, ok := v.(*BlockStat)
			if !ok {
				return e.TypeErr(bs, v)
			}
			_, err := fmt.Fprintf(w, "%s\n", bs.Key)
			return err
		}),
	},
	Type: BlockStat{},
}

func getBlockForKey(req cmds.Request, skey string) (blocks.Block, error) {
	if len(skey) == 0 {
		return nil, fmt.Errorf("zero length cid invalid")
	}

	n, err := req.InvocContext().GetNode()
	if err != nil {
		return nil, err
	}

	c, err := cid.Decode(skey)
	if err != nil {
		return nil, err
	}

	b, err := n.Blocks.GetBlock(req.Context(), c)
	if err != nil {
		return nil, err
	}

	return b, nil
}

var blockRmCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Remove IPFS block(s).",
		ShortDescription: `
'ipfs block rm' is a plumbing command for removing raw ipfs blocks.
It takes a list of base58 encoded multihashs to remove.
`,
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("hash", true, true, "Bash58 encoded multihash of block(s) to remove."),
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("force", "f", "Ignore nonexistent blocks.").Default(false),
		cmdsutil.BoolOption("quiet", "q", "Write minimal output.").Default(false),
	},
	Run: func(req cmds.Request, re cmds.ResponseEmitter) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}
		hashes := req.Arguments()
		force, _, _ := req.Option("force").Bool()
		quiet, _, _ := req.Option("quiet").Bool()
		cids := make([]*cid.Cid, 0, len(hashes))
		for _, hash := range hashes {
			c, err := cid.Decode(hash)
			if err != nil {
				err = fmt.Errorf("invalid content id: %s (%s)", hash, err)
				err2 := re.SetError(err, cmdsutil.ErrNormal)
				if err2 != nil {
					log.Error(err)
				}
				return
			}

			cids = append(cids, c)
		}
		ch, err := util.RmBlocks(n.Blockstore, n.Pinning, cids, util.RmBlocksOpts{
			Quiet: quiet,
			Force: force,
		})

		if err != nil {
			err2 := re.SetError(err, cmdsutil.ErrNormal)
			if err2 != nil {
				log.Error(err)
			}
			return
		}

		err = re.Emit(ch)
		if err != nil {
			log.Error(err)
		}
	},
	PostRun: map[cmds.EncodingType]func(cmds.Request, cmds.ResponseEmitter) cmds.ResponseEmitter{
		cmds.CLI: func(req cmds.Request, re cmds.ResponseEmitter) cmds.ResponseEmitter {
			reNext, res := cmds.NewChanResponsePair(req)

			go func() {
				defer re.Close()

				var (
					err        error
					v          interface{}
					someFailed bool
				)

				for {
					v, err = res.Next()

					if err != nil {
						if err == io.EOF {
							break
						}

						if err == cmds.ErrRcvdError {
							err = res.Error()
						}

						if e, ok := err.(*cmdsutil.Error); ok {
							err := re.SetError(e.Message, e.Code)
							if err != nil {
								log.Error(err)
							}
						} else {
							err2 := re.SetError(err, cmdsutil.ErrNormal)
							if err2 != nil {
								log.Error(err)
							}
						}

						return
					}

					r, ok := v.(*util.RemovedBlock)
					if !ok {
						log.Error(e.New(e.TypeErr(r, v)))
						break
					}

					if r.Hash == "" && r.Error != "" {
						fmt.Fprintf(os.Stderr, "aborted: %s\n", r.Error)
						someFailed = true
						break
					} else if r.Error != "" {
						someFailed = true
						fmt.Fprintf(os.Stderr, "cannot remove %s: %s\n", r.Hash, r.Error)
					} else {
						fmt.Fprintf(os.Stdout, "removed %s\n", r.Hash)
					}
				}

				if someFailed {
					err := re.SetError("some blocks not removed", cmdsutil.ErrNormal)
					if err != nil {
						log.Error(err)
					}
				}
			}()

			return reNext
		},
	},
	Type: util.RemovedBlock{},
}
