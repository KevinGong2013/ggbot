package service

// ProcessUUID impolements UUIDProcessor interface
func (w *Wrapper) ProcessUUID(uuid string) error {
	return w.Forward(w.uuidWebhookURL, map[string]interface{}{
		`uuid`:       uuid,
		`didConfirm`: false,
	})
}

// UUIDDidConfirm impolements UUIDProcessor interface
func (w *Wrapper) UUIDDidConfirm(err error) {
	w.Forward(w.uuidWebhookURL, map[string]interface{}{
		`uuid`:       ``,
		`didConfirm`: true,
	})
}
