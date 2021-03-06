// Package mysqlstore is a MySQL-based session store for the SCS session package.
//
// A working MySQL database is required, containing a sessions table with
// the definition:
//
//	CREATE TABLE sessions (
//	  token CHAR(43) PRIMARY KEY,
//	  data BLOB NOT NULL,
//	  expiry TIMESTAMP(6) NOT NULL
//	);
//	CREATE INDEX sessions_expiry_idx ON sessions (expiry);
//
// The mysqlstore package provides a background 'cleanup' goroutine to delete expired
// session data. This stops the database table from holding on to invalid sessions
// forever and growing unnecessarily large.
package mysqlstore

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	// Register go-sql-driver/mysql with database/sql
	_ "github.com/go-sql-driver/mysql"
)

// MySQLStore represents the currently configured session session store.
type MySQLStore struct {
	*sql.DB
	version     string
	stopCleanup chan bool
}

// New returns a new MySQLStore instance.
//
// The cleanupInterval parameter controls how frequently expired session data
// is removed by the background cleanup goroutine. Setting it to 0 prevents
// the cleanup goroutine from running (i.e. expired sessions will not be removed).
func New(db *sql.DB, cleanupInterval time.Duration) *MySQLStore {
	m := &MySQLStore{
		DB:      db,
		version: getVersion(db),
	}

	if cleanupInterval > 0 {
		go m.startCleanup(cleanupInterval)
	}

	return m
}

// Find returns the data for a given session token from the MySQLStore instance. If
// the session token is not found or is expired, the returned exists flag will be
// set to false.
func (m *MySQLStore) Find(token string) ([]byte, bool, error) {
	var b []byte
	var stmt string

	if compareVersion("5.6.4", m.version) >= 0 {
		stmt = "SELECT data FROM sessions WHERE token = ? AND UTC_TIMESTAMP(6) < expiry"
	} else {
		stmt = "SELECT data FROM sessions WHERE token = ? AND UTC_TIMESTAMP < expiry"
	}

	row := m.DB.QueryRow(stmt, token)
	err := row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func (m *MySQLStore) Loads() ([][]byte, error) {
	var bs [][]byte
	var stmt string

	if compareVersion("5.6.4", m.version) >= 0 {
		stmt = "SELECT data FROM sessions WHERE UTC_TIMESTAMP(6) < expiry"
	} else {
		stmt = "SELECT data FROM sessions WHERE UTC_TIMESTAMP < expiry"
	}

	rows, err := m.DB.Query(stmt)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	for rows.Next() {
		var b []byte
		err = rows.Scan(&b)
		bs = append(bs, b)
	}
	return bs, err
}

// Dumps 数据存储
func (m *MySQLStore) Dumps() (err error) {
	return nil
}

// Save adds a session token and data to the MySQLStore instance with the given expiry
// time. If the session token already exists then the data and expiry time are updated.
func (m *MySQLStore) Save(token string, b []byte, expiry time.Time) error {
	_, err := m.DB.Exec("INSERT INTO sessions (token, data, expiry) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data), expiry = VALUES(expiry)", token, b, expiry.UTC())
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a session token and corresponding data from the MySQLStore instance.
func (m *MySQLStore) Delete(token string) error {
	_, err := m.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (m *MySQLStore) startCleanup(interval time.Duration) {
	m.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := m.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-m.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the MySQLStore instance.
// It's rare to terminate this; generally MySQLStore instances and their cleanup
// goroutines are intended to be long-lived and run for the lifetime of  your
// application.
//
// There may be occasions though when your use of the MySQLStore is transient. An
// example is creating a new MySQLStore instance in a test function. In this scenario,
// the cleanup goroutine (which will run forever) will prevent the MySQLStore object
// from being garbage collected even after the test function has finished. You
// can prevent this by manually calling StopCleanup.
func (m *MySQLStore) StopCleanup() {
	if m.stopCleanup != nil {
		m.stopCleanup <- true
	}
}

func (m *MySQLStore) deleteExpired() error {
	var stmt string

	if compareVersion("5.6.4", m.version) >= 0 {
		stmt = "DELETE FROM sessions WHERE expiry < UTC_TIMESTAMP(6)"
	} else {
		stmt = "DELETE FROM sessions WHERE expiry < UTC_TIMESTAMP"
	}

	_, err := m.DB.Exec(stmt)
	return err
}

func getVersion(db *sql.DB) string {
	var version string
	row := db.QueryRow("SELECT VERSION()")
	err := row.Scan(&version)
	if err != nil {
		return ""
	}
	return strings.Split(version, "-")[0]
}

// Based on https://stackoverflow.com/a/26729704
func compareVersion(a, b string) (ret int) {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	loopMax := len(bs)
	if len(as) > len(bs) {
		loopMax = len(as)
	}
	for i := 0; i < loopMax; i++ {
		var x, y string
		if len(as) > i {
			x = as[i]
		}
		if len(bs) > i {
			y = bs[i]
		}
		xi, _ := strconv.Atoi(x)
		yi, _ := strconv.Atoi(y)
		if xi > yi {
			ret = -1
		} else if xi < yi {
			ret = 1
		}
		if ret != 0 {
			break
		}
	}
	return
}
