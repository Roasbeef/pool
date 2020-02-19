package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
	"github.com/lightninglabs/agora/client/account"
	"github.com/lightninglabs/agora/client/auctioneer"
	"github.com/lightninglabs/agora/client/clientdb"
	"github.com/lightninglabs/agora/client/clmrpc"
	"github.com/lightninglabs/agora/client/order"
	"github.com/lightninglabs/loop/lndclient"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwallet/chainfee"
)

const (
	// getInfoTimeout is the maximum time we allow for the initial getInfo
	// call to the connected lnd node.
	getInfoTimeout = 5 * time.Second
)

// rpcServer implements the gRPC server on the client side and answers RPC calls
// from an end user client program like the command line interface.
type rpcServer struct {
	started uint32 // To be used atomically.
	stopped uint32 // To be used atomically.

	// bestHeight is the best known height of the main chain. This MUST be
	// used atomically.
	bestHeight uint32

	server         *Server
	lndServices    *lndclient.LndServices
	auctioneer     *auctioneer.Client
	db             *clientdb.DB
	accountManager *account.Manager
	orderManager   *order.Manager

	quit chan struct{}
	wg   sync.WaitGroup
}

// newRPCServer creates a new client-side RPC server that uses the given
// connection to the trader's lnd node and the auction server. A client side
// database is created in `serverDir` if it does not yet exist.
func newRPCServer(server *Server, serverDir string) (*rpcServer, error) {

	db, err := clientdb.New(serverDir)
	if err != nil {
		return nil, err
	}

	lnd := &server.lndServices.LndServices
	return &rpcServer{
		server:      server,
		lndServices: lnd,
		auctioneer:  server.AuctioneerClient,
		db:          db,
		accountManager: account.NewManager(&account.ManagerConfig{
			Store:         db,
			Auctioneer:    server.AuctioneerClient,
			Wallet:        lnd.WalletKit,
			Signer:        lnd.Signer,
			ChainNotifier: lnd.ChainNotifier,
			TxSource:      lnd.Client,
		}),
		orderManager: order.NewManager(&order.ManagerConfig{
			Store:     db,
			Lightning: lnd.Client,
			Wallet:    lnd.WalletKit,
			Signer:    lnd.Signer,
		}),
		quit: make(chan struct{}),
	}, nil
}

// Start starts the rpcServer, making it ready to accept incoming requests.
func (s *rpcServer) Start() error {
	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return nil
	}

	log.Infof("Starting trader server")

	ctx := context.Background()

	lndCtx, lndCancel := context.WithTimeout(ctx, getInfoTimeout)
	defer lndCancel()
	info, err := s.lndServices.Client.GetInfo(lndCtx)
	if err != nil {
		return fmt.Errorf("error in GetInfo: %v", err)
	}

	log.Infof("Connected to lnd node %v with pubkey %v", info.Alias,
		hex.EncodeToString(info.IdentityPubkey[:]))

	chainNotifier := s.lndServices.ChainNotifier
	blockChan, blockErrChan, err := chainNotifier.RegisterBlockEpochNtfn(ctx)
	if err != nil {
		return err
	}

	var height int32
	select {
	case height = <-blockChan:
	case err := <-blockErrChan:
		return fmt.Errorf("unable to receive first block "+
			"notification: %v", err)
	case <-ctx.Done():
		return nil
	}

	s.updateHeight(height)

	// Start the auctioneer client first to establish a connection.
	if err := s.auctioneer.Start(); err != nil {
		return fmt.Errorf("unable to start auctioneer client: %v", err)
	}

	// Start managers.
	if err := s.accountManager.Start(); err != nil {
		return fmt.Errorf("unable to start account manager: %v", err)
	}
	if err := s.orderManager.Start(); err != nil {
		return fmt.Errorf("unable to start order manager: %v", err)
	}

	s.wg.Add(1)
	go s.serverHandler(blockChan, blockErrChan)

	log.Infof("Trader server is now active")

	return nil
}

// Stop stops the server.
func (s *rpcServer) Stop() error {
	if !atomic.CompareAndSwapUint32(&s.stopped, 0, 1) {
		return nil
	}

	log.Info("Trader server stopping")
	s.accountManager.Stop()
	s.orderManager.Stop()
	err := s.db.Close()
	if err != nil {
		log.Errorf("Error closing DB: %v")
	}
	err = s.auctioneer.Stop()
	if err != nil {
		log.Errorf("Error closing server stream: %v")
	}

	close(s.quit)
	s.wg.Wait()

	log.Info("Stopped trader server")
	return nil
}

