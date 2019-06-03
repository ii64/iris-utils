package longpoll

type Options struct {
	MaxConnection	uint32 	`json:"maxConnection"`
	TimeoutSeconds	int64 	`json:"timeoutSeconds"`
}

const (
	UNLIMITED_CONN  = uint32(6553655)
	UNSET = int64(-1001)
)