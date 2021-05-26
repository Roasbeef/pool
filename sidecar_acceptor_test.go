package pool

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/lightninglabs/pool/account"
	"github.com/lightninglabs/pool/auctioneer"
	"github.com/lightninglabs/pool/clientdb"
	"github.com/lightninglabs/pool/internal/test"
	"github.com/lightninglabs/pool/order"
	"github.com/lightninglabs/pool/sidecar"
	"github.com/lightningnetwork/lnd/keychain"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_, providerPubKey = btcec.PrivKeyFromBytes(btcec.S256(), []byte{0x02})
	_, ourNodePubKey  = btcec.PrivKeyFromBytes(btcec.S256(), []byte{0x03})
	testOfferSig      = &btcec.Signature{
		R: new(big.Int).SetInt64(44),
		S: new(big.Int).SetInt64(22),
	}
)

func newTestDB(t *testing.T) (*clientdb.DB, func()) {
	tempDir, err := ioutil.TempDir("", "client-db")
	require.NoError(t, err)

	db, err := clientdb.New(tempDir, clientdb.DBFilename)
	if err != nil {
		require.NoError(t, os.RemoveAll(tempDir))
		t.Fatalf("unable to create new db: %v", err)
	}

	return db, func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(tempDir))
	}
}

// TestRegisterSidecar makes sure that registering a sidecar ticket verifies the
// offer signature contained within, adds the recipient node's information to
// the ticket and stores it to the local database.
func TestRegisterSidecar(t *testing.T) {
	t.Parallel()

	mockSigner := test.NewMockSigner()
	mockWallet := test.NewMockWalletKit()
	mockSigner.Signature = testOfferSig.Serialize()

	acceptor := NewSidecarAcceptor(&SidecarAcceptorConfig{
		SidecarDB:  nil,
		AcctDB:     nil,
		Signer:     mockSigner,
		Wallet:     mockWallet,
		NodePubKey: ourNodePubKey,
		ClientCfg:  auctioneer.Config{},
	})

	existingTicket, err := sidecar.NewTicket(
		sidecar.VersionDefault, 1_000_000, 200_000, 2016,
		providerPubKey, false,
	)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		ticket      *sidecar.Ticket
		expectedErr string
		check       func(t *testing.T, ticket *sidecar.Ticket)
	}{{
		name: "ticket with ID exists",
		ticket: &sidecar.Ticket{
			ID:    existingTicket.ID,
			State: sidecar.StateOffered,
			Offer: sidecar.Offer{
				SignPubKey:     providerPubKey,
				SigOfferDigest: testOfferSig,
			},
		},
		expectedErr: "ticket with ID",
	}, {
		name: "invalid sig",
		ticket: &sidecar.Ticket{
			ID:    [8]byte{1, 2, 3, 4},
			State: sidecar.StateOffered,
			Offer: sidecar.Offer{
				SignPubKey: providerPubKey,
				SigOfferDigest: &btcec.Signature{
					R: new(big.Int).SetInt64(33),
					S: new(big.Int).SetInt64(33),
				},
			},
		},
		expectedErr: "signature not valid for public key",
	}, {
		name: "all valid",
		ticket: &sidecar.Ticket{
			ID:    [8]byte{1, 2, 3, 4},
			State: sidecar.StateOffered,
			Offer: sidecar.Offer{
				SignPubKey:     providerPubKey,
				SigOfferDigest: testOfferSig,
			},
		},
		expectedErr: "",
		check: func(t *testing.T, ticket *sidecar.Ticket) {
			_, pubKey := test.CreateKey(0)

			require.NotNil(t, ticket.Recipient)
			require.Equal(
				t, ticket.Recipient.NodePubKey, ourNodePubKey,
			)
			require.Equal(
				t, ticket.Recipient.MultiSigPubKey, pubKey,
			)

			id := [8]byte{1, 2, 3, 4}
			newTicket, err := acceptor.cfg.SidecarDB.Sidecar(
				id, providerPubKey,
			)
			require.NoError(t, err)
			require.Equal(t, newTicket, ticket)
		},
	}}

	for _, tc := range testCases {
		tc := tc

		if tc.ticket != nil {
			digest, err := tc.ticket.OfferDigest()
			require.NoError(t, err)
			mockSigner.SignatureMsg = string(digest[:])
		}

		store, cleanup := newTestDB(t)
		err = store.AddSidecar(existingTicket)
		require.NoError(t, err)

		acceptor.cfg.SidecarDB = store
		ticket, err := acceptor.RegisterSidecar(
			context.Background(), *tc.ticket,
		)

		if tc.expectedErr == "" {
			require.NoError(t, err)

			tc.check(t, ticket)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErr)
		}

		cleanup()
	}
}

