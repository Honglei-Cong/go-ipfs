package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"

	cmds "github.com/ipfs/go-ipfs/commands"
	core "github.com/ipfs/go-ipfs/core"

	"gx/ipfs/QmYiqbfRCkryYvJsxBopy77YEhxNZXTmq5Y2qiKyenc59C/go-ipfs-cmdsutil"
	ma "gx/ipfs/QmcyqRMCAXVtYPS4DiBrA7sezL9rRGfW8Ctx7cywL4TXJj/go-multiaddr"
)

// P2PListenerInfoOutput is output type of ls command
type P2PListenerInfoOutput struct {
	Protocol string
	Address  string
}

// P2PStreamInfoOutput is output type of streams command
type P2PStreamInfoOutput struct {
	HandlerID     string
	Protocol      string
	LocalPeer     string
	LocalAddress  string
	RemotePeer    string
	RemoteAddress string
}

// P2PLsOutput is output type of ls command
type P2PLsOutput struct {
	Listeners []P2PListenerInfoOutput
}

// P2PStreamsOutput is output type of streams command
type P2PStreamsOutput struct {
	Streams []P2PStreamInfoOutput
}

// P2PCmd is the 'ipfs ppp' command
var P2PCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Libp2p stream mounting.",
		ShortDescription: `
Create and use tunnels to remote peers over libp2p

Note: this command is experimental and subject to change as usecases and APIs are refined`,
	},

	Subcommands: map[string]*cmds.Command{
		"listener": p2pListenerCmd,
		"stream":   p2pStreamCmd,
	},
}

// p2pListenerCmd is the 'ipfs p2p listener' command
var p2pListenerCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline:          "P2P listener management.",
		ShortDescription: "Create and manage listener p2p endpoints",
	},

	Subcommands: map[string]*cmds.Command{
		"ls":    p2pListenerLsCmd,
		"open":  p2pListenerListenCmd,
		"close": p2pListenerCloseCmd,
	},
}

// p2pStreamCmd is the 'ipfs p2p stream' command
var p2pStreamCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline:          "P2P stream management.",
		ShortDescription: "Create and manage p2p streams",
	},

	Subcommands: map[string]*cmds.Command{
		"ls":    p2pStreamLsCmd,
		"dial":  p2pStreamDialCmd,
		"close": p2pStreamCloseCmd,
	},
}

var p2pListenerLsCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "List active p2p listeners.",
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("headers", "v", "Print table headers (HandlerID, Protocol, Local, Remote).").Default(false),
	},
	Run: func(req cmds.Request, res cmds.Response) {

		n, err := getNode(req)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		output := &P2PLsOutput{}

		for _, listener := range n.P2P.Listeners.Listeners {
			output.Listeners = append(output.Listeners, P2PListenerInfoOutput{
				Protocol: listener.Protocol,
				Address:  listener.Address.String(),
			})
		}

		res.SetOutput(output)
	},
	Type: P2PLsOutput{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			headers, _, _ := res.Request().Option("headers").Bool()
			list, _ := res.Output().(*P2PLsOutput)
			buf := new(bytes.Buffer)
			w := tabwriter.NewWriter(buf, 1, 2, 1, ' ', 0)
			for _, listener := range list.Listeners {
				if headers {
					fmt.Fprintln(w, "Address\tProtocol")
				}

				fmt.Fprintf(w, "%s\t%s\n", listener.Address, listener.Protocol)
			}
			w.Flush()

			return buf, nil
		},
	},
}

var p2pStreamLsCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "List active p2p streams.",
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("headers", "v", "Print table headers (HagndlerID, Protocol, Local, Remote).").Default(false),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := getNode(req)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		output := &P2PStreamsOutput{}

		for _, s := range n.P2P.Streams.Streams {
			output.Streams = append(output.Streams, P2PStreamInfoOutput{
				HandlerID: strconv.FormatUint(s.HandlerID, 10),

				Protocol: s.Protocol,

				LocalPeer:    s.LocalPeer.Pretty(),
				LocalAddress: s.LocalAddr.String(),

				RemotePeer:    s.RemotePeer.Pretty(),
				RemoteAddress: s.RemoteAddr.String(),
			})
		}

		res.SetOutput(output)
	},
	Type: P2PStreamsOutput{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			headers, _, _ := res.Request().Option("headers").Bool()
			list, _ := res.Output().(*P2PStreamsOutput)
			buf := new(bytes.Buffer)
			w := tabwriter.NewWriter(buf, 1, 2, 1, ' ', 0)
			for _, stream := range list.Streams {
				if headers {
					fmt.Fprintln(w, "HandlerID\tProtocol\tLocal\tRemote")
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", stream.HandlerID, stream.Protocol, stream.LocalAddress, stream.RemotePeer)
			}
			w.Flush()

			return buf, nil
		},
	},
}

