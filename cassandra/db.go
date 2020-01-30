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

// package cassandra provides a way to connect to a cassandra database to
// keep track of device events.
package cassandra

import (
	"errors"
	"time"

	"github.com/InVisionApp/go-health/v2"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/goph/emperror"
	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/codex-db/blacklist"
	"github.com/yugabyte/gocql"
)

var (
	errTableNotExist    = errors.New("Table does not exist")
	errInvaliddeviceID  = errors.New("Invalid device ID")
	errInvalidEventType = errors.New("Invalid event type")
	errNoEvents         = errors.New("no records to be inserted")
)

const (
	defaultOpTimeout             = time.Duration(10) * time.Second
	defaultDatabase              = "devices"
	defaultNumRetries            = 0
	defaultWaitTimeMult          = 1
	defaultMaxNumberConnsPerHost = 2
)

type Config struct {
	// Hosts to  connect to. Must have at least one
	Hosts []string

	// Database aka Keyspace for cassandra
	Database string

	// OpTimeout
	OpTimeout time.Duration

	// SSLRootCert used for enabling tls to the cluster. SSLKey, and SSLCert must also be set.
	SSLRootCert string
	// SSLKey used for enabling tls to the cluster. SSLRootCert, and SSLCert must also be set.
	SSLKey string
	// SSLCert used for enabling tls to the cluster. SSLRootCert, and SSLRootCert must also be set.
	SSLCert string
	// If you want to verify the hostname and server cert (like a wildcard for cass cluster) then you should turn this on
	// This option is basically the inverse of InSecureSkipVerify
	// See InSecureSkipVerify in http://golang.org/pkg/crypto/tls/ for more info
	EnableHostVerification bool

	// Username to authenticate into the cluster. Password must also be provided.
	Username string
	// Password to authenticate into the cluster. Username must also be provided.
	Password string

	// NumRetries for connecting to the db
	NumRetries int

	// WaitTimeMult the amount of time to wait before retrying to connect to the db
	WaitTimeMult time.Duration

	// MaxConnsPerHost max number of connections per host
	MaxConnsPerHost int
}

type Connection struct {
	finder       finder
	findList     findList
	deviceFinder deviceFinder
	multiInsert  multiInserter
	closer       closer
	pinger       pinger

	opTimeout   time.Duration
	health      *health.Health
	measures    Measures
	stopThreads []chan struct{}
}

func CreateDbConnection(config Config, provider provider.Provider, health *health.Health) (*Connection, error) {
	if len(config.Hosts) == 0 {
		return &Connection{}, errors.New("number of hosts must be > 0")
	}

	validateConfig(&config)

	clusterConfig := gocql.NewCluster(config.Hosts...)
	clusterConfig.Consistency = gocql.LocalQuorum
	clusterConfig.Keyspace = config.Database
	clusterConfig.Timeout = config.OpTimeout
	// let retry package handle it
	clusterConfig.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 1}
	// setup ssl
	if config.SSLRootCert != "" && config.SSLCert != "" && config.SSLKey != "" {
		clusterConfig.SslOpts = &gocql.SslOptions{
			CertPath:               config.SSLCert,
			KeyPath:                config.SSLKey,
			CaPath:                 config.SSLRootCert,
			EnableHostVerification: config.EnableHostVerification,
		}
	}
	// setup authentication
	if config.Username != "" && config.Password != "" {
		clusterConfig.Authenticator = gocql.PasswordAuthenticator{
			Username: config.Username,
			Password: config.Password,
		}
	}

	dbConn := Connection{
		health:   health,
		measures: NewMeasures(provider),
	}

	conn, err := connectWithMetrics(clusterConfig, dbConn.measures)

	// retry if it fails
	waitTime := 1 * time.Second
	for attempt := 0; attempt < config.NumRetries && err != nil; attempt++ {
		time.Sleep(waitTime)
		conn, err = connectWithMetrics(clusterConfig, dbConn.measures)
		waitTime = waitTime * config.WaitTimeMult
	}
	if err != nil {
		return &Connection{}, emperror.WrapWith(err, "Connecting to database failed", "hosts", config.Hosts)
	}

	dbConn.finder = conn
	dbConn.findList = conn
	dbConn.deviceFinder = conn
	dbConn.multiInsert = conn
	dbConn.closer = conn
	dbConn.pinger = conn

	return &dbConn, nil
}

func validateConfig(config *Config) {
	zeroDuration := time.Duration(0) * time.Second

	if config.OpTimeout == zeroDuration {
		config.OpTimeout = defaultOpTimeout
	}

	if config.Database == "" {
		config.Database = defaultDatabase
	}
	if config.NumRetries < 0 {
		config.NumRetries = defaultNumRetries
	}
	if config.WaitTimeMult < 1 {
		config.WaitTimeMult = defaultWaitTimeMult
	}
	if config.MaxConnsPerHost <= 0 {
		config.MaxConnsPerHost = defaultMaxNumberConnsPerHost
	}
}

