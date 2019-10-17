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
	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/codex-db/blacklist"
	"github.com/yugabyte/gocql"
	"strconv"
	"time"
)

type dbMeasuresDecorator struct {
	measures Measures

	finder
	findList
	deviceFinder
	multiinserter
	pinger
	closer
}

const CountLabel = "count"

func (b *dbMeasuresDecorator) findRecords(limit int, filter string, where ...interface{}) ([]db.Record, error) {

	b.measures.PoolInUseConnections.Add(1.0)
	now := time.Now()
	records, err := b.finder.findRecords(limit, filter, where...)
	b.measures.SQLDuration.With(db.TypeLabel, db.ReadType, CountLabel, strconv.Itoa(len(records))).Observe(time.Since(now).Seconds())
	b.measures.PoolInUseConnections.Add(-1.0)

	return records, err
}

func (b *dbMeasuresDecorator) getList(startDate time.Time, endDate time.Time, offset int, limit int) ([]string, error) {
	b.measures.PoolInUseConnections.Add(1.0)
	now := time.Now()
	result, err := b.deviceFinder.getList(startDate, endDate, offset, limit)
	b.measures.SQLDuration.With(db.TypeLabel, db.ReadType, CountLabel, strconv.Itoa(len(result))).Observe(time.Since(now).Seconds())
	b.measures.PoolInUseConnections.Add(-1.0)

	return result, err
}

func (b *dbMeasuresDecorator) findBlacklist() ([]blacklist.BlackListedItem, error) {
	b.measures.PoolInUseConnections.Add(1.0)
	now := time.Now()
	records, err := b.findList.findBlacklist()
	b.measures.SQLDuration.With(db.TypeLabel, db.BlacklistReadType, CountLabel, strconv.Itoa(len(records))).Observe(time.Since(now).Seconds())
	b.measures.PoolInUseConnections.Add(-1.0)

	return records, err
}

func (b *dbMeasuresDecorator) insert(records []db.Record) (int, error) {
	b.measures.PoolInUseConnections.Add(1.0)
	now := time.Now()
	count, err := b.multiinserter.insert(records)
	b.measures.SQLDuration.With(db.TypeLabel, db.InsertType, CountLabel, strconv.Itoa(len(records))).Observe(time.Since(now).Seconds())
	b.measures.PoolInUseConnections.Add(-1.0)

	return count, err
}

func (b *dbMeasuresDecorator) ping() error {
	b.measures.PoolInUseConnections.Add(1.0)
	now := time.Now()
	err := b.pinger.ping()
	b.measures.SQLDuration.With(db.TypeLabel, db.PingType).Observe(time.Since(now).Seconds())
	b.measures.PoolInUseConnections.Add(-1.0)

	return err
}
func (b *dbMeasuresDecorator) close() error {

	err := b.closer.close()

	return err
}

func connectWithMetrics(clusterConfig *gocql.ClusterConfig, measures Measures) (*dbMeasuresDecorator, error) {

	db, err := connect(clusterConfig)
	if err != nil {
		return nil, err
	}

	return &dbMeasuresDecorator{
		measures:      measures,
		finder:        db,
		findList:      db,
		deviceFinder:  db,
		multiinserter: db,
		pinger:        db,
		closer:        db,
	}, nil

}
