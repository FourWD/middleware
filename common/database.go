package common

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Database *gorm.DB

func ConnectDatabase(dsn string) error {
	var err error

	Database, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if err != nil {
		panic(err)
	}

	return nil
}

/*func Create(c *fiber.Ctx, obj interface{}) error {
	//objType := reflect.TypeOf(obj).String()
	//println(objType)

	err := c.BodyParser(obj)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}

	aa := new(orm.BaseID)
	aa.ID= ""
	err = Database.Create(&obj).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create note", "data": err})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Created user", "data": obj})
}*/

/*func Update(c *fiber.Ctx) error {
	user := new(orm.User)
	err := c.BodyParser(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}

	user.ID = c.Params("id")

								// Find the note with the given Id
								common.Database.Find(&user, "id = ?", id)

								// If no such note present return an error
								if user.ID == "" {
									return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No note present", "data": nil})
								}
	// Save the Changes
	common.Database.Save(&user)
	return c.JSON(fiber.Map{"status": "success", "message": "success", "data": user})
}

func FindByID(c *fiber.Ctx) error {
	id := c.Params("id")

	user := new(orm.User)
	common.Database.Find(&user, "id = ?", id)
	if user.ID == "" {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No data", "data": nil})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Notes Found", "data": user})
}

func FindAll(c *fiber.Ctx) error {
	var users []orm.User

	common.Database.Find(&users)
	if len(users) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No data", "data": nil})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Notes Found", "data": users})
}

func Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	user := new(orm.User)
	common.Database.Find(&user, "id = ?", id)
	if user.ID == "" {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No note present", "data": nil})
	}

	err := common.Database.Delete(&user, "id = ?", id).Error
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Failed to delete note", "data": nil})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Deleted Note"})
}*/
