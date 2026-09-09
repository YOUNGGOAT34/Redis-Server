package rdb

import (
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

func SaveRDB(path string, replconfig *config.SERVER) error {
	replconfig.DatabaseMutex.RLock()
	replconfig.ExpiryMutex.RLock()
	err := SaveRDBLocked(path, replconfig)
	replconfig.ExpiryMutex.RUnlock()
	replconfig.DatabaseMutex.RUnlock()
	return err
}

// SaveRDBLocked writes an RDB snapshot of replconfig's current Database and
// Expiry to path. The caller MUST already hold at least a read lock on both
// DatabaseMutex and ExpiryMutex (SaveRDB does this for the normal case; the
// PSYNC handler in server.go calls this directly because it needs to hold
// those locks - as write locks - across snapshot creation AND replica
// registration together, to close the window where a write landing between
// "snapshot taken" and "replica registered for propagation" would otherwise
// be silently lost to that replica. See the PSYNC handler for the full
// explanation.
func SaveRDBLocked(path string, replconfig *config.SERVER) error {
	tempPath := path + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	if err := writeRDBHeader(file); err != nil {
		file.Close()
		return err
	}

	if err := writeAuxFileds(file); err != nil {
		file.Close()
		return err
	}

	if err := writeselectdatabase(file, 0); err != nil {
		file.Close()
		return err
	}

	if err := WriteReSizeDB(file, replconfig); err != nil {
		file.Close()
		return err
	}

	if err := writeDatabase(file, replconfig); err != nil {
		file.Close()
		return err
	}

	if _, err := file.Write([]byte{0xFF}); err != nil {
		file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, path)
}

func writeRDBHeader(w io.Writer) error {
	_, err := w.Write([]byte("REDIS0011"))
	if err != nil {
		return err
	}

	return nil
}

func writeAuxFileds(w io.Writer) error {
	if err := writeAuxString(w, "redis-ver", "7.2.0"); err != nil {
		return err
	}

	if err := writeAuxInteger(w, "redis-bits", 64); err != nil {
		return err
	}

	if err := writeAuxInteger(w, "ctime", uint32(time.Now().Unix())); err != nil {
		return err
	}

	if err := writeAuxInteger(w, "used-mem", 0); err != nil {
		return err
	}

	if err := writeAuxInteger(w, "aof-base", 0); err != nil {
		return err
	}

	return nil
}

func encodeSpecialInteger(w io.Writer, value uint32) error {

	if value < 256 {
		buffer := [2]byte{0xc0, byte(value & 0xFF)}
		_, err := w.Write(buffer[:])

		return err
	} else if value < 65536 {
		buffer := [3]byte{0xc1, byte(value >> 8), byte(value & 0xFF)}
		_, err := w.Write(buffer[:])

		return err

	} else {
		buffer := [5]byte{
			0xc2,
			byte(value >> 24),
			byte(value >> 16),
			byte(value >> 8),
			byte(value & 0xFF),
		}

		_, err := w.Write(buffer[:])

		return err
	}

}

func writeselectdatabase(w io.Writer, databaseNumber int) error {
	_, err := w.Write([]byte{0xFE})
	if err != nil {
		return err
	}

	return encodeLength(w, databaseNumber)
}

func WriteReSizeDB(w io.Writer, replconfig *config.SERVER) error {
	_, err := w.Write([]byte{0xFB})

	if err != nil {
		return err
	}

	err = encodeLength(w, len(replconfig.Database))

	if err != nil {
		return err
	}
	err = encodeLength(w, len(replconfig.Expiry))
	if err != nil {
		return err
	}

	return nil

}

func writeAuxInteger(w io.Writer, key string, value uint32) error {

	/*
		AUX opcode
		encoded length of key
		key
		special encoding
		integer bytes
	*/

	if _, err := w.Write([]byte{0xFA}); err != nil {
		return err
	}

	err := encodeLength(w, len(key))

	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(key)); err != nil {
		return err
	}

	err = encodeSpecialInteger(w, value)

	if err != nil {
		return err
	}

	return nil
}

