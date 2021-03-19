package kv

import (
	"context"
	"fmt"

	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) Recreate(ctx context.Context) error {
	numKeys := 2048
	pubKeys := make([][48]byte, numKeys)
	for i := 0; i < len(pubKeys); i++ {
		copy(pubKeys[i][:], fmt.Sprintf("%d", i))
	}

	for i, pubKey := range pubKeys {
		if err := s.view(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(pubKeysBucket)
			pkBucket := bucket.Bucket(pubKey[:])
			sourceEpochsBucket := pkBucket.Bucket(attestationSourceEpochsBucket)

			lowest, _ := sourceEpochsBucket.Cursor().First()
			highest, _ := sourceEpochsBucket.Cursor().Last()
			source := bytesutil.BytesToUint64BigEndian(lowest)
			highestSource := bytesutil.BytesToUint64BigEndian(highest)
			log.Infof("For pubkey %d, lowest source %d, highest source %d", i, source, highestSource)
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
