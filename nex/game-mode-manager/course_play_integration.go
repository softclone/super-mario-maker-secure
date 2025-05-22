package game_mode_manager

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/super-mario-maker/globals"
	nex_datastore_super_mario_maker "github.com/PretendoNetwork/super-mario-maker/nex/datastore/super-mario-maker"
)

// This function should be called when a course play is initiated
func OnCoursePlayInitiated(userID uint32, courseID uint64) bool {
	// Check if the user can attempt the course based on game mode state
	if !Manager.CanAttemptCourse(userID) {
		globals.Logger.Infof("User %d tried to play course %d but has no lives left", userID, courseID)
		return false
	}

	return true
}

// This function should be called when a course play is completed
func OnCoursePlayCompleted(userID uint32, courseID uint64, success bool) {
	// Update the game mode state based on the result
	Manager.UpdateLastResult(userID, success)

	// Check for win/loss conditions
	if Manager.HasWon(userID) {
		globals.Logger.Infof("User %d has won the 100 Man Challenge by clearing %d courses", userID, Manager.GetState(userID).CoursesCleared)
	} else if Manager.HasLost(userID) {
		globals.Logger.Infof("User %d has lost the 100 Man Challenge with %d lives remaining", userID, Manager.GetState(userID).LivesRemaining)
	}
}

// Hook into the recommended course search to check if the user is in a game mode
func RecommendedCourseSearchObjectHook(err error, packet nex.PacketInterface, callID uint32, param nex_datastore_super_mario_maker.DataStoreSearchParam, extraData types.List[types.String]) (*nex.RMCMessage, *nex.Error) {
	// Get the user ID from the packet
	userID := packet.GetSenderID()

	// Check if the user is in a game mode
	state := Manager.GetState(userID)
	if state != nil && state.Mode == GameMode100Man {
		// If the user has no lives left, don't return any courses
		if Manager.HasLost(userID) {
			globals.Logger.Infof("User %d tried to get recommended courses but has no lives left", userID)
			return nex.NewRMCError(nex.ResultCodes.DataStore.NotFound, "No courses found"), nil
		}
	}

	// Call the original function
	return nex_datastore_super_mario_maker.RecommendedCourseSearchObject(err, packet, callID, param, extraData)
}