package database

import (
	"database/sql/driver"
	"github.com/google/uuid"
)

// NullUUID represents a uuid.UUID that may be null. It is
// similar to sql.NullString and other sql Null* types.
type NullUUID struct {
	UUID  uuid.UUID
	Valid bool // Valid is true if UUID is not NULL
}

// Scan implements sql.Scanner and wraps uuid.Scan
func (n *NullUUID) Scan(value interface{}) error {
	n.Valid = value != nil
	if n.Valid {
		return n.UUID.Scan(value)
	} else {
		return nil
	}
}

// Value implements sql.Valuer and wraps uuid.Value
func (n NullUUID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.UUID.Value()
}
