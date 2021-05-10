package pool

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/lightninglabs/pool/account"
	"github.com/lightninglabs/pool/clientdb"
	"github.com/lightninglabs/pool/order"
	"github.com/lightninglabs/pool/sidecar"
	"github.com/lightningnetwork/lnd/lnwire"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SidecarPacket encapsulates the current state of an auto sidecar negotiator.
// Note that the state of the negotiator, and the ticket may differ, this is
// what will trigger a state transition.
type SidecarPacket struct {
	// CurrentState is the current state of the negotiator.
	//
	// TODO(roasbeef): remove??
	CurrentState sidecar.State

	// ReceiverTicket is the current ticket of the receiver.
	ReceiverTicket *sidecar.Ticket

	// ProviderTicket is the current ticket of the provider.
	ProviderTicket *sidecar.Ticket
}

// deriveProviderStreamID derives the stream ID of the provider's cipher box,
// we'll use this to allow the recipient to send messages to the provider.
func deriveProviderStreamID(ticket *sidecar.Ticket) ([64]byte, error) {

	var streamID [64]byte

	// This stream ID will simply be the fixed 64-byte signature of our
	// sidecar ticket offer.
	wireSig, err := lnwire.NewSigFromRawSignature(
		ticket.Offer.SigOfferDigest.Serialize(),
	)
	if err != nil {
		return streamID, err
	}

	copy(streamID[:], wireSig[:])

	return streamID, nil
}

// deriveRecipientStreamID derives the stream ID of the cipher box that the
// provider of the sidecar ticket will use to send messages to the receiver.
func deriveRecipientStreamID(ticket *sidecar.Ticket) [64]byte {
	receiverMultisig := ticket.Recipient.NodePubKey.SerializeCompressed()
	receiverNode := ticket.Recipient.MultiSigPubKey.SerializeCompressed()

	// The stream ID will be the concentration of the receiver's multi-sig
	// and node keys, ignoring the first byte of each key that essentially
	// communicates parity information.
	var (
		streamID [64]byte
		n        int
	)
	n += copy(streamID[:], receiverMultisig[1:])
	copy(streamID[n:], receiverNode[1:])

	return streamID
}

// deriveStreamID derives corresponding stream ID for the provider of the
// receiver based on the passed sidecar ticket.
func deriveStreamID(ticket *sidecar.Ticket, provider bool) ([64]byte, error) {
	if provider {
		return deriveProviderStreamID(ticket)
	}

	return deriveRecipientStreamID(ticket), nil
}

// sendSidecarPkt attempts to send a sidecar packet to the opposite party using
// their registered cipherbox stream.
func (a *SidecarAcceptor) sendSidecarPkt(pkt *sidecar.Ticket,
	provider bool) error {

	var ticketBuf bytes.Buffer
	err := sidecar.SerializeTicket(&ticketBuf, pkt)
	if err != nil {
		return err
	}

	streamID, err := deriveStreamID(pkt, provider)
	if err != nil {
		return err
	}

	target := "receiver"
	if provider {
		target = "provider"
	}

	log.Infof("Sending ticket(state=%v, id=%x) to %v stream_id=%x",
		pkt.State, pkt.ID[:], target, streamID[:])

	return a.client.SendCipherBoxMsg(
		context.Background(), streamID, ticketBuf.Bytes(),
	)
}

// recvSidecarPkt attempts to receive a new sidecar packet from the opposite
// party using their registered cipherbox stream.
func (a *SidecarAcceptor) recvSidecarPkt(ticket *sidecar.Ticket,
	provider bool) (*sidecar.Ticket, error) {

	streamID, err := deriveStreamID(ticket, provider)
	if err != nil {
		return nil, err
	}

	log.Infof("Waiting for ticket (id=%x) using stream_id=%x, provider=%v",
		ticket.ID[:], streamID[:], provider)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msg, err := a.client.RecvCipherBoxMsg(ctx, streamID)
	if err != nil {
		return nil, fmt.Errorf("unable to recv cipher box "+
			"msg: %w", err)
	}

	log.Infof("Receive new message for ticket (id=%x) "+
		"via stream_id=%x, provider=%v", ticket.ID[:], streamID,
		provider)

	return sidecar.DeserializeTicket(bytes.NewReader(msg))
}

// isErrAlreadyExists returns true if the passed error is the "already exists"
// error within the error wrapped error which is returned by the hash mail
// server when a stream we're attempting to create already exists.
func isErrAlreadyExists(err error) bool {
	statusCode, ok := status.FromError(err)
	if !ok {
		return false
	}

	return statusCode.Code() == codes.AlreadyExists
}

