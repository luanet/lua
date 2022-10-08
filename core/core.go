/*
Package core implements the IpfsNode object and related methods.

Packages underneath core/ provide a (relatively) stable, low-level API
to carry out most IPFS-related tasks.  For more details on the other
interfaces and how core/... fits into the bigger IPFS picture, see:

	$ godoc github.com/ipfs/go-ipfs
*/
package core

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ipfs/go-filestore"
	pin "github.com/ipfs/go-ipfs-pinner"

	bserv "github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-fetcher"
	"github.com/ipfs/go-graphsync"
	bstore "github.com/ipfs/go-ipfs-blockstore"
	exchange "github.com/ipfs/go-ipfs-exchange-interface"
	provider "github.com/ipfs/go-ipfs-provider"
	ipld "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log"
	mfs "github.com/ipfs/go-mfs"
	goprocess "github.com/jbenet/goprocess"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	psrouter "github.com/libp2p/go-libp2p-pubsub-router"
	record "github.com/libp2p/go-libp2p-record"
	connmgr "github.com/libp2p/go-libp2p/core/connmgr"
	ic "github.com/libp2p/go-libp2p/core/crypto"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	metrics "github.com/libp2p/go-libp2p/core/metrics"
	"github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	pstore "github.com/libp2p/go-libp2p/core/peerstore"
	routing "github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	p2pbhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"

	"github.com/ipfs/go-namesys"
	ipnsrp "github.com/ipfs/go-namesys/republisher"
	"github.com/ipfs/kubo/core/bootstrap"
	"github.com/ipfs/kubo/core/node"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/fuse/mount"
	"github.com/ipfs/kubo/p2p"
	"github.com/ipfs/kubo/peering"
	"github.com/ipfs/kubo/repo"
	irouting "github.com/ipfs/kubo/routing"

	"github.com/luanet/lua-proto/proto"
	"github.com/lucas-clemente/quic-go"
)

var log = logging.Logger("core")

// IpfsNode is IPFS Core module. It represents an IPFS instance.
type IpfsNode struct {

	// Self
	Identity peer.ID // the local node's identity

	// Quic connection to luanet
	Stream *quic.Stream `optional:"true"`

	Repo repo.Repo

	// config root
	ConfigRoot string `optional:"true"`

	// Local node
	Pinning         pin.Pinner             // the pinning manager
	Mounts          Mounts                 `optional:"true"` // current mount state, if any.
	PrivateKey      ic.PrivKey             `optional:"true"` // the local node's private Key
	PNetFingerprint libp2p.PNetFingerprint `optional:"true"` // fingerprint of private network

	// Services
	Peerstore            pstore.Peerstore          `optional:"true"` // storage for other Peer instances
	Blockstore           bstore.GCBlockstore       // the block store (lower level)
	Filestore            *filestore.Filestore      `optional:"true"` // the filestore blockstore
	BaseBlocks           node.BaseBlocks           // the raw blockstore, no filestore wrapping
	GCLocker             bstore.GCLocker           // the locker used to protect the blockstore during gc
	Blocks               bserv.BlockService        // the block service, get/add blocks.
	DAG                  ipld.DAGService           // the merkle dag service, get/add objects.
	IPLDFetcherFactory   fetcher.Factory           `name:"ipldFetcher"`   // fetcher that paths over the IPLD data model
	UnixFSFetcherFactory fetcher.Factory           `name:"unixfsFetcher"` // fetcher that interprets UnixFS data
	Reporter             *metrics.BandwidthCounter `optional:"true"`
	Discovery            mdns.Service              `optional:"true"`
	FilesRoot            *mfs.Root
	RecordValidator      record.Validator

	// Online
	PeerHost        p2phost.Host               `optional:"true"` // the network host (server+client)
	Peering         *peering.PeeringService    `optional:"true"`
	Filters         *ma.Filters                `optional:"true"`
	Bootstrapper    io.Closer                  `optional:"true"` // the periodic bootstrapper
	Routing         irouting.ProvideManyRouter `optional:"true"` // the routing system. recommend ipfs-dht
	DNSResolver     *madns.Resolver            // the DNS resolver
	Exchange        exchange.Interface         // the block exchange + strategy (bitswap)
	Namesys         namesys.NameSystem         // the name system, resolves paths to hashes
	Provider        provider.System            // the value provider system
	IpnsRepub       *ipnsrp.Republisher        `optional:"true"`
	GraphExchange   graphsync.GraphExchange    `optional:"true"`
	ResourceManager network.ResourceManager    `optional:"true"`

	PubSub   *pubsub.PubSub             `optional:"true"`
	PSRouter *psrouter.PubsubValueStore `optional:"true"`

	DHT       *ddht.DHT       `optional:"true"`
	DHTClient routing.Routing `name:"dhtc" optional:"true"`

	P2P *p2p.P2P `optional:"true"`

	Process goprocess.Process
	ctx     context.Context

	stop func() error

	// Flags
	IsOnline bool `optional:"true"` // Online is set when networking is enabled.
	IsDaemon bool `optional:"true"` // Daemon is set when running on a long-running daemon.
}

