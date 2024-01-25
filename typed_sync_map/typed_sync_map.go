package typed_sync_map

import "sync"

type TypedSyncMap[K comparable, V any] struct {
	syncMap sync.Map
}

func (m *TypedSyncMap[K, V]) Load(key K) (V, bool) {
	value, hasKey := m.syncMap.Load(key)

	castedValue, isOfCorrectType := value.(V)

	return castedValue, hasKey && isOfCorrectType
}

func (m *TypedSyncMap[K, V]) Store(key K, val V) {
	m.syncMap.Store(key, val)
}