type mockMailBox struct {
	providerChan     chan *sidecar.Ticket
	providerMsgAck   chan struct{}
	providerDel      chan struct{}
	providerDropChan chan struct{}

	receiverChan     chan *sidecar.Ticket
	receiverMsgAck   chan struct{}
	receiverDel      chan struct{}
	receiverDropChan chan struct{}
}

func newMockMailBox() *mockMailBox {
	return &mockMailBox{
		providerChan:     make(chan *sidecar.Ticket),
		providerMsgAck:   make(chan struct{}),
		providerDel:      make(chan struct{}),
		providerDropChan: make(chan struct{}, 1),

		receiverChan:     make(chan *sidecar.Ticket),
		receiverMsgAck:   make(chan struct{}),
		receiverDel:      make(chan struct{}),
		receiverDropChan: make(chan struct{}, 1),
	}
}

func (m *mockMailBox) RecvSidecarPkt(ctx context.Context, pkt *sidecar.Ticket,
	provider bool) (*sidecar.Ticket, error) {

	var (
		recvChan chan *sidecar.Ticket
		dropChan chan struct{}
		ackChan  chan struct{}
	)
	if provider {
		recvChan = m.providerChan
		ackChan = m.providerMsgAck
		dropChan = m.providerDropChan
	} else {
		recvChan = m.receiverChan
		ackChan = m.receiverMsgAck
		dropChan = m.receiverDropChan
	}

recvMsg:
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("mailbox shutting down")

	case tkt := <-recvChan:
		tktCopy := *tkt

		select {
		case ackChan <- struct{}{}:

		// If we get a signal to drop the message, then we'll just go
		// back to receiving as normal.
		case <-dropChan:
			goto recvMsg
		}

		return &tktCopy, nil
	}
}

func (m *mockMailBox) SendSidecarPkt(ctx context.Context, pkt *sidecar.Ticket,
	provider bool) error {
	var sendChan chan *sidecar.Ticket
	if provider {
		sendChan = m.providerChan
	} else {
		sendChan = m.receiverChan
	}

	select {
	case <-ctx.Done():
	case sendChan <- pkt:
	}

	return nil
}

func (m *mockMailBox) InitSidecarMailbox(streamID [64]byte, ticket *sidecar.Ticket) error {
	return nil
}

func (m *mockMailBox) InitAcctMailbox(streamID [64]byte, pubKey *keychain.KeyDescriptor) error {
	return nil
}

func (m *mockMailBox) DelSidecarMailbox(streamID [64]byte, ticket *sidecar.Ticket) error {
	m.receiverDel <- struct{}{}
	return nil
}

func (m *mockMailBox) DelAcctMailbox(streamID [64]byte, pubKey *keychain.KeyDescriptor) error {
	m.providerDel <- struct{}{}
	return nil
}

type mockDriver struct {
	stateUpdates    chan sidecar.State
	bidSubmitted    chan struct{}
	ticketValidated chan struct{}
	channelExpected chan struct{}
}

func newMockDriver() *mockDriver {
	return &mockDriver{
		stateUpdates:    make(chan sidecar.State),
		bidSubmitted:    make(chan struct{}),
		ticketValidated: make(chan struct{}),
		channelExpected: make(chan struct{}),
	}
}

func (m *mockDriver) ValidateOrderedTicket(tkt *sidecar.Ticket) error {
	if tkt.State != sidecar.StateOrdered {
		return fmt.Errorf("sidecar not in state ordered: %v", tkt.State)
	}

	m.ticketValidated <- struct{}{}

	return nil
}

func (m *mockDriver) ExpectChannel(ctx context.Context, tkt *sidecar.Ticket) error {
	tkt.State = sidecar.StateExpectingChannel

	m.channelExpected <- struct{}{}

	return nil
}

func (m *mockDriver) UpdateSidecar(tkt *sidecar.Ticket) error {
	m.stateUpdates <- tkt.State

	return nil
}

func (m *mockDriver) SubmitSidecarOrder(tkt *sidecar.Ticket, bid *order.Bid,
	acct *account.Account) (*sidecar.Ticket, error) {

	tkt.State = sidecar.StateOrdered

	m.bidSubmitted <- struct{}{}

	return tkt, nil
}

