package modules



// ModuleInfo to DTO dla modułów dostępnych danemu użytkownikowi
type ModuleInfo struct {
    ID          int    `json:"id"`
    ModuleName  string `json:"module_name"`
    Description string `json:"description"`
    Role        string `json:"role"`
    Icon 	string `json:"icon"`
}




type SystemModules struct {
    ID          int    `json:"id"`
    ModuleName  string `json:"module_name"`
    Description string `json:"description"`
	Icon 	string `json:"icon"`

}


