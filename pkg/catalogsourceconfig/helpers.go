package catalogsourceconfig

import "strings"

func RemoveNamespaces(packages string) string {
	packageList := strings.Split(packages, ",")
	for i := range packageList {
		if strings.Contains(packageList[i], "/") {
			packageList[i] = strings.Split(packageList[i], "/")[1]
		}
	}
	packages = strings.Join(packageList, ",")

	return packages
}
