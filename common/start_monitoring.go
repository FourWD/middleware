package common

func StartMonitoring() {
	runLatestVersionOnly()
	monitorDatabaseConnectionPool()
}
