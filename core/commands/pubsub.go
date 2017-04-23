package commands

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	blocks "github.com/ipfs/go-ipfs/blocks"
	cmds "github.com/ipfs/go-ipfs/commands"
	core "github.com/ipfs/go-ipfs/core"
	e "github.com/ipfs/go-ipfs/core/commands/e"

	floodsub "gx/ipfs/QmUpeULWfmtsgCnfuRN3BHsfhHvBxNphoYh4La4CMxGt2Z/floodsub"
	pstore "gx/ipfs/QmXZSd1qR5BxZkPyuwfT5jpqQFScZccoZvDneXsKzCNHWX/go-libp2p-peerstore"
	cid "gx/ipfs/QmYhQaCYEcaPPjxJX7YcPcVKkQfRy6sJ7B3XmGFk82XYdQ/go-cid"
	"gx/ipfs/QmYiqbfRCkryYvJsxBopy77YEhxNZXTmq5Y2qiKyenc59C/go-ipfs-cmdkit"

)

var PubsubCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "An experimental publish-subscribe system on ipfs.",
		ShortDescription: `
ipfs pubsub allows you to publish messages to a given topic, and also to
subscribe to new messages on a given topic.

This is an experimental feature. It is not intended in its current state
to be used in a production environment.

To use, the daemon must be run with '--enable-pubsub-experiment'.
`,
	},
	Subcommands: map[string]*cmds.Command{
		"pub":   PubsubPubCmd,
		"sub":   PubsubSubCmd,
		"ls":    PubsubLsCmd,
		"peers": PubsubPeersCmd,
	},
}

var PubsubSubCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Subscribe to messages on a given topic.",
		ShortDescription: `
ipfs pubsub sub subscribes to messages on a given topic.

This is an experimental feature. It is not intended in its current state
to be used in a production environment.

To use, the daemon must be run with '--enable-pubsub-experiment'.
`,
		LongDescription: `
ipfs pubsub sub subscribes to messages on a given topic.

This is an experimental feature. It is not intended in its current state
to be used in a production environment.

To use, the daemon must be run with '--enable-pubsub-experiment'.

This command outputs data in the following encodings:
  * "json"
(Specified by the "--encoding" or "--enc" flag)
`,
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("topic", true, false, "String name of topic to subscribe to."),
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("discover", "try to discover other peers subscribed to the same topic"),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		// Must be online!
		if !n.OnlineMode() {
			res.SetError(errNotOnline, cmdsutil.ErrClient)
			return
		}

		if n.Floodsub == nil {
			res.SetError(fmt.Errorf("experimental pubsub feature not enabled. Run daemon with --enable-pubsub-experiment to use."), cmdsutil.ErrNormal)
			return
		}

		topic := req.Arguments()[0]
		sub, err := n.Floodsub.Subscribe(topic)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		discover, _, _ := req.Option("discover").Bool()
		if discover {
			go func() {
				blk := blocks.NewBlock([]byte("floodsub:" + topic))
				cid, err := n.Blocks.AddBlock(blk)
				if err != nil {
					log.Error("pubsub discovery: ", err)
					return
				}

				connectToPubSubPeers(req.Context(), n, cid)
			}()
		}

		out := make(chan interface{})
		res.SetOutput((<-chan interface{})(out))

		go func() {
			defer sub.Cancel()
			defer close(out)

			for {
				msg, err := sub.Next(req.Context())
				if err == io.EOF || err == context.Canceled {
					return
				} else if err != nil {
					res.SetError(err, cmdsutil.ErrNormal)
					return
				}

				out <- msg
			}
		}()
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: getPsMsgMarshaler(func(m *floodsub.Message) (io.Reader, error) {
			r := bytes.NewReader(m.Data)
			return r, nil
		}),
		"ndpayload": getPsMsgMarshaler(func(m *floodsub.Message) (io.Reader, error) {
			m.Data = append(m.Data, '\n')
			return bytes.NewReader(m.Data), nil
		}),
		"lenpayload": getPsMsgMarshaler(func(m *floodsub.Message) (io.Reader, error) {
			buf := make([]byte, 8)

			n := binary.PutUvarint(buf, uint64(len(m.Data)))
			return io.MultiReader(bytes.NewReader(buf[:n]), bytes.NewReader(m.Data)), nil
		}),
	},
	Type: floodsub.Message{},
}

