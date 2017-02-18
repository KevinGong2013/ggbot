package bridge

import "fmt"

// ProcessUUID impolements UUIDProcessor interface
func (w *Wrapper) ProcessUUID(uuid string) error {

	r := w.login(uuid)
	if r.IsFailure() {
		return fmt.Errorf(`bridge uuid processor errror: %v`, r.Err)
	}

	return nil
}

// UUIDDidConfirm impolements UUIDProcessor interface
func (w *Wrapper) UUIDDidConfirm(err error) {
	if err != nil {
		logger.Errorf(`uuid has been invalidated`)
	} else {
		logger.Info(`uuid did confirmed`)
	}
}
