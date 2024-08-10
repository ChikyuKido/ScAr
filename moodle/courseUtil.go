package moodle

// GetAllModules
// Returns all the modules from a course.
func GetAllModules(course *Course) []CourseModule {
	var modules []CourseModule
	for _, section := range course.Sections {
		for _, module := range section.Modules {
			modules = append(modules, module)
		}
	}
	return modules
}