// autoSidecarReceiver is a goroutine that will attempt to advance a new
// sidecar ticket through the process until it reaches its final state.
func (a *SidecarAcceptor) autoSidecarReceiver(startingPkt *SidecarPacket) {
	defer a.wg.Done()

	packetChan := make(chan *sidecar.Ticket, 1)
	cancelChan := make(chan struct{})

	currentState := startingPkt.CurrentState
	localTicket := startingPkt.ReceiverTicket

	// We'll start with a simulated starting message from the sidecar
	// provider.
	packetChan <- startingPkt.ProviderTicket

	// Before we enter our main read loop below, we'll attempt to re-create
	// out mailbox as the recipient.
	recipientStreamID := deriveRecipientStreamID(
		localTicket,
	)

	log.Infof("Creating receiver reply mailbox for ticket=%x, "+
		"stream_id=%x", startingPkt.ReceiverTicket.ID[:],
		recipientStreamID[:])

	err := a.client.InitTicketCipherBox(
		context.Background(), recipientStreamID,
		startingPkt.ReceiverTicket,
	)
	if err != nil && !isErrAlreadyExists(err) {
		log.Errorf("unable to init cipher box: %v", err)
		return
	}

	// Launch a goroutine to continually read new packets off the wire and
	// send them to our state step routine. We'll always read packets until
	// things are finished, as the other side may retransmit messages until
	// the process has been finalized.
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		// We'll continue to read out new messages from the cipherbox
		// stream and deliver them to the main gorotuine until we
		// receive a message over the cancel channel.
		for {
			newTicket, err := a.recvSidecarPkt(
				startingPkt.ReceiverTicket, false,
			)
			if err != nil {
				log.Error(err)
				return
			}

			select {
			case packetChan <- newTicket:

			case <-cancelChan:
				return
			case <-a.quit:
				return
			}

		}
	}()

	for {
		select {

		case newTicket := <-packetChan:
			newPktState, err := a.stateStepRecipient(&SidecarPacket{
				CurrentState:   currentState,
				ProviderTicket: newTicket,
				ReceiverTicket: localTicket,
			})
			if err != nil {
				log.Errorf("unable to transition state: %v", err)
				continue
			}

			currentState = newPktState.CurrentState
			localTicket = newPktState.ReceiverTicket

			// If our next target state is the completion state,
			// then our job here is done, and we can safely exit
			// this main goroutine.
			if newPktState.CurrentState == sidecar.StateCompleted {
				log.Infof("Receiver negotiation for " +
					"SidecarTicket(%x) complete!")

				close(cancelChan)
				return
			}

		case <-a.quit:
			return
		}
	}
}

// stateStepRecipient is a state transition function that will walk the
// receiver through the sidecar negotiation process. It takes the current state
// (the state of the goroutine, and the incoming ticket) and maps that into a
// new state, with a possibly modified ticket.
func (a *SidecarAcceptor) stateStepRecipient(pkt *SidecarPacket,
) (*SidecarPacket, error) {

	switch {

	// If the state of the ticket shows up as offered, then this is the
	// remote party restarting and requesting we re-send our registered
	// ticket. So we'll fall through to our "starting" state below to
	// re-send them the packet.
	case pkt.ProviderTicket.State == sidecar.StateOffered:
		log.Infof("Provider retransmitted initial offer, re-sending "+
			"registered ticket=%x", pkt.ProviderTicket.ID[:])

		fallthrough

	// In this state, they've just sent us their version of the ticket w/o
	// our node information (and processed it adding our information),
	// we'll populate it then send it to them over the cipherbox they've
	// created for this purpose.
	case pkt.CurrentState == sidecar.StateRegistered &&
		pkt.ReceiverTicket.State == sidecar.StateRegistered &&
		pkt.ProviderTicket.State == sidecar.StateRegistered:

		log.Infof("Transmitting registered ticket=%x to provider",
			pkt.ProviderTicket.ID[:])

		err := a.sendSidecarPkt(pkt.ReceiverTicket, true)
		if err != nil {
			return nil, fmt.Errorf("unable to send pkt: %w", err)
		}

		// We'll return a new packet that should reflect our state
		// after the above message is sent: both parties have the
		// ticket in the registered state.
		return &SidecarPacket{
			CurrentState:   sidecar.StateRegistered,
			ReceiverTicket: pkt.ReceiverTicket,
			ProviderTicket: pkt.ReceiverTicket,
		}, nil

	// This is effectively our final state transition: we're waiting with a
	// local registered ticket and receive a ticket in the ordered state.
	// We'll validate the ticket and start expecting the channel and
	// transition to our final state.
	case pkt.CurrentState == sidecar.StateRegistered &&
		(pkt.ProviderTicket.State == sidecar.StateOrdered ||
			pkt.ProviderTicket.State == sidecar.StateExpectingChannel):

		// At this point, we'll finish validating the ticket, then
		// await the ticket on the side lines if it's valid.
		ctx := context.Background()
		err := validateOrderedTicket(
			ctx, pkt.ProviderTicket, a.cfg.Signer, a.cfg.SidecarDB,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to verify ticket: "+
				"%w", err)
		}

		log.Infof("Auto negotiation for ticket=%x complete! Expecting "+
			"channel...", pkt.ProviderTicket.ID[:])

		// Now that we know the channel is valid, we'll wait for the
		// channel to show up at our node, and allow things to advance
		// to the completion state.
		err = a.ExpectChannel(ctx, pkt.ProviderTicket)
		if err != nil {
			return nil, fmt.Errorf("failed to expect "+
				"channel: %w", err)
		}

		// TODO(roasbeef): set state to expecting channel?

		return &SidecarPacket{
			CurrentState:   sidecar.StateExpectingChannel,
			ReceiverTicket: pkt.ProviderTicket,
			ProviderTicket: pkt.ProviderTicket,
		}, nil

	// If we fall through here, then either we read a buffered message or
	// the remote party isn't following the protocol, so we'll just ignore
	// it.
	default:
		return nil, fmt.Errorf("unhandled receiver state transition "+
			"for ticket=%v, state=%v", pkt.ProviderTicket.ID[:],
			pkt.ProviderTicket.State)
	}
}

