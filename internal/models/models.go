package models

type Site struct {
	ID          int    `json:"id"`
	Domain      string `json:"domain"`
	Type        string `json:"type"`
	Target      string `json:"target"`
	SSLEnabled  bool   `json:"ssl_enabled"`
	Environment string `json:"environment"`
	PHPVersion  string `json:"php_version"`
}

type Project struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	ProjectType      string `json:"project_type"`
	RootDir          string `json:"root_dir"`
	ExecPath         string `json:"exec_path"`
	Port             int    `json:"port"`
	StartCommand     string `json:"start_command"`
	AutoStart        bool   `json:"auto_start"`
	Status           string `json:"status"`
	Domains          string `json:"domains"`
	SSLEnabled       bool   `json:"ssl_enabled"`
	SSLEmail         string `json:"ssl_email"`
	ReverseProxyPath string `json:"reverse_proxy_path"`
	ExtraHeaders     string `json:"extra_headers"`
	Description      string `json:"description"`
}

type Task struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Command   string `json:"command"`
	Schedule  string `json:"schedule"`
	IsLoop    bool   `json:"is_loop"`
	Status    string `json:"status"`
	LastRun   string `json:"last_run"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}
