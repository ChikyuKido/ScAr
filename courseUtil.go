package main

func getAllResourcesModulesByModName(course *Course, modName string) []CourseModule {
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
func getAllResourceModules(course *Course) []CourseModule {
	return getAllResourcesModulesByModName(course, "resource")
}
func getAllURLModules(course *Course) []CourseModule {
	return getAllResourcesModulesByModName(course, "url")
}
func getAllAssignModules(course *Course) []CourseModule {
	return getAllResourcesModulesByModName(course, "assign")
}
func getAllLabelModules(course *Course) []CourseModule {
	return getAllResourcesModulesByModName(course, "label")
}
