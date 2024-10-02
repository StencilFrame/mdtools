package ordered

import "iter"

// Key Value iterator
func (om *OrderedMap) KVIter() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for e := om.l.Front(); e != nil; e = e.Next() {
			key := e.Value.(string)
			value := om.m[key]
			if !yield(key, value) {
				return
			}
		}
	}
}