// GetRecords returns a list of records for a given device.
func (c *Connection) GetRecords(deviceID string, limit int, stateHash string) ([]db.Record, error) {
	filterString := "WHERE device_id=?"
	items := []interface{}{deviceID}
	if stateHash != "" {
		filterString = "WHERE device_id = ? AND row_id > ?"
		items = []interface{}{deviceID, stateHash}
	}
	deviceInfo, err := c.finder.findRecords(limit, filterString, items...)
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []db.Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	c.measures.SQLReadRecords.Add(float64(len(deviceInfo)))
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return deviceInfo, nil
}

// GetRecords returns a list of records for a given device and event type.
func (c *Connection) GetRecordsOfType(deviceID string, limit int, eventType db.EventType, stateHash string) ([]db.Record, error) {
	filterString := "WHERE device_id = ? AND record_type = ?"
	items := []interface{}{deviceID, eventType}
	if stateHash != "" {
		filterString = "WHERE device_id = ? AND record_type = ? AND row_id > ?"
		items = []interface{}{deviceID, eventType, stateHash}
	}
	deviceInfo, err := c.finder.findRecords(limit, filterString, items...)
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []db.Record{}, emperror.WrapWith(err, "Getting records from database failed", "device id", deviceID)
	}
	c.measures.SQLReadRecords.Add(float64(len(deviceInfo)))
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return deviceInfo, nil
}

// GetStateHash returns a hash for the latest record added to the database.
func (c *Connection) GetStateHash(records []db.Record) (string, error) {
	if len(records) == 0 {
		return "", errors.New("record slice is empty")
	} else if len(records) == 1 && records[0].RowID != "" {
		return records[0].RowID, nil
	}
	original := gocql.UUIDFromTime(time.Time{})
	latest := original
	originalBirthDate := int64(0)
	latestBirthDate := originalBirthDate
	for _, elem := range records {
		uuid, err := gocql.ParseUUID(elem.RowID)
		if err != nil {
			continue
		}
		if uuid.Time().UnixNano() > latest.Time().UnixNano() {
			latest = uuid
		}
		if elem.BirthDate > latestBirthDate {
			latestBirthDate = elem.BirthDate
		}
	}
	if latest == original {
		if latestBirthDate == originalBirthDate {
			return gocql.TimeUUID().String(), errors.New("no hash or birthdate found")
		}
		return gocql.UUIDFromTime(time.Unix(0, latestBirthDate)).String(), errors.New("no hash found")
	}
	return latest.String(), nil
}

// GetBlacklist returns a list of blacklisted devices.
func (c *Connection) GetBlacklist() (list []blacklist.BlackListedItem, err error) {
	list, err = c.findList.findBlacklist()
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.BlacklistReadType).Add(1.0)
		return []blacklist.BlackListedItem{}, emperror.WrapWith(err, "Getting records from database failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.BlacklistReadType).Add(1.0)
	return
}

// GetDeviceList returns a list of device ids where the device id is greater
// than the offset device id.
func (c *Connection) GetDeviceList(startDate time.Time, endDate time.Time, offset int, limit int) ([]string, error) {
	list, err := c.deviceFinder.getList(startDate, endDate, offset, limit)
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.ReadType).Add(1.0)
		return []string{}, emperror.WrapWith(err, "Getting list of devices from database failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.ReadType).Add(1.0)
	return list, nil
}

// InsertEvent adds a list of records to the table.
func (c *Connection) InsertRecords(records ...db.Record) error {
	rowsAffected, err := c.multiInsert.insert(records)
	c.measures.SQLInsertedRecords.Add(float64(rowsAffected))
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.InsertType).Add(1.0)
		return emperror.Wrap(err, "Inserting records failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.InsertType).Add(1.0)
	return nil
}

// Ping is for pinging the database to verify that the connection is still good.
func (c *Connection) Ping() error {
	err := c.pinger.ping()
	if err != nil {
		c.measures.SQLQueryFailureCount.With(db.TypeLabel, db.PingType).Add(1.0)
		return emperror.WrapWith(err, "Pinging connection failed")
	}
	c.measures.SQLQuerySuccessCount.With(db.TypeLabel, db.PingType).Add(1.0)
	return nil
}

// Close closes the database connection.
func (c *Connection) Close() error {
	for _, stopThread := range c.stopThreads {
		stopThread <- struct{}{}
	}

	err := c.closer.close()
	if err != nil {
		return emperror.WrapWith(err, "Closing connection failed")
	}
	return nil
}
