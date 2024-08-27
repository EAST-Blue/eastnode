package tss

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	signer "eastnode/tss/signer"
	"eastnode/tss/verifier"

	"github.com/bytemare/frost"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

type PeerNode struct {
	ctx          context.Context
	Frost        *signer.Signer
	Host         host.Host
	Topic        *pubsub.Topic
	Subscription *pubsub.Subscription
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var addressList arrayFlags

var (
	topicNameFlag = flag.String("topicName", "eastnode", "name of topic to join")
)

func Run(ctx context.Context) PeerNode {
	flag.Var(&addressList, "peerid", "List of peer ID")
	flag.Parse()

	privKey, err := GetPeerKey()
	if err != nil {
		panic(err)
	}
	h, err := libp2p.New(libp2p.Identity(privKey))
	if err != nil {
		panic(err)
	}

	t, n, p, g, s, err := LoadFrostKey()
	if err != nil {
		panic(err)
	}
	frostInstance := signer.NewFromStaticKeys(h.ID().String(), t, n, p, g, s)

	discoverPeers(ctx, h)

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}
	topic, err := ps.Join(*topicNameFlag)
	if err != nil {
		panic(err)
	}
	go streamConsoleTo(ctx, topic)

	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)

	return PeerNode{
		ctx:          ctx,
		Frost:        &frostInstance,
		Host:         h,
		Topic:        topic,
		Subscription: sub,
	}
}

func initDHT(ctx context.Context, h host.Host) *dht.IpfsDHT {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		panic(err)
	}
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				log.Println("Bootstrap warning:", err)
			}
		}()
	}
	wg.Wait()

	return kademliaDHT
}

func discoverPeers(ctx context.Context, h host.Host) {
	kademliaDHT := initDHT(ctx, h)
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, *topicNameFlag)

	// Look for others who have announced and attempt to connect to them
	anyConnected := false
	for !anyConnected {
		log.Println("Searching for peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, *topicNameFlag)
		if err != nil {
			panic(err)
		}
		for peer := range peerChan {
			if peer.ID == h.ID() {
				continue // No self connection
			}

			if len(addressList) > 0 && !slices.Contains(addressList, peer.ID.String()) {
				continue
			}

			err := h.Connect(ctx, peer)
			if err != nil {
				log.Printf("Failed connecting to %s, error: %s\n", peer.ID, err)
			} else {
				log.Println("Connected to:", peer.ID)
				anyConnected = true
			}
		}
	}
	log.Println("Peer discovery complete")
}

func streamConsoleTo(ctx context.Context, topic *pubsub.Topic) {
	reader := bufio.NewReader(os.Stdin)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if err := topic.Publish(ctx, []byte(s)); err != nil {
			log.Println("Publish error:", err)
		}
	}
}

func (node *PeerNode) Publish(p Payload) {

	data, err := json.Marshal(p)
	if err != nil {
		log.Println("Publish error:", err)
		return
	}

	if err := node.Topic.Publish(node.ctx, data); err != nil {
		log.Println("Publish error:", err)
	}
}

func (node *PeerNode) Listen() {
	for {
		m, err := node.Subscription.Next(node.ctx)
		if err != nil {
			panic(err)
		}
		log.Println(m.ReceivedFrom, ": ", string(m.Message.Data))
		node.processData(m.Message.Data)
	}
}

func (node *PeerNode) StartFROST(message []byte) []byte {
	signatureShares = nil
	commitmentList = nil
	doneFROST = nil
	messageToSign = message

	p := Payload{
		Command: "COMMITMENT",
		Sender:  node.Host.ID().String(),
		Message: string(message),
		Package: []byte{},
	}

	log.Println("Starting FROST signing process")
	node.Publish(p)

	node.waitForCommitment()
	node.waitForSignatureShare()

	ok := <-doneFROST
	if ok {
		return finalSignature
	} else {
		return []byte{}
	}
}

