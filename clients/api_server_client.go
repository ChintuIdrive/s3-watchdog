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
)

const (
	get_datacenters = "api/admin/view_storage/get_datacenters"
	get_credential  = "api/admin/update_user/storage/debug_credentials"
	server_details  = "api/admin/view_storage/server_details"
	login           = "api/admin/login"
	renewToken      = "api/session/admin/renew"
)

type APIserverClient struct {
	apiserverConfig *dto.ApiServerConfig
	login           dto.Login
	token           dto.Token
}

func NewApiServerClient(config *conf.Config) *APIserverClient {
	return &APIserverClient{
		apiserverConfig: config.ApiServerConfig,
		login:           config.Login,
	}
}

func (asc *APIserverClient) Notify(payload []byte) {
	log.Println("Notifying to API server")
	log.Printf("Notification Payload: %s", string(payload))
}
func (asc *APIserverClient) RenewToken() error {
	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, renewToken)
	method := "POST"
	renewreq := dto.RenewReq{
		RT: asc.token.RT,
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
	} else {
		asc.token.ST = renewResp.ST
	}
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
	} else {
		asc.token = dto.Token{
			ST:   loginResp.ST,
			RT:   loginResp.RT,
			User: loginResp.User,
		}
	}
	return nil

}

func (asc *APIserverClient) GetTenatsListFromApiServer() ([]dto.Tenant, error) {

	var tenatList []dto.Tenant

	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, asc.apiserverConfig.TenantListApi)
	method := "POST"

	payload := []byte(fmt.Sprintf(`{"NodeId":"%s"}`, asc.apiserverConfig.NodeId))

	res, err := asc.FireRequest(method, asc.token.ST, url, payload)
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
	cred := dto.Cred{}
	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, get_credential)
	method := "POST"
	credReq := dto.CredReq{
		UserID: user.UserID,
		RDNS:   user.StorageDNS,
	}
	payload, err := json.Marshal(credReq)
	if err != nil {
		return cred, err
	}
	res, err := asc.FireRequest(method, url, asc.token.ST, payload)
	if err != nil {
		return cred, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return cred, err
	}
	var credResp dto.CredentialResponse

	err = json.Unmarshal(body, &credResp)
	if err != nil {
		return cred, err
	} else {
		cred = credResp.Data
	}
	return cred, nil
}

func (asc *APIserverClient) GetNodeDetails(node string) ([]dto.User, error) {
	var users []dto.User
	url := fmt.Sprintf("https://%s/%s", asc.apiserverConfig.APIServerDNS, server_details)
	method := "POST"
	payload := []byte(fmt.Sprintf(`{"sn_id":"%s"}`, node))

	res, err := asc.FireRequest(method, url, asc.token.ST, payload)
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
	res, err := asc.FireRequest(method, url, asc.token.ST, []byte{})
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
func (asc *APIserverClient) FireRequest(method, url, token string, payload []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(string(payload)))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", asc.token.ST)

	res, err := client.Do(req)
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		body, _ := io.ReadAll(res.Body) // Read response body for logging
		log.Printf("HTTP %d: %s - Response: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))
		asc.RenewToken()
		req.Header.Add("Authorization", asc.token.ST)
		res, err = client.Do(req)
	}

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return res, nil
}
