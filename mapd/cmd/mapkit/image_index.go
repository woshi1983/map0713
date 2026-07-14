// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/imgpkg/pkg/imgpkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// IndexImageCmd represents the index command.
var IndexImageCmd = &cobra.Command{
	Use:   "index",
	Short: "Create index image and push to registry.",
	Long:  `Create index image and push to registry.`,
	Example: `
Create plain image index:
$ mapkit image index -i myregistry/myproject/myimage1:mytag1 \
    -i myregistry/myproject/myimage2:mytag2 -o myregistry/myproject:mytag

Create image bundle index:
$ mapkit image index -b myregistry/myproject/mybundle1:mytag1 \
    -b myregistry/myproject/mybundle2:mytag2 -o myregistry/myproject:mytag

Create plain image index from Helm charts:
$ mapkit image index -c ./postgresql-11.8.1.tgz \
    -c ./mariadb-11.3.0.tgz -o myregistry/myproject:mytag

Create hybrid image index:
$ mapkit image index -i myregistry/myproject/myimage1:mytag1 \
    -b myregistry/myproject/mybundle2:mytag2 \
    --registry-username myuser --registry-password mypassword \
    --registry-ca-cert-path rootca.crt -o myregistry/myproject:mytag`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			exitCode int
			err      error
		)
		if indexImage == "" {
			exitCode = 1
			err = fmt.Errorf("specify a location to push the index image.")
			goto DONE
		}

		if len(plainImages)+len(bundleImages)+len(helmCharts) < 1 {
			exitCode = 1
			err = fmt.Errorf("at least one plain image, image bundle, or helm chart must be specified.")
			goto DONE
		}

		if err = generateIndexAndPush(indexImage, plainImages, bundleImages, helmCharts); err != nil {
			exitCode = 2
		}
	DONE:
		if exitCode > 0 {
			fmt.Fprintf(os.Stderr, "FATAL(%d): %v.\n", exitCode, err)
		}
		os.Exit(exitCode)
	},
}

