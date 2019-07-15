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

package postgresql

import (
	// Import GORM-related packages.

	"database/sql"
	"fmt"
	"strings"

	"github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/codex-db/blacklist"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type (
	finder interface {
		findRecords(out *[]db.Record, limit int, where ...interface{}) error
		findRecordsToDelete(limit int, shard int, deathDate int64) ([]db.RecordToDelete, error)
	}
	findList interface {
		findBlacklist(out *[]blacklist.BlackListedItem) error
	}
	deviceFinder interface {
		getList(offset string, limit int, where ...interface{}) ([]string, error)
	}
	multiinserter interface {
		insert(records []db.Record) (int64, error)
	}
	deleter interface {
		delete(value *db.Record, limit int, where ...interface{}) (int64, error)
	}
	pinger interface {
		ping() error
	}
	closer interface {
		close() error
	}
	stats interface {
		getStats() sql.DBStats
	}
)

type dbDecorator struct {
	*gorm.DB
}

func (b *dbDecorator) findRecords(out *[]db.Record, limit int, where ...interface{}) error {
	db := b.Order("birth_date desc").Limit(limit).Find(out, where...)
	return db.Error
}

func (b *dbDecorator) findRecordsToDelete(limit int, shard int, deathDate int64) ([]db.RecordToDelete, error) {
	var (
		out []db.RecordToDelete
	)
	db := b.Raw("SELECT death_date, record_id from devices.events WHERE shard = ? AND death_date < ? LIMIT ?", shard, deathDate, limit).Scan(&out)
	// db := b.Order("birth_date desc").Limit(limit).Find(&records, where...).Pluck("record_id", out)
	return out, db.Error
}

func (b *dbDecorator) findBlacklist(out *[]blacklist.BlackListedItem) error {
	db := b.Find(out)
	return db.Error
}

func (b *dbDecorator) getList(offset string, limit int, where ...interface{}) ([]string, error) {
	var result []string
	// Raw SQL
	db := b.Raw("SELECT device_id from devices.events WHERE device_id > ? GROUP BY device_id LIMIT ?", offset, limit).Pluck("device_id", &result)
	//db := b.Limit(limit).Select("device_id").Find(&[]Record{}, where).Group("device_id").Where("device_id > ?", offset).Pluck("device_id", &result)
	return result, db.Error
}

func (b *dbDecorator) insert(records []db.Record) (int64, error) {
	if len(records) == 0 {
		return 0, errNoEvents
	}
	mainScope := b.DB.NewScope(records[0])
	mainFields := mainScope.Fields()
	quoted := make([]string, 0, len(mainFields))
	for i := range mainFields {
		// If primary key has blank value (0 for int, "" for string, nil for interface ...), skip it.
		// If field is ignore field, skip it.
		if (mainFields[i].IsPrimaryKey && mainFields[i].IsBlank) || (mainFields[i].IsIgnored) {
			continue
		}
		quoted = append(quoted, mainScope.Quote(mainFields[i].DBName))
	}
	placeholdersArr := make([]string, 0, len(records))

	for _, obj := range records {
		scope := b.DB.NewScope(obj)
		fields := scope.Fields()
		placeholders := make([]string, 0, len(fields))
		for i := range fields {
			if (fields[i].IsPrimaryKey && fields[i].IsBlank) || (fields[i].IsIgnored) {
				continue
			}
			// the trick it to use mainScope instead of scope so the number keeps on increasing
			// aka $1, $2, $2, etc.
			placeholders = append(placeholders, mainScope.AddToVars(fields[i].Field.Interface()))
		}
		placeholdersStr := "(" + strings.Join(placeholders, ", ") + ")"
		placeholdersArr = append(placeholdersArr, placeholdersStr)
		// add real variables for the replacement of placeholders' '?' letter later.
		mainScope.SQLVars = append(mainScope.SQLVars, scope.SQLVars...)
	}

	mainScope.Raw(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		mainScope.QuotedTableName(),
		strings.Join(quoted, ", "),
		strings.Join(placeholdersArr, ", "),
	))

	result, err := mainScope.SQLDB().Exec(mainScope.SQL, mainScope.SQLVars...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (b *dbDecorator) delete(value *db.Record, limit int, where ...interface{}) (int64, error) {
	var db *gorm.DB
	if limit > 0 {
		db = b.Limit(limit).Delete(value, where...)
	} else {
		db = b.Delete(value, where...)
	}
	return db.RowsAffected, db.Error
}

func (b *dbDecorator) ping() error {
	return b.DB.DB().Ping()
}

func (b *dbDecorator) close() error {
	return b.DB.Close()
}

func (b *dbDecorator) getStats() sql.DBStats {
	return b.DB.DB().Stats()
}

func connect(connSpecStr string) (*dbDecorator, error) {
	c, err := gorm.Open("postgres", connSpecStr)

	if err != nil {
		return nil, err
	}

	db := &dbDecorator{c}

	return db, nil
}
