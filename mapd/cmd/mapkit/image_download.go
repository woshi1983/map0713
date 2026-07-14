// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/google/uuid"
	"github.com/k14s/imgpkg/pkg/imgpkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/compressutil"
)

// DownloadImageCmd represents the download command.
var DownloadImageCmd = &cobra.Command{
	Use:   "download",
	Short: "Download images or image bundles.",
	Long:  `Download images or image bundles.`,
	Example: `
Download plain images:
$ mapkit image download -i myregistry/myproject/myimage1:mytag1 \
    -i myregistry/myproject/myimage2:mytag2 -o mysave.tar

Download image bundles:
$ mapkit image download -b myregistry/myproject/mybundle1:mytag1 \
    -b myregistry/myproject/mybundle2:mytag2 -o mysave.tar

Download plain images from Helm charts:
$ mapkit image download -c ./postgresql-11.8.1.tgz \
    -c ./mariadb-11.3.0.tgz -o mysave.tar

Download images from index:
$ mapkit image download -I myregistry/myproject/myindex:mytag \
    -o mysave.tar

Download images from multiple sources:
$ mapkit image download -i myregistry/myproject/myimage1:mytag1 \
    -b myregistry/myproject/mybundle2:mytag2 \
    --registry-username myuser --registry-password mypassword \
    --registry-ca-cert-path rootca.crt -o mysave.tar`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			exitCode int
			err      error
		)
		if tarFilePath == "" {
			exitCode = 1
			err = fmt.Errorf("specify a path to save the exported images")
			goto DONE
		}
		if indexImage == "" && len(plainImages)+len(bundleImages)+len(helmCharts) < 1 {
			exitCode = 1
			err = fmt.Errorf("at least one plain image, image bundle, or Helm chart must be specified")
			goto DONE
		}
		if err = downloadAndSaveImages(tarFilePath, indexImage, plainImages, bundleImages, helmCharts); err != nil {
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
	plainImages  []string
	bundleImages []string
	helmCharts   []string
	indexImage   string
)

func init() {
	DownloadImageCmd.Flags().StringSliceVarP(&plainImages, "image", "i", plainImages, "Plain images to download")
	DownloadImageCmd.Flags().StringSliceVarP(&bundleImages, "image-bundle", "b", bundleImages, "Bundle images to download")
	DownloadImageCmd.Flags().StringVarP(&indexImage, "image-index", "I", indexImage, "Index image to download from")
	DownloadImageCmd.Flags().StringSliceVarP(&helmCharts, "helm-chart", "c", helmCharts, "(Experimental) Paths of the helm chart to download images from")
	DownloadImageCmd.Flags().StringVarP(&tarFilePath, "output", "o", tarFilePath, "Tar file to download the images to")

	DownloadImageCmd.Flags().StringSliceVar(&registryCACertPaths, "registry-ca-cert-path", registryCACertPaths, "Add CA certificates for registry API (format: /tmp/foo)")
	DownloadImageCmd.Flags().BoolVar(&registryVerifyCerts, "registry-verify-certs", registryVerifyCerts, "Set whether to verify server's certificate chain and host name")
	DownloadImageCmd.Flags().BoolVar(&registryInsecure, "registry-insecure", registryInsecure, "Allow the use of http when interacting with registries")
	DownloadImageCmd.Flags().StringVar(&registryUsername, "registry-username", registryUsername, "Set username for registry auth")
	DownloadImageCmd.Flags().StringVar(&registryPassword, "registry-password", registryPassword, "Set password for registry auth")
	DownloadImageCmd.Flags().StringVar(&registryToken, "registry-token", registryToken, "Set token for registry auth")
}

