package database

import (
	"bot/internal/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Get a command by name in a chat.
func GetCommand(db *pgxpool.Pool, chatid int, name string) (models.Command, bool, error) {
	var cmd models.Command
	row := db.QueryRow(context.Background(), "SELECT chatid, name, reply FROM commands WHERE chatid = $1 AND name = $2", chatid, name)
	err := row.Scan(&cmd.ChatID, &cmd.Name, &cmd.Reply)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Command{}, false, nil
		}
		return models.Command{}, false, models.NewDatabaseError(err)
	}
	return cmd, true, nil
}

func CreateCommand(db *pgxpool.Pool, cmd models.Command) error {
	_, err := db.Exec(context.Background(), "INSERT INTO commands (chatid, name, reply) VALUES ($1, $2, $3)", cmd.ChatID, cmd.Name, cmd.Reply)
	if err != nil {
		return models.NewDatabaseError(err)
	}
	return nil
}

func DeleteCommand(db *pgxpool.Pool, cmd models.Command) error {
	_, err := db.Exec(context.Background(), "DELETE FROM commands WHERE chatid = $1 AND name = $2", cmd.ChatID, cmd.Name)
	if err != nil {
		return models.NewDatabaseError(err)
	}
	return nil
}
