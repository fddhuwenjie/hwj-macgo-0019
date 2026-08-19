package clockidcodec

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidEncodedID = errors.New("clockidcodec: invalid encoded id")

func EncodeID(prefix, id string, version int64) string {
	payload := prefix + ":" + id + ":" + strconv.FormatInt(version, 10)
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func DecodeID(encoded string) (prefix, id string, version int64, err error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", 0, fmt.Errorf("%w: %v", ErrInvalidEncodedID, err)
	}
	parts := strings.Split(string(raw), ":")
	if len(parts) != 3 {
		return "", "", 0, fmt.Errorf("%w: expected prefix:id:version", ErrInvalidEncodedID)
	}
	prefix = parts[0]
	id = parts[1]
	version, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("%w: bad version: %v", ErrInvalidEncodedID, err)
	}
	if prefix == "" || id == "" {
		return "", "", 0, fmt.Errorf("%w: empty prefix or id", ErrInvalidEncodedID)
	}
	return prefix, id, version, nil
}
