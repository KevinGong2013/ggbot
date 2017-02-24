package service

// ProcessUUID impolements UUIDProcessor interface
func (w *Wrapper) ProcessUUID(uuid string) error {
	return w.f.forward(w.f.uuidWebhookURL, map[string]interface{}{
		`uuid`:       uuid,
		`didConfirm`: false,
	})
}

// UUIDDidConfirm impolements UUIDProcessor interface
func (w *Wrapper) UUIDDidConfirm(err error) {
	w.f.forward(w.f.uuidWebhookURL, map[string]interface{}{
		`uuid`:       ``,
		`didConfirm`: true,
	})
}