type sidecarTestCtx struct {
	t *testing.T

	provider       *SidecarNegotiator
	providerDriver *mockDriver

	recipient       *SidecarNegotiator
	recipientDriver *mockDriver

	mailbox *mockMailBox
}

func (s *sidecarTestCtx) startNegotiators() error {
	if err := s.provider.Start(); err != nil {
		return err
	}

	return s.recipient.Start()
}

func (s *sidecarTestCtx) restartAllNegotiators() error {
	s.provider.Stop()
	s.recipient.Stop()

	s.provider.quit = make(chan struct{})
	s.recipient.quit = make(chan struct{})

	s.provider.cfg.StartingPkt.CurrentState = sidecar.State(s.provider.currentState)
	s.recipient.cfg.StartingPkt.CurrentState = sidecar.State(s.recipient.currentState)

	if err := s.provider.Start(); err != nil {
		return err
	}

	return s.recipient.Start()
}

func (s *sidecarTestCtx) restartProvider() error {
	s.provider.Stop()

	s.provider.quit = make(chan struct{})

	// When we restart the provider in isolation, we'll have their state be
	// mapped to the _created_ state (as the SidecarAcceptor would),
	// which'll cause them to retransmit their last message.
	s.provider.cfg.StartingPkt.CurrentState = sidecar.StateCreated

	return s.provider.Start()
}

func (s *sidecarTestCtx) restartRecipient() error {
	s.recipient.Stop()

	s.recipient.quit = make(chan struct{})

	s.recipient.cfg.StartingPkt.CurrentState = sidecar.State(s.recipient.currentState)

	return s.recipient.Start()
}

func (s *sidecarTestCtx) assertProviderMsgRecv() {
	s.t.Helper()

	select {
	case <-s.mailbox.providerMsgAck:
	case <-time.After(time.Second * 5):
		s.t.Fatalf("no provider msg received")
	}
}

func (s *sidecarTestCtx) assertRecipientMsgRecv() {
	s.t.Helper()

	select {
	case <-s.mailbox.receiverMsgAck:
	case <-time.After(time.Second * 5):
		s.t.Fatalf("no recipient msg received")
	}
}

func (s *sidecarTestCtx) dropReceiverMessage() {
	s.mailbox.receiverDropChan <- struct{}{}
}

func (s *sidecarTestCtx) dropProviderMessage() {
	s.mailbox.providerDropChan <- struct{}{}
}

func (s *sidecarTestCtx) assertNoProviderMsgsRecvd() {
	select {
	case <-s.mailbox.providerMsgAck:
		s.t.Fatalf("provider should've received no messages")
	case <-time.After(time.Second * 1):
	}
}

func (s *sidecarTestCtx) assertNoReceiverMsgsRecvd() {
	select {
	case <-s.mailbox.receiverMsgAck:
		s.t.Fatalf("receiver should've received no messages")
	case <-time.After(time.Second * 1):
	}
}

func (s *sidecarTestCtx) assertProviderTicketUpdated(expectedState sidecar.State) {
	select {
	case stateUpdate := <-s.providerDriver.stateUpdates:

		if stateUpdate != expectedState {
			s.t.Fatalf("expected state=%v, got: %v", expectedState,
				stateUpdate)
		}

	case <-time.After(time.Second * 5):
		s.t.Fatalf("provider ticket never updated")
	}
}

func (s *sidecarTestCtx) assertBidSubmited() {
	select {
	case <-s.providerDriver.bidSubmitted:
	case <-time.After(time.Second * 5):
		s.t.Fatalf("provider bid never submitted")
	}
}

func (s *sidecarTestCtx) assertRecipientTicketValidated() {
	select {
	case <-s.recipientDriver.ticketValidated:
	case <-time.After(time.Second * 5):
		s.t.Fatalf("recipient ticket never validated")
	}
}

func (s *sidecarTestCtx) assertRecipientExpectsChannel() {
	select {
	case <-s.recipientDriver.channelExpected:
	case <-time.After(time.Second * 5):
		s.t.Fatalf("recipient channel never expected")
	}
}

func (s *sidecarTestCtx) assertRecipientTicketUpdated(expectedState sidecar.State) {
	s.t.Helper()

	select {
	case stateUpdate := <-s.recipientDriver.stateUpdates:

		if stateUpdate != expectedState {
			s.t.Fatalf("expected state=%v, got: %v", expectedState,
				stateUpdate)
		}

	case <-time.After(time.Second * 5):
		s.t.Fatalf("recipient ticket never updated")
	}
}

