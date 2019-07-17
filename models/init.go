package models

func init() {
	initDB()

	initSystem()

	initParam()

	qs_task = o.QueryTable("pms_task").OrderBy("id")

	initProcess()
}
