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

//import (
//	"fmt"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	db "github.com/xmidt-org/codex-db"
//	"github.com/yugabyte/gocql"
//	"log"
//	"testing"
//	"time"
//)
//
//var createStmt = `CREATE TABLE devices.events (device_id  varchar,
//                                                          type int,
//                                                          birthdate bigint,
//                                                          deathdate bigint,
//                                                          data blob,
//                                                          nonce blob,
//                                                          alg varchar,
//                                                          kid varchar,
//                                                          PRIMARY KEY ((device_id,  type), birthdate)
//                                                  ) WITH CLUSTERING ORDER BY (birthdate DESC) AND transactions = { 'enabled' : true } AND default_time_to_live = 0;`
//
//func TestBasicCommands(t *testing.T) {
//	assert := assert.New(t)
//	require := require.New(t)
//
//	d, err := connect(gocql.NewCluster("127.0.0.1"))
//	defer d.close()
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Set up the keyspace and table.
//	if err := d.session.Query("CREATE KEYSPACE IF NOT EXISTS devices").Exec(); err != nil {
//	}
//	fmt.Println("Created keyspace devices")
//
//	if err := d.session.Query(`DROP TABLE IF EXISTS devices.events`).Exec(); err != nil {
//		log.Fatal(err)
//	}
//
//	if err := d.session.Query(createStmt).Exec(); err != nil {
//		log.Fatal(err)
//	}
//
//	goodRecord := db.Record{
//		DeviceID:  "neat",
//		Type:      db.Default,
//		BirthDate: time.Now().UnixNano() + time.Minute.Nanoseconds(),
//		DeathDate: time.Now().Add(time.Second * 30).UnixNano() + time.Minute.Nanoseconds(),
//		Data:      []byte("hello"),
//		Nonce:     []byte("hello"),
//		KID:       "3",
//		Alg:       "3",
//	}
//
//	_, err = d.insert([]db.Record{
//		goodRecord,
//	})
//	assert.NoError(err)
//
//	records, err := d.findRecords(100, "WHERE device_id=?", "neat")
//	require.Len(records, 1)
//	assert.Equal(goodRecord, records[0])
//}
//
//func TestMultiRecord(t *testing.T) {
//	assert := assert.New(t)
//	require := require.New(t)
//
//	d, err := connect(gocql.NewCluster("127.0.0.1"))
//	defer d.close()
//	require.NoError(err)
//	// Set up the keyspace and table.
//	if err := d.session.Query("CREATE KEYSPACE IF NOT EXISTS devices").Exec(); err != nil {
//		require.NoError(err)
//	}
//	fmt.Println("Created keyspace devices")
//
//	if err := d.session.Query(`DROP TABLE IF EXISTS devices.events`).Exec(); err != nil {
//		require.NoError(err)
//	}
//
//	if err := d.session.Query(createStmt).Exec(); err != nil {
//		require.NoError(err)
//	}
//
//	recordA := db.Record{
//		DeviceID:  "neat",
//		Type:      db.State,
//		BirthDate: time.Now().UnixNano() + time.Minute.Nanoseconds(),
//		DeathDate: time.Now().Add(time.Second * 30).UnixNano() + time.Minute.Nanoseconds(),
//		Data:      []byte("hello"),
//		Nonce:     []byte("hello"),
//		KID:       "A",
//		Alg:       "A",
//	}
//	recordB := db.Record{
//		DeviceID:  "neat",
//		Type:      db.Default,
//		BirthDate: time.Now().UnixNano(),
//		DeathDate: time.Now().Add(time.Second * 2).UnixNano(),
//		Data:      []byte("hello world"),
//		Nonce:     []byte("hello world"),
//		KID:       "B",
//		Alg:       "B",
//	}
//	recordC := db.Record{
//		DeviceID:  "neat",
//		Type:      db.Default,
//		BirthDate: time.Now().UnixNano() + 100,
//		DeathDate: time.Now().Add(time.Second * 2).UnixNano() + 100,
//		Data:      []byte("hello world C"),
//		Nonce:     []byte("hello world C"),
//		KID:       "C",
//		Alg:       "C",
//	}
//
//	_, err = d.insert([]db.Record{
//		recordC,
//		recordB,
//		recordA,
//	})
//	assert.NoError(err)
//	time.Sleep(10)
//
//	records, err := d.findRecords(100, "WHERE device_id=?", "neat")
//	assert.NoError(err)
//
//	for _, r := range records {
//		fmt.Println(r.KID)
//	}
//	require.Len(records, 3)
//	assert.Equal([]db.Record{recordA, recordC, recordB}, records)
//
//	time.Sleep(30)
//	// testing TTL
//	records, err = d.findRecords(100, "WHERE device_id=?", "neat")
//
//	fmt.Println(records)
//	time.Sleep(10)
//	// testing TTL
//	records, err = d.findRecords(100, "WHERE device_id=?", "neat")
//	fmt.Println(records)
//
//	require.Len(records, 1)
//	assert.Equal([]db.Record{recordA}, records)
//}
