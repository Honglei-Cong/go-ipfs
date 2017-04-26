package commands

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	cmds "github.com/ipfs/go-ipfs/commands"
	e "github.com/ipfs/go-ipfs/core/commands/e"

	ci "gx/ipfs/QmP1DfoUjiWH2ZBo1PBH6FupdBucbDepx3HpWmEY6JMUpY/go-libp2p-crypto"
	"gx/ipfs/Qmf7G7FikwUsm48Jm4Yw4VBGNZuyRaAMzpWDJcW8V71uV2/go-ipfs-cmdkit"
	peer "gx/ipfs/QmdS9KpbDyPrieswibZhkod1oXqRwZJrUPzxCofAMWpFGq/go-libp2p-peer"
)

var KeyCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Create and list IPNS name keypairs",
		ShortDescription: `
'ipfs key gen' generates a new keypair for usage with IPNS and 'ipfs name publish'.

  > ipfs key gen --type=rsa --size=2048 mykey
  > ipfs name publish --key=mykey QmSomeHash

'ipfs key list' lists the available keys.

  > ipfs key list
  self
  mykey
		`,
	},
	Subcommands: map[string]*cmds.Command{
		"gen":    keyGenCmd,
		"list":   keyListCmd,
		"rename": keyRenameCmd,
		"rm":     keyRmCmd,
	},
}

type KeyOutput struct {
	Name string
	Id   string
}

type KeyOutputList struct {
	Keys []KeyOutput
}

// KeyRenameOutput define the output type of keyRenameCmd
type KeyRenameOutput struct {
	Was       string
	Now       string
	Id        string
	Overwrite bool
}

var keyGenCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Create a new keypair",
	},
	Options: []cmdsutil.Option{
		cmdsutil.StringOption("type", "t", "type of the key to create [rsa, ed25519]"),
		cmdsutil.IntOption("size", "s", "size of the key to generate"),
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("name", true, false, "name of key to create"),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		typ, f, err := req.Option("type").String()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		if !f {
			res.SetError(fmt.Errorf("please specify a key type with --type"), cmdsutil.ErrNormal)
			return
		}

		size, sizefound, err := req.Option("size").Int()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		name := req.Arguments()[0]
		if name == "self" {
			res.SetError(fmt.Errorf("cannot create key with name 'self'"), cmdsutil.ErrNormal)
			return
		}

		var sk ci.PrivKey
		var pk ci.PubKey

		switch typ {
		case "rsa":
			if !sizefound {
				res.SetError(fmt.Errorf("please specify a key size with --size"), cmdsutil.ErrNormal)
				return
			}

			priv, pub, err := ci.GenerateKeyPairWithReader(ci.RSA, size, rand.Reader)
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}

			sk = priv
			pk = pub
		case "ed25519":
			priv, pub, err := ci.GenerateEd25519Key(rand.Reader)
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}

			sk = priv
			pk = pub
		default:
			res.SetError(fmt.Errorf("unrecognized key type: %s", typ), cmdsutil.ErrNormal)
			return
		}

		err = n.Repo.Keystore().Put(name, sk)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		pid, err := peer.IDFromPublicKey(pk)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		res.SetOutput(&KeyOutput{
			Name: name,
			Id:   pid.Pretty(),
		})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			k, ok := v.(*KeyOutput)
			if !ok {
				return nil, e.TypeErr(k, v)
			}

			return strings.NewReader(k.Id + "\n"), nil
		},
	},
	Type: KeyOutput{},
}

var keyListCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "List all local keypairs",
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("l", "Show extra information about keys."),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		keys, err := n.Repo.Keystore().List()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		sort.Strings(keys)

		list := make([]KeyOutput, 0, len(keys)+1)

		list = append(list, KeyOutput{Name: "self", Id: n.Identity.Pretty()})

		for _, key := range keys {
			privKey, err := n.Repo.Keystore().Get(key)
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}

			pubKey := privKey.GetPublic()

			pid, err := peer.IDFromPublicKey(pubKey)
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}

			list = append(list, KeyOutput{Name: key, Id: pid.Pretty()})
		}

		res.SetOutput(&KeyOutputList{list})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: keyOutputListMarshaler,
	},
	Type: KeyOutputList{},
}

var keyRenameCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Rename a keypair",
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("name", true, false, "name of key to rename"),
		cmdsutil.StringArg("newName", true, false, "new name of the key"),
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("force", "f", "Allow to overwrite an existing key."),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		ks := n.Repo.Keystore()

		name := req.Arguments()[0]
		newName := req.Arguments()[1]

		if name == "self" {
			res.SetError(fmt.Errorf("cannot rename key with name 'self'"), cmdsutil.ErrNormal)
			return
		}

		if newName == "self" {
			res.SetError(fmt.Errorf("cannot overwrite key with name 'self'"), cmdsutil.ErrNormal)
			return
		}

		oldKey, err := ks.Get(name)
		if err != nil {
			res.SetError(fmt.Errorf("no key named %s was found", name), cmdsutil.ErrNormal)
			return
		}

		pubKey := oldKey.GetPublic()

		pid, err := peer.IDFromPublicKey(pubKey)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		overwrite := false
		force, _, _ := res.Request().Option("f").Bool()
		if force {
			exist, err := ks.Has(newName)
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}

			if exist {
				overwrite = true
				err := ks.Delete(newName)
				if err != nil {
					res.SetError(err, cmdsutil.ErrNormal)
					return
				}
			}
		}

		err = ks.Put(newName, oldKey)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		err = ks.Delete(name)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		res.SetOutput(&KeyRenameOutput{
			Was:       name,
			Now:       newName,
			Id:        pid.Pretty(),
			Overwrite: overwrite,
		})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			k, ok := res.Output().(*KeyRenameOutput)
			if !ok {
				return nil, fmt.Errorf("expected a KeyRenameOutput as command result")
			}

			buf := new(bytes.Buffer)

			if k.Overwrite {
				fmt.Fprintf(buf, "Key %s renamed to %s with overwriting\n", k.Id, k.Now)
			} else {
				fmt.Fprintf(buf, "Key %s renamed to %s\n", k.Id, k.Now)
			}
			return buf, nil
		},
	},
	Type: KeyRenameOutput{},
}

var keyRmCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Remove a keypair",
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("name", true, true, "names of keys to remove").EnableStdin(),
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("l", "Show extra information about keys."),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		names := req.Arguments()

		list := make([]KeyOutput, 0, len(names))
		for _, name := range names {
			if name == "self" {
				res.SetError(fmt.Errorf("cannot remove key with name 'self'"), cmdsutil.ErrNormal)
				return
			}

			removed, err := n.Repo.Keystore().Get(name)
			if err != nil {
				res.SetError(fmt.Errorf("no key named %s was found", name), cmdsutil.ErrNormal)
				return
			}

			pubKey := removed.GetPublic()

			pid, err := peer.IDFromPublicKey(pubKey)
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}

			list = append(list, KeyOutput{Name: name, Id: pid.Pretty()})
		}

		for _, name := range names {
			err = n.Repo.Keystore().Delete(name)
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}
		}

		res.SetOutput(&KeyOutputList{list})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: keyOutputListMarshaler,
	},
	Type: KeyOutputList{},
}

func keyOutputListMarshaler(res cmds.Response) (io.Reader, error) {
	withId, _, _ := res.Request().Option("l").Bool()

	v, err := unwrapOutput(res.Output())
	if err != nil {
		return nil, err
	}
	list, ok := v.(*KeyOutputList)
	if !ok {
		return nil, e.TypeErr(list, v)
	}

	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 1, 2, 1, ' ', 0)
	for _, s := range list.Keys {
		if withId {
			fmt.Fprintf(w, "%s\t%s\t\n", s.Id, s.Name)
		} else {
			fmt.Fprintf(w, "%s\n", s.Name)
		}
	}
	w.Flush()
	return buf, nil
}
