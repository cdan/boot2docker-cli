package vsphere

import (
	"fmt"
	"strings"

	"code.google.com/p/gopass"
	"github.com/boot2docker/boot2docker-cli/vsphere/errors"
)

type VcConn struct {
	cfg      *DriverCfg
	password string
}

func NewVcConn(cfg *DriverCfg) VcConn {
	return VcConn{
		cfg:      cfg,
		password: "",
	}
}

func (conn VcConn) Login() error {
	err := conn.queryAboutInfo()
	if err == nil {
		return nil
	}

	// need to login here
	conn.password, err = gopass.GetPass("Enter vCenter Password: ")
	if err != nil {
		return err
	}
	err = conn.queryAboutInfo()
	if err == nil {
		return nil
	}
	return err
}

func (conn VcConn) AppendConnectionString(args []string) []string {
	if conn.password == "" {
		args = append(args, fmt.Sprintf("--u=%s@%s", conn.cfg.VcenterUser, cfg.VcenterIp))
	} else {
		args = append(args, fmt.Sprintf("--u=%s:%s@%s", conn.cfg.VcenterUser, conn.password, conn.cfg.VcenterIp))
	}
	args = append(args, "--k=true")
	return args
}

func (conn VcConn) queryAboutInfo() error {
	args := []string{"about"}
	args = conn.AppendConnectionString(args)
	stdout, _, _ := govcOutErr(args...)
	if strings.Contains(stdout, "Name") {
		return nil
	}
	return errors.NewInvalidLoginError()
}
