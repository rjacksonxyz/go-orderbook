package orderbook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderBook(t *testing.T) {
	// Test cases
	oi := OrderId(1)
	ob := NewOrderbook()
	ob.AddOrder(
		Order{
			orderType:       GoodTillCancel,
			orderId:         1,
			side:            Buy,
			price:           100,
			initialQuantity: 10,
		},
	)
	t.Logf("Orderbook Size: %d", ob.Size())
	assert.Equal(t, 1, ob.Size())
	t.Logf("Orderbook Orders: %#v", ob.OrderInfo())
	assert.NoError(t, ob.CancelOrder(oi))
	t.Logf("Orderbook Size: %d", ob.Size())
	assert.Equal(t, 0, ob.Size())
}