// serverHandler is the main event loop of the server.
func (s *rpcServer) serverHandler(blockChan chan int32, blockErrChan chan error) {
	defer s.wg.Done()

	for {
		select {
		case msg := <-s.auctioneer.FromServerChan:
			// An empty message means the client is shutting down.
			if msg == nil {
				continue
			}

			log.Debugf("Received message from the server: %v", msg)
			err := s.handleServerMessage(msg)
			if err != nil {
				log.Errorf("Error handling server message: %v",
					err)
				err := s.server.Stop()
				if err != nil {
					log.Errorf("Error shutting down: %v",
						err)
				}
			}

		case err := <-s.auctioneer.StreamErrChan:
			// If the server is shutting down, then the client has
			// already scheduled a restart. We only need to handle
			// other errors here.
			if err != nil && err != auctioneer.ErrServerShutdown {
				log.Errorf("Error in server stream: %v", err)
				err := s.auctioneer.HandleServerShutdown(err)
				if err != nil {
					log.Errorf("Error closing stream: %v",
						err)
				}
			}

		case height := <-blockChan:
			log.Infof("Received new block notification: height=%v",
				height)
			s.updateHeight(height)

		case err := <-blockErrChan:
			if err != nil {
				log.Errorf("Unable to receive block "+
					"notification: %v", err)
				err := s.server.Stop()
				if err != nil {
					log.Errorf("Error shutting down: %v",
						err)
				}
			}

		// In case the server is shutting down.
		case <-s.quit:
			return
		}
	}
}

func (s *rpcServer) updateHeight(height int32) {
	// Store height atomically so the incoming request handler can access it
	// without locking.
	atomic.StoreUint32(&s.bestHeight, uint32(height))
}

// handleServerMessage reads a gRPC message received in the stream from the
// auctioneer server and passes it to the correct manager.
func (s *rpcServer) handleServerMessage(rpcMsg *clmrpc.ServerAuctionMessage) error {
	switch msg := rpcMsg.Msg.(type) {
	// A new batch has been assembled with some of our orders.
	case *clmrpc.ServerAuctionMessage_Prepare:
		log.Tracef("Received prepare msg from server, batch_id=%x: %v",
			msg.Prepare.BatchId, spew.Sdump(msg))

		// TODO(guggero): Add real batch validation here.
		// For now, we just send the accept back.
		err := s.auctioneer.SendAuctionMessage(&clmrpc.ClientAuctionMessage{
			Msg: &clmrpc.ClientAuctionMessage_Accept{
				Accept: &clmrpc.OrderMatchAccept{
					BatchId: msg.Prepare.BatchId,
				},
			},
		})
		if err != nil {
			return err
		}

		// TODO(guggero): Initiate channel opening negotiation with
		// remote peer here.
		err = s.auctioneer.SendAuctionMessage(&clmrpc.ClientAuctionMessage{
			Msg: &clmrpc.ClientAuctionMessage_Sign{
				Sign: &clmrpc.OrderMatchSign{
					BatchId: msg.Prepare.BatchId,
				},
			},
		})
		if err != nil {
			return err
		}

	case *clmrpc.ServerAuctionMessage_Finalize:
		log.Tracef("Received finalize msg from server, batch_id=%x: %v",
			msg.Finalize.BatchId, spew.Sdump(msg))

	default:
		return fmt.Errorf("unknown server message: %v", msg)
	}

	return nil
}

func (s *rpcServer) InitAccount(ctx context.Context,
	req *clmrpc.InitAccountRequest) (*clmrpc.Account, error) {

	account, err := s.accountManager.InitAccount(
		ctx, btcutil.Amount(req.AccountValue), req.AccountExpiry,
		atomic.LoadUint32(&s.bestHeight),
	)
	if err != nil {
		return nil, err
	}

	return marshallAccount(account)
}

func (s *rpcServer) ListAccounts(ctx context.Context,
	req *clmrpc.ListAccountsRequest) (*clmrpc.ListAccountsResponse, error) {

	accounts, err := s.db.Accounts()
	if err != nil {
		return nil, err
	}

	rpcAccounts := make([]*clmrpc.Account, 0, len(accounts))
	for _, account := range accounts {
		rpcAccount, err := marshallAccount(account)
		if err != nil {
			return nil, err
		}
		rpcAccounts = append(rpcAccounts, rpcAccount)
	}

	return &clmrpc.ListAccountsResponse{
		Accounts: rpcAccounts,
	}, nil
}

