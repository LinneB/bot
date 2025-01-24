package database

import (
	"bot/internal/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetChats(db *pgxpool.Pool) ([]models.Chat, error) {
	rows, err := db.Query(context.Background(), "SELECT chatid, chatname FROM chats")
	if err != nil {
		return nil, err
	}
	chats, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Chat])
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func GetChatByName(db *pgxpool.Pool, chatname string) (models.Chat, bool, error) {
	var chat models.Chat
	row := db.QueryRow(context.Background(), "SELECT chatid, chatname FROM chats WHERE chatname = $1", chatname)
	err := row.Scan(&chat.ChatID, &chat.ChatName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Chat{}, false, nil
		}
		return models.Chat{}, false, err
	}
	return chat, true, nil
}

func DeleteChat(db *pgxpool.Pool, chat models.Chat) error {
	_, err := db.Exec(context.Background(), "DELETE FROM chats WHERE chatid = $1", chat.ChatID)
	return err
}

func InsertChat(db *pgxpool.Pool, chat models.Chat) error {
	_, err := db.Exec(context.Background(), "INSERT INTO chats (chatid, chatname) VALUES ($1, $2)", chat.ChatID, chat.ChatName)
	return err
}
