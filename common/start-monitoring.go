package common

func StartMonitoring(name string, logic interface{}) {
	runLatestVersionOnly()
	registerPrometheus(name, logic)
	monitorDatabaseConnectionPool()
}
