package vsphere

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/boot2docker/boot2docker-cli/driver"
	"github.com/boot2docker/boot2docker-cli/vsphere/errors"
	flag "github.com/ogier/pflag"
)

type DriverCfg struct {
	Govc        string // Path to govc binary
	VcenterIp   string // vCenter URL
	VcenterUser string // vCenter User
	VcenterDC   string // target vCenter Datacenter
	VcenterDS   string // target vCenter Datastore
	VcenterNet  string // vCenter VM Network
}

var (
	verbose bool // Verbose mode (Local copy of B2D.Verbose).
	cfg     DriverCfg

	// TODO make all these below errors a class
	ErrVmIpNotFound = errors.New("VM IP Address not found")
)

func init() {
	if err := driver.Register("vsphere", InitFunc); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize driver. Error : %s", err.Error())
		os.Exit(1)
	}
	if err := driver.RegisterConfig("vsphere", ConfigFlags); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize driver config. Error : %s", err.Error())
		os.Exit(1)
	}
}

// Initialize the Machine.
func InitFunc(mc *driver.MachineConfig) (driver.Machine, error) {
	verbose = mc.Verbose

	m, err := GetMachine(mc)
	//if err != nil && mc.Init == true {
	//return CreateMachine(mc)
	//}
	return m, err
}

// Add cmdline params for this driver
func ConfigFlags(B2D *driver.MachineConfig, flags *flag.FlagSet) error {
	flags.StringVar(&cfg.Govc, "govc", "govc", "path to govc binary")
	flags.StringVar(&cfg.VcenterIp, "vcenter-ip", "", "vCenter URL")
	flags.StringVar(&cfg.VcenterUser, "vcenter-user", "", "vCenter User")
	flags.StringVar(&cfg.VcenterDC, "vcenter-datacenter", "", "vCenter Datacenter")
	flags.StringVar(&cfg.VcenterDS, "vcenter-datastore", "", "vCenter Datastore")
	flags.StringVar(&cfg.VcenterNet, "vcenter-vm-network", "", "vCenter VM network")

	return nil
}

// GetMachine fetches the machine information from a vCenter
func GetMachine(mc *driver.MachineConfig) (*Machine, error) {
	err := GetDriverCfg(mc)
	if err != nil {
		return nil, err
	}

	args := []string{"vm.info"}
	args = append(args, fmt.Sprintf("--u=%s@%s", cfg.VcenterUser, cfg.VcenterIp))
	args = append(args, "--k=true")
	args = append(args, fmt.Sprintf("--dc=%s", cfg.VcenterDC))
	args = append(args, mc.VM)

	m := &Machine{Name: mc.VM, State: driver.Poweroff}
	stdout, _, err := govcOutErr(args...)
	if err != nil {
		fmt.Println("errors! %s", err)
		return nil, err
	}
	if strings.Contains(stdout, "Name") {
		currentCpu := strings.Trim(strings.Split(strings.Split(stdout, "CPU:")[1], "vCPU")[0], " ")
		if cpus, err := strconv.Atoi(currentCpu); err == nil {
			m.CPUs = cpus
		}
		currentMem := strings.Trim(strings.Split(strings.Split(stdout, "Memory:")[1], "MB")[0], " ")
		if mem, err := strconv.Atoi(currentMem); err == nil {
			m.Memory = mem
		}
		if strings.Contains(stdout, "poweredOn") {
			m.State = driver.Running
			m.VmIp = strings.Trim(strings.Split(stdout, "IP address:")[1], " ")
		}
		m.VcenterIp, _ = mc.DriverCfg["VcenterIp"].(string)
		m.VcenterUser, _ = mc.DriverCfg["VcenterUser"].(string)
		return m, nil
	}
	return nil, errors.NewVmNotFoundError()
}

func GetDriverCfg(mc *driver.MachineConfig) error {
	vcenterIp := mc.DriverCfg["vsphere"].(map[string]interface{})["VcenterIp"]
	if vcenterIp == nil {
		if cfg.VcenterIp == "" {
			return errors.NewIncompleteVcConfigError("vCenter IP")
		}
	} else {
		cfg.VcenterIp = vcenterIp.(string)
	}
	vcenterUser := mc.DriverCfg["vsphere"].(map[string]interface{})["VcenterUser"]
	if vcenterUser == nil {
		if cfg.VcenterUser == "" {
			return errors.NewIncompleteVcConfigError("vCenter User")
		}
	} else {
		cfg.VcenterUser = vcenterUser.(string)
	}
	vcenterDC := mc.DriverCfg["vsphere"].(map[string]interface{})["VcenterDC"]
	if vcenterDC == nil {
		if cfg.VcenterDC == "" {
			return errors.NewIncompleteVcConfigError("vCenter Datacenter")
		}
	} else {
		cfg.VcenterDC = vcenterDC.(string)
	}
	return nil
}

