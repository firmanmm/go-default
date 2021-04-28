package godefault

type _KeySet map[string]bool

func (k _KeySet) SetKey(key string) {
	k[key] = true
}

func (k _KeySet) HasKey(key string) bool {
	_, ok := k[key]
	return ok
}
