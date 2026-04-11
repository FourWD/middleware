package common

import "github.com/google/uuid"

func SyncStart(name string, isManual bool) string {
	return uuid.NewString()
}

func SyncEnd(id string, isComplete bool) {

}
