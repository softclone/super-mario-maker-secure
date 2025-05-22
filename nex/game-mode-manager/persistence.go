package game_mode_manager

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/PretendoNetwork/super-mario-maker/database"
	"github.com/PretendoNetwork/super-mario-maker/globals"
)

const (
	tableName = "game_mode_state"
)

type PersistedGameModeState struct {
	UserID           uint32    `json:"user_id"`
	Mode             string    `json:"mode"`
	LivesRemaining   int       `json:"lives_remaining"`
	CoursesCleared   int       `json:"courses_cleared"`
	CoursesAttempted int       `json:"courses_attempted"`
	LastResult       bool      `json:"last_result"`
	LastUpdated      time.Time `json:"last_updated"`
}

func init() {
	// Create table if it doesn't exist
	query := `
	CREATE TABLE IF NOT EXISTS ` + tableName + ` (
		user_id BIGINT PRIMARY KEY,
		mode TEXT NOT NULL,
		lives_remaining INTEGER NOT NULL,
		courses_cleared INTEGER NOT NULL,
		courses_attempted INTEGER NOT NULL,
		last_result BOOLEAN NOT NULL,
		last_updated TIMESTAMPTZ NOT NULL
	);`

	_, err := database.Postgres.Exec(query)
	if err != nil {
		globals.Logger.Error("Failed to create game mode state table: " + err.Error())
	}
}

func SaveState(userID uint32, state *GameModeState) error {
	persistedState := PersistedGameModeState{
		UserID:           userID,
		Mode:             string(state.Mode),
		LivesRemaining:   state.LivesRemaining,
		CoursesCleared:   state.CoursesCleared,
		CoursesAttempted: state.CoursesAttempted,
		LastResult:       state.LastResult,
		LastUpdated:      time.Now(),
	}

	jsonData, err := json.Marshal(persistedState)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO ` + tableName + ` (user_id, mode, lives_remaining, courses_cleared, courses_attempted, last_result, last_updated)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (user_id) DO UPDATE SET
		mode = EXCLUDED.mode,
		lives_remaining = EXCLUDED.lives_remaining,
		courses_cleared = EXCLUDED.courses_cleared,
		courses_attempted = EXCLUDED.courses_attempted,
		last_result = EXCLUDED.last_result,
		last_updated = EXCLUDED.last_updated`

	_, err = database.Postgres.Exec(query,
		persistedState.UserID,
		persistedState.Mode,
		persistedState.LivesRemaining,
		persistedState.CoursesCleared,
		persistedState.CoursesAttempted,
		persistedState.LastResult,
		persistedState.LastUpdated)

	return err
}

func LoadState(userID uint32) (*GameModeState, error) {
	var jsonData []byte
	var lastUpdated time.Time

	query := `
	SELECT mode, lives_remaining, courses_cleared, courses_attempted, last_result, last_updated
	FROM ` + tableName + `
	WHERE user_id = $1`

	err := database.Postgres.QueryRow(query, userID).Scan(
		&jsonData,
		&lastUpdated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No state found for this user
		}
		return nil, err
	}

	var persistedState PersistedGameModeState
	err = json.Unmarshal(jsonData, &persistedState)
	if err != nil {
		return nil, err
	}

	state := &GameModeState{
		Mode:           GameMode(persistedState.Mode),
		LivesRemaining: persistedState.LivesRemaining,
		CoursesCleared: persistedState.CoursesCleared,
		CoursesAttempted: persistedState.CoursesAttempted,
		LastResult:      persistedState.LastResult,
	}

	return state, nil
}

func DeleteState(userID uint32) error {
	query := `
	DELETE FROM ` + tableName + `
	WHERE user_id = $1`

	_, err := database.Postgres.Exec(query, userID)
	return err
}

func SaveAllStates() error {
	Manager.RLock()
	defer Manager.RUnlock()

	for userID, state := range Manager.states {
		err := SaveState(userID, state)
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadAllStates() error {
	rows, err := database.Postgres.Query(`
	SELECT user_id, mode, lives_remaining, courses_cleared, courses_attempted, last_result, last_updated
	FROM ` + tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	Manager.Lock()
	defer Manager.Unlock()

	for rows.Next() {
		var userID uint32
		var mode string
		var livesRemaining int
		var coursesCleared int
		var coursesAttempted int
		var lastResult bool
		var lastUpdated time.Time

		err := rows.Scan(
			&userID,
			&mode,
			&livesRemaining,
			&coursesCleared,
			&coursesAttempted,
			&lastResult,
			&lastUpdated,
		)
		if err != nil {
			return err
		}

		state := &GameModeState{
			Mode:           GameMode(mode),
			LivesRemaining: livesRemaining,
			CoursesCleared: coursesCleared,
			CoursesAttempted: coursesAttempted,
			LastResult:      lastResult,
		}

		Manager.states[userID] = state
	}

	return nil
}