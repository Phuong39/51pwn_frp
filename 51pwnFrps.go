package main

import (
	"math/rand"
	"time"

	"github.com/fatedier/golib/crypto"

	"fmt"
	"os"

	_ "github.com/hktalent/frp/pkg/metrics"

	"github.com/hktalent/frp/pkg/auth"
	"github.com/hktalent/frp/pkg/config"
	"github.com/hktalent/frp/pkg/util/log"
	"github.com/hktalent/frp/pkg/util/util"
	"github.com/hktalent/frp/pkg/util/version"
	"github.com/hktalent/frp/server"

	"github.com/spf13/cobra"
)

const (
	CfgFileTypeIni = iota
	CfgFileTypeCmd
)

var (
	cfgFile     string
	showVersion bool

	bindAddr          string
	bindPort          int
	bindUDPPort       int
	kcpBindPort       int
	proxyBindAddr     string
	vhostHTTPPort     int
	vhostHTTPSPort    int
	vhostHTTPTimeout  int64
	dashboardAddr     string
	dashboardPort     int
	dashboardUser     string
	dashboardPwd      string
	enablePrometheus  bool
	assetsDir         string
	logFile           string
	logLevel          string
	logMaxDays        int64
	disableLogColor   bool
	token             string
	subDomainHost     string
	tcpMux            bool
	allowPorts        string
	maxPoolCount      int64
	maxPortsPerClient int64
	tlsOnly           bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file of frps")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of frps")

	rootCmd.PersistentFlags().StringVarP(&bindAddr, "bind_addr", "", "0.0.0.0", "bind address")
	rootCmd.PersistentFlags().IntVarP(&bindPort, "bind_port", "p", 7000, "bind port")
	rootCmd.PersistentFlags().IntVarP(&bindUDPPort, "bind_udp_port", "", 0, "bind udp port")
	rootCmd.PersistentFlags().IntVarP(&kcpBindPort, "kcp_bind_port", "", 0, "kcp bind udp port")
	rootCmd.PersistentFlags().StringVarP(&proxyBindAddr, "proxy_bind_addr", "", "0.0.0.0", "proxy bind address")
	rootCmd.PersistentFlags().IntVarP(&vhostHTTPPort, "vhost_http_port", "", 0, "vhost http port")
	rootCmd.PersistentFlags().IntVarP(&vhostHTTPSPort, "vhost_https_port", "", 0, "vhost https port")
	rootCmd.PersistentFlags().Int64VarP(&vhostHTTPTimeout, "vhost_http_timeout", "", 60, "vhost http response header timeout")
	rootCmd.PersistentFlags().StringVarP(&dashboardAddr, "dashboard_addr", "", "0.0.0.0", "dasboard address")
	rootCmd.PersistentFlags().IntVarP(&dashboardPort, "dashboard_port", "", 0, "dashboard port")
	rootCmd.PersistentFlags().StringVarP(&dashboardUser, "dashboard_user", "", "admin", "dashboard user")
	rootCmd.PersistentFlags().StringVarP(&dashboardPwd, "dashboard_pwd", "", "admin", "dashboard password")
	rootCmd.PersistentFlags().BoolVarP(&enablePrometheus, "enable_prometheus", "", false, "enable prometheus dashboard")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log_file", "", "console", "log file")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log_level", "", "info", "log level")
	rootCmd.PersistentFlags().Int64VarP(&logMaxDays, "log_max_days", "", 3, "log max days")
	rootCmd.PersistentFlags().BoolVarP(&disableLogColor, "disable_log_color", "", false, "disable log color in console")

	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "auth token")
	rootCmd.PersistentFlags().StringVarP(&subDomainHost, "subdomain_host", "", "", "subdomain host")
	rootCmd.PersistentFlags().StringVarP(&allowPorts, "allow_ports", "", "", "allow ports")
	rootCmd.PersistentFlags().Int64VarP(&maxPortsPerClient, "max_ports_per_client", "", 0, "max ports per client")
	rootCmd.PersistentFlags().BoolVarP(&tlsOnly, "tls_only", "", false, "frps tls only")
}

