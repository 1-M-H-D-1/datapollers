package repositories

import (
	"DataPoller/internal/common/domain/entities"
)

type DataSourcesRepository interface {
	FindAll() ([]*entities.DataSource, error)
	FindById(id int) (*entities.DataSource, error)
}
