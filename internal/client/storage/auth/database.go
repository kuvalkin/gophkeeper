package auth

import "context"

func NewDatabaseRepository() (*DatabaseRepository, error) {

}

type DatabaseRepository struct {
}

func (d *DatabaseRepository) GetToken(ctx context.Context) (string, bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DatabaseRepository) SetToken(ctx context.Context, token string) error {
	//TODO implement me
	panic("implement me")
}
