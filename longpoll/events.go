package longpoll
type lpEvent struct {
	timestamp	[]int64
	size		uint32
	data		chan interface{}
}