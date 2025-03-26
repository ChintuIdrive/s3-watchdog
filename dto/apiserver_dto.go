package dto

type TenantList struct {
	RegionName               string        `json:"region_name"`
	RegionID                 int           `json:"region_id"`
	TenantList               []Tenant      `json:"TenantList"`
	Pool                     []interface{} `json:"pool"`
	SelfConfig               interface{}   `json:"self_config"`
	AvgLoad                  int           `json:"AvgLoad"`
	HealingAvgLoadLimit      int           `json:"healing_avg_load_limit"`
	HealingThreadPerTenant   int           `json:"healing_thread_per_tenant"`
	HealingConcurrentTenants int           `json:"healing_concurrent_tenants"`
}

type Tenant struct {
	DNS                       string        `json:"dns"`
	UserID                    string        `json:"userId"`
	Password                  Password      `json:"Password"`
	E2UserID                  string        `json:"e2_userid"`
	UserType                  int           `json:"user_type"`
	PublicBucketsEnabled      bool          `json:"public_buckets_enabled"`
	EnablePublicAccessOnE2URL bool          `json:"enable_public_access_on_e2_url"`
	UserStorageNodeID         int           `json:"user_storage_node_id"`
	CnameList                 []interface{} `json:"cname_list"`
	AllowedOrigin             string        `json:"AllowedOrigin"`
	Compression               bool          `json:"compression"`
	MaxAPIRequests            int           `json:"MaxApiRequests"`
	APIRequestsDeadline       int           `json:"ApiRequestsDeadline"`
	Whitelist                 []string      `json:"whitelist"`
	Blacklist                 interface{}   `json:"blacklist"`
	UseDEC                    bool          `json:"UseDEC"`
	Restart                   bool          `json:"Restart"`
	ForceRestart              bool          `json:"ForceRestart"`
	DownloadLimit             interface{}   `json:"DownloadLimit"`
	UploadLimit               interface{}   `json:"UploadLimit"`
}

type Password struct {
	CString string `json:"CString"`
}

type CredReq struct {
	UserID string `json:"user_id"`
	RDNS   string `json:"rdns"`
}
type Cred struct {
	AccessKey string `json:"user"`
	SecretKey string `json:"pass"`
}

type CredentialResponse struct {
	RespCode int    `json:"resp_code"`
	RespMsg  string `json:"resp_msg"`
	Data     Cred   `json:"data"`
}

type User struct {
	Email          string `json:"email"`
	AllowMigration bool   `json:"allow_migration"`
	UserID         string `json:"user_id"`
	LastActivityTS string `json:"last_activity_ts"`
	StorageUsed    int    `json:"storage_used"`
	FileCount      int    `json:"file_count"`
	PlanName       string `json:"plan_name"`
	StorageDNS     string `json:"storage_dns"`
	IndexingStatus int    `json:"indexing_status"`
}

type UserListResponse struct {
	RespCode    int     `json:"resp_code"`
	RespMsg     string  `json:"resp_msg"`
	NodeDetails []*User `json:"node_details"`
}

type Region struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Address        string `json:"address"`
	Zipcode        string `json:"zipcode"`
	RegionLocation string `json:"region_location"`
	City           string `json:"city"`
	Country        string `json:"country"`
	ContactPersons string `json:"contact_persons"`
	GatewayIPs     string `json:"gateway_ips"`
	StorageNodes   string `json:"storage_nodes"`
	Region         string `json:"region"`
}

type RegionListResponse struct {
	RespCode int      `json:"resp_code"`
	RespMsg  string   `json:"resp_msg"`
	Regions  []Region `json:"data"`
}

type Login struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Recaptcha string `json:"recaptcha"`
}

type Token struct {
	ST   string `json:"st"`
	RT   string `json:"rt"`
	User User   `json:"user"`
}

type LoginResponse struct {
	RespCode int    `json:"resp_code"`
	RespMsg  string `json:"resp_msg"`
	ST       string `json:"st"`
	RT       string `json:"rt"`
	User     User   `json:"user"`
}

type RenewReq struct {
	RT string `json:"rt"`
}

type RenewResponse struct {
	RespCode int    `json:"resp_code"`
	RespMsg  string `json:"resp_msg"`
	ST       string `json:"st"`
	RT       string `json:"rt"`
}