// autoSidecarProvider is a goroutine that will attempt to advance a new
// sidecar ticket through the negotiation process until it reaches its final
// state.
func (a *SidecarAcceptor) autoSidecarProvider(startingPkt *SidecarPacket,
	bid *order.Bid, acct *account.Account) {

	defer a.wg.Done()

	// TODO(roasbeef): subscribe to order state so know when things are
	// done, use that to send the extra msg

	packetChan := make(chan *sidecar.Ticket, 1)
	cancelChan := make(chan struct{})

	currentState := startingPkt.CurrentState
	localTicket := startingPkt.ProviderTicket

	// We'll start with a simulated starting message from the sidecar
	// receiver.
	packetChan <- startingPkt.ReceiverTicket

	// First, we'll need to derive the stream ID that we'll use to receive
	// new messages from the recipient.
	streamID, err := deriveProviderStreamID(localTicket)
	if err != nil {
		log.Errorf("unable to derive stream_id for ticket=%x",
			localTicket.ID[:])
		return
	}

	log.Infof("Creating provider mailbox for ticket=%x, w/ stream_id=%x",
		localTicket.ID[:], streamID[:])

	err = a.client.InitAccountCipherBox(
		context.Background(), streamID, acct.TraderKey,
	)
	if err != nil && !isErrAlreadyExists(err) {
		log.Errorf("unable to init cipher box: %v", err)
		return
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		// We'll continue to read out new messages from the cipherbox
		// stream and deliver them to the main gorotuine until we
		// receive a message over the cancel channel.
		for {
			newTicket, err := a.recvSidecarPkt(
				startingPkt.ProviderTicket, true,
			)
			if err != nil {
				log.Error(err)
				return
			}

			select {
			case packetChan <- newTicket:

			case <-cancelChan:
				return
			case <-a.quit:
				return
			}

		}
	}()

	for {
		select {
		case newTicket := <-packetChan:
			// The provider has more states it needs to transition
			// through, so we'll continue until we end up at the
			// same state (a noop)
			for {
				priorState := currentState

				log.Infof("step=%v", currentState)

				newPktState, err := a.stateStepProvider(&SidecarPacket{
					CurrentState:   currentState,
					ReceiverTicket: newTicket,
					ProviderTicket: localTicket,
				}, bid, acct)
				if err != nil {
					log.Errorf("unable to transition state: %v", err)
					break
				}

				currentState = newPktState.CurrentState
				localTicket = newPktState.ProviderTicket

				switch {
				case priorState == currentState:
					fallthrough
				case currentState == sidecar.StateExpectingChannel:
					break

				// If our next target state is the completion
				// state, then our job here is done, and we can
				// safely exit this main goroutine.
				case newPktState.CurrentState ==
					sidecar.StateCompleted:

					log.Infof("Receiver negotiation for " +
						"SidecarTicket(%x) complete!")

					close(cancelChan)
					return
				}
			}

		case <-a.quit:
			return
		}
	}
}

