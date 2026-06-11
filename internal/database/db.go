package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/moritys/world_of_gophers/internal/models"
)

const createTablesQuery = `
CREATE TABLE IF NOT EXISTS player (
	id BIGINT PRIMARY KEY,
	name TEXT,
	level INTEGER DEFAULT 1
);
`

// создать ф-ю для создания пользователя
// получение пользователя по айди

func GetDBPool(connString string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func CreateTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, createTablesQuery)
	return err
}

func CreatePlayer(
	ctx context.Context,
	pool *pgxpool.Pool,
	playerID int,
	name string,
) error {
	query := `
	INSERT INTO player (id, name)
	VALUES ($1, $2);
	`

	_, err := pool.Exec(ctx, query, playerID, name)
	return err
}

func GetUserByID(
	ctx context.Context,
	pool *pgxpool.Pool,
	playerID int,
) (models.Player, error) {
	query := `
	SELECT id, name, level FROM player
	WHERE id=$1;
	`

	player := models.Player{}

	err := pool.QueryRow(ctx, query, playerID).Scan(&player.ID, &player.Name, &player.Level)
	if err != nil {
		fmt.Println("Ошибка поиска игрока", err)
	}

	return player, err
}