func writeAuxString(w io.Writer, key string, value string) error {
	/*
		      AUX opcode
				length of key
				key
				length of value
				value
	*/
	if _, err := w.Write([]byte{0xFA}); err != nil {
		return err
	}

	err := encodeLength(w, len(key))

	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(key)); err != nil {
		return err
	}

	err = encodeLength(w, len(value))

	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(value)); err != nil {
		return err
	}

	return nil
}

func encodeLength(w io.Writer, length int) error {

	if length < 0 {
		return errors.New("length cannot be negative")
	}
	if length < 64 {
		buffer := [1]byte{byte(length)}
		_, err := w.Write(buffer[:])

		return err

	} else if length < 16384 {

		buffer := [2]byte{
			byte((length >> 8) | 0x40),
			byte(length & 0xFF),
		}

		_, err := w.Write(buffer[:])

		return err

	} else {

		buffer := [5]byte{
			0x80,
			byte(length >> 24),
			byte(length >> 16),
			byte(length >> 8),
			byte(length & 0xFF),
		}

		_, err := w.Write(buffer[:])

		return err

	}

}

func writeDatabase(w io.Writer, replconfig *config.SERVER) error {

	for key, value := range replconfig.Database {

		// Always write exactly len(replconfig.Database) entries, one per
		// key - WriteReSizeDB's count is NOT just an advisory hint, the
		// loader's 0xFB handler reads exactly that many entries via
		// readEntry, so skipping a key here would desync the count and
		// corrupt parsing of everything after it. Whether an
		// already-expired key should actually be loaded back is instead
		// left to LoadFileToMemory, which already correctly skips any
		// entry whose expiry has passed - so it's enough to always write
		// the expiry metadata when present (even for an already-expired
		// key) and let that existing load-side check do the filtering.
		if expiresAt, hasExpiry := replconfig.Expiry[key]; hasExpiry {
			if err := writeExpiry(w, expiresAt); err != nil {
				return err
			}
		}

		err := writeObjectType(w, value.Type)
		if err != nil {
			return err
		}

		err = writeKey(w, key)
		if err != nil {
			return err
		}

		err = writeValue(w, value)

		if err != nil {
			return err
		}

	}

	return nil
}

// writeExpiry writes the millisecond-precision expiry opcode (0xFC) that
// ReadRDBFile's readEntry already knows how to parse, immediately before the
// object-type/key/value for the same entry.
func writeExpiry(w io.Writer, expiresAt time.Time) error {
	if _, err := w.Write([]byte{0xFC}); err != nil {
		return err
	}

	millis := uint64(expiresAt.UnixMilli())
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, millis)

	_, err := w.Write(buf)
	return err
}

func writeKey(w io.Writer, key string) error {
	err := encodeLength(w, len(key))

	if err != nil {
		return err
	}

	_, err = w.Write([]byte(key))

	return err
}

func writeValue(w io.Writer, data storage.Data) error {
	switch data.Type {
	case storage.STRING:
		value, ok := data.Value.([]byte)

		if !ok {
			return errors.New("Wrong data stored in a string type")
		}

		err := encodeLength(w, len(value))

		if err != nil {
			return err
		}

		_, err = w.Write(value)

		return err

	case storage.LIST:
		values, ok := data.Value.(*storage.List)

		if !ok {
			return errors.New("Wrong data stored in list type")
		}

		err := encodeLength(w, values.Len)

		if err != nil {
			return err
		}

		current := values.Head

		for {

			if current == nil {
				break
			}

			nodeData := current.Data

			err = encodeLength(w, len(nodeData))

			if err != nil {
				return err
			}

			_, err = w.Write(nodeData)

			if err != nil {
				return err
			}

			current = current.Next

		}

	default:
		return errors.New("unknown data type")

	}

	return nil
}

func writeObjectType(w io.Writer, objectType storage.TYPE) error {
	switch objectType {

	case storage.STRING:
		_, err := w.Write([]byte{0x00})

		return err
	case storage.LIST:
		_, err := w.Write([]byte{0x01})
		return err
	default:
		return errors.New("Unkown object type")
	}
}
