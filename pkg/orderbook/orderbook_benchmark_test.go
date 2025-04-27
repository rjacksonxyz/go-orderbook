package orderbook

import (
	"testing"
)

// setupOrderbook creates a new orderbook filled with asks and bids similar to the C++ example
func setupOrderbook(n int) *Orderbook {
	ob := NewOrderbook()

	// Add asks (sell orders)
	for i := 0; i < n; i++ {
		quantity := Quantity(35 + i%10)
		price := Price(5950 + i%10*10) // Using integers, but simulating 59.50 + i * 0.10

		// Add two orders at same price point (like in C++ example)
		ob.AddOrder(NewOrder(
			GoodTillCancel,
			OrderId(i*2+1),
			Sell,
			price,
			quantity,
		))

		ob.AddOrder(NewOrder(
			GoodTillCancel,
			OrderId(i*2+2),
			Sell,
			price,
			quantity,
		))
	}

	// Add bids (buy orders)
	for i := 0; i < n; i++ {
		quantity := Quantity(70 + i%10)
		price := Price(5990 - i%10*10) // Using integers, but simulating 59.90 - i * 0.10

		// Add two orders at same price point (like in C++ example)
		ob.AddOrder(NewOrder(
			GoodTillCancel,
			OrderId(2*n+i*2+1),
			Buy,
			price,
			quantity,
		))

		ob.AddOrder(NewOrder(
			GoodTillCancel,
			OrderId(2*n+i*2+2),
			Buy,
			price,
			quantity,
		))
	}

	return &ob
}

// makeAsks creates sell orders following the pattern from the C++ example
func makeAsks(ob *Orderbook, startId OrderId, count int) {
	for i := 0; i < count; i++ {
		quantity := Quantity(35 + i%10)
		price := Price(59.5 + float64(i%10)*0.1) // Using integers, but simulating 59.50 + i * 0.10

		// Add two orders at same price point (like in C++ example)
		ob.AddOrder(NewOrder(
			GoodTillCancel,
			startId+OrderId(i*2),
			Sell,
			price,
			quantity,
		))

		ob.AddOrder(NewOrder(
			GoodTillCancel,
			startId+OrderId(i*2+1),
			Sell,
			price,
			quantity,
		))
	}
}

// makeBids creates buy orders following the pattern from the C++ example
func makeBids(ob *Orderbook, startId OrderId, count int) {
	for i := 0; i < count; i++ {
		quantity := Quantity(70 + i%10)
		price := Price(59.9 + float64(i%10)*0.1) // Using integers, but simulating 59.90 - i * 0.10

		// Add two orders at same price point (like in C++ example)
		ob.AddOrder(NewOrder(
			GoodTillCancel,
			startId+OrderId(i*2),
			Buy,
			price,
			quantity,
		))

		ob.AddOrder(NewOrder(
			GoodTillCancel,
			startId+OrderId(i*2+1),
			Buy,
			price,
			quantity,
		))
	}
}

// BenchmarkOrderbookAddMultipleOrders benchmarks adding multiple orders with different prices
func BenchmarkOrderbookAddMultipleOrders(b *testing.B) {
	ob := NewOrderbook()
	for i := 0; i < b.N; i++ {
		makeAsks(&ob, 1, 10)
		makeBids(&ob, 21, 10) // starting ID after 20 ask orders (10 * 2)
	}
}
