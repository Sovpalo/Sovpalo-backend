package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyPostgres struct {
	pool *pgxpool.Pool
}

func NewCompanyRepository(pool *pgxpool.Pool) *CompanyPostgres {
	return &CompanyPostgres{pool: pool}
}
