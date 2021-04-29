//nolint:exhaustivestruct,gochecknoglobals
package prepare

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"text/template"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/assets"
)

const (
	mopidyGroup        = "audio"
	mopidyUser         = "mopidy"
	serviceFileName    = "schnutibox.service"
	serviceLocation    = "/etc/systemd/system"
	timesyncGroup      = "systemd-timesync"
	timesyncUser       = "systemd-timesync"
	schnutiboxUser     = "schnutibox"
	schnutboxConfigDir = "/etc/schnutibox"
	upmpdcliUser       = "upmpdcli"
	upmpdcliGroup      = "nogroup"
	snapserverUser     = "snapserver"
	snapserverGroup    = "snapserver"
	snapclientUser     = "snapclient"
	snapclientGroup    = "snapclient"
)

var Cfg = struct {
	RFIDReader          string
	ReadOnly            bool
	Spotify             bool
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyPassword     string
	SpotifyUsername     string
	StopID              string
	System              string
}{}

// boxService creates a systemd service for schnutibox.
func boxService(filename string, enable bool) error {
	logger := log.With().Str("stage", "BoxService").Logger()

	if err := createUser(); err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}

	// Create config dir.
	if err := os.MkdirAll(schnutboxConfigDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not create config dir: %w", err)
	}

	schnutiboxService, err := assets.Assets.ReadFile("files/schnutibox.service")
	if err != nil {
		return fmt.Errorf("could not get service file: %w", err)
	}

	//nolint:gosec
	if err := ioutil.WriteFile(filename, schnutiboxService, 0o644); err != nil {
		return fmt.Errorf("could not write service file: %w", err)
	}

	if enable {
		cmd := exec.Command("systemctl", "daemon-reload")
		logger.Info().Str("cmd", cmd.String()).Msg("running")

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not reload service files: %w", err)
		}

		cmd = exec.Command("systemctl", "enable", "schnutibox.service")
		logger.Info().Str("cmd", cmd.String()).Msg("running")

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not enable service: %w", err)
		}
	}

	return nil
}

func ntp() error {
	logger := log.With().Str("stage", "NTP").Logger()

	cmd := exec.Command("apt-get", "install", "-y", "ntp", "ntpdate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install ntp: %w", err)
	}

	ntpService, err := assets.Assets.ReadFile("files/ntp.service")
	if err != nil {
		return fmt.Errorf("could not get ntp service file: %w", err)
	}

	// nolint:gosec
	if err := ioutil.WriteFile("/etc/systemd/system/ntp.service", ntpService, 0o644); err != nil {
		return fmt.Errorf("could not copy ntp service file: %w", err)
	}

	cmd = exec.Command("systemctl", "daemon-reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not reload systemd service files: %w", err)
	}

	return nil
}

