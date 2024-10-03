package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/northmule/gophermart/config"
	"github.com/northmule/gophermart/internal/app/repository/models"
	"github.com/northmule/gophermart/internal/app/services/logger"
	"github.com/northmule/gophermart/internal/app/storage"
	"time"
)

type BalanceRepository struct {
	store                      storage.DBQuery
	sqlFindByUserUUID          *sql.Stmt
	sqlCreateBalanceByUserUUID *sql.Stmt
}

func NewBalanceRepository(store storage.DBQuery) *BalanceRepository {
	instance := BalanceRepository{
		store: store,
	}

	var err error
	instance.sqlFindByUserUUID, err = store.Prepare(`
				select b.id, b.value, b.updated_at, u.id, u.login, u.password, u.created_at, u.uuid from user_balance as b
				join users as u on u.id = b.user_id
				where u.uuid = $1
				limit 1
				`)
	if err != nil {
		logger.LogSugar.Error(err)
		return nil
	}
	instance.sqlCreateBalanceByUserUUID, err = store.Prepare(`insert into user_balance (user_id, value) values ((select u.id from users u where u.uuid = $1 limit 1), 0) returning id;`)
	if err != nil {
		logger.LogSugar.Error(err)
		return nil
	}

	return &instance
}

func (br *BalanceRepository) FindOneByUserUUID(ctx context.Context, userUUID string) (*models.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DataBaseConnectionTimeOut*time.Second)
	defer cancel()
	rows, err := br.sqlFindByUserUUID.QueryContext(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("при вызове FindOneByUserUUID(%s) произошла ошибка %w", userUUID, err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("при вызове FindOneByUserUUID(%s) произошла ошибка %w", userUUID, err)
	}
	var balance models.Balance
	var user models.User
	if rows.Next() {
		err = rows.Scan(&balance.ID, &balance.Value, &balance.UpdatedAt, &user.ID, &user.Login, &user.Password, &user.CreatedAt, &user.UUID)
		if err != nil {
			return nil, fmt.Errorf("при обработке значений в FindOneByUserUUID(%s) произошла ошибка %w", userUUID, err)
		}
	}
	balance.User = user
	return &balance, nil
}

func (br *BalanceRepository) CreateBalanceByUserUUID(ctx context.Context, userUUID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DataBaseConnectionTimeOut*time.Second)
	defer cancel()
	rows := br.sqlCreateBalanceByUserUUID.QueryRowContext(ctx, userUUID)
	err := rows.Err()
	if err != nil {
		return 0, fmt.Errorf("при вызове CreateBalanceByUserUUID(%s) произошла ошибка %w", userUUID, err)
	}

	var id int64
	err = rows.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (br *BalanceRepository) TxCreateBalanceByUserUUID(ctx context.Context, tx storage.TxDBQuery, userUUID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DataBaseConnectionTimeOut*time.Second)
	defer cancel()
	rows := tx.Tx().QueryRowContext(ctx, `insert into user_balance (user_id, value) values ((select u.id from users u where u.uuid = $1 limit 1), 0) returning id;`, userUUID)
	err := rows.Err()
	if err != nil {
		tx.AddError(err)
		return 0, fmt.Errorf("при вызове CreateBalanceByUserUUID(%s) произошла ошибка %w", userUUID, err)
	}

	var id int64
	err = rows.Scan(&id)
	if err != nil {
		tx.AddError(err)
		return 0, err
	}

	return id, nil
}
