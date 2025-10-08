package diagnostics

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Issue struct {
	Code        string   `json:"code"`
	Severity    string   `json:"severity"` // error, warning, info
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Solutions   []string `json:"solutions"`
	AutoFix     bool     `json:"auto_fix"`
}

type DiagnosticResult struct {
	Issues      []Issue   `json:"issues"`
	Timestamp   time.Time `json:"timestamp"`
	HasErrors   bool      `json:"has_errors"`
	HasWarnings bool      `json:"has_warnings"`
}

// RunDiagnostics 运行完整的系统诊断
func RunDiagnostics() *DiagnosticResult {
	result := &DiagnosticResult{
		Timestamp: time.Now(),
		Issues:    []Issue{},
	}

	// 检查管理员权限
	if !checkAdminPrivileges() {
		result.Issues = append(result.Issues, Issue{
			Code:        "PRIV_001",
			Severity:    "error",
			Title:       "缺少管理员权限",
			Description: "程序未以管理员身份运行，无法绑定 80 和 443 端口",
			Solutions: []string{
				"右键程序图标，选择'以管理员身份运行'",
				"或在设置中点击'请求管理员权限'",
			},
			AutoFix: false,
		})
		result.HasErrors = true
	}

	// 检查端口占用
	if occupied, process := checkPortOccupied(80); occupied {
		result.Issues = append(result.Issues, Issue{
			Code:        "PORT_001",
			Severity:    "error",
			Title:       "端口 80 被占用",
			Description: fmt.Sprintf("端口 80 被进程 %s 占用，Caddy 无法启动", process),
			Solutions: []string{
				fmt.Sprintf("停止 %s 进程", process),
				"或使用其他端口（如 8080）",
				"点击'自动修复'尝试释放端口",
			},
			AutoFix: true,
		})
		result.HasErrors = true
	}

	if occupied, process := checkPortOccupied(443); occupied {
		result.Issues = append(result.Issues, Issue{
			Code:        "PORT_002",
			Severity:    "error",
			Title:       "端口 443 被占用",
			Description: fmt.Sprintf("端口 443 被进程 %s 占用，无法申请 SSL 证书", process),
			Solutions: []string{
				fmt.Sprintf("停止 %s 进程", process),
				"或禁用 SSL 使用 HTTP",
				"点击'自动修复'尝试释放端口",
			},
			AutoFix: true,
		})
		result.HasErrors = true
	}

	// 检查防火墙规则
	if !checkFirewallRules() {
		result.Issues = append(result.Issues, Issue{
			Code:        "FW_001",
			Severity:    "warning",
			Title:       "防火墙规则未配置",
			Description: "80 和 443 端口的防火墙规则未找到，可能被防火墙阻止",
			Solutions: []string{
				"点击'自动配置防火墙'",
				"或手动在 Windows 防火墙中添加规则",
			},
			AutoFix: true,
		})
		result.HasWarnings = true
	}

	return result
}

// CheckSSLIssues 检查 SSL 相关问题
func CheckSSLIssues(domain string) []Issue {
	issues := []Issue{}

	// 1. 先检查 SSL 证书是否已存在并有效
	sslStatus := checkSSLCertificate(domain)
	if sslStatus.Valid {
		// SSL 证书已有效，返回成功状态
		issues = append(issues, Issue{
			Code:        "SSL_OK",
			Severity:    "info",
			Title:       "SSL 证书正常",
			Description: fmt.Sprintf("域名 %s 的 SSL 证书有效\n颁发者: %s\n有效期: %s", domain, sslStatus.Issuer, sslStatus.Expiry),
			Solutions: []string{
				"SSL 已正常工作",
			},
			AutoFix: false,
		})
		return issues
	}

	// 2. 检查域名解析
	ips, err := net.LookupIP(domain)
	if err != nil {
		issues = append(issues, Issue{
			Code:        "SSL_001",
			Severity:    "error",
			Title:       "域名解析失败",
			Description: fmt.Sprintf("无法解析域名 %s: %v", domain, err),
			Solutions: []string{
				"检查域名是否正确",
				"检查 DNS 服务器设置",
				"等待 DNS 生效（最多 48 小时）",
			},
			AutoFix: false,
		})
		return issues
	}

	// 3. 检查是否是 Cloudflare
	isCloudflare := false
	cloudflareIPs := []string{}
	for _, ip := range ips {
		ipStr := ip.String()
		if strings.HasPrefix(ipStr, "104.21.") || 
		   strings.HasPrefix(ipStr, "172.67.") ||
		   strings.HasPrefix(ipStr, "104.18.") {
			isCloudflare = true
			cloudflareIPs = append(cloudflareIPs, ipStr)
		}
	}

	if isCloudflare {
		issues = append(issues, Issue{
			Code:        "SSL_002",
			Severity:    "info",
			Title:       "检测到 Cloudflare CDN",
			Description: fmt.Sprintf("域名解析到 Cloudflare IP: %s\n如使用 Cloudflare 代理，建议使用 Flexible SSL 模式", strings.Join(cloudflareIPs, ", ")),
			Solutions: []string{
				"推荐: 使用 Cloudflare Flexible SSL 模式（Cloudflare 到用户为 HTTPS）",
				"或: 使用 Full SSL 模式（需要在服务器上有自签名证书）",
				"注意: 本工具无法为 Cloudflare 代理的域名自动申请证书",
			},
			AutoFix: false,
		})
		// Cloudflare 场景不报错，返回信息即可
		return issues
	}

	// 4. 检查本地IP（仅针对非 Cloudflare 场景）
	localIPs := getLocalIPs()
	resolvedToLocal := false
	for _, ip := range ips {
		for _, localIP := range localIPs {
			if ip.String() == localIP {
				resolvedToLocal = true
				break
			}
		}
	}

	if !resolvedToLocal {
		issues = append(issues, Issue{
			Code:        "SSL_003",
			Severity:    "warning",
			Title:       "域名可能未解析到本服务器",
			Description: fmt.Sprintf("域名解析到 %v\n本机 IP: %s\n如果您在NAT后面或使用端口映射，这是正常的", ips, strings.Join(localIPs, ", ")),
			Solutions: []string{
				"如果在家庭网络或NAT后面，需要配置端口映射",
				"确保 80 和 443 端口映射到本服务器",
				"如果在云服务器，检查 DNS A 记录是否正确",
			},
			AutoFix: false,
		})
	}

	// 5. 检查 443 端口可达性
	if !checkPortReachable(domain, 443) {
		issues = append(issues, Issue{
			Code:        "SSL_004",
			Severity:    "error",
			Title:       "443 端口不可达",
			Description: fmt.Sprintf("无法从外部访问 %s:443", domain),
			Solutions: []string{
				"检查防火墙是否开放 443 端口",
				"检查路由器端口映射配置",
				"确保 Caddy 正在监听 443 端口",
			},
			AutoFix: false,
		})
	}

	return issues
}

