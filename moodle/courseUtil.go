package moodle

func GetAllResourcesModulesByModName(course *Course, modName string) []CourseModule {
	var modules []CourseModule
	for _, module := range GetAllModules(course) {
		if module.ModName == modName {
			modules = append(modules, module)
		}
	}

	return modules
}
func GetAllModules(course *Course) []CourseModule {
	var modules []CourseModule
	for _, section := range course.Sections {
		for _, module := range section.Modules {
			modules = append(modules, module)

		}
	}
	return modules
}
