package tss

import (
	"context"
	"encoding/json"
	"flag"
	"log"
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

// TODO: make these configurable
const commitmentTimeout time.Duration = 10 * time.Second
const signatureTimeout time.Duration = 20 * time.Second
const receivingTimeout time.Duration = commitmentTimeout + signatureTimeout

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

	t, n, s, p, g, err := LoadFrostKey()
	if err != nil {
		panic(err)
	}
	frostInstance := signer.NewFromStaticKeys(h.ID().String(), t, n, s, p, g)

	discoverPeers(ctx, h)

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}
	topic, err := ps.Join(*topicNameFlag)
	if err != nil {
		panic(err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}

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
		node.processData(m.Message.Data, m.ReceivedFrom.String())
	}
}

var messageToSign []byte    // message as coordinator
var messageToReceive []byte // message as participant
var commitmentList frost.CommitmentList
var signatureShares []*frost.SignatureShare
var finalSignature []byte
var doneWaitCommitment chan bool
var doneWaitSignature chan bool
var doneFROST chan bool
var doneReceiving chan bool
var mu sync.Mutex

func (node *PeerNode) StartFROST(message []byte) []byte {

	doneFROST = make(chan bool)
	doneWaitCommitment = make(chan bool)
	doneWaitSignature = make(chan bool)

	mu.Lock()
	signatureShares = nil
	commitmentList = nil
	messageToSign = message
	mu.Unlock()

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

func (node *PeerNode) processData(data []byte, source string) {
	var decoded Payload
	err := json.Unmarshal(data, &decoded)
	if err != nil {
		log.Println("Invalid message:", err)
		return
	}

	// ignore if sender ID is different with original source
	if decoded.Sender != source {
		log.Printf("Impersonation from %s to %s detected, ignoring message", source, decoded.Sender)
		return
	}

	// ignore if sender is self
	if decoded.Sender == node.Host.ID().String() {
		return
	}

	switch decoded.Command {
	case "COMMITMENT":

		if messageToReceive == nil {
			log.Printf("Receiving message %s to be signed", decoded.Message)
			mu.Lock()
			messageToReceive = []byte(decoded.Message)
			mu.Unlock()
		} else {
			log.Printf("Receiving %s while processing %s, ignoring request", decoded.Message, string(messageToReceive))
			return
		}

		// TODO: implement rejection scenario in case of invalid message

		doneReceiving = make(chan bool)
		node.waitAsParticipant()

		commitment := node.Frost.Commit()

		dataToPublish := Payload{
			Command: "COMMITMENT_SHARE",
			Sender:  node.Host.ID().String(),
			Message: decoded.Message,
			Package: commitment.Encode(),
		}

		node.Publish(dataToPublish)

	case "COMMITMENT_SHARE":

		if decoded.Message != string(messageToSign) {
			log.Printf("Receiving %s while processing %s, ignoring request", decoded.Message, string(messageToSign))
			return
		}

		commitment, err := signer.DecodeCommitment(decoded.Package)
		if err != nil {
			log.Println(err)
			return
		}

		mu.Lock()
		commitmentList = append(commitmentList, &commitment)
		mu.Unlock()

	case "SIGN":

		if decoded.Message != string(messageToReceive) {
			log.Printf("Receiving %s while processing %s, ignoring request", decoded.Message, string(messageToReceive))
			return
		}

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
		doneReceiving <- true

	case "SIGNATURE_SHARE":

		if decoded.Message != string(messageToSign) {
			log.Printf("Receiving %s while processing %s, ignoring request", decoded.Message, string(messageToSign))
			return
		}

		configuration := frost.Secp256k1.Configuration()

		sigShare, err := configuration.DecodeSignatureShare(decoded.Package)
		if err != nil {
			log.Println(err)
			return
		}

		mu.Lock()
		signatureShares = append(signatureShares, sigShare)
		mu.Unlock()

	default:
		log.Println("Ignoring message: ", string(decoded.Command))
	}
}

func (node *PeerNode) broadcastCommitmentList() {
	selfCommitment := node.Frost.Commit()
	mu.Lock()
	commitmentList = append(commitmentList, &selfCommitment)
	mu.Unlock()

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
			if len(commitmentList) >= node.Frost.N-1 {
				doneWaitCommitment <- true
				break
			}
		}
	}()

	go func() {
		select {
		case <-doneWaitCommitment:
			node.broadcastCommitmentList()
		case <-time.After(commitmentTimeout):
			// After certain seconds, process should continue
			// if collected commitments are above threshold (including self), continue
			// else raise timeout
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
			if len(signatureShares) >= node.Frost.N-1 {
				doneWaitSignature <- true
				break
			}
		}
	}()

	go func() {
		select {
		case <-doneWaitSignature:
			node.aggregateSignature()
		case <-time.After(signatureTimeout):
			// After certain seconds, process should continue
			// if collected signatureShares are above threshold (including self), continue
			// else raise timeout
			if len(signatureShares) >= node.Frost.T-1 {
				node.aggregateSignature()
			} else {
				log.Println("Timeout for waiting signature share from other peers")
				doneFROST <- false
			}
		}
	}()
}

func (node *PeerNode) waitAsParticipant() {
	go func() {
		select {
		case <-doneReceiving:
			mu.Lock()
			messageToReceive = nil
			mu.Unlock()
		case <-time.After(receivingTimeout):
			mu.Lock()
			messageToReceive = nil
			mu.Unlock()
		}
	}()
}
