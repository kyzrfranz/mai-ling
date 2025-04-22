# mai-ling
mailing backend to store print jobs & provide simple stats.

Used by https://www.stoppt-scheinselbststaendigkeit.de

## Configuration

- `MONGO_URL` is the only required environment variable. It should be set to the MongoDB connection string.
- `MONGO_DBNAME` is the name of the database to use. It defaults to `test`.
- `MONGO_COLLECTION` is the name of the collection to use. It defaults to `test`.


## How to run

run it like this:
```
make run
```