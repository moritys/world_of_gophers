package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/moritys/world_of_gophers/internal/models"
)

const createTablesQuery = `
CREATE TABLE IF NOT EXISTS player (
	id BIGINT PRIMARY KEY,
	name TEXT NOT NULL,
	level INTEGER NOT NULL DEFAULT 1,
	xp INTEGER NOT NULL DEFAULT 0,
	gold INTEGER NOT NULL DEFAULT 0,
	strength INTEGER NOT NULL DEFAULT 0,
	knowledge INTEGER NOT NULL DEFAULT 0,
	focus INTEGER NOT NULL DEFAULT 0
);
`

var ErrPlayerNotFound = errors.New("игрок не найден")

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) CreatePlayer(
	ctx context.Context,
	id int64,
	name string,
) error {
	query := `
	INSERT INTO player (id, name)
	VALUES ($1, $2)
	ON CONFLICT (id) DO NOTHING;
	`
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx, query, id, name)
	return err
}

func (s *Storage) GetPlayer(
	ctx context.Context,
	id int64,
) (models.Player, error) {
	query := `
	SELECT id, name, level, xp, gold, strength, knowledge, focus FROM player
	WHERE id=$1;
	`
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	player := models.Player{}

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&player.ID, &player.Name, &player.Level, &player.XP, &player.Gold, &player.Strength, &player.Knowledge, &player.Focus)

	if errors.Is(err, pgx.ErrNoRows) {
		return player, ErrPlayerNotFound
	}
	if err != nil {
		return player, fmt.Errorf("поиск игрока %d: %w", id, err)
	}

	return player, nil
}

func GetDBPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func CreateTables(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, createTablesQuery)
	return err
}