// fstab creates a fstab for a read-only system.
// nolint:funlen
func fstab(system string) error {
	logger := log.With().Str("stage", "Fstab").Logger()

	// Getting timesync user and group informations.
	timesyncUser, err := user.Lookup(timesyncUser)
	if err != nil {
		return fmt.Errorf("could not lookup timesync user: %w", err)
	}

	timesyncGroup, err := user.LookupGroup(timesyncGroup)
	if err != nil {
		return fmt.Errorf("could not lookup timesync group: %w", err)
	}

	logger.Debug().Str("uid", timesyncUser.Uid).Str("gid", timesyncGroup.Gid).Msg("timesyncd")

	// Getting mopidy user and group informations.
	mopidyUser, err := user.Lookup(mopidyUser)
	if err != nil {
		return fmt.Errorf("could not lookup mopidy user: %w", err)
	}

	mopidyGroup, err := user.LookupGroup(mopidyGroup)
	if err != nil {
		return fmt.Errorf("could not lookup mopidy group: %w", err)
	}

	logger.Debug().Str("uid", mopidyUser.Uid).Str("gid", mopidyGroup.Gid).Msg("mopidy")

	// Getting upmpd user and group informations.
	upmpdcliUser, err := user.Lookup(upmpdcliUser)
	if err != nil {
		return fmt.Errorf("could not lookup upmpdcli user: %w", err)
	}

	upmpdcliGroup, err := user.LookupGroup(upmpdcliGroup)
	if err != nil {
		return fmt.Errorf("could not lookup upmpdcli group: %w", err)
	}

	logger.Debug().Str("uid", upmpdcliUser.Uid).Str("gid", upmpdcliGroup.Gid).Msg("upmpdcli")

	// Getting snapserver user and group informations.
	snapserverUser, err := user.Lookup(snapserverUser)
	if err != nil {
		return fmt.Errorf("could not lookup snapserver user: %w", err)
	}

	snapserverGroup, err := user.LookupGroup(snapserverGroup)
	if err != nil {
		return fmt.Errorf("could not lookup snapserver group: %w", err)
	}

	logger.Debug().Str("uid", snapserverUser.Uid).Str("gid", snapserverGroup.Gid).Msg("snapserver")

	snapclientUser, err := user.Lookup(snapclientUser)
	if err != nil {
		return fmt.Errorf("could not lookup snapclient user: %w", err)
	}

	snapclientGroup, err := user.LookupGroup(snapclientGroup)
	if err != nil {
		return fmt.Errorf("could not lookup snapclient group: %w", err)
	}

	logger.Debug().Str("uid", snapclientUser.Uid).Str("gid", snapclientGroup.Gid).Msg("snapclient")

	// Chose the right template.
	// In future it should be a switch statement.
	tmpl, err := assets.Assets.ReadFile("templates/fstab.raspbian.tmpl")
	if err != nil {
		return fmt.Errorf("could not get fstab template: %w", err)
	}

	// Parse template.
	t := template.Must(template.New("fstab").Parse(string(tmpl)))

	// Open fstab.
	f, err := os.Create("/etc/fstab")
	if err != nil {
		return fmt.Errorf("could not create file to write: %w", err)
	}
	defer f.Close()

	// Create and write.
	if err := t.Execute(f, struct {
		TimesyncUID   string
		TimesyncGID   string
		MopidyUID     string
		MopidyGID     string
		UpmpdcliUID   string
		UpmpdcliGID   string
		SnapserverUID string
		SnapserverGID string
		SnapclientUID string
		SnapclientGID string
	}{
		timesyncUser.Uid,
		timesyncGroup.Gid,
		mopidyUser.Uid,
		mopidyGroup.Gid,
		upmpdcliUser.Uid,
		upmpdcliGroup.Gid,
		snapserverUser.Uid,
		snapserverGroup.Gid,
		snapclientUser.Uid,
		snapclientGroup.Gid,
	}); err != nil {
		return fmt.Errorf("could not write templated fstab: %w", err)
	}

	return nil
}

// removePkgs removes not needed software in read-only mode.
func removePkgs(system string) error {
	logger := log.With().Str("stage", "RemovePkgs").Logger()
	if system != "raspbian" {
		logger.Info().Msg("nothing to do")

		return nil
	}

	pkgs := []string{
		"cron",
		"logrotate",
		"triggerhappy",
		"dphys-swapfile",
		"fake-hwclock",
		"samba-common",
	}

	for _, i := range pkgs {
		logger.Debug().Str("pkg", i).Msg("remove package")
		cmd := exec.Command("apt-get", "remove", "-y", i)
		logger.Debug().Str("cmd", cmd.String()).Msg("running")

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not remove pkg: %w", err)
		}
	}

	cmd := exec.Command("apt-get", "autoremove", "--purge", "-y")
	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not do an autoremove: %w", err)
	}

	return nil
}

func udevRules() error {
	logger := log.With().Str("stage", "CreateUDEVrules").Logger()
	logger.Info().Msg("writing udev rule file")

	// Parse template.
	tmpl, err := assets.Assets.ReadFile("templates/50-neuftech.rules.tmpl")
	if err != nil {
		return fmt.Errorf("could not get udev rules file: %w", err)
	}

	t := template.Must(template.New("udev").Parse(string(tmpl)))

	// Open file.
	f, err := os.Create("/etc/udev/rules.d/50-neuftech.rules")
	if err != nil {
		return fmt.Errorf("could not create file to write: %w", err)
	}
	defer f.Close()

	// Create and write.
	if err := t.Execute(f, struct {
		SchnutiboxGroup string
	}{
		schnutiboxUser,
	}); err != nil {
		return fmt.Errorf("could not write templated udev rules: %w", err)
	}

	return nil
}

// createUser creates schnutibox system user and group.
func createUser() error {
	logger := log.With().Str("stage", "CreateUser").Logger()

	cmd := exec.Command("adduser", "--system", "--group", "--no-create-home", schnutiboxUser)
	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}

	return nil
}