func (s *sidecarTestCtx) confirmSidecarBatch() {
	s.provider.TicketExecuted()

	s.recipient.TicketExecuted()
}

func (s *sidecarTestCtx) assertProviderMailboxDel() {
	select {
	case <-s.mailbox.providerDel:
	case <-time.After(time.Second * 5):
		s.t.Fatalf("provider mailbox not deleted")
	}
}

func (s *sidecarTestCtx) assertRecipientMailboxDel() {
	select {
	case <-s.mailbox.receiverDel:
	case <-time.After(time.Second * 5):
		s.t.Fatalf("provider mailbox not deleted")
	}
}

func (s *sidecarTestCtx) assertNegotiatorStates(providerState, recepientState sidecar.State) {
	err := wait.Predicate(func() bool {
		return s.provider.CurrentState() == providerState
	}, time.Second*5)
	assert.NoError(s.t, err)
	err = wait.Predicate(func() bool {
		return s.recipient.CurrentState() == recepientState
	}, time.Second*5)
	assert.NoError(s.t, err)
}

func newSidecarTestCtx(t *testing.T) *sidecarTestCtx {
	mailBox := newMockMailBox()
	providerDriver := newMockDriver()
	recipientDriver := newMockDriver()

	ticketID := [8]byte{1}

	provider := NewSidecarNegotiator(AutoAcceptorConfig{
		Provider: true,
		ProviderBid: &order.Bid{
			Kit: order.Kit{
				Version:          order.VersionSelfChanBalance,
				LeaseDuration:    144,
				MaxBatchFeeRate:  253,
				MinUnitsMatch:    1,
				Amt:              0,
				UnitsUnfulfilled: 0,
			},
			SelfChanBalance: 1,
		},
		ProviderAccount: &account.Account{},
		StartingPkt: &SidecarPacket{
			CurrentState: sidecar.StateOffered,
			ProviderTicket: &sidecar.Ticket{
				ID:    ticketID,
				State: sidecar.StateOffered,
				Offer: sidecar.Offer{
					SigOfferDigest: testOfferSig,
				},
			},
			ReceiverTicket: &sidecar.Ticket{
				ID:    ticketID,
				State: sidecar.StateOffered,
				Offer: sidecar.Offer{
					SigOfferDigest: testOfferSig,
				},
			},
		},
		Driver:  providerDriver,
		MailBox: mailBox,
	})

	recipient := NewSidecarNegotiator(AutoAcceptorConfig{
		Provider: false,
		StartingPkt: &SidecarPacket{
			CurrentState: sidecar.StateRegistered,
			ProviderTicket: &sidecar.Ticket{
				ID:    ticketID,
				State: sidecar.StateRegistered,
				Offer: sidecar.Offer{
					SigOfferDigest: testOfferSig,
				},
			},
			ReceiverTicket: &sidecar.Ticket{
				ID:    ticketID,
				State: sidecar.StateRegistered,
				Offer: sidecar.Offer{
					SigOfferDigest: testOfferSig,
				},
			},
		},
		Driver:  recipientDriver,
		MailBox: mailBox,
	})

	return &sidecarTestCtx{
		t:               t,
		provider:        provider,
		providerDriver:  providerDriver,
		recipient:       recipient,
		recipientDriver: recipientDriver,
		mailbox:         mailBox,
	}
}