var rootCmd = &cobra.Command{
	// Use:   "frps",
	// Short: "frps is the server of frp (https://github.com/fatedier/frp)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}

		var cfg config.ServerCommonConf
		var err error
		if cfgFile != "" {
			var content []byte
			content, err = config.GetRenderedConfFromFile(cfgFile)
			if err != nil {
				return err
			}
			cfg, err = parseServerCommonCfg(CfgFileTypeIni, content)
		} else {
			cfg, err = parseServerCommonCfg(CfgFileTypeCmd, nil)
		}
		if err != nil {
			return err
		}

		err = runServer(cfg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func parseServerCommonCfg(fileType int, source []byte) (cfg config.ServerCommonConf, err error) {
	if fileType == CfgFileTypeIni {
		cfg, err = config.UnmarshalServerConfFromIni(source)
	} else if fileType == CfgFileTypeCmd {
		cfg, err = parseServerCommonCfgFromCmd()
	}
	if err != nil {
		return
	}
	cfg.Complete()
	err = cfg.Validate()
	if err != nil {
		err = fmt.Errorf("Parse config error: %v", err)
		return
	}
	return
}

func parseServerCommonCfgFromCmd() (cfg config.ServerCommonConf, err error) {
	cfg = config.GetDefaultServerConf()

	cfg.BindAddr = bindAddr
	cfg.BindPort = bindPort
	cfg.BindUDPPort = bindUDPPort
	cfg.KCPBindPort = kcpBindPort
	cfg.ProxyBindAddr = proxyBindAddr
	cfg.VhostHTTPPort = vhostHTTPPort
	cfg.VhostHTTPSPort = vhostHTTPSPort
	cfg.VhostHTTPTimeout = vhostHTTPTimeout
	cfg.DashboardAddr = dashboardAddr
	cfg.DashboardPort = dashboardPort
	cfg.DashboardUser = dashboardUser
	cfg.DashboardPwd = dashboardPwd
	cfg.EnablePrometheus = enablePrometheus
	cfg.LogFile = logFile
	cfg.LogLevel = logLevel
	cfg.LogMaxDays = logMaxDays
	cfg.SubDomainHost = subDomainHost
	cfg.TLSOnly = tlsOnly

	// Only token authentication is supported in cmd mode
	cfg.ServerConfig = auth.GetDefaultServerConf()
	cfg.Token = token
	if len(allowPorts) > 0 {
		// e.g. 1000-2000,2001,2002,3000-4000
		ports, errRet := util.ParseRangeNumbers(allowPorts)
		if errRet != nil {
			err = fmt.Errorf("Parse conf error: allow_ports: %v", errRet)
			return
		}

		for _, port := range ports {
			cfg.AllowPorts[int(port)] = struct{}{}
		}
	}
	cfg.MaxPortsPerClient = maxPortsPerClient
	cfg.DisableLogColor = disableLogColor
	return
}

func runServer(cfg config.ServerCommonConf) (err error) {
	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel, cfg.LogMaxDays, cfg.DisableLogColor)

	if cfgFile != "" {
		log.Info("frps uses config file: %s", cfgFile)
	} else {
		log.Info("frps uses command line arguments for config")
	}

	svr, err := server.NewService(cfg)
	if err != nil {
		return err
	}
	log.Info("frps started successfully")
	svr.Run()
	return
}

/////////////////////////////////
func init() {
	rootCmd.AddCommand(verifyCmd)
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify that the configures is valid",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgFile == "" {
			fmt.Println("no config file is specified")
			return nil
		}
		iniContent, err := config.GetRenderedConfFromFile(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		_, err = parseServerCommonCfg(CfgFileTypeIni, iniContent)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("frps: the configuration file %s syntax is ok\n", cfgFile)
		return nil
	},
}

////////////////////////////////
type FrpServer struct{}

func (r *FrpServer) StartFrpServer() {
	crypto.DefaultSalt = "frp"
	rand.Seed(time.Now().UnixNano())

	Execute()
}

func NewFrpServer() *FrpServer {
	return &FrpServer{}
}

// func main() {
// 	NewFrpServer().StartFrpServer()
// }
