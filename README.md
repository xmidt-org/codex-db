# codex-db

Codex-db provides the database packages needed for the codex project.

[![Build Status](https://travis-ci.com/xmidt-org/codex-db.svg?branch=main)](https://travis-ci.com/xmidt-org/codex-db)
[![codecov.io](http://codecov.io/github/xmidt-org/codex-db/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/codex-db?branch=main)
[![Code Climate](https://codeclimate.com/github/xmidt-org/codex-db/badges/gpa.svg)](https://codeclimate.com/github/xmidt-org/codex-db)
[![Issue Count](https://codeclimate.com/github/xmidt-org/codex-db/badges/issue_count.svg)](https://codeclimate.com/github/xmidt-org/codex-db)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/codex-db)](https://goreportcard.com/report/github.com/xmidt-org/codex-db)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/codex-db/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/release/xmidt-org/codex-db.svg)](CHANGELOG.md)
[![GoDoc](https://godoc.org/github.com/xmidt-org/codex-db?status.svg)](https://godoc.org/github.com/xmidt-org/codex-db)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_codex-db&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_codex-db)

## Summary

Codex-db provides the database packages needed for the [codex project](https://github.com/xmidt-org/codex-deploy).

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Install](#install)
- [Cassandra DB Setup](#cassandra-db-setup)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Install
This repo is a library of packages.  There is no installation.

## Cassandra DB Setup
```cassandraql
CREATE KEYSPACE IF NOT EXISTS devices;
CREATE TABLE devices.events (device_id  varchar,
    record_type INT,
    birthdate BIGINT,
    deathdate BIGINT,
    data BLOB,
    nonce BLOB,
    alg VARCHAR,
    kid VARCHAR,
    row_id TIMEUUID,
    PRIMARY KEY (device_id, birthdate, record_type))
    WITH CLUSTERING ORDER BY (birthdate DESC, record_type ASC)
    AND default_time_to_live = 2768400
    AND transactions = {'enabled': 'false'};
CREATE INDEX search_by_record_type ON devices.events
    (device_id, record_type, birthdate) 
    WITH CLUSTERING ORDER BY (record_type ASC, birthdate DESC)
    AND default_time_to_live = 2768400
    AND transactions = {'enabled': 'false', 'consistency_level':'user_enforced'};
CREATE INDEX search_by_row_id ON devices.events
    (device_id, row_id) 
    WITH CLUSTERING ORDER BY (row_id DESC)
    AND default_time_to_live = 2768400
    AND transactions = {'enabled': 'false', 'consistency_level':'user_enforced'};
CREATE TABLE devices.blacklist (device_id varchar PRIMARY KEY, reason varchar);
```

## Contributing
Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