// symlinks creates all needed symlinks.
func symlinks(system string) error {
	logger := log.With().Str("stage", "Symlinks").Logger()

	links := []struct {
		symlink string
		dest    string
	}{
		{
			"/var/lib/dhcp",
			"/tmp/dhcp",
		},
		{
			"/var/spool",
			"/tmp/spool",
		},
		{
			"/var/lock",
			"/tmp/lock",
		},
		{
			"/etc/resolv.conf",
			"/tmp/resolv.conf",
		},
	}

	removeFiles := []string{
		"/var/lib/dhcp",
		"/var/spool",
		"/var/lock",
		"/etc/resolv.conf",
	}

	for _, i := range removeFiles {
		logger.Debug().Str("item", i).Msg("remove file/directory")

		if err := os.RemoveAll(i); err != nil {
			return fmt.Errorf("could not remove: %w", err)
		}
	}

	for _, i := range links {
		logger.Debug().Str("symlink", i.symlink).Str("dest", i.dest).Msg("linking")

		dest, err := os.Readlink(i.symlink)
		if err == nil {
			if dest == i.dest {
				logger.Debug().Str("dest", dest).Str("expected dest", i.dest).Msg("matches")

				continue
			}
		}

		if err := os.Symlink(i.dest, i.symlink); err != nil {
			return fmt.Errorf("could not create symlink: %w", err)
		}
	}

	return nil
}

// cmdlineTxt modifies the /boot/cmdline.txt.
func cmdlineTxt() error {
	// Read.
	oldLine, err := ioutil.ReadFile("/boot/cmdline.txt")
	if err != nil {
		return fmt.Errorf("could not read cmdline.txt: %w", err)
	}

	newLine := strings.TrimSuffix(string(oldLine), "\n") + " " + "fastboot" + " " + "noswap"

	// Write.
	// nolint:gosec
	if err := ioutil.WriteFile("/boot/cmdline.txt", []byte(newLine), 0o644); err != nil {
		return fmt.Errorf("could not write cmdline.txt: %w", err)
	}

	return nil
}

// readOnly executes stuff if a read-only system is wanted.
func readOnly(system string) error {
	if err := removePkgs(system); err != nil {
		return fmt.Errorf("could not remove pkgs: %w", err)
	}

	if err := symlinks(system); err != nil {
		return fmt.Errorf("could not create symlinks: %w", err)
	}

	if err := fstab(system); err != nil {
		return fmt.Errorf("could not create fstab: %w", err)
	}

	if err := cmdlineTxt(); err != nil {
		return fmt.Errorf("could not modify cmdline.txt: %w", err)
	}

	return nil
}

// mopidy setups mopidy.
//nolint:funlen,cyclop
func mopidy() error {
	logger := log.With().Str("stage", "Mopidy").Logger()

	// GPG Key.
	cmd := exec.Command("/bin/sh", "-c", "wget -q -O - https://apt.mopidy.com/mopidy.gpg | apt-key add -")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not add mopidy key: %w", err)
	}

	// Repo.
	cmd = exec.Command(
		"/bin/sh", "-c",
		"wget -q -O /etc/apt/sources.list.d/mopidy.list https://apt.mopidy.com/buster.list",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not download apt repo: %w", err)
	}

	// Update.
	cmd = exec.Command("apt-get", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not update apt: %w", err)
	}

	// Install.
	cmd = exec.Command(
		"apt-get", "install", "-y",
		"mopidy",
		"mopidy-alsamixer",
		"mopidy-mpd",
		"mopidy-spotify",
		"python3-pip",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install mopidy: %w", err)
	}

	// Extensions.
	cmd = exec.Command(
		"pip3", "install", "--upgrade",
		"Mopidy-YouTube",
		"requests>=2.22",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install extensions: %w", err)
	}

	// Enable service.
	cmd = exec.Command("systemctl", "enable", "mopidy.service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not enable mopidy service: %w", err)
	}

	// Config.
	if Cfg.SpotifyUsername != "" &&
		Cfg.SpotifyPassword != "" &&
		Cfg.SpotifyClientID != "" &&
		Cfg.SpotifyClientSecret != "" {
		Cfg.Spotify = true
	}

	tmpl, err := assets.Assets.ReadFile("templates/mopidy.conf.tmpl")
	if err != nil {
		return fmt.Errorf("could not get mopidy.conf: %w", err)
	}

	t := template.Must(template.New("mopidyConf").Parse(string(tmpl)))

	f, err := os.Create("/etc/mopidy/mopidy.conf")
	if err != nil {
		return fmt.Errorf("could not create file to write: %w", err)
	}
	defer f.Close()

	if err := t.Execute(f, Cfg); err != nil {
		return fmt.Errorf("could not write compiled mopidy config: %w", err)
	}

	return nil
}

