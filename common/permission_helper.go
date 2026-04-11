package common

type Permission struct {
	Read   int `json:"read"`
	Create int `json:"create"`
	Update int `json:"update"`
	Delete int `json:"delete"`
}

func checkPermissionByAction(permission Permission, action string) bool {
	switch action {
	case "READ":
		return permission.Read == 1
	case "CREATE":
		return permission.Create == 1
	case "UPDATE":
		return permission.Update == 1
	case "DELETE":
		return permission.Delete == 1
	default:
		return false
	}
}