func downloadAndSaveImages(saveTo, indexImage string, plainImages, bundleImages, helmCharts []string) error {
	tempDir, err := os.MkdirTemp("", "mapkit-download-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	confUI := ui.NewConfUI(ui.NewNoopLogger())
	defer confUI.Flush()

	if indexImage != "" {
		var images, bundles []string
		images, bundles, err = downloadAndGetIndex(indexImage, confUI)
		if err != nil {
			return err
		}
		plainImages = append(plainImages, images...)
		bundleImages = append(bundleImages, bundles...)
	}
	for _, each := range helmCharts {
		confUI.PrintLinef("helm | evaluating chart %s with default values.", each)
		images, err := getReferencedImagesOfHelmChart(each)
		if err != nil {
			return err
		}
		plainImages = append(plainImages, images...)
	}

	tempSaveDir, err := os.MkdirTemp("", "mapkit-download-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempSaveDir)

	entries, err := saveImages(tempSaveDir, plainImages, bundleImages, confUI)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		filename := "sha256-" + entry.ID + ".tar"
		fp := filepath.Join(tempDir, filename)
		if err := os.Rename(entry.Path, fp); err != nil {
			return err
		}
		entry.Path = "./" + filename
	}

	b, err := json.Marshal(entries)
	if err != nil {
		return errors.Wrapf(err, "error generating manifest.json")
	}
	manifest := filepath.Join(tempDir, "manifest.json")
	if err := os.WriteFile(manifest, b, 0644); err != nil {
		return errors.Wrapf(err, "error writing manifest.json")
	}

	if err := compressutil.CompressTAR(saveTo, tempDir); err != nil {
		return err
	}

	confUI.PrintLinef("Succeeded.")
	return nil
}

func saveImages(saveDir string, plainImages, bundleImages []string, confUI *ui.ConfUI) ([]*ImageEntry, error) {
	digestTable := make(map[string]*ImageEntry) // SHA256 -> image
	var n int
	total := len(plainImages) + len(bundleImages)
	for _, each := range plainImages {
		confUI.PrintLinef("copy | image(s) to export: [%d/%d]", n+1, total)
		confUI.PrintLinef("copy | downloading image: %s", each)
		fp := filepath.Join(saveDir, uuid.NewString()+".tar")
		digest, err := saveImageTo(fp, each, confUI)
		if err != nil {
			return nil, err
		}
		entry, ok := digestTable[digest]
		if !ok {
			digestTable[digest] = &ImageEntry{
				ID:   digest,
				Path: fp,
				Type: "image",
				Refs: []string{each},
			}
		} else {
			if entry.Type != "image" {
				return nil, fmt.Errorf("image %s has the same digest with %s %s", each, entry.Type, entry.Refs[0])
			}
			os.Remove(fp)
			entry.Refs = append(entry.Refs, each)
		}
		n++
	}
	for _, each := range bundleImages {
		confUI.PrintLinef("copy | image(s) to export: [%d/%d]", n+1, total)
		confUI.PrintLinef("copy | downloading image: %s", each)
		filename := filepath.Join(saveDir, uuid.NewString()+".tar")
		digest, err := saveImageBundleTo(filename, each, confUI)
		if err != nil {
			return nil, err
		}
		entry, ok := digestTable[digest]
		if !ok {
			digestTable[digest] = &ImageEntry{
				ID:   digest,
				Path: filename,
				Type: "bundle",
				Refs: []string{each},
			}
		} else {
			if entry.Type != "bundle" {
				return nil, fmt.Errorf("bundle %s has the same digest with %s %s", each, entry.Type, entry.Refs[0])
			}
			os.Remove(filename)
			entry.Refs = append(entry.Refs, each)
		}
		n++
	}
	var saved []*ImageEntry
	for i := range digestTable {
		saved = append(saved, digestTable[i])
	}
	confUI.PrintLinef("copy | exported %d plain image(s) and %d bundle image(s)", len(plainImages), len(bundleImages))
	return saved, nil
}

type savedImage struct {
	Ref    string
	Digest string
	Path   string
	Type   string
}

func saveImageTo(saveTo, imageRef string, confUI *ui.ConfUI) (string, error) {
	opts := &cmd.CopyOptions{
		ImageFlags: cmd.ImageFlags{
			Image: imageRef,
		},
		TarFlags: cmd.TarFlags{
			TarDst: saveTo,
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
		Concurrency: 8,
	}
	if opts.RegistryFlags.Username == "" && opts.RegistryFlags.Token == "" {
		username, password, err := getCrendentialForImageFromDockerConfig(imageRef)
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
	return downloadWithImgpkg(saveTo, opts)
}

func saveImageBundleTo(saveTo, imageRef string, confUI *ui.ConfUI) (string, error) {
	opts := &cmd.CopyOptions{
		BundleFlags: cmd.BundleFlags{
			Bundle: imageRef,
		},
		TarFlags: cmd.TarFlags{
			TarDst: saveTo,
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
		Concurrency: 8,
	}
	if opts.RegistryFlags.Username == "" && opts.RegistryFlags.Token == "" {
		username, password, err := getCrendentialForImageFromDockerConfig(imageRef)
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
	return downloadWithImgpkg(saveTo, opts)
}

func downloadWithImgpkg(saveTo string, opts *cmd.CopyOptions) (string, error) {
	if err := opts.Run(); err != nil {
		return "", err
	}

	h := newSHA256Writer(io.Discard)
	f, err := os.Open(saveTo)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return h.SHA256(), nil
}

func getCrendentialForImageFromDockerConfig(imageRef string) (username, password string, err error) {
	registryName, err := getImageRegistry(imageRef)
	if err != nil {
		return username, password, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return username, password, nil
	}
	p, err := os.ReadFile(filepath.Join(homeDir, ".docker", "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return username, password, nil
		}
		return username, password, err
	}
	cred := &DockerConfigJSON{}
	if err := json.Unmarshal(p, cred); err != nil {
		return username, password, errors.Wrap(err, "error reading docker config file")
	}

	match := func(s string) bool {
		return s == registryName || strings.HasPrefix(s, "https://"+registryName+"/")
	}
	if registryName == "docker.io" || registryName == "index.docker.io" {
		match = func(s string) bool {
			return strings.HasPrefix(s, "https://index.docker.io/")
		}
	}

	for reg, each := range cred.Auths {
		if match(reg) {
			if each.Auth != nil && len(each.Auth) > 0 {
				ss := strings.Split(string(each.Auth), ":")
				if len(ss) > 1 {
					username, password = ss[0], ss[1]
				}
			}
			if username == "" && each.Username != "" && each.Password != "" {
				username, password = each.Username, each.Password
			}
			break
		}
	}
	return username, password, nil
}

func getImageRegistry(imageRef string) (string, error) {
	name, _, _, err := splitImageRef(imageRef)
	if err != nil {
		return "", err
	}
	return qualifyImageName(name)[0], nil
}

func qualifyImageName(imageName string) []string {
	ss := strings.Split(imageName, "/")
	switch len(ss) {
	case 1:
		ss = []string{"docker.io", "library", ss[0]}
	case 2:
		ss = []string{"docker.io", ss[0], ss[1]}
	}
	return ss
}

func splitImageRef(imageRef string) (name, digest, tag string, err error) {
	i := strings.IndexRune(imageRef, '@')
	if i > -1 {
		if i == 0 {
			return name, digest, tag, fmt.Errorf("invalid image: %s", imageRef)
		}
		digest = imageRef[i+1:]
		if digest == "" {
			return name, digest, tag, fmt.Errorf("invalid image: %s", imageRef)
		}
		return imageRef[:i], digest, tag, nil
	}
	i = strings.IndexRune(imageRef, ':')
	if i < 0 {
		return imageRef, digest, "latest", nil
	} else if i == 0 {
		return name, digest, tag, fmt.Errorf("invalid image: %s", imageRef)
	}
	tag = imageRef[i+1:]
	if tag == "" {
		return name, digest, tag, fmt.Errorf("invalid image: %s", imageRef)
	}
	return imageRef[:i], digest, tag, nil
}

type DockerConfigJSON struct {
	Auths map[string]*DockerConfigJSONAuth `json:"auths"`
}

type DockerConfigJSONAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     []byte `json:"auth,omitempty"`
}

type sha256Writer struct {
	hash.Hash
	w io.Writer
}

func (w *sha256Writer) Write(p []byte) (int, error) {
	w.Hash.Write(p)
	return w.w.Write(p)
}

func (w *sha256Writer) SHA256() string {
	return fmt.Sprintf("%x", w.Hash.Sum(nil))
}

func newSHA256Writer(w io.Writer) *sha256Writer {
	return &sha256Writer{Hash: sha256.New(), w: w}
}