// Machine information.
type Machine struct {
	Name        string
	State       driver.MachineState
	CPUs        int
	Memory      int
	VcenterIp   string // the vcenter the machine belongs to
	VcenterUser string // the vcenter user/admin to own the machine
	Datastore   string // the datastore for the ISO file
	Datacenter  string // the datacenter the machine locates
	Network     string // the network the machine is using
	VmIp        string // the Ip address of the machine
}

// Refresh reloads the machine information.
func (m *Machine) Refresh() error {
	fmt.Printf("Refresh %s: %s\n", m.Name, m.State)
	return nil
}

// Start starts the machine.
func (m *Machine) Start() error {
	args := []string{"vm.power"}
	m.appendConnectionString(args)
	args = append(args, "-on")
	args = append(args, m.Name)

	if err := govc(args...); err != nil {
		return err
	}

	if ip, _, err := m.fetchIp(); ip == "" && err != nil {
		return err
	}
	fmt.Printf("Start %s: %s\n", m.Name, m.State)
	return nil
}

// Suspend suspends the machine and saves its state to disk.
func (m *Machine) Save() error {
	m.State = driver.Saved
	fmt.Printf("Save %s: %s\n", m.Name, m.State)
	return nil
}

// Pause pauses the execution of the machine.
func (m *Machine) Pause() error {
	m.State = driver.Paused
	fmt.Printf("Pause %s: %s\n", m.Name, m.State)
	return nil
}

// Stop gracefully stops the machine.
func (m *Machine) Stop() error {
	m.State = driver.Poweroff
	fmt.Printf("Stop %s: %s\n", m.Name, m.State)
	return nil
}

// Poweroff forcefully stops the machine. State is lost and might corrupt the disk image.
func (m *Machine) Poweroff() error {
	m.State = driver.Poweroff
	fmt.Printf("Poweroff %s: %s\n", m.Name, m.State)
	return nil
}

// Restart gracefully restarts the machine.
func (m *Machine) Restart() error {
	m.State = driver.Running
	fmt.Printf("Restart %s: %s\n", m.Name, m.State)
	return nil
}

// Reset forcefully restarts the machine. State is lost and might corrupt the disk image.
func (m *Machine) Reset() error {
	m.State = driver.Running
	fmt.Printf("Reset %s: %s\n", m.Name, m.State)
	return nil
}

// Get current name
func (m *Machine) GetName() string {
	return m.Name
}

// Get current state
func (m *Machine) GetState() driver.MachineState {
	return m.State
}

// Get serial file
func (m *Machine) GetSerialFile() string {
	return ""
}

// Get Docker port
func (m *Machine) GetDockerPort() uint {
	return 2375
}

// Get SSH port
func (m *Machine) GetSSHPort() uint {
	return 22
}

// Delete deletes the machine and associated disk images.
func (m *Machine) Delete() error {
	fmt.Printf("Delete %s: %s\n", m.Name, m.State)
	return nil
}

// Modify changes the settings of the machine.
func (m *Machine) Modify() error {
	fmt.Printf("Modify %s: %s\n", m.Name, m.State)
	return m.Refresh()
}

// AddNATPF adds a NAT port forarding rule to the n-th NIC with the given name.
func (m *Machine) AddNATPF(n int, name string, rule driver.PFRule) error {
	fmt.Println("Add NAT PF")
	return nil
}

// DelNATPF deletes the NAT port forwarding rule with the given name from the n-th NIC.
func (m *Machine) DelNATPF(n int, name string) error {
	fmt.Println("Del NAT PF")
	return nil
}

// SetNIC set the n-th NIC.
func (m *Machine) SetNIC(n int, nic driver.NIC) error {
	fmt.Println("Set NIC")
	return nil
}

// AddStorageCtl adds a storage controller with the given name.
func (m *Machine) AddStorageCtl(name string, ctl driver.StorageController) error {
	fmt.Println("Add storage ctl")
	return nil
}

// DelStorageCtl deletes the storage controller with the given name.
func (m *Machine) DelStorageCtl(name string) error {
	fmt.Println("Del storage ctl")
	return nil
}

// AttachStorage attaches a storage medium to the named storage controller.
func (m *Machine) AttachStorage(ctlName string, medium driver.StorageMedium) error {
	fmt.Println("Attach storage")
	return nil
}

func (m *Machine) appendConnectionString(args []string) []string {
	args = append(args, fmt.Sprintf("--u=", m.VcenterUser, "@", m.VcenterIp))
	args = append(args, "--k=true")
	return args
}

func (m *Machine) fetchIp() (string, string, error) {
	args := []string{"vm.ip"}
	m.appendConnectionString(args)
	args = append(args, m.Name)
	return govcOutErr(args...)
}
