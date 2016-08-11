package vscale

import (
	"fmt"
	"net"
	"time"
	"io/ioutil"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/ssh"
	api "github.com/vscale/go-vscale"
)

type Driver struct {
	*drivers.BaseDriver
	AccessToken string
	ScaletID    int64
	ScaletName  string
	Rplan       string
	MadeFrom    string
	Location    string
	SSHKeyID    int64
	SwapFile    int
}

const (
	defaultRplan = "small"
	defaultLocation = "spb0"
	defaultMadeFrom = "ubuntu_14.04_64_002_master"
	defaultSwapFile = 0
)

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "VSCALE_ACCESS_TOKEN",
			Name:   "vscale-access-token",
			Usage:  "Vscale access token",
		},
		mcnflag.StringFlag{
			EnvVar: "VSCALE_LOCATION",
			Name:   "vscale-location",
			Usage:  "Vscale location",
			Value:  defaultLocation,
		},
		mcnflag.StringFlag{
			EnvVar: "VSCALE_RPLAN",
			Name:   "vscale-rplan",
			Usage:  "Vscale rplan",
			Value:  defaultRplan,
		},
		mcnflag.StringFlag{
			EnvVar: "VSCALE_MADE_FROM",
			Name:   "vscale-made-from",
			Usage:  "Vscale made from",
			Value:  defaultMadeFrom,
		},
		mcnflag.IntFlag{
			EnvVar: "VSCALE_SWAP_FILE",
			Name:   "vscale-swap-file",
			Usage:  "Vscale swap file",
			Value:  defaultSwapFile,
		},
	}
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		Rplan:    defaultRplan,
		Location: defaultLocation,
		MadeFrom: defaultMadeFrom,
		SwapFile: defaultSwapFile,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "vscale"
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) createSSHKey() (*api.SSHKey, error) {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	key, _, err := d.getClient().SSHKey.Create(string(publicKey), d.MachineName)
	return key, err
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessToken = flags.String("vscale-access-token")
	d.Location = flags.String("vscale-location")
	d.Rplan = flags.String("vscale-rplan")
	d.MadeFrom = flags.String("vscale-made-from")
	d.SwapFile = flags.Int("vscale-swap-file")

	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHPort = 22

	if d.AccessToken == "" {
		return fmt.Errorf("vscale driver requres the --vscale-access-token option")
	}

	return nil
}

func (d *Driver) getClient() *api.WebClient {
	return api.NewClient(d.AccessToken)
}

func (d *Driver) PreCreateCheck() error {
	client := d.getClient()
	if client == nil {
		return fmt.Errorf("Cannot create Vscale client. Check --vscale-access-token option")
	}

	return nil
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")
	key, err := d.createSSHKey()
	if err != nil {
		return err
	}
	d.SSHKeyID = key.ID

	log.Infof("Creating Vscale scalet...")

	client := d.getClient()
	newScalet, _, err := client.Scalet.CreateWithoutPassword(d.MadeFrom, d.Rplan, d.MachineName, d.Location, true, []int64{d.SSHKeyID}, true)

	if err != nil {
		return err
	}

	d.ScaletID = newScalet.CTID

	log.Info("Waiting for IP address to be assigned to the Scalet...")

	for {
		newScalet, _, err = client.Scalet.Get(d.ScaletID)
		if err != nil {
			return err
		}

		if newScalet.Active == true {
			d.IPAddress = newScalet.PublicAddresses.Address
		}

		if d.IPAddress != "" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Info(fmt.Sprintf("Scalet created! ID: %v, IPAddress: %v", d.ScaletID, d.IPAddress))
	if d.SwapFile > 0 {
		for {
			ssh := drivers.WaitForSSH(d)
			if ssh == nil {
				log.Info(fmt.Sprintf("Creating SWAP file %d MB", d.SwapFile))
				_, err := drivers.RunSSHCommandFromDriver(d, fmt.Sprintf(`touch /var/swap.img && \
					chmod 600 /var/swap.img && \
					dd if=/dev/zero of=/var/swap.img bs=1MB count=%d && \
					mkswap /var/swap.img && swapon /var/swap.img && \
					echo '/var/swap.img    none    swap    sw    0    0' >> /etc/fstab`, d.SwapFile))

				if err != nil {
					return err
				}
				break
			}

			if err != nil {
				return err
			}
			time.Sleep(3 * time.Second)
		}
	}
	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	scalet, _, err := d.getClient().Scalet.Get(d.ScaletID)
	if err != nil {
		return state.Error, err
	}

	switch scalet.Status {
	case "started":
		return state.Running, nil
	case "stopped":
		return state.Stopped, nil
	case "defined":
		return state.Starting, nil
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	_, _, err := d.getClient().Scalet.Start(d.ScaletID, true)
	return err
}

func (d *Driver) Stop() error {
	_, _, err := d.getClient().Scalet.Stop(d.ScaletID, true)
	return err
}

func (d *Driver) Remove() error {
	client := d.getClient()
	_, _, err := client.Scalet.Remove(d.ScaletID, true)
	if err != nil {
		return err
	}

	_, _ , _ = client.SSHKey.Remove(d.SSHKeyID)

	return nil
}

func (d *Driver) Restart() error {
	_, _, err := d.getClient().Scalet.Restart(d.ScaletID, true)
	return err
}

func (d *Driver) Kill() error {
	_, _, err := d.getClient().Scalet.Remove(d.ScaletID, true)
	return err
}