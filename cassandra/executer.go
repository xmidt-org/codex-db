/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cassandra

import (
	"errors"
	"fmt"
	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/codex-db/blacklist"
	"github.com/yugabyte/gocql"
)

type (
	finder interface {
		findRecords(limit int, filter string, where ...interface{}) ([]db.Record, error)
	}
	findList interface {
		findBlacklist() ([]blacklist.BlackListedItem, error)
	}
	deviceFinder interface {
		getList(offset string, limit int) ([]string, error)
	}
	multiinserter interface {
		insert(records []db.Record) (int, error)
	}
	pinger interface {
		ping() error
	}
	closer interface {
		close() error
	}
)

type dbDecorator struct {
	session *gocql.Session
}

func (b *dbDecorator) findRecords(limit int, filter string, where ...interface{}) ([]db.Record, error) {
	var (
		records []db.Record
	)

	// row fields for the record
	var (
		device    string
		eventType int
		birthdate int64
		deathdate int64
		data      []byte
		nonce     []byte
		alg       string
		kid       string
	)

	iter := b.session.Query(fmt.Sprintf("SELECT device_id, record_type, birthdate, deathdate, data, nonce, alg, kid FROM devices.events %s LIMIT ?", filter), append(where, limit)...).Iter()

	for iter.Scan(&device, &eventType, &birthdate, &deathdate, &data, &nonce, &alg, &kid) {
		records = append(records, db.Record{
			DeviceID:  device,
			Type:      db.EventType(eventType),
			BirthDate: birthdate,
			DeathDate: deathdate,
			Data:      data,
			Nonce:     nonce,
			Alg:       alg,
			KID:       kid,
		})
	}

	err := iter.Close()
	return records, err
}

func (b *dbDecorator) getList(offset string, limit int) ([]string, error) {
	var result []string

	var device string

	iter := b.session.Query("SELECT device_id from devices.events WHERE device_id > ? GROUP BY device_id LIMIT ?", offset, limit).Iter()
	for iter.Scan(&device) {
		result = append(result, device)
	}

	err := iter.Close()
	return result, err
}

func (b *dbDecorator) findBlacklist() ([]blacklist.BlackListedItem, error) {
	var (
		records []blacklist.BlackListedItem
	)

	var device string
	var reason string

	iter := b.session.Query("SELECT device_id, reason FROM devices.blacklist;").Iter()

	for iter.Scan(&device, &reason) {
		records = append(records, blacklist.BlackListedItem{
			ID:     device,
			Reason: reason,
		})
	}

	err := iter.Close()
	return records, err
}

func (b *dbDecorator) insert(records []db.Record) (int, error) {

	batch := b.session.NewBatch(gocql.UnloggedBatch)

	for _, record := range records {
		// TTL is not supported at the row level
		// DO NOT DELETE
		//ttl := int(time.Unix(0, record.DeathDate).Sub(time.Now()).Seconds())
		//// yugabyte bounds checking
		//if ttl < 0 {
		//	ttl = 0
		//}
		//if ttl > 2147483647 {
		//	ttl = 2147483647
		//}

		// there can be no spaces for some weird reason. Otherwise the database returns and error.
		batch.Query("INSERT INTO devices.events (device_id, record_type, birthdate, deathdate, data, nonce, alg, kid) VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
			record.DeviceID,
			record.Type,
			record.BirthDate,
			record.DeathDate,
			record.Data,
			record.Nonce,
			record.Alg,
			record.KID,
			//ttl,
		)
	}
	err := b.session.ExecuteBatch(batch)
	return batch.Size(), err
}

func (b *dbDecorator) ping() error {
	if b.session.Closed() {
		return errors.New("server is closed")
	}
	return nil
}
func (b *dbDecorator) close() error {
	b.session.Close()
	return nil
}

func connect(clusterConfig *gocql.ClusterConfig) (*dbDecorator, error) {
	session, err := clusterConfig.CreateSession()
	if err != nil {
		return nil, err
	}

	return &dbDecorator{session: session}, nil

}
