package arg

const (
	// Token ...
	Token = 0

	// Login ...
	Login = 1

	// RedPacket ..
	RedPacket = 2
)

//Arg send to client arg struct
type Arg struct {
	Seq   uint64
	CMD   int
	Value map[string]interface{}
}

// NewArg ...
func NewArg(cmd int) *Arg {
	return &Arg{
		CMD:   cmd,
		Value: make(map[string]interface{}),
	}
}

// Append ..
func (arg *Arg) Append(k string, v interface{}) *Arg {
	arg.Value[k] = v
	return arg
}
