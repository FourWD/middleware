package common

import (
	"errors"
	"strings"

	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/kit"
	"github.com/gofiber/fiber/v3"
)

func CheckRolePermission(c fiber.Ctx, path string, action string) error {
	userID := infra.GetSessionUserID(c)

	if !kit.StringExistsInList(action, []string{"READ", "CREATE", "UPDATE", "DELETE"}) {
		return errors.New("Unauthorized")
	}

	var permission kit.Permission

	// is_read/is_create/is_update/is_delete เป็น bool ใน orm ส่วน kit.Permission เป็น int
	// COALESCE(bool, 0) ใช้ไม่ได้บน PostgreSQL ("COALESCE types boolean and integer
	// cannot be matched") — ห่อด้วย CASE จึงคืน int ให้ตรงกับ struct และรันได้ทั้งสอง engine
	sql := `SELECT
	CASE WHEN COALESCE(rup.is_read, false)   THEN 1 ELSE 0 END AS "read",
	CASE WHEN COALESCE(rup.is_create, false) THEN 1 ELSE 0 END AS "create",
	CASE WHEN COALESCE(rup.is_update, false) THEN 1 ELSE 0 END AS "update",
	CASE WHEN COALESCE(rup.is_delete, false) THEN 1 ELSE 0 END AS "delete"
	FROM role_menus AS rm
	LEFT JOIN role_user_permissions rup ON rup.role_menu_id = rm.id AND rup.user_id = ?
	WHERE LOWER(rm.path) = ?`

	Database.Raw(sql, userID, strings.ToLower(path)).Scan(&permission)

	if kit.CheckPermissionByAction(permission, action) {
		return nil
	}

	return errors.New("Unauthorized")
}
