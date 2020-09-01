package health

import (
	"context"
	"database/sql"
)

type DbIndicator struct {
	db             *sql.DB
	driverName     string
	validatorQuery string
	versionQuery   string
}

func (r *DbIndicator) Health(ctx context.Context) Health {
	h := NewHealth()
	var (
		ok      string
		version string
		err     error
	)

	if r.db == nil {
		return h
	}
	h.WithDetail("database", r.driverName)

	if r.validatorQuery != "" {
		h.WithDetail("validationQuery", r.validatorQuery)
		err = r.db.QueryRowContext(ctx, r.validatorQuery).Scan(&ok)
	} else {
		h.WithDetail("validationQuery", "ping()")
		err = r.db.PingContext(ctx)
	}

	if err != nil {
		return *h.WithDown().WithDetail("err", err.Error())
	}

	if r.versionQuery != "" {
		err = r.db.QueryRowContext(ctx, r.versionQuery).Scan(&version)
		if err != nil {
			h.WithDetail("version", err.Error())
		} else {
			h.WithDetail("version", version)
		}
	}
	return *h.WithUp()
}

func NewDbHealthChecker(db *sql.DB, driverName, validationQuery, versionQuery string) Indicator {
	return &DbIndicator{db: db, driverName: driverName, validatorQuery: validationQuery, versionQuery: versionQuery}
}