// Mounts defines what the node's mount state is. This should
// perhaps be moved to the daemon or mount. It's here because
// it needs to be accessible across daemon requests.
type Mounts struct {
	Ipfs mount.Mount
	Ipns mount.Mount
}

// Close calls Close() on the App object
func (n *IpfsNode) Close() error {
	return n.stop()
}

// Context returns the IpfsNode context
func (n *IpfsNode) Context() context.Context {
	if n.ctx == nil {
		n.ctx = context.TODO()
	}
	return n.ctx
}

// Bootstrap will set and call the IpfsNodes bootstrap function.
func (n *IpfsNode) Bootstrap(cfg bootstrap.BootstrapConfig) error {
	// TODO what should return value be when in offlineMode?
	if n.Routing == nil {
		return nil
	}

	if n.Bootstrapper != nil {
		n.Bootstrapper.Close() // stop previous bootstrap process.
	}

	// if the caller did not specify a bootstrap peer function, get the
	// freshest bootstrap peers from config. this responds to live changes.
	if cfg.BootstrapPeers == nil {
		cfg.BootstrapPeers = func() []peer.AddrInfo {
			ps, err := n.loadBootstrapPeers()
			if err != nil {
				log.Warn("failed to parse bootstrap peers from config")
				return nil
			}
			return ps
		}
	}

	var err error
	n.Bootstrapper, err = bootstrap.Bootstrap(n.Identity, n.PeerHost, n.Routing, cfg)
	return err
}

func (n *IpfsNode) loadBootstrapPeers() ([]peer.AddrInfo, error) {
	cfg, err := n.Repo.Config()
	if err != nil {
		return nil, err
	}

	return cfg.BootstrapPeers()
}

