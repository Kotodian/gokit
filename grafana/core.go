package grafana

import (
	"errors"
	"fmt"
	"os"

	"github.com/parnurzeal/gorequest"
)

type KindOrganisationRole string

var (
	KindOrganisationRoleView   KindOrganisationRole = "Viewer"
	KindOrganisationRoleAdmin  KindOrganisationRole = "Admin"
	KindOrganisationRoleEditor KindOrganisationRole = "Editor"
)

type User struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Password string
	Cookie   string
}

func GetGrafanaHost(id string, t string) string {
	return fmt.Sprintf("%s.%s.grafana%s", id, t, os.Getenv("GRAFANA_DOMIN_SUFFIX"))
}

func NewSuperBaseAuthRequest() *gorequest.SuperAgent {
	return gorequest.New().SetBasicAuth("admin", "goiot2018")
}

func NewBaseAuthRequest(name, password string) *gorequest.SuperAgent {
	return gorequest.New().SetBasicAuth(name, password)
}

func GetUserByName(name string) (*User, error) {
	var u User
	if err := Request(NewSuperBaseAuthRequest().Get("http://grafana:3000/api/users/lookup?loginOrEmail="+name), nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

type Organisation struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func GetOrganisationByName(name string) (*Organisation, error) {
	var u Organisation
	if err := Request(NewSuperBaseAuthRequest().Get("http://grafana:3000/api/orgs/name/"+name), nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateOrganisation(name string) (*Organisation, error) {
	// ret := map[string]string{}
	var ret map[string]interface{}
	if err := Request(NewSuperBaseAuthRequest().Post("http://grafana:3000/api/orgs"), map[string]interface{}{
		"name": name,
	}, &ret); err != nil {
		return nil, err
	}
	// id, _ := strconv.Atoi(ret["orgId"])
	return &Organisation{
		ID:   uint64(ret["orgId"].(float64)),
		Name: name,
	}, nil
}

func CreateOrganisationIfNotExists(name string) (*Organisation, error) {
	if org, err := GetOrganisationByName(name); err == nil {
		return org, nil
	} else if org, err := CreateOrganisation(name); err != nil {
		return nil, err
	} else {
		return org, nil
	}
}

type Team struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	OrganisationID uint64 `json:"orgId"`
}

func CreateTeamIfNotExists(name string) (*Team, error) {
	if org, err := GetTeamByName(name); err == nil {
		return org, nil
	} else if org, err := CreateTeam(name); err != nil {
		return nil, err
	} else {
		return org, nil
	}
}

func CreateTeam(name string) (*Team, error) {
	var ret map[string]interface{}
	if err := Request(NewSuperBaseAuthRequest().Post("http://grafana:3000/api/teams"), map[string]interface{}{
		"name": name,
	}, &ret); err != nil {
		return nil, err
	}
	// id, _ := strconv.Atoi(ret["orgId"])
	return &Team{
		ID:   uint64(ret["teamId"].(float64)),
		Name: name,
	}, nil
}

func GetTeamByName(name string) (*Team, error) {
	// var ret map[string]interface{}
	var ret struct {
		Teams []Team `json:"teams"`
	}

	if err := Request(NewSuperBaseAuthRequest().Get("http://grafana:3000/api/teams/search?name="+name), nil, &ret); err != nil {
		return nil, err
	} else if len(ret.Teams) == 0 {
		return nil, errors.New("not found")
	}
	return &ret.Teams[0], nil
}

func CreateUser(name string, password string) (*User, error) {
	var ret map[string]interface{}
	if err := Request(NewSuperBaseAuthRequest().Post("http://grafana:3000/api/admin/users"), map[string]interface{}{
		"name":     name,
		"password": password,
		"login":    name,
		"theme":    "dark",
	}, &ret); err != nil {
		return nil, err
	}
	// id, _ := strconv.Atoi(ret["orgId"])
	return &User{
		ID:   uint64(ret["id"].(float64)),
		Name: name,
	}, nil
}

func DeleteUserInOrganisation(userID uint64, orgID uint64) error {
	return Request(NewSuperBaseAuthRequest().Delete(fmt.Sprintf("http://grafana:3000/api/orgs/%v/users/%v", orgID, userID)), nil, nil)
}

//绑定用户到组织
func CreateUserIfNotExists(name, password string) (user *User, err error) {
	//先创建用户
	if user, err = CreateUser(name, password); err != nil {
		// return err
		if user, err = GetUserByName(name); err != nil {
			//创建用户
			return
		}
	}
	return
}
