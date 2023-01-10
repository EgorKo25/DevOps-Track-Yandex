package hashing

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/storage"
	"log"
)

type Hash struct {
	Key []byte
}

func MewHash(cfg string) *Hash {
	return &Hash{
		Key: []byte(cfg),
	}
}

func (h *Hash) Run(metric *storage.Metric) (hash string, err error) {

	var src string

	switch metric.MType {
	case "gauge":
		src = fmt.Sprintf("%s:%s:%f", metric.ID, metric.MType, *metric.Value)
	case "counter":
		src = fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType, *metric.Delta)
	}

	log.Println("Data: ", src)
	log.Printf("Key: %x", h.Key)

	if h.Key == nil {
		return "", nil
	}

	hm := hmac.New(sha256.New, h.Key)
	hm.Write([]byte(src))
	hash = fmt.Sprintf("%x", hm.Sum(nil))

	if metric.Hash != "" && !hmac.Equal([]byte(metric.Hash), []byte(hash)) {
		return "", fmt.Errorf("not equal hash")
	}

	return hash, nil
}
