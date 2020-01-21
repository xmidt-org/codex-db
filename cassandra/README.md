# Cassandra DB driver
This implementation is geared towards yugabyte.

# Migration from  v0.4.0 to v0.5.0
The addition of row_id as a TIMEUUID as a simplictic verison of state hash.
Since row_id can be null gungnir will work with both databasescheme.
In order to do long polling, gungnir db driver will need to be updated.
Svalinn on the other hand is not backwards compatible as the insert statment has changed to include the
TIMEUUID.

The following is the migration script from v0.4.0 to v0.5.0
```cassandraql
ALTER TABLE devices.events ADD row_id TIMEUUID;
CREATE INDEX search_by_row_id ON devices.events
    (device_id, row_id) 
    WITH CLUSTERING ORDER BY (row_id DESC)
    AND default_time_to_live = 2768400
    AND transactions = {'enabled': 'false', 'consistency_level':'user_enforced'};
```