func connectToPubSubPeers(ctx context.Context, n *core.IpfsNode, cid *cid.Cid) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	provs := n.Routing.FindProvidersAsync(ctx, cid, 10)
	wg := &sync.WaitGroup{}
	for p := range provs {
		wg.Add(1)
		go func(pi pstore.PeerInfo) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			err := n.PeerHost.Connect(ctx, pi)
			if err != nil {
				log.Info("pubsub discover: ", err)
				return
			}
			log.Info("connected to pubsub peer:", pi.ID)
		}(p)
	}

	wg.Wait()
}

func getPsMsgMarshaler(f func(m *floodsub.Message) (io.Reader, error)) func(cmds.Response) (io.Reader, error) {
	return func(res cmds.Response) (io.Reader, error) {
		v, err := unwrapOutput(res.Output())
		if err != nil {
			return nil, err
		}

		obj, ok := v.(*floodsub.Message)
		if !ok {
			return nil, e.TypeErr(obj, v)
		}
		if obj.Message == nil {
			return strings.NewReader(""), nil
		}

		return f(obj)
	}
}

var PubsubPubCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Publish a message to a given pubsub topic.",
		ShortDescription: `
ipfs pubsub pub publishes a message to a specified topic.

This is an experimental feature. It is not intended in its current state
to be used in a production environment.

To use, the daemon must be run with '--enable-pubsub-experiment'.
`,
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("topic", true, false, "Topic to publish to."),
		cmdsutil.StringArg("data", true, true, "Payload of message to publish.").EnableStdin(),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		// Must be online!
		if !n.OnlineMode() {
			res.SetError(errNotOnline, cmdsutil.ErrClient)
			return
		}

		if n.Floodsub == nil {
			res.SetError(fmt.Errorf("experimental pubsub feature not enabled. Run daemon with --enable-pubsub-experiment to use."), cmdsutil.ErrNormal)
			return
		}

		topic := req.Arguments()[0]

		for _, data := range req.Arguments()[1:] {
			if err := n.Floodsub.Publish(topic, []byte(data)); err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}
		}
	},
}

var PubsubLsCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "List subscribed topics by name.",
		ShortDescription: `
ipfs pubsub ls lists out the names of topics you are currently subscribed to.

This is an experimental feature. It is not intended in its current state
to be used in a production environment.

To use, the daemon must be run with '--enable-pubsub-experiment'.
`,
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		// Must be online!
		if !n.OnlineMode() {
			res.SetError(errNotOnline, cmdsutil.ErrClient)
			return
		}

		if n.Floodsub == nil {
			res.SetError(fmt.Errorf("experimental pubsub feature not enabled. Run daemon with --enable-pubsub-experiment to use."), cmdsutil.ErrNormal)
			return
		}

		res.SetOutput(&stringList{n.Floodsub.GetTopics()})
	},
	Type: stringList{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: stringListMarshaler,
	},
}

var PubsubPeersCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "List peers we are currently pubsubbing with.",
		ShortDescription: `
ipfs pubsub peers with no arguments lists out the pubsub peers you are
currently connected to. If given a topic, it will list connected
peers who are subscribed to the named topic.

This is an experimental feature. It is not intended in its current state
to be used in a production environment.

To use, the daemon must be run with '--enable-pubsub-experiment'.
`,
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("topic", false, false, "topic to list connected peers of"),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		// Must be online!
		if !n.OnlineMode() {
			res.SetError(errNotOnline, cmdsutil.ErrClient)
			return
		}

		if n.Floodsub == nil {
			res.SetError(fmt.Errorf("experimental pubsub feature not enabled. Run daemon with --enable-pubsub-experiment to use."), cmdsutil.ErrNormal)
			return
		}

		var topic string
		if len(req.Arguments()) == 1 {
			topic = req.Arguments()[0]
		}

		var out []string
		for _, p := range n.Floodsub.ListPeers(topic) {
			out = append(out, p.Pretty())
		}
		res.SetOutput(&stringList{out})
	},
	Type: stringList{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: stringListMarshaler,
	},
}
