package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/kubo/core/commands"
	"github.com/showwin/speedtest-go/speedtest"
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

		// port forwarding checking...
		ip4Opened := ipv4Test()
		ip6Opened := ipv6Test()
		if !ip4Opened && !ip6Opened {
			// return fmt.Errorf(portForwardError)
		}

		// speedtest
		user, _ := speedtest.FetchUserInfo()
		serverList, _ := speedtest.FetchServers(user)
		targets, _ := serverList.FindServer([]int{})
		for _, s := range targets {
			err := testDownload(s, false)
			if err != nil {
				return err
			}

			err = testUpload(s, false)
			if err != nil {
				return err
			}

			showSpeedResult(s)
		}

		return nil
	},
}

func ipv4Test() bool {
	URL := "http://ip4.luanet.io"
	resp, err := http.Get(URL)
	if err != nil {
		log.Warn("Failed to check ipv4 port forwarding.")
		return false
	}

	defer resp.Body.Close()
	var cResp IpTest
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		log.Warn("Failed to decode ipv4 port forwarding data.")
		return false
	}

	if !cResp.Swarm {
		log.Warn("[Ipv4] Swarm port 4001 is not opened to the internet.")
	}

	if !cResp.Gateway {
		log.Warn("[Ipv4] Gateway port 443 is not opened to the internet.")
	}

	return cResp.Swarm && cResp.Gateway
}

func ipv6Test() bool {
	URL := "http://ip6.luanet.io"
	resp, err := http.Get(URL)
	if err != nil {
		log.Warn("Failed to check ipv6 port forwarding.")
		return false
	}

	defer resp.Body.Close()
	var cResp IpTest
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		log.Warn("Failed to decode ipv6 port forwarding data.")
		return false
	}

	if !cResp.Swarm {
		log.Warn("[Ipv6] Swarm port 4001 is not opened to the internet.")
	}

	if !cResp.Gateway {
		log.Warn("[Ipv6] Gateway port 443 is not opened to the internet.")
	}

	return cResp.Swarm && cResp.Gateway
}

func testDownload(server *speedtest.Server, savingMode bool) error {
	quit := make(chan bool)
	fmt.Printf("[Speedtest] Downloading: ")
	go dots(quit)
	err := server.DownloadTest(savingMode)
	quit <- true
	if err != nil {
		return err
	}
	fmt.Println()
	return err
}

func testUpload(server *speedtest.Server, savingMode bool) error {
	quit := make(chan bool)
	fmt.Printf("[Speedtest] Uploading: ")
	go dots(quit)
	err := server.UploadTest(savingMode)
	quit <- true
	if err != nil {
		return err
	}
	fmt.Println()
	return nil
}

func dots(quit chan bool) {
	for {
		select {
		case <-quit:
			return
		default:
			time.Sleep(time.Second)
			fmt.Print(".")
		}
	}
}

func showSpeedResult(server *speedtest.Server) {
	fmt.Printf("Download Speed: %5.2f Mbit/s\n", server.DLSpeed)
	fmt.Printf("Upload Speed: %5.2f Mbit/s\n\n", server.ULSpeed)
	valid := server.CheckResultValid()
	if !valid {
		fmt.Println("Warning: Result seems to be wrong. Please speedtest again.")
	}
}
