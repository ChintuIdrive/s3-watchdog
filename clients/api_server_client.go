package clients

import (
	"ChintuIdrive/s3-watchdog/conf"
	"ChintuIdrive/s3-watchdog/dto"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	get_datacenters      = "api/admin/view_storage/get_datacenters"
	get_credential       = "api/admin/update_user/storage/debug_credentials"
	server_details       = "api/admin/view_storage/server_details"
	login                = "api/admin/login"
	renewToken           = "api/session/admin/renew"
	tokenRefreshInterval = 15 * time.Minute
	credentialTTL        = 15 * time.Minute
)

type APIserverClient struct {
	apiserverConfig *dto.ApiServerConfig
	login           dto.Login
	token           dto.Token
	tokenMutex      sync.RWMutex
	credCache       map[string]cachedCred
	credMutex       sync.RWMutex
}

// New type to store credential with its timestamp
type cachedCred struct {
	cred      dto.Cred
	timestamp time.Time
}

func NewApiServerClient(config *conf.Config) *APIserverClient {
	client := &APIserverClient{
		apiserverConfig: config.ApiServerConfig,
		login:           config.Login,
		tokenMutex:      sync.RWMutex{},
		credCache:       make(map[string]cachedCred),
		credMutex:       sync.RWMutex{},
	}

	// Initial login to get the first token
	err := client.Login()
	if err != nil {
		log.Printf("Initial login failed: %v", err)
	}

	// Start credential cleanup goroutine
	go client.startCredentialCleanup()

	// Start token refresh goroutine
	go client.startTokenRefresh()

	return client
}

func (asc *APIserverClient) startCredentialCleanup() {
	ticker := time.NewTicker(credentialTTL)
	for range ticker.C {
		asc.cleanupExpiredCredentials()
	}
}

func (asc *APIserverClient) cleanupExpiredCredentials() {
	asc.credMutex.Lock()
	defer asc.credMutex.Unlock()

	now := time.Now()
	for dns, cached := range asc.credCache {
		if now.Sub(cached.timestamp) >= credentialTTL {
			delete(asc.credCache, dns)
		}
	}
}

func (asc *APIserverClient) startTokenRefresh() {
	ticker := time.NewTicker(tokenRefreshInterval)
	for range ticker.C {
		err := asc.RenewToken()
		if err != nil {
			log.Printf("Token refresh failed: %v", err)
			// If token renewal fails, try logging in again
			err = asc.Login()
			if err != nil {
				log.Printf("Re-login failed: %v", err)
			}
		}
	}
}

func (asc *APIserverClient) Notify(payload []byte) {
	log.Println("Notifying to API server")
	log.Printf("Notification Payload: %s", string(payload))
}

func (asc *APIserverClient) getToken() dto.Token {
	asc.tokenMutex.RLock()
	defer asc.tokenMutex.RUnlock()
	return asc.token
}

func (asc *APIserverClient) setToken(token dto.Token) {
	asc.tokenMutex.Lock()
	defer asc.tokenMutex.Unlock()
	asc.token = token
}

func (asc *APIserverClient) getSessionToken() string {
	asc.tokenMutex.RLock()
	defer asc.tokenMutex.RUnlock()
	return asc.token.ST
}

func (asc *APIserverClient) setSessionToken(st string) {
	asc.tokenMutex.Lock()
	defer asc.tokenMutex.Unlock()
	asc.token.ST = st
}

func (asc *APIserverClient) RenewToken() error {
	token := asc.getToken()
	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, renewToken)
	method := "POST"
	renewreq := dto.RenewReq{
		RT: token.RT,
	}
	payload, err := json.Marshal(renewreq)
	if err != nil {
		log.Println(err)
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var renewResp dto.RenewResponse
	err = json.Unmarshal(body, &renewResp)
	if err != nil {
		log.Println(err)
		return err
	}

	asc.setSessionToken(renewResp.ST)
	return nil
}