// TestAutoSidecarNegotiation tests the routine sidecar negotiation process
// including that both sides are able to properly handle retransmissions and
// also restarts assuming persistent storage is durable.
func TestAutoSidecarNegotiation(t *testing.T) {
	t.Parallel()

	testCtx := newSidecarTestCtx(t)

	// First, we'll start both negotiators. The provider should no-op, but
	// then the receiver should send over the ticket and complete
	// execution. At the end of the exchange, we expect that both sides are
	// waiting for the channel in its expectation state.
	err := testCtx.startNegotiators()
	assert.NoError(t, err, fmt.Errorf("unable to start negotiators: %v", err))

	// The recipient should send a new message to the provider with their
	// ticket in the registered state.
	testCtx.assertProviderMsgRecv()

	// Upon receiving the new ticket, the provider should write the new
	// registered state to disk, submit the bid, then then the new ticket
	// over to the recipient.
	testCtx.assertProviderTicketUpdated(sidecar.StateRegistered)
	testCtx.assertBidSubmited()
	testCtx.assertRecipientMsgRecv()

	// After sending the ticket, the provider should update the ticket in
	// its database as it waits in the expected state.
	testCtx.assertProviderTicketUpdated(sidecar.StateExpectingChannel)

	// The recipient, should now validate the ticket, then wait and expect
	// the channel.
	testCtx.assertRecipientTicketValidated()
	testCtx.assertRecipientExpectsChannel()

	// At this point, both sides should be waiting for the channel in its
	// expected state.
	testCtx.assertNegotiatorStates(
		sidecar.StateExpectingChannel, sidecar.StateExpectingChannel,
	)

	// We'll now simulate a restart on both sides by signalling their
	// goroutines to exit, then re-starting them anew with their persisted
	// state.
	err = testCtx.restartAllNegotiators()
	assert.NoError(t, err, fmt.Errorf("unable to restart negotiators: %v", err))

	// After the start, both sides should still show that they're expecting
	// the channel
	assert.Equal(
		t, sidecar.StateExpectingChannel, testCtx.provider.CurrentState(),
	)
	assert.Equal(
		t, sidecar.StateExpectingChannel, testCtx.recipient.CurrentState(),
	)

	// Finally there should be no additional message sent either since both
	// sides should now be in a terminal state
	testCtx.assertNoProviderMsgsRecvd()
	testCtx.assertNoReceiverMsgsRecvd()

	// The recipient of the ticket should re-expect the channel to
	// re-register with the auctioneer to ensure the channel can be
	// executed amidst their restarts.
	testCtx.assertRecipientExpectsChannel()

	// We'll now signal to both goroutines that the channel has been
	// finalized, at this point, we expect both ticket to transition to the
	// terminal state and the goroutines to exit.
	go testCtx.confirmSidecarBatch()

	// We expect that both sides now update their state one last time to
	// transition the ticket to a completed state, afterwards, they should
	// move clean up their mailboxes.
	testCtx.assertProviderTicketUpdated(sidecar.StateCompleted)
	testCtx.assertProviderMailboxDel()

	testCtx.assertRecipientTicketUpdated(sidecar.StateCompleted)
	testCtx.assertRecipientMailboxDel()

	// Once again, no messages should be received by either side.
	testCtx.assertNoProviderMsgsRecvd()
	testCtx.assertNoReceiverMsgsRecvd()
}

// TestAutoSidecarNegotiationRetransmission tests that if either side restarts,
// then the proper message is sent in order to ensure the negotiation state
// machine continues to be progressed.
func TestAutoSidecarNegotiationRetransmission(t *testing.T) {
	t.Parallel()

	testCtx := newSidecarTestCtx(t)

	// We'll start our negotiators as usual, however before we start them
	// we'll make sure that the message sent by the receiver is never
	// received by the provider.
	testCtx.dropProviderMessage()

	err := testCtx.startNegotiators()
	assert.NoError(t, err, fmt.Errorf("unable to start negotiators: %v", err))

	// At this point, both sides should still be in their starting state as
	// the initial message was never received.
	testCtx.assertNegotiatorStates(
		sidecar.StateOffered, sidecar.StateRegistered,
	)

	// We'll now restart only the provider. This should cause the provider
	// to retransmit a message of their offered ticket, which should cause
	// the recipient to re-send their registered ticket.
	//
	// In order to test the other retransmission case, we'll drop the
	// provider's message which carries the ticket in the ordered state.
	require.NoError(t, testCtx.restartProvider())
	testCtx.assertRecipientMsgRecv()

	// The provider receive the ticket, then update their local state as
	// normal.
	testCtx.assertProviderMsgRecv()
	testCtx.dropReceiverMessage()
	testCtx.assertProviderTicketUpdated(sidecar.StateRegistered)
	testCtx.assertBidSubmited()
	testCtx.assertProviderTicketUpdated(sidecar.StateExpectingChannel)

	// The provider should now have transitioned to the final state,
	// however the recipient should still be in their initial registered
	// state as they haven't received any messages yet.
	testCtx.assertNegotiatorStates(
		sidecar.StateExpectingChannel, sidecar.StateRegistered,
	)

	// We'll now restart the recipient, which should cause them to re-send
	// their registered ticket that'll cause the provider to re-send
	// _their_ ticket which should conclude the process with the ticket
	// being fully finalized.
	require.NoError(t, testCtx.restartRecipient())
	testCtx.assertProviderMsgRecv()

	testCtx.assertRecipientMsgRecv()
	testCtx.assertRecipientTicketValidated()
	testCtx.assertRecipientExpectsChannel()
}
