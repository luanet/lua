package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/luanet/lua-proto/proto"
	"github.com/showwin/speedtest-go/speedtest"
)

func NodeTest(testResult chan proto.TestResult) {
	result := proto.TestResult{}
	result.Ports = make(map[string]proto.IpTest)
	// port forwarding checking...
	v4 := ipv4Test()
	v6 := ipv6Test()
	if !v4.IsOpen() && !v6.IsOpen() {
		logger.Error("Your node's ports is not open to the internet.")
	}

	result.Ports["v4"] = v4
	result.Ports["v6"] = v6

	// speedtest
	user, _ := speedtest.FetchUserInfo()
	serverList, _ := speedtest.FetchServers(user)
	targets, _ := serverList.FindServer([]int{})
	for _, s := range targets {
		if err := testDownload(s, false); err == nil {
			result.Download = s.DLSpeed
		}

		if err := testUpload(s, false); err == nil {
			result.Upload = s.ULSpeed
		}

		showSpeedResult(s)
		break
	}

	testResult <- result
	return
}
func ipv4Test() proto.IpTest {
	var cResp proto.IpTest
	URL := "http://ip4.luanet.io"
	resp, err := http.Get(URL)
	if err != nil {
		logger.Warn("Failed to check ipv4 port forwarding.")
		return cResp
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		logger.Warn("Failed to decode ipv4 port forwarding data.")
		return cResp
	}

	if !cResp.Swarm {
		logger.Warn("[Ipv4] Swarm port 4001 is not opened to the internet.")
	}

	if !cResp.Gateway {
		logger.Warn("[Ipv4] Gateway port 443 is not opened to the internet.")
	}

	return cResp
}

func ipv6Test() proto.IpTest {
	var cResp proto.IpTest
	URL := "http://ip6.luanet.io"
	resp, err := http.Get(URL)
	if err != nil {
		logger.Warn("Failed to check ipv6 port forwarding.")
		return cResp
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		logger.Warn("Failed to decode ipv6 port forwarding data.")
		return cResp
	}

	if !cResp.Swarm {
		logger.Warn("[Ipv6] Swarm port 4001 is not opened to the internet.")
	}

	if !cResp.Gateway {
		logger.Warn("[Ipv6] Gateway port 443 is not opened to the internet.")
	}

	return cResp
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
	fmt.Printf("[Speedtest] Download Speed: %5.2f Mbit/s\n", server.DLSpeed)
	fmt.Printf("[Speedtest] Upload Speed: %5.2f Mbit/s\n\n", server.ULSpeed)
	valid := server.CheckResultValid()
	if !valid {
		fmt.Println("Warning: Result seems to be wrong. Please speedtest again.")
	}
}
