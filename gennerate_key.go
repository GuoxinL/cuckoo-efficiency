/*
   Created by guoxin in 2022/8/31 3:38 PM
*/
package main

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	guuid "github.com/google/uuid"
)

const Separator = 202

func GenTimestampKeyByNano(nano int64) []byte {
	b := make([]byte, 16, 32)
	binary.BigEndian.PutUint64(b, uint64(nano))
	/*
		Read generates len(p) random bytes from the default Source and
		writes them into p. It always returns len(p) and a nil error.
		Read, unlike the Rand.Read method, is safe for concurrent use.
	*/
	b[8] = Separator
	// nolint: gosec
	_, _ = rand.Read(b[9:16])
	u := guuid.New()
	b = append(b, u[:]...)
	return b
}

func GenTimestampKey() []byte {
	return GenTimestampKeyByNano(time.Now().UnixNano())
}
