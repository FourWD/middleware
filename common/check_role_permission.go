package common

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func CheckRolePermission(c *fiber.Ctx, path string, action string) error {
	userID := GetSessionUserID(c)

	if !StringExistsInList(action, []string{"READ", "CREATE", "UPDATE", "DELETE"}) {
		return errors.New("Unauthorized")
	}

	var permission Permission

	sql := `SELECT
	COALESCE(rup.is_read, 0) "read",
	COALESCE(rup.is_create, 0) "create",
	COALESCE(rup.is_update, 0) "update",
	COALESCE(rup.is_delete, 0) "delete"
	FROM role_menus AS rm
	LEFT JOIN role_user_permissions rup ON rup.role_menu_id = rm.id AND rup.user_id = ?
	WHERE LOWER(rm.path) = ?`

	Database.Raw(sql, userID, strings.ToLower(path)).Scan(&permission)

	if checkPermissionByAction(permission, action) {
		return nil
	}

	return errors.New("Unauthorized")
}
