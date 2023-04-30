package sheets

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var (
	SheetUrl     = "https://superquacken.com/openkhleaders"
	SheetStatUrl = "https://superquacken.com/pushkhstats"
	// you should probably change this secret to something else
	EncryptionSecret = "super secret encryption key"
)

func PushRowsToSheet(values [][]interface{}) error {
	jsonBytes, err := json.Marshal(values)
	if err != nil {
		return err
	}
	encrypted, err := Encrypt(jsonBytes, EncryptionSecret)
	if err != nil {
		return err
	}
	_, err = http.Post(SheetStatUrl, "application/octet-stream", bytes.NewReader(encrypted))
	return err
}

func PushRowToSheet(values ...interface{}) error {
	return PushRowsToSheet([][]interface{}{values})
}
