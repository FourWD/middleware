package common

import (
	"errors"
	"strings"

	"github.com/FourWD/middleware/orm"
	"github.com/gofiber/fiber/v2"
)

func CheckRolePermission(c *fiber.Ctx, path string, action string) error {
	userID := GetSessionUserID(c)

	sqlEmployee := `SELECT role_id FROM employees WHERE id = ?`
	employee := new(orm.Employee)
	Database.Raw(sqlEmployee, userID).Scan(&employee)

	isStrExists := StringExistsInList(action, []string{"READ", "CREATE", "UPDATE", "DELETE"})

	if isStrExists {
		type Permission struct {
			Read   int `json:"read"`
			Create int `json:"create"`
			Update int `json:"update"`
			Delete int `json:"delete"`
		}

		var permission Permission

		sql := `SELECT
		COALESCE(rp.is_read, 0) "read",
		COALESCE(rp.is_create, 0) "create",
		COALESCE(rp.is_update, 0) "update",
		COALESCE(rp.is_delete, 0) "delete"
		FROM role_menus AS rm
		LEFT JOIN role_permissions rp ON rp.role_menu_id = rm.id AND rp.role_id = ?
		WHERE LOWER(rm.path) = ?`

		Database.Raw(sql, employee.RoleID, strings.ToLower(path)).Scan(&permission)

		if action == "READ" && permission.Read == 0 {
			return nil
		} else if action == "CREATE" && permission.Create == 0 {
			return nil
		} else if action == "UPDATE" && permission.Update == 0 {
			return nil
		} else if action == "DELETE" && permission.Delete == 0 {
			return nil
		}
	}
	return errors.New("Unauthorized")
}