// Upmpdcli setups upmpdcli.
//nolint:funlen
func upmpdcli() error {
	logger := log.With().Str("stage", "Upmpdcli").Logger()

	// GPG Key.
	cmd := exec.Command(
		"/bin/sh", "-c",
		"wget https://www.lesbonscomptes.com/pages/lesbonscomptes.gpg -O /usr/share/keyrings/lesbonscomptes.gpg",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not add upmpdcli key: %w", err)
	}

	// Repo.
	cmd = exec.Command(
		"curl",
		"-o",
		"/etc/apt/sources.list.d/upmpdcli.list",
		"https://www.lesbonscomptes.com/upmpdcli/pages/upmpdcli-rbuster.list",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not add sources list for upmpdcli: %w", err)
	}

	// Update.
	cmd = exec.Command("apt-get", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not update apt: %w", err)
	}

	// Install.
	cmd = exec.Command(
		"apt-get", "install", "-y",
		"upmpdcli",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install mopidy: %w", err)
	}

	// Enable service.
	cmd = exec.Command("systemctl", "enable", "upmpdcli.service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not enable upmpdcli service: %w", err)
	}

	// Create config.
	upmpdcliConf, err := assets.Assets.ReadFile("files/upmpdcli.conf")
	if err != nil {
		return fmt.Errorf("could not get upmpdcli.conf: %w", err)
	}

	// nolint:gosec
	if err := ioutil.WriteFile("/etc/upmpdcli.conf", upmpdcliConf, 0o644); err != nil {
		return fmt.Errorf("could not copy upmpdcli config: %w", err)
	}

	return nil
}

// nolint:funlen
func snapcast() error {
	logger := log.With().Str("stage", "snapcast").Logger()

	// Download deb.
	cmd := exec.Command(
		"wget",
		"https://github.com/badaix/snapcast/releases/download/v0.24.0/snapclient_0.24.0-1_without-pulse_armhf.deb",
		"-O", "/tmp/snapclient.deb",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not download snapclient deb: %w", err)
	}

	// Install deb
	cmd = exec.Command(
		"/bin/sh", "-c",
		"dpkg -i /tmp/snapclient.deb; apt --fix-broken install -y; rm /tmp/snapclient.deb",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install snapclient deb: %w", err)
	}

	// Download deb.
	cmd = exec.Command(
		"wget",
		"https://github.com/badaix/snapcast/releases/download/v0.24.0/snapserver_0.24.0-1_armhf.deb",
		"-O", "/tmp/snapserver.deb",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not download snapserver deb: %w", err)
	}

	// Install deb
	cmd = exec.Command(
		"/bin/sh", "-c",
		"dpkg -i /tmp/snapserver.deb; apt --fix-broken install -y; rm /tmp/snapserver.deb",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install snapserver deb: %w", err)
	}

	return nil
}

func schnutiboxConfig() error {
	logger := log.With().Str("stage", "schnutiboxConfig").Logger()
	logger.Info().Msg("writing schnutibox config")

	// Parse template.
	tmpl, err := assets.Assets.ReadFile("templates/schnutibox.yml.tmpl")
	if err != nil {
		return fmt.Errorf("could not get template: %w", err)
	}

	t := template.Must(template.New("config").Parse(string(tmpl)))

	// Open file.
	f, err := os.Create("/etc/schnutibox/schnutibox.yml")
	if err != nil {
		return fmt.Errorf("could not create file to write: %w", err)
	}
	defer f.Close()

	// Create and write.
	if err := t.Execute(f, Cfg); err != nil {
		return fmt.Errorf("could not write templated udev rules: %w", err)
	}

	return nil
}

func Run(cmd *cobra.Command, args []string) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Install schnutibox service.
	if err := boxService(serviceLocation+"/"+serviceFileName, true); err != nil {
		log.Fatal().Err(err).Msg("could not create schnutibox service")
	}

	// Create schnutibox config.
	if err := schnutiboxConfig(); err != nil {
		log.Fatal().Err(err).Msg("could not create schnutibox config.")
	}

	// Install udev file.
	if err := udevRules(); err != nil {
		log.Fatal().Err(err).Msg("could not install udev rules")
	}

	// Setup NTP.
	if err := ntp(); err != nil {
		log.Fatal().Err(err).Msg("could not setup ntp")
	}

	// Setup mopidy.
	if err := mopidy(); err != nil {
		log.Fatal().Err(err).Msg("could not setup mopidy")
	}

	// Setup upmpdcli.
	if err := upmpdcli(); err != nil {
		log.Fatal().Err(err).Msg("could not setup upmpdcli")
	}

	// Setup snapcast.
	if err := snapcast(); err != nil {
		log.Fatal().Err(err).Msg("could not setup snapclient")
	}

	// Making system read-only.
	if Cfg.ReadOnly {
		if err := readOnly(Cfg.System); err != nil {
			log.Fatal().Err(err).Msg("could not make system read-only")
		}
	}
}