func marshallAccount(a *account.Account) (*clmrpc.Account, error) {
	var rpcState clmrpc.AccountState
	switch a.State {
	case account.StateInitiated, account.StatePendingOpen:
		rpcState = clmrpc.AccountState_PENDING_OPEN

	case account.StateOpen:
		rpcState = clmrpc.AccountState_OPEN

	case account.StateExpired:
		rpcState = clmrpc.AccountState_EXPIRED

	case account.StatePendingClosed:
		rpcState = clmrpc.AccountState_PENDING_CLOSED

	case account.StateClosed:
		rpcState = clmrpc.AccountState_CLOSED

	default:
		return nil, fmt.Errorf("unknown state %v", a.State)
	}

	var closeTxHash chainhash.Hash
	if a.CloseTx != nil {
		closeTxHash = a.CloseTx.TxHash()
	}

	return &clmrpc.Account{
		TraderKey: a.TraderKey.PubKey.SerializeCompressed(),
		Outpoint: &clmrpc.OutPoint{
			Txid:        a.OutPoint.Hash[:],
			OutputIndex: a.OutPoint.Index,
		},
		Value:            uint32(a.Value),
		ExpirationHeight: a.Expiry,
		State:            rpcState,
		CloseTxid:        closeTxHash[:],
	}, nil
}

func (s *rpcServer) CloseAccount(ctx context.Context,
	req *clmrpc.CloseAccountRequest) (*clmrpc.CloseAccountResponse, error) {

	traderKey, err := btcec.ParsePubKey(req.TraderKey, btcec.S256())
	if err != nil {
		return nil, err
	}

	var closeOutputs []*wire.TxOut
	if len(req.Outputs) > 0 {
		closeOutputs = make([]*wire.TxOut, 0, len(req.Outputs))
		for _, output := range req.Outputs {
			// Make sure they've provided a valid output script.
			_, err := txscript.ParsePkScript(output.Script)
			if err != nil {
				return nil, err
			}

			closeOutputs = append(closeOutputs, &wire.TxOut{
				Value:    int64(output.Value),
				PkScript: output.Script,
			})
		}
	}

	closeTx, err := s.accountManager.CloseAccount(
		ctx, traderKey, closeOutputs, atomic.LoadUint32(&s.bestHeight),
	)
	if err != nil {
		return nil, err
	}
	closeTxHash := closeTx.TxHash()

	return &clmrpc.CloseAccountResponse{
		CloseTxid: closeTxHash[:],
	}, nil
}

func (s *rpcServer) ModifyAccount(ctx context.Context,
	req *clmrpc.ModifyAccountRequest) (
	*clmrpc.ModifyAccountResponse, error) {

	return nil, fmt.Errorf("unimplemented")
}

// SubmitOrder assembles all the information that is required to submit an order
// from the trader's lnd node, signs it and then sends the order to the server
// to be included in the auctioneer's order book.
func (s *rpcServer) SubmitOrder(ctx context.Context,
	req *clmrpc.SubmitOrderRequest) (*clmrpc.SubmitOrderResponse, error) {

	var o order.Order
	switch requestOrder := req.Details.(type) {
	case *clmrpc.SubmitOrderRequest_Ask:
		a := requestOrder.Ask
		kit, err := parseRPCOrder(a.Version, a.Details)
		if err != nil {
			return nil, err
		}
		o = &order.Ask{
			Kit:         *kit,
			MaxDuration: uint32(a.MaxDurationBlocks),
		}

	case *clmrpc.SubmitOrderRequest_Bid:
		b := requestOrder.Bid
		kit, err := parseRPCOrder(b.Version, b.Details)
		if err != nil {
			return nil, err
		}
		o = &order.Bid{
			Kit:         *kit,
			MinDuration: uint32(b.MinDurationBlocks),
		}

	default:
		return nil, fmt.Errorf("invalid order request")
	}

	// Verify that the account exists.
	acct, err := s.db.Account(o.Details().AcctKey)
	if err != nil {
		return nil, fmt.Errorf("cannot accept order: %v", err)
	}

	// Collect all the order data and sign it before sending it to the
	// auction server.
	serverParams, err := s.orderManager.PrepareOrder(ctx, o, acct)
	if err != nil {
		return nil, err
	}

	// Send the order to the server. If this fails, then the order is
	// certain to never get into the order book. We don't need to keep it
	// around in that case.
	err = s.auctioneer.SubmitOrder(ctx, o, serverParams)
	if err != nil {
		// TODO(guggero): Put in state failed instead of removing?
		if err2 := s.db.DelOrder(o.Nonce()); err2 != nil {
			log.Errorf("Could not delete failed order: %v", err2)
		}

		// If there was something wrong with the information the user
		// provided, then return this as a nice string instead of an
		// error type.
		if userErr, ok := err.(*order.UserError); ok {
			return &clmrpc.SubmitOrderResponse{
				Details: &clmrpc.SubmitOrderResponse_InvalidOrder{
					InvalidOrder: userErr.Details,
				},
			}, nil
		}

		// Any other error we return normally as a gRPC status level
		// error.
		return nil, fmt.Errorf("error submitting order to auctioneer: "+
			"%v", err)
	}

	// ServerOrder is accepted.
	return &clmrpc.SubmitOrderResponse{
		Details: &clmrpc.SubmitOrderResponse_Accepted{
			Accepted: true,
		},
	}, nil
}