var messageToSign []byte
var commitmentList frost.CommitmentList
var signatureShares []*frost.SignatureShare
var finalSignature []byte
var doneWaitCommitment chan bool
var doneWaitSignature chan bool
var doneFROST chan bool

func (node *PeerNode) processData(data []byte) {
	var decoded Payload
	err := json.Unmarshal(data, &decoded)
	if err != nil {
		log.Println("Invalid message:", err)
		return
	}

	if decoded.Sender == node.Host.ID().String() {
		return
	}

	switch decoded.Command {
	case "COMMITMENT":

		commitment := node.Frost.Commit()

		dataToPublish := Payload{
			Command: "COMMITMENT_SHARE",
			Sender:  node.Host.ID().String(),
			Message: decoded.Message,
			Package: commitment.Encode(),
		}

		node.Publish(dataToPublish)

	case "COMMITMENT_SHARE":

		commitment, err := signer.DecodeCommitment(decoded.Package)
		if err != nil {
			log.Println(err)
			return
		}

		commitmentList = append(commitmentList, &commitment)

	case "SIGN":
		commitmentList, err := signer.DecodeCommitmentList(decoded.Package)
		if err != nil {
			log.Println(err)
			return
		}

		sig, err := node.Frost.SignAsParticipant([]byte(decoded.Message), commitmentList)
		if err != nil {
			log.Println(err)
			return
		}

		dataToPublish := Payload{
			Command: "SIGNATURE_SHARE",
			Sender:  node.Host.ID().String(),
			Message: decoded.Message,
			Package: sig.Encode(),
		}

		node.Publish(dataToPublish)

	case "SIGNATURE_SHARE":
		configuration := frost.Secp256k1.Configuration()

		sigShare, err := configuration.DecodeSignatureShare(decoded.Package)
		if err != nil {
			log.Println(err)
			return
		}

		signatureShares = append(signatureShares, sigShare)

	default:
		log.Println("Ignoring message: ", string(decoded.Command))
	}
}

func (node *PeerNode) broadcastCommitmentList() {
	selfCommitment := node.Frost.Commit()
	commitmentList = append(commitmentList, &selfCommitment)

	dataToPublish := Payload{
		Command: "SIGN",
		Sender:  node.Host.ID().String(),
		Message: string(messageToSign),
		Package: commitmentList.Encode(),
	}

	node.Publish(dataToPublish)
}

func (node *PeerNode) aggregateSignature() {
	signature, err := node.Frost.SignAsCoordinator(messageToSign, commitmentList, signatureShares)
	if err != nil {
		log.Println(err)
		doneFROST <- false
		return
	}

	log.Printf("Aggregated signature: %x", signature.Encode())
	log.Println("Verify signature:", verifier.Verify(messageToSign, &signature, node.Frost.GroupPublicKey.Encode()))

	finalSignature = signature.Encode()
	doneFROST <- true
}

func (node *PeerNode) waitForCommitment() {
	go func() {
		for {
			if len(commitmentList) == node.Frost.N-1 {
				doneWaitCommitment <- true
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		select {
		case <-doneWaitCommitment:
			node.broadcastCommitmentList()
		case <-time.After(5 * time.Second):
			if len(commitmentList) >= node.Frost.T-1 {
				node.broadcastCommitmentList()
			} else {
				log.Println("Timeout for waiting commitment from other peers")
			}
		}
	}()
}

func (node *PeerNode) waitForSignatureShare() {
	go func() {
		for {
			if len(commitmentList) == node.Frost.N-1 {
				doneWaitSignature <- true
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		select {
		case <-doneWaitSignature:
			node.aggregateSignature()
		case <-time.After(10 * time.Second):
			if len(commitmentList) >= node.Frost.T-1 {
				node.aggregateSignature()
			} else {
				log.Println("Timeout for waiting signature share from other peers")
				doneFROST <- false
			}
		}
	}()
}
