package typecast

import "testing"

type BeforeHook interface {
	BeforeCreate() error
}

type AfterHook interface {
	AfterCreate()
}

func generateNoHookItems(count int) []*NoHookItem {
	items := make([]*NoHookItem, count)
	for i := 0; i < count; i++ {
		items[i] = &NoHookItem{}
	}
	return items
}

func generateHookItems(count int) []*HookItem {
	items := make([]*HookItem, count)
	for i := 0; i < count; i++ {
		items[i] = &HookItem{}
	}
	return items
}

func Benchmark10000(b *testing.B) {
	count := 10000

	b.Run("origin", func(b *testing.B) {
		items := generateNoHookItems(count)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for _, item := range items {
				do(item)
			}
		}
	})

	b.Run("hookitems", func(b *testing.B) {
		items := generateHookItems(count)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for _, item := range items {
				doWithHook(item)
			}
		}
	})

	b.Run("nohookitems", func(b *testing.B) {
		items := generateNoHookItems(count)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for _, item := range items {
				doWithHook(item)
			}
		}
	})
}

type NoHookItem struct{}

type HookItem struct{}

func (h *HookItem) BeforeCreate() error { return nil }
func (h *HookItem) AfterCreate()        {}

func do(_ interface{}) int {
	return 1 + 2
}

func doWithHook(v interface{}) int {
	if _, ok := v.(BeforeHook); ok {
		//_ = h.BeforeCreate()
	}
	res := 1 + 2
	if _, ok := v.(AfterHook); ok {
		// h.AfterCreate()
	}
	return res
}
