package main

import (
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/kubo/core/commands"
	"github.com/ipfs/kubo/core/node"
	"github.com/luanet/lua-proto/proto"
)

type IpTest struct {
	Ip      string `json:"ip"`
	Swarm   bool   `json:"swarm"`
	Gateway bool   `json:"gateway"`
}

var portForwardError = "Node port forwarding is not accessible. Please change your router's configuration and try again."
var testCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Luanet configuration tests.",
		ShortDescription: `
Before join Luanet, your node have to pass several tests.
Include port forwarding, internet speedtest and socket connection test.
`,
	},
	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	NoRemote:  true,
	Extra:     commands.CreateCmdExtras(commands.SetDoesNotUseRepo(true), commands.SetDoesNotUseConfigAsInput(true)),
	PreRun:    commands.DaemonRunning,
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		// socket connection test
		//TODO
		result := make(chan proto.TestResult)
		go node.NodeTest(result)
		_ = <-result
		return nil
	},
}