func (n *IpfsNode) JoinLuanet() (*proto.JoinRes, error) {
	cfg, err := n.Repo.Config()
	if err != nil {
		return nil, err
	}

	gob.Register(proto.Ip{})
	gob.Register(proto.JoinReq{})
	gob.Register(proto.JoinRes{})
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"wq-vvv-01"},
	}

	log.Info("Connecting to node address: ", cfg.Luanet.Api)
	conn, err := quic.DialAddr(cfg.Luanet.Api, tlsConf, nil)
	if err != nil {
		return nil, err
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	expires := time.Now().Unix() + cfg.Luanet.ExpiresTime
	bytes := []byte(n.Identity.String() + "." + strconv.FormatInt(expires, 10))
	signature, err := n.PrivateKey.Sign(bytes)
	if err != nil {
		return nil, err
	}

	ip4 := n.GetIpInfo("ip4")
	ip6 := n.GetIpInfo("ip6")
	message := proto.Proto{
		Service: proto.JoinService,
		Data: proto.JoinReq{
			Address:   n.Identity.String(),
			Ipv4:      *ip4,
			Ipv6:      *ip6,
			Signature: signature,
			Expires:   expires,
		},
	}

	n.Stream = &stream
	n.SendQuicMsg(message)

	msg := n.ReadQuicMsg()
	joinRes := msg.Data.(proto.JoinRes)
	if !joinRes.Success {
		return nil, fmt.Errorf("Failed to join lua network: %s", joinRes.Message)
	}

	// write certs to file
	for ip, cert := range joinRes.Certs {
		if err := os.MkdirAll(filepath.Join(n.ConfigRoot, "certs", ip), os.ModePerm); err != nil {
			return nil, err
		}

		if err = ioutil.WriteFile(filepath.Join(n.ConfigRoot, "certs", ip, "private.pem"), []byte(cert.Pems.Privkey), os.ModePerm); err != nil {
			return nil, err
		}

		_ = ioutil.WriteFile(filepath.Join(n.ConfigRoot, "certs", ip, "cert.pem"), []byte(cert.Pems.Cert), os.ModePerm)
	}

	return &joinRes, nil
}

func (n *IpfsNode) CmdHandlers() error {
	ticker := time.NewTicker(250 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			msg := n.ReadQuicMsg()
			message := proto.Proto{
				Service: msg.Service,
			}

			switch msg.Service {
			case proto.SpeedTestService:
				// result := make(chan proto.TestResult)
				// go node.NodeTest(result)
				// message.Data = <-result
			}

			if message.Data != nil {
				fmt.Println("Sending cmd response....")
				n.SendQuicMsg(message)
			}
		}
	}

	return nil
}

func (n *IpfsNode) GetIpInfo(version string) *proto.Ip {
	var ip proto.Ip = proto.Ip{}
	cfg, err := n.Repo.Config()
	if err != nil {
		return &ip
	}

	var client = &http.Client{Timeout: 2 * time.Second}
	r, err := client.Get("http://" + version + "." + cfg.Luanet.Domain)
	if err != nil {
		return &ip
	}

	defer r.Body.Close()
	json.NewDecoder(r.Body).Decode(&ip)
	return &ip
}

func (n *IpfsNode) HeartBeat() {
	gob.Register(proto.HeartBeatReq{})
	gob.Register(proto.HeartBeatRes{})
	gob.Register(proto.Stats{})
	gob.Register(proto.TestResult{})
	gob.Register(proto.IpTest{})
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			// TODO get stats from gateway only, not IPFS
			totals := n.Reporter.GetBandwidthTotals()
			message := proto.Proto{
				Service: proto.HeartBeatService,
				Data: proto.HeartBeatReq{
					Stats: proto.Stats{
						Storage: 0,
						In:      totals.TotalIn,
						Out:     totals.TotalOut,
						Ingress: totals.RateIn,
						Egress:  totals.RateOut,
					},
				},
			}

			n.SendQuicMsg(message)
		}
	}
}

func (n *IpfsNode) SendQuicMsg(msg proto.Proto) {
	enc := gob.NewEncoder(*n.Stream) // Will write to network.
	err := enc.Encode(msg)
	if err != nil {
		log.Error("Failed to send quic message: %v", err)
		n.JoinLuanet()
	}
}

func (n *IpfsNode) ReadQuicMsg() (message proto.Proto) {
	dec := gob.NewDecoder(*n.Stream)
	err := dec.Decode(&message)
	if err != nil {
		log.Error("Failed to read quic message: %v", err)
	}

	return
}

type ConstructPeerHostOpts struct {
	AddrsFactory      p2pbhost.AddrsFactory
	DisableNatPortMap bool
	DisableRelay      bool
	EnableRelayHop    bool
	ConnectionManager connmgr.ConnManager
}
