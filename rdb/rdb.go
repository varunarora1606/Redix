package rdb

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/varunarora1606/My-Redis/store"
)

type rdbEntry struct {
	key string
	val string
	ttl int64
}

func SaveRDB(filename string, snapShot store.SnapShot) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create RDB file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write([]byte("REDIS")); err != nil {
		return fmt.Errorf("could not write to RDB file: %v", err)
	}

	for k, v := range snapShot.Data {
		ttl, ok := snapShot.Expiry[k]
		if !ok {
			ttl = 0
		}
		if err := WriteRDB(file, k, v, ttl); err != nil {
			return err
		}
	}

	return nil
}

func WriteRDB(file *os.File, key string, val string, ttl int64) error {
	keyLen := uint32(len(key))
	valLen := uint32(len(val))

	if err := binary.Write(file, binary.LittleEndian, keyLen); err != nil {
		return err
	}
	if _, err := file.Write([]byte(key)); err != nil {
		return nil
	}

	if err := binary.Write(file, binary.LittleEndian, valLen); err != nil {
		return err
	}
	if _, err := file.Write([]byte(val)); err != nil {
		return nil
	}

	if err := binary.Write(file, binary.LittleEndian, ttl); err != nil {
		return err
	}

	return nil
}

func LoadRDB(filepath string, kv store.Store) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("could not open RDB file: %v", err)
	}
	defer file.Close()

	header := make([]byte, 5)
	if _, err := file.Read(header); err != nil || string(header) != "REDIS" {
		return fmt.Errorf("invalid RDB file")
	}

	for {
		entry, err := ReadRDB(file)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		kv.Add(entry.key, entry.val, entry.ttl)
	}

	return nil
}

func ReadRDB(file *os.File) (rdbEntry, error) {
	var keyLen uint32
	if err := binary.Read(file, binary.LittleEndian, &keyLen); err != nil {
		return rdbEntry{}, err
	}

	key := make([]byte, keyLen)
	if _, err := file.Read(key); err != nil {
		return rdbEntry{}, err
	}

	var valLen uint32
	if err := binary.Read(file, binary.LittleEndian, &valLen); err != nil {
		return rdbEntry{}, err
	}

	val := make([]byte, valLen)
	if _, err := file.Read(val); err != nil {
		return rdbEntry{}, err
	}

	var ttl int64
	if err := binary.Read(file, binary.LittleEndian, &ttl); err != nil {
		return rdbEntry{}, err
	}

	return rdbEntry{
		key: string(key),
		val: string(val),
		ttl: ttl,
	}, nil
}
