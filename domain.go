package discovery

type Permission struct {
	Label string `json:"label"`
	Code  string `json:"code"`
}
type DiscoveryModule struct {
	Label       string         `json:"label"`
	Code        string         `json:"code"`
	Permissions PermissionList `json:"permissions"`
}
type PermissionList []Permission
type DiscoveryModulesList []DiscoveryModule