func generateIndexAndPush(saveTo string, plainImages, bundleImages, helmCharts []string) error {
	tempDir, err := os.MkdirTemp("", "mapkit-upload-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	confUI := ui.NewConfUI(ui.NewNoopLogger())
	defer confUI.Flush()

	enc := newImageIndexV1Encoder()
	for _, each := range plainImages {
		if err = enc.AddImage(each); err != nil {
			return err
		}
	}
	for _, each := range bundleImages {
		if err = enc.AddBundle(each); err != nil {
			return err
		}
	}
	for _, each := range helmCharts {
		confUI.PrintLinef("helm | evaluating chart %s with default values.", each)
		images, err := getReferencedImagesOfHelmChart(each)
		if err != nil {
			return errors.Wrap(err, "failed to get referenced image")
		}
		for _, image := range images {
			if err = enc.AddImage(image); err != nil {
				return err
			}
		}
	}

	var buf bytes.Buffer
	if err = enc.Encode(&buf); err != nil {
		return err
	}
	filename := filepath.Join(tempDir, "index.json")
	if err = os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write index file")
	}

	opts := cmd.NewPushOptions(confUI)
	opts.ImageFlags.Image = saveTo
	opts.RegistryFlags = cmd.RegistryFlags{
		CACertPaths:           registryCACertPaths,
		VerifyCerts:           registryVerifyCerts,
		Insecure:              registryInsecure,
		Username:              registryUsername,
		Password:              registryPassword,
		Token:                 registryToken,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	opts.FileFlags.Files = []string{filename}
	if opts.RegistryFlags.Username == "" && opts.RegistryFlags.Token == "" {
		username, password, err := getCrendentialForImageFromDockerConfig(saveTo)
		if err == nil {
			opts.RegistryFlags.Username = username
			opts.RegistryFlags.Password = password
			confUI.PrintLinef("push | using credential in docker config.")
		} else {
			opts.RegistryFlags.Anon = true
			confUI.PrintLinef("push | using anonymous auth.")
		}
	} else {
		confUI.PrintLinef("push | using user-specified credential.")
	}

	if err = opts.Run(); err != nil {
		return err
	}
	confUI.PrintLinef("Succeeded.")
	return nil
}

func downloadAndGetIndex(indexRef string, confUI *ui.ConfUI) (plainImages, bundleImages []string, err error) {
	var tempDir string
	tempDir, err = os.MkdirTemp("", "mapkit-download-*")
	if err != nil {
		return
	}
	defer os.RemoveAll(tempDir)

	opts := cmd.NewPullOptions(confUI)
	opts.ImageFlags.Image = indexRef
	opts.RegistryFlags = cmd.RegistryFlags{
		CACertPaths:           registryCACertPaths,
		VerifyCerts:           registryVerifyCerts,
		Insecure:              registryInsecure,
		Username:              registryUsername,
		Password:              registryPassword,
		Token:                 registryToken,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	opts.OutputPath = tempDir
	if opts.RegistryFlags.Username == "" && opts.RegistryFlags.Token == "" {
		username, password, err := getCrendentialForImageFromDockerConfig(indexRef)
		if err == nil {
			opts.RegistryFlags.Username = username
			opts.RegistryFlags.Password = password
		}
	}

	if err = opts.Run(); err != nil {
		return
	}

	var (
		p     []byte
		index ImageIndex
	)
	p, err = os.ReadFile(filepath.Join(tempDir, "index.json"))
	if err != nil {
		return plainImages, bundleImages, errors.Wrapf(err, "error extracting index.json from index image")
	}
	if err = json.Unmarshal(p, &index); err != nil {
		return plainImages, bundleImages, errors.Wrapf(err, "error parsing index from index image")
	}
	for _, each := range index.Images {
		switch each.Type {
		case "image":
			plainImages = append(plainImages, each.Ref)
		case "bundle":
			bundleImages = append(bundleImages, each.Ref)
		}
	}
	return plainImages, bundleImages, nil
}

func init() {
	IndexImageCmd.Flags().StringSliceVarP(&plainImages, "image", "i", plainImages, "Plain images to download")
	IndexImageCmd.Flags().StringSliceVarP(&bundleImages, "image-bundle", "b", bundleImages, "Bundle images to download")
	IndexImageCmd.Flags().StringSliceVarP(&helmCharts, "helm-chart", "c", helmCharts, "(Experimental) Paths of the helm chart to download images from")
	IndexImageCmd.Flags().StringVarP(&indexImage, "output", "o", indexImage, "Location to push the index image to")

	IndexImageCmd.Flags().StringSliceVar(&registryCACertPaths, "registry-ca-cert-path", registryCACertPaths, "Add CA certificates for registry API (format: /tmp/foo)")
	IndexImageCmd.Flags().BoolVar(&registryVerifyCerts, "registry-verify-certs", registryVerifyCerts, "Set whether to verify server's certificate chain and host name")
	IndexImageCmd.Flags().BoolVar(&registryInsecure, "registry-insecure", registryInsecure, "Allow the use of http when interacting with registries")
	IndexImageCmd.Flags().StringVar(&registryUsername, "registry-username", registryUsername, "Set username for registry auth")
	IndexImageCmd.Flags().StringVar(&registryPassword, "registry-password", registryPassword, "Set password for registry auth")
	IndexImageCmd.Flags().StringVar(&registryToken, "registry-token", registryToken, "Set token for registry auth")
}

type ImageIndex struct {
	Images  []*ImageIndexEntry `json:"images,omitempty"`
	Version string             `json:"_version,omitempty"`
}

type ImageIndexEntry struct {
	Ref  string `json:"ref,omitempty"`
	Type string `json:"type,omitempty"`
}

type imageIndexV1Encoder struct {
	Images map[string]*ImageIndexEntry
}

func (g *imageIndexV1Encoder) AddBundle(b string) error {
	_, _, _, err := splitImageRef(b)
	if err != nil {
		return err
	}
	g.Images[b] = &ImageIndexEntry{
		Ref:  b,
		Type: "bundle",
	}
	return nil
}

func (g *imageIndexV1Encoder) AddImage(i string) error {
	_, _, _, err := splitImageRef(i)
	if err != nil {
		return err
	}
	g.Images[i] = &ImageIndexEntry{
		Ref:  i,
		Type: "image",
	}
	return nil
}

func (g *imageIndexV1Encoder) Index() *ImageIndex {
	index := &ImageIndex{Version: "v1"}
	for i := range g.Images {
		index.Images = append(index.Images, g.Images[i])
	}
	return index
}

func (g *imageIndexV1Encoder) Encode(w io.Writer) error {
	b, err := json.Marshal(g.Index())
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func newImageIndexV1Encoder() *imageIndexV1Encoder {
	return &imageIndexV1Encoder{Images: make(map[string]*ImageIndexEntry)}
}

func getReferencedImagesOfHelmChart(fpath string) ([]string, error) {
	_, manifest, err := k8s.RenderChart("default", "default", fpath, make(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	var images []string
	buf := bytes.NewBuffer(nil)
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if err = k8s.WalkManifest(manifest, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, _, err := ser.Decode(p, nil, obj)
		if err != nil {
			if runtime.IsMissingKind(err) || runtime.IsMissingVersion(err) {
				return nil
			}
			return err
		}

		switch obj.GetKind() {
		case "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet", "Job":
			spec, found, err := unstructured.NestedMap(obj.Object, "spec", "template", "spec")
			if !found || err != nil {
				return nil
			}
			ss, err := getReferencedImagesFromPodSpec(spec)
			if err != nil {
				return nil
			}
			images = append(images, ss...)
		case "CronJob":
			spec, found, err := unstructured.NestedMap(obj.Object, "spec", "jobTemplate", "spec", "template", "spec")
			if !found || err != nil {
				return nil
			}
			ss, err := getReferencedImagesFromPodSpec(spec)
			if err != nil {
				return nil
			}
			images = append(images, ss...)
		case "Pod":
			spec, found, err := unstructured.NestedMap(obj.Object, "spec")
			if !found || err != nil {
				return nil
			}
			ss, err := getReferencedImagesFromPodSpec(spec)
			if err != nil {
				return nil
			}
			images = append(images, ss...)
		}
		return ser.Encode(obj, buf)
	}); err != nil {
		return nil, err
	}

	return images, nil
}

func getReferencedImagesFromPodSpec(spec map[string]interface{}) ([]string, error) {
	b, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	obj := &corev1.PodSpec{}
	if err := json.Unmarshal(b, obj); err != nil {
		return nil, err
	}
	var images []string
	for _, container := range obj.InitContainers {
		images = append(images, container.Image)
	}
	for _, container := range obj.Containers {
		images = append(images, container.Image)
	}
	return images, nil
}
