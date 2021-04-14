//nolint:exhaustivestruct,gochecknoglobals
package prepare

//nolint:golint
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
)

// Config.
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

// BoxService creates a systemd service for schnutibox.
func BoxService(filename string, enable bool) error {
	logger := log.With().Str("stage", "BoxService").Logger()

	if err := CreateUser(); err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}

	// Create config dir.
	if err := os.MkdirAll(schnutboxConfigDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not create config dir: %w", err)
	}

	//nolint:gosec
	if err := ioutil.WriteFile(filename, assets.SchnutiboxService, 0o644); err != nil {
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

func NTP() error {
	logger := log.With().Str("stage", "NTP").Logger()

	cmd := exec.Command("apt-get", "install", "-y", "ntp", "ntpdate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install ntp: %w", err)
	}

	if err := ioutil.WriteFile("/etc/systemd/system/ntp.service", assets.NtpService, 0o644); err != nil {
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

// Fstab creates a fstab for a read-only system.
// nolint:funlen
func Fstab(system string) error {
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

	// Chose the right template.
	// In future it should be a switch statement.
	tmpl := assets.FstabRaspbianTemplate

	// Parse template.
	t := template.Must(template.New("fstab").Parse(tmpl))

	// Open fstab.
	f, err := os.Create("/etc/fstab")
	if err != nil {
		return fmt.Errorf("could not create file to write: %w", err)
	}
	defer f.Close()

	// Create and write.
	if err := t.Execute(f, struct {
		TimesyncUID string
		TimesyncGID string
		MopidyUID   string
		MopidyGID   string
		UpmpdcliUID string
		UpmpdcliGID string
	}{
		timesyncUser.Uid,
		timesyncGroup.Gid,
		mopidyUser.Uid,
		mopidyGroup.Gid,
		upmpdcliUser.Uid,
		upmpdcliGroup.Gid,
	}); err != nil {
		return fmt.Errorf("could not write templated fstab: %w", err)
	}

	return nil
}

// RemovePkgs removes not needed software in read-only mode.
func RemovePkgs(system string) error {
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

func CreateUDEVrules() error {
	logger := log.With().Str("stage", "CreateUDEVrules").Logger()
	logger.Info().Msg("writing udev rule file")

	// Parse template.
	t := template.Must(template.New("udev").Parse(assets.UDEVRules))

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

// CreateUser creates schnutibox system user and group.
func CreateUser() error {
	logger := log.With().Str("stage", "CreateUser").Logger()

	cmd := exec.Command("adduser", "--system", "--group", "--no-create-home", schnutiboxUser)
	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}

	return nil
}

// CreateSymlinks creates all needed symlinks.
func CreateSymlinks(system string) error {
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
			return err
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
func cmdlineTxt(system string) error {
	// Read.
	oldLine, err := ioutil.ReadFile("/boot/cmdline.txt")
	if err != nil {
		return fmt.Errorf("could not read cmdline.txt: %w", err)
	}

	newLine := strings.TrimSuffix(string(oldLine), "\n") + " " + "fastboot" + " " + "noswap"

	// Write.
	if err := ioutil.WriteFile("/boot/cmdline.txt", []byte(newLine), 0o644); err != nil {
		return fmt.Errorf("could not write cmdline.txt: %w", err)
	}

	return nil
}

// makeReadOnly executes stuff if a read-only system is wanted.
func makeReadOnly(system string) error {
	if err := RemovePkgs(system); err != nil {
		return fmt.Errorf("could not remove pkgs: %w", err)
	}

	if err := CreateSymlinks(system); err != nil {
		return fmt.Errorf("could not create symlinks: %w", err)
	}

	if err := Fstab(system); err != nil {
		return fmt.Errorf("could not create fstab: %w", err)
	}

	if err := cmdlineTxt(system); err != nil {
		return fmt.Errorf("could not modify cmdline.txt: %w", err)
	}

	return nil
}

// Mopidy setups mopidy.
//nolint:funlen
func Mopidy() error {
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
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug().Str("cmd", cmd.String()).Msg("running")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install mopidy: %w", err)
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

	t := template.Must(template.New("mopidyConf").Parse(assets.MopidyConf))

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
func Upmpdcli() error {
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
	if err := ioutil.WriteFile("/etc/upmpdcli.conf", assets.UpmpdcliConf, 0o644); err != nil {
		return fmt.Errorf("could not copy upmpdcli config: %w", err)
	}

	return nil
}

func SchnutiboxConfig() error {
	logger := log.With().Str("stage", "Upmpdcli").Logger()
	logger.Info().Msg("writing schnutibox config")

	// Parse template.
	t := template.Must(template.New("config").Parse(assets.SchnutiboxConfig))

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
	if err := BoxService(serviceLocation+"/"+serviceFileName, true); err != nil {
		log.Fatal().Err(err).Msg("could not create schnutibox service")
	}

	// Create schnutibox config.
	if err := SchnutiboxConfig(); err != nil {
		log.Fatal().Err(err).Msg("could not create schnutibox config.")
	}

	// Install udev file.
	if err := CreateUDEVrules(); err != nil {
		log.Fatal().Err(err).Msg("could not install udev rules")
	}

	// Setup NTP.
	if err := NTP(); err != nil {
		log.Fatal().Err(err).Msg("could not setup ntp")
	}

	// Setup mopidy.
	if err := Mopidy(); err != nil {
		log.Fatal().Err(err).Msg("could not setup mopidy")
	}

	// Setup upmpdcli.
	if err := Upmpdcli(); err != nil {
		log.Fatal().Err(err).Msg("could not setup upmpdcli")
	}

	// Making system read-only.
	if Cfg.ReadOnly {
		if err := makeReadOnly(Cfg.System); err != nil {
			log.Fatal().Err(err).Msg("could not make system read-only")
		}
	}
}
