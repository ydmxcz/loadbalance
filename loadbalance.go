package loadbalance

type Instance[T Hashable] interface {
	InstanceID() T
	InstanceWeight() int
}

type base[T Hashable, I Instance[T]] interface {
	Add(instances ...I) int
	Del(instances ...I) int
	Get(T) (I, bool)
	Size() int
	ForEach(func(T, I) bool)
}

type SelectorBy[T Hashable, I Instance[T]] interface {
	base[T, I]
	SelectBy(string) I
}

type Selector[T Hashable, I Instance[T]] interface {
	base[T, I]
	Select() I
}
