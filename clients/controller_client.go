 package clients

// import (
// 	"ChintuIdrive/s3-watchdog/cryption"
// 	"ChintuIdrive/s3-watchdog/dto"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

// type ControllerClient struct {
// 	controllerConfig *dto.ControllerConfig
// }

// func NewControllerClientt(controllerConfig *dto.ControllerConfig) *ControllerClient {
// 	return &ControllerClient{
// 		controllerConfig: controllerConfig,
// 	}
// }
// func (cc *ControllerClient) GetTenantListFromController() ([]dto.TenatWithProcessInfo, error) {
// 	var tenants []dto.TenatWithProcessInfo
// 	minioProcessPath := "/opt/e2-node-controller-1/running_processes"
// 	dirEntries, err := os.ReadDir(minioProcessPath)
// 	if err != nil {
// 		fmt.Println(err)
// 		return tenants, err
// 	}

// 	for _, entry := range dirEntries {
// 		if !entry.IsDir() {
// 			if strings.HasSuffix(entry.Name(), ".info") {
// 				minioProcessFile := filepath.Join(minioProcessPath, entry.Name())
// 				file, err := os.Open(minioProcessFile)
// 				if err != nil {
// 					fmt.Println(err)
// 					continue
// 				}
// 				defer file.Close()

// 				var serverInfo dto.TenatWithProcessInfo
// 				decoder := json.NewDecoder(file)
// 				err = decoder.Decode(&serverInfo)
// 				if err != nil {
// 					fmt.Println(err)
// 					continue
// 				}

// 				tenants = append(tenants, serverInfo)
// 				fmt.Println(minioProcessFile)
// 			}
// 		}

// 	}

// 	return tenants, err

// }

// func (cc *ControllerClient) GetAccessKeys(tenat dto.Tenant) (*cryption.SecretData, error) {

// 	url := fmt.Sprintf("https://%s/%s", cc.controllerConfig.ControllerDNS, cc.controllerConfig.AddServiceAccountApi)
// 	method := "POST"
// 	// url := "https://localhost:44344/admin/v1/add_service_account"
// 	// method := "POST"
// 	addSrvAcctReq := dto.ServiceAccountReq{
// 		Name: "wtachdog",
// 		BaseReq: dto.BaseReq{
// 			DNS: tenat.DNS,
// 			SID: tenat.UserID,
// 			SK: dto.SecretKey{
// 				CString: tenat.Password.CString,
// 			},
// 		},
// 		IsInternal:    false,
// 		IsTestAccount: false,
// 		Permissions:   2,
// 	}

// 	payload, err := json.Marshal(addSrvAcctReq)
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, err := FireRequest(method, url, payload)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer res.Body.Close()

// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	var accesskey *cryption.SecretData

// 	err = json.Unmarshal(body, &accesskey)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	} else {
// 		return accesskey, nil
// 	}

// }

// func (cc *ControllerClient) GetTenantWithProcessInfo(tenat dto.Tenant) (*dto.TenatWithProcessInfo, error) {

// 	url := fmt.Sprintf("https://%s/%s", cc.controllerConfig.ControllerDNS, cc.controllerConfig.GetTenantInfoApi)
// 	method := "POST"

// 	clientreq := dto.TenantProcessInfoReq{
// 		BaseReq: dto.BaseReq{
// 			DNS: tenat.DNS,
// 			SID: tenat.UserID,
// 			SK:  dto.SecretKey{CString: tenat.Password.CString},
// 		},
// 	}
// 	payload, err := json.Marshal(clientreq)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}

// 	res, err := FireRequest(method, url, payload)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	defer res.Body.Close()

// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	var resp dto.TenantProcessInfoResponse
// 	err = json.Unmarshal(body, &resp)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &resp.TenatWithProcessInfo, err
// }

// func (cc *ControllerClient) GetSavedAccessKey(tenant dto.Tenant) (*cryption.SecretData, error) {
// 	var accessKeys *cryption.SecretData
// 	s3credentialsPath := filepath.Join(cc.controllerConfig.AccessKeyDir, tenant.DNS, "s3-credentials.json")
// 	data, err := os.ReadFile(s3credentialsPath)
// 	if err != nil {
// 		// If the file does not exist, create a default S3Config
// 		log.Printf("S3 configuration not available for tenant %s, adding default configuration", tenant.DNS)
// 		return cc.LoadS3Credentials(tenant)
// 	}
// 	err = json.Unmarshal(data, &accessKeys)
// 	if err != nil {
// 		fmt.Println(err)
// 		return accessKeys, err
// 	}
// 	return accessKeys, nil
// }
// func (cc *ControllerClient) LadAccessKeys(tenantsFromApiServer []dto.Tenant) {

// 	for _, tenant := range tenantsFromApiServer {
// 		cc.LoadS3Credentials(tenant)
// 	}

// }

// func (cc *ControllerClient) LoadS3Credentials(tenant dto.Tenant) (*cryption.SecretData, error) {
// 	accKey, err := cc.GetAccessKeys(tenant)
// 	if err != nil {
// 		log.Printf("Failed to get access keys for tenant %s: %v", tenant.DNS, err)
// 		return nil, err
// 	}
// 	ds, err := accKey.SecretKey.GetDString()
// 	if err != nil {
// 		log.Printf("Failed to get DString for tenant %s: %v", tenant.DNS, err)
// 		return nil, err
// 	}
// 	accKey.SecretKey.DString = ds
// 	// if err != nil {
// 	// 	log.Printf("Failed to set secret key for tenant %s: %v", tenant.DNS, err)
// 	// }
// 	accKeyDir := filepath.Join(cc.controllerConfig.AccessKeyDir, tenant.DNS)
// 	if _, err := os.Stat(accKeyDir); os.IsNotExist(err) {
// 		err := os.MkdirAll(cc.controllerConfig.AccessKeyDir, os.ModePerm)
// 		if err != nil {
// 			log.Printf("Failed to create access key directory: %v", err)
// 			return nil, err
// 		}
// 	}
// 	accKeyFilePath := filepath.Join(accKeyDir, "s3-credentials.json")
// 	accessKeyFile, err := os.Create(accKeyFilePath)
// 	if err != nil {
// 		log.Printf("Failed to create access key file for tenant %s: %v", tenant.DNS, err)
// 		return nil, err
// 	}
// 	defer accessKeyFile.Close()

// 	accessKeyData, err := json.MarshalIndent(accKey, "", "  ")
// 	if err != nil {
// 		log.Printf("Failed to marshal access key data for tenant %s: %v", tenant.DNS, err)
// 		return nil, err
// 	}

// 	accessKeyFile.Write(accessKeyData)
// 	return accKey, nil
// }
