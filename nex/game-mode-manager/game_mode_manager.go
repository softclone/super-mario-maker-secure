package game_mode_manager

import (
	"sync"

	"github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/super-mario-maker/globals"
)

const (
	// 100 Man Challenge constants
	InitialLives = 100
	WinConditionCoursesCleared = 100 // Win after clearing this many courses
)

type GameMode string

const (
	GameModeNone GameMode = "None"
	GameMode100Man GameMode = "100Man"
)

type GameModeState struct {
	Mode           GameMode
	LivesRemaining int
	CoursesCleared int
	CoursesAttempted int
	LastResult      bool // true = success, false = failure
}

type GameModeManager struct {
	sync.RWMutex
	states map[uint32]*GameModeState // keyed by user ID
}

var Manager *GameModeManager

func Init() {
	Manager = &GameModeManager{
		states: make(map[uint32]*GameModeState),
	}
	globals.Logger.Info("GameModeManager initialized")
}

func (m *GameModeManager) GetState(userID uint32) *GameModeState {
	m.RLock()
	defer m.RUnlock()

	state, exists := m.states[userID]
	if !exists {
		return nil
	}
	return state
}

func (m *GameModeManager) CreateState(userID uint32, mode GameMode) *GameModeState {
	m.Lock()
	defer m.Unlock()

	state := &GameModeState{
		Mode:           mode,
		LivesRemaining: InitialLives,
		CoursesCleared: 0,
		CoursesAttempted: 0,
		LastResult:      false,
	}

	m.states[userID] = state
	return state
}

func (m *GameModeManager) DeleteState(userID uint32) {
	m.Lock()
	defer m.Unlock()

	delete(m.states, userID)
}

func (m *GameModeManager) UpdateLastResult(userID uint32, success bool) {
	m.Lock()
	defer m.Unlock()

	state, exists := m.states[userID]
	if !exists {
		return
	}

	state.LastResult = success
	state.CoursesAttempted++

	if success {
		state.CoursesCleared++
		state.LivesRemaining = min(state.LivesRemaining+1, InitialLives) // Gain a life for each cleared course
	} else {
		state.LivesRemaining--
	}
}

func (m *GameModeManager) HasWon(userID uint32) bool {
	m.RLock()
	defer m.RUnlock()

	state, exists := m.states[userID]
	if !exists {
		return false
	}

	return state.CoursesCleared >= WinConditionCoursesCleared
}

func (m *GameModeManager) HasLost(userID uint32) bool {
	m.RLock()
	defer m.RUnlock()

	state, exists := m.states[userID]
	if !exists {
		return false
	}

	return state.LivesRemaining <= 0
}

func (m *GameModeManager) CanAttemptCourse(userID uint32) bool {
	m.RLock()
	defer m.RUnlock()

	state, exists := m.states[userID]
	if !exists {
		return true // Not in a game mode, so can attempt course
	}

	return !m.HasLost(userID)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// OnCourseCompleted is called when a user completes a course
// It updates the game mode state based on the course completion
func OnCourseCompleted(courseID uint64, success bool) {
	// In a real implementation, we would get the userID from the context
	// For now, we'll just use a dummy userID
	userID := uint32(1) // TODO: Replace with actual user ID

	// Update the last result for the user
	Manager.UpdateLastResult(userID, success)

	// Check if the user has won or lost
	if Manager.HasWon(userID) {
		globals.Logger.Infof("User %d has won the 100 Man Challenge!", userID)
		// TODO: Handle win condition (e.g., show achievement, notify user)
	} else if Manager.HasLost(userID) {
		globals.Logger.Infof("User %d has lost the 100 Man Challenge!", userID)
		// TODO: Handle loss condition (e.g., show game over screen)
	}
}