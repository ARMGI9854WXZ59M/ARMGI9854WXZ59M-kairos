package hook

import (
	"fmt"
	"os"
	"strings"
	"time"

	config "github.com/kairos-io/kairos/pkg/config"
	"github.com/kairos-io/kairos/pkg/machine"
	"github.com/kairos-io/kairos/pkg/utils"
	cp "github.com/otiai10/copy"

	kcryptconfig "github.com/kairos-io/kcrypt/pkg/config"
)

type Kcrypt struct{}

func (k Kcrypt) Run(c config.Config) error {

	if len(c.Install.Encrypt) == 0 {
		return nil
	}

	machine.Mount("COS_OEM", "/oem") //nolint:errcheck
	defer func() {
		machine.Umount("/oem") //nolint:errcheck
	}()

	_ = os.MkdirAll("/oem/system/discovery", 0650)

	kcryptc, err := kcryptconfig.GetConfiguration(kcryptconfig.ConfigScanDirs)
	if err != nil {
		fmt.Println("Failed getting kcrypt configuration: ", err.Error())
		if c.FailOnBundleErrors {
			return err
		}
	}

	for _, p := range c.Install.Encrypt {
		out, err := utils.SH(fmt.Sprintf("kcrypt encrypt %s", p))
		if err != nil {
			fmt.Printf("could not encrypt partition: %s\n", out+err.Error())
			if c.FailOnBundleErrors {
				return err
			}
			// Give time to show the error
			time.Sleep(10 * time.Second)
			return nil // do not error out
		}

		err = kcryptc.SetMapping(strings.TrimSpace(out))
		if err != nil {
			fmt.Println("Failed updating the kcrypt configuration file: ", err.Error())
			if c.FailOnBundleErrors {
				return err
			}
		}
	}

	err = kcryptc.WriteMappings(kcryptconfig.MappingsFile)
	if err != nil {
		fmt.Println("Failed writing kcrypt partition mappings: ", err.Error())
		if c.FailOnBundleErrors {
			return err
		}
	}

	if c.Install.SkipEncryptCopyPlugins {
		fmt.Println("Skip discovery plugin copy")
		return nil
	}

	err = cp.Copy("/system/discovery", "/oem/system/discovery")
	if err != nil {
		fmt.Println("Failed during copying discovery plugins: ", err.Error())
		if c.FailOnBundleErrors {
			return err
		}
	}

	return nil
}
