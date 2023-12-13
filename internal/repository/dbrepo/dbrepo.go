package dbrepo

import (
	"github.com/elidotexe/backend_byteurl/internal/config"
	"gorm.io/gorm"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *gorm.DB
}

type testDBRepo struct {
	App *config.AppConfig
	DB  *gorm.DB
}

func NewPostgresRepo(db *gorm.DB, a *config.AppConfig) *postgresDBRepo {
	return &postgresDBRepo{
		App: a,
		DB:  db,
	}
}

func NewTestingRepo(a *config.AppConfig) *testDBRepo {
	return &testDBRepo{
		App: a,
	}
}
