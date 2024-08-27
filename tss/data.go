package tss

type Payload struct {
	Sender  string `json:"sender"`
	Command string `json:"command"`
	Message string `json:"message"`
	Package []byte `json:"package"`
}