func (asc *APIserverClient) Login() error {
	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, login)
	method := "POST"
	payload, err := json.Marshal(asc.login)
	if err != nil {
		log.Println(err)
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var loginResp dto.LoginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		log.Println(err)
		return err
	}

	asc.setToken(dto.Token{
		ST:   loginResp.ST,
		RT:   loginResp.RT,
		User: loginResp.User,
	})
	return nil
}

func (asc *APIserverClient) GetTenatsListFromApiServer() ([]dto.Tenant, error) {

	var tenatList []dto.Tenant

	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, asc.apiserverConfig.TenantListApi)
	method := "POST"

	payload := []byte(fmt.Sprintf(`{"NodeId":"%s"}`, asc.apiserverConfig.NodeId))

	res, err := asc.FireRequest(method, url, payload)
	if err != nil {
		return tenatList, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return tenatList, err
	}
	var nodeInfo dto.TenantList

	err = json.Unmarshal(body, &nodeInfo)
	if err != nil {
		return tenatList, err
	} else {
		tenatList = nodeInfo.TenantList
	}
	return tenatList, err

}

func (asc *APIserverClient) GetCredential(user dto.User) (dto.Cred, error) {
	// Try to get credentials from cache first
	asc.credMutex.RLock()
	if cached, exists := asc.credCache[user.StorageDNS]; exists {
		// Check if credentials are still valid
		if time.Since(cached.timestamp) < credentialTTL {
			asc.credMutex.RUnlock()
			return cached.cred, nil
		}
	}
	asc.credMutex.RUnlock()

	// If not in cache or expired, fetch from API
	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, get_credential)
	method := "POST"
	credReq := dto.CredReq{
		UserID: user.UserID,
		RDNS:   user.StorageDNS,
	}
	payload, err := json.Marshal(credReq)
	if err != nil {
		return dto.Cred{}, err
	}
	res, err := asc.FireRequest(method, url, payload)
	if err != nil {
		return dto.Cred{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return dto.Cred{}, err
	}
	var credResp dto.CredentialResponse

	err = json.Unmarshal(body, &credResp)
	if err != nil {
		return dto.Cred{}, err
	}

	// Store credentials in cache with current timestamp
	asc.credMutex.Lock()
	asc.credCache[user.StorageDNS] = cachedCred{
		cred:      credResp.Data,
		timestamp: time.Now(),
	}
	asc.credMutex.Unlock()

	return credResp.Data, nil
}

func (asc *APIserverClient) GetNodeDetails(node string) ([]dto.User, error) {
	var users []dto.User
	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, server_details)
	method := "POST"
	payload := []byte(fmt.Sprintf(`{"sn_id":"%s"}`, node))

	res, err := asc.FireRequest(method, url, payload)
	if err != nil {
		return users, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return users, err
	}
	var userListResp dto.UserListResponse

	err = json.Unmarshal(body, &userListResp)
	if err != nil {
		return users, err
	} else {
		users = userListResp.NodeDetails
	}
	return users, nil
}
func (asc *APIserverClient) GetRegions() ([]dto.Region, error) {
	var regions []dto.Region
	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, get_datacenters)
	method := "POST"
	res, err := asc.FireRequest(method, url, []byte{})
	if err != nil {
		return regions, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return regions, err
	}
	var regionListResp dto.RegionListResponse

	err = json.Unmarshal(body, &regionListResp)
	if err != nil {
		return regions, err
	} else {
		regions = regionListResp.Regions
	}
	return regions, nil
}
func (asc *APIserverClient) FireRequest(method, url string, payload []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", asc.getSessionToken())

	res, err := client.Do(req)
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		body, _ := io.ReadAll(res.Body)
		log.Printf("HTTP %d: %s - Response: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))

		err = asc.RenewToken()
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", asc.getSessionToken())
		res, err = client.Do(req)
	}

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return res, nil
}
