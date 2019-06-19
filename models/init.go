package models

func init() {
	initDB()

	initSystem()

	qs_task = o.QueryTable("pms_task").OrderBy("id")

	initProcess()
}