// ListOrders returns a list of all orders that is currently known to the trader
// client's local store. The state of each order is queried on the auction
// server and returned as well.
func (s *rpcServer) ListOrders(ctx context.Context, _ *clmrpc.ListOrdersRequest) (
	*clmrpc.ListOrdersResponse, error) {

	// Get all orders from our local store first.
	dbOrders, err := s.db.GetOrders()
	if err != nil {
		return nil, err
	}

	// The RPC is split by order type so we have to separate them now.
	asks := make([]*clmrpc.Ask, 0, len(dbOrders))
	bids := make([]*clmrpc.Bid, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		nonce := dbOrder.Nonce()

		// Ask the server about the order's current status.
		state, unitsUnfullfilled, err := s.auctioneer.OrderState(
			ctx, nonce,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to query order state on"+
				"server for order %v: %v", nonce.String(), err)
		}

		dbDetails := dbOrder.Details()
		details := &clmrpc.Order{
			UserSubKey:       dbDetails.AcctKey.SerializeCompressed(),
			RateFixed:        int64(dbDetails.FixedRate),
			Amt:              int64(dbDetails.Amt),
			FundingFeeRate:   int64(dbDetails.FixedRate),
			OrderNonce:       nonce[:],
			State:            state.String(),
			Units:            uint32(dbDetails.Units),
			UnitsUnfulfilled: unitsUnfullfilled,
		}

		switch o := dbOrder.(type) {
		case *order.Ask:
			rpcAsk := &clmrpc.Ask{
				Details:           details,
				MaxDurationBlocks: int64(o.MaxDuration),
				Version:           uint32(o.Version),
			}
			asks = append(asks, rpcAsk)

		case *order.Bid:
			rpcBid := &clmrpc.Bid{
				Details:           details,
				MinDurationBlocks: int64(o.MinDuration),
				Version:           uint32(o.Version),
			}
			bids = append(bids, rpcBid)

		default:
			return nil, fmt.Errorf("unknown order type: %v",
				o)
		}
	}
	return &clmrpc.ListOrdersResponse{
		Asks: asks,
		Bids: bids,
	}, nil
}

// CancelOrder cancels the order on the server and updates the state of the
// local order accordingly.
func (s *rpcServer) CancelOrder(ctx context.Context,
	req *clmrpc.CancelOrderRequest) (*clmrpc.CancelOrderResponse, error) {

	var nonce order.Nonce
	copy(nonce[:], req.OrderNonce)
	err := s.auctioneer.CancelOrder(ctx, nonce)
	if err != nil {
		return nil, err
	}
	return &clmrpc.CancelOrderResponse{}, nil
}

// parseRPCOrder parses the incoming raw RPC order into the go native data
// types used in the order struct.
func parseRPCOrder(version uint32, details *clmrpc.Order) (*order.Kit, error) {
	var nonce order.Nonce
	copy(nonce[:], details.OrderNonce)
	kit := order.NewKit(nonce)

	// If the user didn't provide a nonce, we generate one.
	if nonce == order.ZeroNonce {
		preimageBytes, err := randomPreimage()
		if err != nil {
			return nil, fmt.Errorf("cannot generate nonce: %v", err)
		}
		var preimage lntypes.Preimage
		copy(preimage[:], preimageBytes)
		kit = order.NewKitWithPreimage(preimage)
	}

	pubKey, err := btcec.ParsePubKey(details.UserSubKey, btcec.S256())
	if err != nil {
		return nil, fmt.Errorf("error parsing account key: %v", err)
	}

	kit.AcctKey = pubKey
	kit.Version = order.Version(version)
	kit.FixedRate = uint32(details.RateFixed)
	kit.Amt = btcutil.Amount(details.Amt)
	kit.FundingFeeRate = chainfee.SatPerKWeight(details.FundingFeeRate)
	kit.Units = order.NewSupplyFromSats(kit.Amt)
	kit.UnitsUnfulfilled = kit.Units
	return kit, nil
}

// randomPreimage creates a new preimage from a random number generator.
func randomPreimage() ([]byte, error) {
	var nonce order.Nonce
	_, err := rand.Read(nonce[:])
	if err != nil {
		return nil, err
	}
	return nonce[:], nil
}