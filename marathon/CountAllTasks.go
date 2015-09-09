package marathon

////////////////////////////////////////////////////////////////////////////////////////
// CountAllTasks
//
// This does nothing for Marathon because we just don't
// use this data. We do use it for ECS though so this function is provided for interface consistency.
//
func CountAllTasks() (int, int, error) {
	return 0, 0, nil
}