// stateStepProvider is the state transition function for the provider of a
// sidecar ticket. It takes the current transcript state, the provider's
// account, and canned bid and returns a new transition to a new ticket state.
func (a *SidecarAcceptor) stateStepProvider(pkt *SidecarPacket, bid *order.Bid,
	acct *account.Account) (*SidecarPacket, error) {

	switch {
	// In this case, we've just restarted, so we'll attempt to start from
	// scratch by sending the recipient a packet that has our ticket in the
	// offered state. This signals to them we never wrote the registered
	// ticket and need it again.
	case pkt.CurrentState == sidecar.StateCreated &&
		pkt.ProviderTicket.State == sidecar.StateOffered:

		log.Infof("Resuming negotiation for ticket=%x, requesting "+
			"registered ticket", pkt.ProviderTicket.ID[:])

		err := a.sendSidecarPkt(pkt.ProviderTicket, false)
		if err != nil {
			return nil, err
		}

		return &SidecarPacket{
			CurrentState:   sidecar.StateOffered,
			ReceiverTicket: pkt.ProviderTicket,
			ProviderTicket: pkt.ReceiverTicket,
		}, nil

	// In this state, we've just started anew, and have received a ticket
	// from the receiver with their node information. We'll write this to
	// disk, then transition to the next state.
	//
	// Transition: -> StateRegistered
	case pkt.CurrentState == sidecar.StateOffered &&
		pkt.ReceiverTicket.State == sidecar.StateRegistered:

		log.Infof("Received registered ticket=%x from recipient",
			pkt.ReceiverTicket.ID[:])

		// Now that we have the ticket, we'll update the state on disk
		// to checkpoint the new state.
		err := a.cfg.SidecarDB.UpdateSidecar(pkt.ReceiverTicket)
		if err != nil {
			return nil, fmt.Errorf("unable to update ticket: %w",
				err)
		}

		return &SidecarPacket{
			CurrentState:   sidecar.StateRegistered,
			ReceiverTicket: pkt.ReceiverTicket,
			ProviderTicket: pkt.ReceiverTicket,
		}, nil

	// If we're in this state (possibly after a restart), we have all the
	// information we need to submit the order, so we'll do that, then send
	// the finalized ticket back to the recipient.
	//
	// Transition: -> StateOrdered
	case pkt.CurrentState == sidecar.StateRegistered:

		log.Infof("Submitting bid order for ticket=%x",
			pkt.ProviderTicket.ID[:])

		// Now we have the recipient's information, we can attach it to
		// our bid, and submit it as normal.
		updatedTicket, err := a.submitSidecarOrder(
			context.Background(), pkt.ProviderTicket, bid, acct,
		)
		switch {
		// If the order has already been submitted, then we'll catch
		// this error and go to the next state. Submitting the order
		// doesn't persist the state update to the ticket, so we don't
		// risk a split brain state.
		case err == nil:
		case errors.Is(err, clientdb.ErrOrderExists):

		case err != nil:
			return nil, fmt.Errorf("unable to submit sidecar "+
				"order: %v", err)
		}

		return &SidecarPacket{
			CurrentState:   sidecar.StateOrdered,
			ReceiverTicket: updatedTicket,
			ProviderTicket: updatedTicket,
		}, nil

	// In this state, we've already sent over the final ticket, but the
	// other party is requesting a re-transmission.
	case pkt.CurrentState == sidecar.StateExpectingChannel &&
		pkt.ReceiverTicket.State == sidecar.StateRegistered:

		fallthrough

	// In this state, we've submitted the order and now need to send back
	// the completed order to the recipient so they can expect the ultimate
	// sidecar channel. Notice that we don't persist this state, as upon
	// restart we'll always re-send the ticket to the other party until
	// things are finalized.
	//
	// Transition: -> StateExpectingChannel
	case pkt.CurrentState == sidecar.StateOrdered:

		log.Infof("Sending finalize ticket=%x to receiver, entering "+
			"final stage", pkt.ProviderTicket.ID[:])

		err := a.sendSidecarPkt(pkt.ProviderTicket, false)
		if err != nil {
			return nil, fmt.Errorf("unable to send sidecar "+
				"pkt: %v", err)
		}

		updatedTicket := *pkt.ProviderTicket
		updatedTicket.State = sidecar.StateExpectingChannel

		// Now that we have the final ticket, we'll update the state on
		// disk to checkpoint the new state. If the remote party ends
		// us any messages after we persist this state, then we'll
		// simply re-send the latest ticket.
		err = a.cfg.SidecarDB.UpdateSidecar(&updatedTicket)
		if err != nil {
			return nil, fmt.Errorf("unable to update ticket: %w",
				err)
		}

		log.Infof("Negotiation for ticket=%x has been "+
			"completed!", pkt.ProviderTicket.ID[:])

		return &SidecarPacket{
			CurrentState:   sidecar.StateExpectingChannel,
			ReceiverTicket: &updatedTicket,
			ProviderTicket: &updatedTicket,
		}, nil

	default:
		return nil, fmt.Errorf("unhandled provider state "+
			"transition ticket=%x, state=%v",
			pkt.ReceiverTicket.ID[:], pkt.ReceiverTicket.State)
	}
}
