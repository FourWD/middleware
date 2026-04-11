package common

import (
	"errors"
	"strings"

	"github.com/FourWD/middleware/orm"
	"github.com/gofiber/fiber/v2"
)

func CheckRolePermissionByUser(c *fiber.Ctx, role string, action string) error {
	userID := GetSessionUserID(c)

	sqlEmployee := `SELECT * FROM employees WHERE id = ?`
	employee := new(orm.Employee)
	if err := Database.Raw(sqlEmployee, userID).Scan(&employee).Error; err != nil {
		LogError("CHECK_ROLE_EMPLOYEE_ERROR", map[string]interface{}{"error": err.Error(), "userID": userID}, "")
		return errors.New("Unauthorized")
	}

	var roleName string
	sqlRole := `SELECT name FROM roles WHERE id = ?`
	if err := Database.Raw(sqlRole, employee.RoleID).Scan(&roleName).Error; err != nil {
		LogError("CHECK_ROLE_NAME_ERROR", map[string]interface{}{"error": err.Error(), "roleID": employee.RoleID}, "")
		return errors.New("Unauthorized")
	}

	if !StringExistsInList(action, []string{"READ", "CREATE", "UPDATE", "DELETE"}) {
		return errors.New("Unauthorized")
	}

	if !strings.EqualFold(role, roleName) {
		return errors.New("Unauthorized")
	}

	var permission Permission

	sql := `SELECT
	COALESCE(rp.is_read, 0) "read",
	COALESCE(rp.is_create, 0) "create",
	COALESCE(rp.is_update, 0) "update",
	COALESCE(rp.is_delete, 0) "delete"
	FROM role_menus AS rm
	LEFT JOIN role_permissions rp ON rp.role_menu_id = rm.id
	WHERE rp.role_id = ?`

	Database.Raw(sql, employee.RoleID).Scan(&permission)

	if checkPermissionByAction(permission, action) {
		return nil
	}

	return errors.New("Unauthorized")
}
