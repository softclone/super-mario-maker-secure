package game_mode_manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameModeManager(t *testing.T) {
	// Initialize the manager
	Init()

	// Test user ID
	userID := uint32(1)

	// Test creating a game mode state
	state := Manager.CreateState(userID, GameMode100Man)
	assert.NotNil(t, state)
	assert.Equal(t, GameMode100Man, state.Mode)
	assert.Equal(t, InitialLives, state.LivesRemaining)
	assert.Equal(t, 0, state.CoursesCleared)
	assert.Equal(t, 0, state.CoursesAttempted)
	assert.Equal(t, false, state.LastResult)

	// Test updating last result with success
	Manager.UpdateLastResult(userID, true)
	state = Manager.GetState(userID)
	assert.True(t, state.LastResult)
	assert.Equal(t, 1, state.CoursesCleared)
	assert.Equal(t, 1, state.CoursesAttempted)
	assert.Equal(t, InitialLives, state.LivesRemaining) // Should gain a life

	// Test updating last result with failure
	Manager.UpdateLastResult(userID, false)
	state = Manager.GetState(userID)
	assert.False(t, state.LastResult)
	assert.Equal(t, 1, state.CoursesCleared)
	assert.Equal(t, 2, state.CoursesAttempted)
	assert.Equal(t, InitialLives-1, state.LivesRemaining) // Should lose a life

	// Test win condition
	for i := 0; i < WinConditionCoursesCleared-1; i++ {
		Manager.UpdateLastResult(userID, true)
	}
	state = Manager.GetState(userID)
	assert.False(t, Manager.HasWon(userID))
	assert.True(t, Manager.HasLost(userID)) // Should have lost due to running out of lives

	// Test deleting state
	Manager.DeleteState(userID)
	state = Manager.GetState(userID)
	assert.Nil(t, state)
}