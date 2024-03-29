package clientdb

import (
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
	"github.com/lightninglabs/llm/order"
	"github.com/lightningnetwork/lnd/keychain"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwallet/chainfee"
)

// TestSubmitOrder tests that orders can be stored and retrieved correctly.
func TestSubmitOrder(t *testing.T) {
	t.Parallel()

	store, cleanup := newTestDB(t)
	defer cleanup()

	// Store a dummy order and see if we can retrieve it again.
	o := &order.Bid{
		Kit:         *dummyOrder(t, 500000),
		MinDuration: 1337,
	}
	err := store.SubmitOrder(o)
	if err != nil {
		t.Fatalf("unable to store order: %v", err)
	}
	storedOrder, err := store.GetOrder(o.Nonce())
	if err != nil {
		t.Fatalf("unable to retrieve order: %v", err)
	}
	if !reflect.DeepEqual(o, storedOrder) {
		t.Fatalf("expected order: %v\ngot: %v", spew.Sdump(o),
			spew.Sdump(storedOrder))
	}

	// Check that we got the correct type back.
	if storedOrder.Type() != order.TypeBid {
		t.Fatalf("unexpected order type. got %d expected %d",
			storedOrder.Type(), order.TypeBid)
	}

	// Get all orders and check that we get the same as when querying a
	// specific one.
	allOrders, err := store.GetOrders()
	if err != nil {
		t.Fatalf("unable to get all asks: %v", err)
	}
	if len(allOrders) != 1 {
		t.Fatalf("unexpected number of asks. got %d expected %d",
			len(allOrders), 1)
	}
	if allOrders[0].Type() != order.TypeBid {
		t.Fatalf("unexpected order type. got %d expected %d",
			allOrders[0].Type(), order.TypeBid)
	}
	if !reflect.DeepEqual(o, allOrders[0]) {
		t.Fatalf("expected order: %v\ngot: %v", spew.Sdump(o),
			spew.Sdump(allOrders[0]))
	}

	// Check that we got the correct type back.
	if allOrders[0].Type() != order.TypeBid {
		t.Fatalf("unexpected order type. got %d expected %d",
			allOrders[0].Type(), order.TypeBid)
	}
	
	// Delete the order and make sure it's gone.
	err = store.DelOrder(o.Nonce())
	if err != nil {
		t.Fatalf("could not delete order: %v", err)
	}
	allOrders, err = store.GetOrders()
	if err != nil {
		t.Fatalf("unable to get all asks: %v", err)
	}
	if len(allOrders) != 0 {
		t.Fatalf("unexpected number of asks. got %d expected %d",
			len(allOrders), 0)
	}
}

// TestUpdateOrders tests that orders can be updated correctly.
func TestUpdateOrders(t *testing.T) {
	t.Parallel()

	store, cleanup := newTestDB(t)
	defer cleanup()

	// Store two dummy orders that we are going to update later.
	o1 := &order.Bid{
		Kit:         *dummyOrder(t, 500000),
		MinDuration: 1337,
	}
	err := store.SubmitOrder(o1)
	if err != nil {
		t.Fatalf("unable to store order: %v", err)
	}
	o2 := &order.Ask{
		Kit:         *dummyOrder(t, 500000),
		MaxDuration: 1337,
	}
	err = store.SubmitOrder(o2)
	if err != nil {
		t.Fatalf("unable to store order: %v", err)
	}

	// Update the state of the first order and check that it is persisted.
	err = store.UpdateOrder(
		o1.Nonce(), order.StateModifier(order.StatePartiallyFilled),
	)
	if err != nil {
		t.Fatalf("unable to update order: %v", err)
	}
	storedOrder, err := store.GetOrder(o1.Nonce())
	if err != nil {
		t.Fatalf("unable to retrieve order: %v", err)
	}
	if storedOrder.Details().State != order.StatePartiallyFilled {
		t.Fatalf("unexpected order state. got %d expected %d",
			storedOrder.Details().State, order.StatePartiallyFilled)
	}

	// Bulk update the state of both orders and check that they are
	// persisted correctly.
	stateModifier := order.StateModifier(order.StateCleared)
	err = store.UpdateOrders(
		[]order.Nonce{o1.Nonce(), o2.Nonce()},
		[][]order.Modifier{{stateModifier}, {stateModifier}},
	)
	if err != nil {
		t.Fatalf("unable to update orders: %v", err)
	}
	allOrders, err := store.GetOrders()
	if err != nil {
		t.Fatalf("unable to get all orders: %v", err)
	}
	if len(allOrders) != 2 {
		t.Fatalf("unexpected number of orders. got %d expected %d",
			len(allOrders), 2)
	}
	for _, o := range allOrders {
		if o.Details().State != order.StateCleared {
			t.Fatalf("unexpected order state. got %d expected %d",
				o.Details().State, order.StateCleared)
		}
	}
}

func dummyOrder(t *testing.T, amt btcutil.Amount) *order.Kit {
	var testPreimage lntypes.Preimage
	if _, err := rand.Read(testPreimage[:]); err != nil {
		t.Fatalf("could not create private key: %v", err)
	}
	kit := order.NewKitWithPreimage(testPreimage)
	kit.Version = order.VersionDefault
	kit.State = order.StateExecuted
	kit.FixedRate = 21
	kit.Amt = amt
	kit.MultiSigKeyLocator = keychain.KeyLocator{
		Family: 123,
		Index:  345,
	}
	kit.FundingFeeRate = chainfee.FeePerKwFloor
	copy(kit.AcctKey[:], testTraderKey.SerializeCompressed())
	kit.UnitsUnfulfilled = 741
	return kit
}
