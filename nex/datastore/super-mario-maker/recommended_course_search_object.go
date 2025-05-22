package nex_datastore_super_mario_maker

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore_super_mario_maker "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/super-mario-maker"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
	datastore_smm_db "github.com/PretendoNetwork/super-mario-maker/database/datastore/super-mario-maker"
	"github.com/PretendoNetwork/super-mario-maker/globals"
	"strconv"
	"github.com/PretendoNetwork/super-mario-maker/nex/game-mode-manager"
)

func RecommendedCourseSearchObject(err error, packet nex.PacketInterface, callID uint32, param datastore_types.DataStoreSearchParam, extraData types.List[types.String]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, err.Error())
	}

	// * This method is used in 100 Mario and Course World
	// *
	// * extraData seems to be a set of filters defining a
	// * range for a courses success rate
	// *
	// * Course World (All)          ["",  "",   "",    "0", "0"]
	// * Course World (Easy)         ["1", "0",  "34",  "0", "0"]
	// * Course World (Normal)       ["1", "35", "74",  "0", "0"]
	// * Course World (Expert)       ["1", "75", "95",  "0", "0"]
	// * Course World (Super Expert) ["1", "96", "100", "0", "0"]
	// *
	// * Indexes 1 and 2 seem to be a min and max for the *failure*
	// * rate of the courses. This is now taken into account with
	// * the difficulty filtering.

	// HACK The database load is exponential here and
	length := int(param.ResultRange.Length)
	maxLength := 25
	if length < 0 || length > maxLength {
		globals.Logger.Warningf("Limiting request to %d courses (was %d)", maxLength, length)
		length = maxLength
	}

	// Check if the user is in a game mode
	userID := packet.GetSenderID()
	state := game_mode_manager.Manager.GetState(userID)
	if state != nil && state.Mode == game_mode_manager.GameMode100Man {
		// If the user has no lives left, don't return any courses
		if game_mode_manager.Manager.HasLost(userID) {
			globals.Logger.Infof("User %d tried to get recommended courses but has no lives left", userID)
			return nex.NewRMCError(nex.ResultCodes.DataStore.NotFound, "No courses found"), nil
		}
	}

	// Determine difficulty based on extraData
	var difficulty datastore_smm_db.Difficulty
	if len(extraData) >= 1 && extraData[0].Value == "1" {
		if len(extraData) >= 2 && len(extraData) >= 3 {
			minFailureRate, _ := strconv.Atoi(extraData[1].Value)
			maxFailureRate, _ := strconv.Atoi(extraData[2].Value)

			if minFailureRate == 0 && maxFailureRate == 34 {
				difficulty = datastore_smm_db.DifficultyEasy
			} else if minFailureRate == 35 && maxFailureRate == 74 {
				difficulty = datastore_smm_db.DifficultyNormal
			} else if minFailureRate == 75 && maxFailureRate == 95 {
				difficulty = datastore_smm_db.DifficultyExpert
			} else if minFailureRate == 96 && maxFailureRate == 100 {
				difficulty = datastore_smm_db.DifficultySuperExpert
			} else {
				// Unknown difficulty, default to All
				difficulty = datastore_smm_db.DifficultyAll
			}
		} else {
			// Not enough data to determine difficulty, default to All
			difficulty = datastore_smm_db.DifficultyAll
		}
	} else {
		// First element is not "1" or missing, treat as All
		difficulty = datastore_smm_db.DifficultyAll
	}

	globals.Logger.Infof("Selected difficulty: %s", difficulty)

	// TODO - Use the offset? Real client never uses it, but might be nice for completeness sake?
	pRankingResults, nexError := datastore_smm_db.GetRandomCoursesWithLimit(length, difficulty)
	if nexError != nil {
		return nil, nexError
	}

	rmcResponseStream := nex.NewByteStreamOut(globals.SecureServer.LibraryVersions, globals.SecureServer.ByteStreamSettings)

	pRankingResults.WriteTo(rmcResponseStream)

	rmcResponse := nex.NewRMCSuccess(globals.SecureEndpoint, rmcResponseStream.Bytes())
	rmcResponse.ProtocolID = datastore_super_mario_maker.ProtocolID
	rmcResponse.MethodID = datastore_super_mario_maker.MethodRecommendedCourseSearchObject
	rmcResponse.CallID = callID

	return rmcResponse, nil
}
