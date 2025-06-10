package model

type OrderHeap []*Order

func (h OrderHeap) Len() int { return len(h) }
func (h OrderHeap) Less(i, j int) bool {
	if h[i].Price == h[j].Price {
		return h[i].OrderId < h[j].OrderId
	}
	if h[i].Direction == OrderDirectionBID {
		return h[i].Price > h[j].Price
	}

	return h[i].Price < h[j].Price
}

func (h OrderHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}

func (h *OrderHeap) Push(x any) {
	order := x.(*Order)
	order.index = len(*h)
	*h = append(*h, order)
}

func (h *OrderHeap) Pop() any {
	old := *h
	n := len(old)
	order := old[n-1]
	order.index = -1
	*h = old[0 : n-1]

	return order
}
