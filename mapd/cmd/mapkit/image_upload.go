// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/imgpkg/pkg/imgpkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/compressutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/harbor"
)

// UploadImageCmd represents the upload command.
var UploadImageCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload images or image bundles.",
	Long:  `Upload images or image bundles.`,
	Example: `
$ mapkit image upload -i mysave.tar \
    -o myregistry

$ mapkit image upload -i mysave.tar \
    -o myregistry/myproject/a/b/c

$ mapkit image upload -i mysave.tar \
    --registry-username myuser --registry-password mypassword \
    --registry-ca-cert-path rootca.crt \
    -o myregistry/myproject	
`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			exitCode int
			err      error
		)
		if tarFilePath == "" {
			exitCode = 1
			err = fmt.Errorf("specify a path to save the exported images.")
			goto DONE
		}
		if relocateTo == "" {
			exitCode = 1
			err = fmt.Errorf("specify a location to relocate images.")
			goto DONE
		}
		if err = uploadImages(relocateTo, tarFilePath); err != nil {
			exitCode = 2
		}
	DONE:
		if exitCode > 0 {
			fmt.Fprintf(os.Stderr, "FATAL(%d): %v.\n", exitCode, err)
		}
		os.Exit(exitCode)
	},
}

var (
	relocateTo            string
	registryCreateProject bool
)

func init() {
	UploadImageCmd.Flags().StringVarP(&tarFilePath, "input", "i", tarFilePath, "Tar file to upload the images from")
	UploadImageCmd.Flags().StringVarP(&relocateTo, "relocate-to", "o", relocateTo, "Where the image will be relocated/uploaded to")
	UploadImageCmd.Flags().BoolVar(&registryCreateProject, "registry-create-project", registryCreateProject, "Create project if it does not exist on Harbor 2.0 registry")

	UploadImageCmd.Flags().StringSliceVar(&registryCACertPaths, "registry-ca-cert-path", registryCACertPaths, "Add CA certificates for registry API (format: /tmp/foo)")
	UploadImageCmd.Flags().BoolVar(&registryVerifyCerts, "registry-verify-certs", registryVerifyCerts, "Set whether to verify server's certificate chain and host name")
	UploadImageCmd.Flags().BoolVar(&registryInsecure, "registry-insecure", registryInsecure, "Allow the use of http when interacting with registries")
	UploadImageCmd.Flags().StringVar(&registryUsername, "registry-username", registryUsername, "Set username for registry auth")
	UploadImageCmd.Flags().StringVar(&registryPassword, "registry-password", registryPassword, "Set password for registry auth")
	UploadImageCmd.Flags().StringVar(&registryToken, "registry-token", registryToken, "Set token for registry auth")

	mapd.GlobalConfig.HTTPClientTimeout = 10 * time.Second
}

func uploadImages(relocateTo, uploadFrom string) error {
	tempDir, err := os.MkdirTemp("", "mapkit-upload-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	confUI := ui.NewConfUI(ui.NewNoopLogger())
	defer confUI.Flush()

	if err = compressutil.DecompressTAR(tempDir, uploadFrom); err != nil {
		return err
	}

	manifest := filepath.Join(tempDir, "manifest.json")
	b, err := os.ReadFile(manifest)
	if err != nil {
		return errors.Wrap(err, "error reading image manifest")
	}

	var toUpload []*ImageEntry
	if err = json.Unmarshal(b, &toUpload); err != nil {
		return errors.Wrap(err, "error parsing image manifest")
	}

	total := len(toUpload)
	var plainImages, bundleImages int
	for i, each := range toUpload {
		confUI.PrintLinef("copy | image(s) to import: [%d/%d]", i+1, total)

		fp := filepath.Join(tempDir, each.Path)

		for j, ref := range each.Refs {
			newRepo, err := getRelocatedImageName(ref, relocateTo)
			if err != nil {
				return errors.Wrapf(err, "error relocate image %s", ref)
			}
			confUI.PrintLinef("copy | relocate %s => %s [%d/%d]", ref, newRepo, j+1, len(each.Refs))
			if err := uploadImageTo(newRepo, fp, confUI); err != nil {
				return errors.Wrapf(err, "error uploading image %s", newRepo)
			}
		}

		switch each.Type {
		case "image":
			plainImages++
		case "bundle":
			bundleImages++
		}
	}

	confUI.PrintLinef("copy | imported %d plain image(s) and %d bundle image(s)", plainImages, bundleImages)
	confUI.PrintLinef("Succeeded.")
	return nil
}

func getRelocatedImageName(imageRef, relocateTo string) (string, error) {
	name, _, _, err := splitImageRef(imageRef)
	if err != nil {
		return "", err
	}
	ss := qualifyImageName(name)

	newSs := strings.Split(relocateTo, "/")
	if len(newSs) >= len(ss) {
		return relocateTo, nil
	}
	copy(ss, newSs)
	return strings.Join(ss, "/"), nil
}

func uploadImageTo(toRepo, fromFile string, confUI *ui.ConfUI) error {
	opts := &cmd.CopyOptions{
		TarFlags: cmd.TarFlags{
			TarSrc: fromFile,
		},
		RegistryFlags: cmd.RegistryFlags{
			CACertPaths:           registryCACertPaths,
			VerifyCerts:           registryVerifyCerts,
			Insecure:              registryInsecure,
			Username:              registryUsername,
			Password:              registryPassword,
			Token:                 registryToken,
			ResponseHeaderTimeout: 30 * time.Second,
		},
		RepoDst:     toRepo,
		Concurrency: 8,
	}
	if opts.RegistryFlags.Username == "" && opts.RegistryFlags.Token == "" {
		username, password, err := getCrendentialForImageFromDockerConfig(toRepo)
		if err == nil {
			opts.RegistryFlags.Username = username
			opts.RegistryFlags.Password = password
			confUI.PrintLinef("copy | using credential in docker config.")
		} else {
			opts.RegistryFlags.Anon = true
			confUI.PrintLinef("copy | using anonymous auth.")
		}
	} else {
		confUI.PrintLinef("copy | using user-specified credential.")
	}

	if registryCreateProject {
		registry, project := getImageRegistryAndProject(toRepo)
		confUI.PrintLinef("copy | head project: %s/%s", registry, project)

		var endpoint string
		if registryInsecure {
			endpoint = "http://" + registry
		} else {
			endpoint = "https://" + registry
		}

		if opts.RegistryFlags.Username == "" {
			username, password, err := getCrendentialForImageFromDockerConfig(toRepo)
			if err == nil {
				opts.RegistryFlags.Username = username
				opts.RegistryFlags.Password = password
			}
		}
		cli, err := harbor.NewV2Client(endpoint, opts.RegistryFlags.Username, []byte(opts.RegistryFlags.Password), !registryVerifyCerts)
		if err != nil {
			return err
		}
		if err = cli.HeadProject(project); err != nil {
			if !harbor.IsErrObjectNotExists(err) {
				return err
			}

			if err = cli.PostProject(&harbor.CreateProject{
				ProjectName:  project,
				StorageLimit: -1,
				Public:       true,
			}); err != nil {
				return err
			}
			confUI.PrintLinef("copy | created project: %s/%s", registry, project)
		}
	}

	return opts.Run()
}

func getImageRegistryAndProject(imageName string) (string, string) {
	ss := qualifyImageName(imageName)
	if len(ss) > 1 {
		return ss[0], ss[1]
	}
	return ss[0], "library"
}
