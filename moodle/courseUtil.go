package moodle

func GetAllResourcesModulesByModName(course *Course, modName string) []CourseModule {
	var modules []CourseModule
	for _, section := range course.Sections {
		for _, module := range section.Modules {
			if module.ModName == modName {
				modules = append(modules, module)
			}
		}
	}
	return modules
}
func GetAllResourceModules(course *Course) []CourseModule {
	return GetAllResourcesModulesByModName(course, "resource")
}
func GetAllURLModules(course *Course) []CourseModule {
	return GetAllResourcesModulesByModName(course, "url")
}
func GetAllAssignModules(course *Course) []CourseModule {
	return GetAllResourcesModulesByModName(course, "assign")
}
func GetAllLabelModules(course *Course) []CourseModule {
	return GetAllResourcesModulesByModName(course, "label")
}