var p2pListenerListenCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Forward p2p connections to a network multiaddr.",
		ShortDescription: `
Register a p2p connection handler and forward the connections to a specified address.

Note that the connections originate from the ipfs daemon process.
		`,
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("Protocol", true, false, "Protocol identifier."),
		cmdsutil.StringArg("Address", true, false, "Request handling application address."),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := getNode(req)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		proto := "/p2p/" + req.Arguments()[0]
		if n.P2P.CheckProtoExists(proto) {
			res.SetError(errors.New("protocol handler already registered"), cmdsutil.ErrNormal)
			return
		}

		addr, err := ma.NewMultiaddr(req.Arguments()[1])
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		_, err = n.P2P.NewListener(n.Context(), proto, addr)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		// Successful response.
		res.SetOutput(&P2PListenerInfoOutput{
			Protocol: proto,
			Address:  addr.String(),
		})
	},
}

var p2pStreamDialCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Dial to a p2p listener.",

		ShortDescription: `
Establish a new connection to a peer service.

When a connection is made to a peer service the ipfs daemon will setup one time
TCP listener and return it's bind port, this way a dialing application can
transparently connect to a p2p service.
		`,
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("Peer", true, false, "Remote peer to connect to"),
		cmdsutil.StringArg("Protocol", true, false, "Protocol identifier."),
		cmdsutil.StringArg("BindAddress", false, false, "Address to listen for connection/s (default: /ip4/127.0.0.1/tcp/0)."),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := getNode(req)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		addr, peer, err := ParsePeerParam(req.Arguments()[0])
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		proto := "/p2p/" + req.Arguments()[1]

		bindAddr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
		if len(req.Arguments()) == 3 {
			bindAddr, err = ma.NewMultiaddr(req.Arguments()[2])
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}
		}

		listenerInfo, err := n.P2P.Dial(n.Context(), addr, peer, proto, bindAddr)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		output := P2PListenerInfoOutput{
			Protocol: listenerInfo.Protocol,
			Address:  listenerInfo.Address.String(),
		}

		res.SetOutput(&output)
	},
}

var p2pListenerCloseCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Close active p2p listener.",
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("Protocol", false, false, "P2P listener protocol"),
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("all", "a", "Close all listeners.").Default(false),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := getNode(req)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		closeAll, _, _ := req.Option("all").Bool()
		var proto string

		if !closeAll {
			if len(req.Arguments()) == 0 {
				res.SetError(errors.New("no protocol name specified"), cmdsutil.ErrNormal)
				return
			}

			proto = "/p2p/" + req.Arguments()[0]
		}

		for _, listener := range n.P2P.Listeners.Listeners {
			if !closeAll && listener.Protocol != proto {
				continue
			}
			listener.Close()
			if !closeAll {
				break
			}
		}
	},
}

var p2pStreamCloseCmd = &cmds.Command{
	Helptext: cmdsutil.HelpText{
		Tagline: "Close active p2p stream.",
	},
	Arguments: []cmdsutil.Argument{
		cmdsutil.StringArg("HandlerID", false, false, "Stream HandlerID"),
	},
	Options: []cmdsutil.Option{
		cmdsutil.BoolOption("all", "a", "Close all streams.").Default(false),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := getNode(req)
		if err != nil {
			res.SetError(err, cmdsutil.ErrNormal)
			return
		}

		closeAll, _, _ := req.Option("all").Bool()
		var handlerID uint64

		if !closeAll {
			if len(req.Arguments()) == 0 {
				res.SetError(errors.New("no HandlerID specified"), cmdsutil.ErrNormal)
				return
			}

			handlerID, err = strconv.ParseUint(req.Arguments()[0], 10, 64)
			if err != nil {
				res.SetError(err, cmdsutil.ErrNormal)
				return
			}
		}

		for _, stream := range n.P2P.Streams.Streams {
			if !closeAll && handlerID != stream.HandlerID {
				continue
			}
			stream.Close()
			if !closeAll {
				break
			}
		}
	},
}

func getNode(req cmds.Request) (*core.IpfsNode, error) {
	n, err := req.InvocContext().GetNode()
	if err != nil {
		return nil, err
	}

	config, err := n.Repo.Config()
	if err != nil {
		return nil, err
	}

	if !config.Experimental.Libp2pStreamMounting {
		return nil, errors.New("libp2p stream mounting not enabled")
	}

	if !n.OnlineMode() {
		return nil, errNotOnline
	}

	return n, nil
}