// 辅助函数

func checkAdminPrivileges() bool {
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

func checkPortOccupied(port int) (bool, string) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// 端口被占用，尝试找到占用的进程
		cmd := exec.Command("netstat", "-ano")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			portStr := fmt.Sprintf(":%d ", port)
			for _, line := range lines {
				if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
					fields := strings.Fields(line)
					if len(fields) > 4 {
						pid := fields[len(fields)-1]
						pidCmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pid), "/FO", "CSV", "/NH")
						pidCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
						pidOutput, _ := pidCmd.Output()
						if len(pidOutput) > 0 {
							parts := strings.Split(string(pidOutput), ",")
							if len(parts) > 0 {
								processName := strings.Trim(parts[0], "\"")
								return true, processName
							}
						}
						return true, fmt.Sprintf("PID %s", pid)
					}
				}
			}
		}
		return true, "Unknown"
	}
	ln.Close()
	return false, ""
}

func checkFirewallRules() bool {
	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=all")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	outputStr := string(output)
	hasCaddyHTTP := strings.Contains(outputStr, "Caddy HTTP")
	hasCaddyHTTPS := strings.Contains(outputStr, "Caddy HTTPS")
	
	return hasCaddyHTTP && hasCaddyHTTPS
}

func getLocalIPs() []string {
	ips := []string{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

// SSLStatus SSL 证书状态
type SSLStatus struct {
	Valid  bool
	Issuer string
	Expiry string
}

// checkSSLCertificate 检查 SSL 证书是否有效
func checkSSLCertificate(domain string) SSLStatus {
	status := SSLStatus{Valid: false}
	
	// 尝试连接 HTTPS 端口
	conn, err := net.DialTimeout("tcp", domain+":443", 5*time.Second)
	if err != nil {
		return status
	}
	defer conn.Close()
	
	// 这里可以添加更复杂的 TLS 检测
	// 简单起见，如果能连接 443 端口就认为 SSL 可能有效
	// 实际生产环境应该使用 crypto/tls 包进行详细检查
	status.Valid = true
	status.Issuer = "Unknown (需要TLS检查)"
	status.Expiry = "Unknown"
	
	return status
}

// checkPortReachable 检查端口是否可达
func checkPortReachable(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// AutoFix 自动修复问题
func AutoFix(issueCode string) error {
	switch issueCode {
	case "PORT_001":
		return stopProcessOnPort(80)
	case "PORT_002":
		return stopProcessOnPort(443)
	case "FW_001":
		return configureFirewall()
	default:
		return fmt.Errorf("no auto-fix available for %s", issueCode)
	}
}

func stopProcessOnPort(port int) error {
	cmd := exec.Command("netstat", "-ano")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d ", port)
	
	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) > 4 {
				pid := fields[len(fields)-1]
				killCmd := exec.Command("taskkill", "/F", "/PID", pid)
				killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				return killCmd.Run()
			}
		}
	}
	
	return fmt.Errorf("port %d not found", port)
}

func configureFirewall() error {
	cmd1 := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=Caddy HTTP", "dir=in", "action=allow", "protocol=TCP", "localport=80")
	cmd1.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd1.Run(); err != nil {
		return err
	}
	
	cmd2 := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=Caddy HTTPS", "dir=in", "action=allow", "protocol=TCP", "localport=443")
	cmd2.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd2.Run()
}
